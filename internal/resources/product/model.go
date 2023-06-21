package product

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

const (
	PriceModeEmbedded   = "Embedded"
	PriceModeStandalone = "Standalone"
)

type Product struct {
	ID          types.String `tfsdk:"id"`
	Key         types.String `tfsdk:"key"`
	Version     types.Int64  `tfsdk:"version"`
	ProductType types.String `tfsdk:"product_type"`
	TaxCategory types.String `tfsdk:"tax_category"`
	// TODO State
	PriceMode types.String `tfsdk:"price_mode"`

	// These items all have maximal one item. We don't use SingleNestedBlock
	// here since it isn't quite robust currently.
	// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
	MasterData []ProductCatalogData `tfsdk:"master_data"`
}

func NewProductFromNative(n platform.Product) Product {
	return Product{
		ID:          types.StringValue(n.ID),
		Key:         utils.FromOptionalString(n.Key),
		Version:     types.Int64Value(int64(n.Version)),
		ProductType: types.StringValue(n.ProductType.ID),
		TaxCategory: types.StringValue(n.TaxCategory.ID),
		PriceMode:   utils.FromOptionalString((*string)(n.PriceMode)),
		MasterData:  []ProductCatalogData{NewProductCatalogDataFromNative(n.MasterData)},
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ProductCatalogData struct {
	Published        types.Bool `tfsdk:"published"`
	HasStagedChanges types.Bool `tfsdk:"has_staged_changes"`
	// These items all have maximal one item. We don't use SingleNestedBlock
	// here since it isn't quite robust currently.
	// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
	Current []ProductData `tfsdk:"current"`
	Staged  []ProductData `tfsdk:"staged"`
}

func NewProductCatalogDataFromNative(n platform.ProductCatalogData) ProductCatalogData {
	return ProductCatalogData{
		Published:        types.BoolValue(n.Published),
		HasStagedChanges: types.BoolValue(n.HasStagedChanges),
		Current:          []ProductData{NewProductDataFromNative(n.Current)},
		Staged:           []ProductData{NewProductDataFromNative(n.Staged)},
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ProductData struct {
	Name            customtypes.LocalizedStringValue `tfsdk:"name"`
	Categories      types.List                       `tfsdk:"categories"`
	Description     customtypes.LocalizedStringValue `tfsdk:"description"`
	Slug            customtypes.LocalizedStringValue `tfsdk:"slug"`
	MetaTitle       customtypes.LocalizedStringValue `tfsdk:"meta_title"`
	MetaDescription customtypes.LocalizedStringValue `tfsdk:"meta_description"`
	MetaKeywords    customtypes.LocalizedStringValue `tfsdk:"meta_keywords"`
	MasterVariant   []ProductVariant                 `tfsdk:"master_variant"`
	Variants        []ProductVariant                 `tfsdk:"variant"`
	// TODO CategoryOrderHints
	// TODO searchKeywords
}

func NewProductDataFromNative(n platform.ProductData) ProductData {
	res := ProductData{
		Name:            utils.FromLocalizedString(n.Name),
		Categories:      types.ListNull(types.StringType),
		Description:     utils.FromOptionalLocalizedString(n.Description),
		Slug:            utils.FromLocalizedString(n.Slug),
		MetaTitle:       utils.FromOptionalLocalizedString(n.MetaTitle),
		MetaDescription: utils.FromOptionalLocalizedString(n.MetaDescription),
		MetaKeywords:    utils.FromOptionalLocalizedString(n.MetaKeywords),
		MasterVariant:   []ProductVariant{NewProductVariantFromNative(n.MasterVariant)},
		Variants:        pie.Map(n.Variants, NewProductVariantFromNative),
	}

	// If the categories is empty we want to keep the value as null and not an empty
	// list
	if len(n.Categories) > 0 {
		var diagnostic diag.Diagnostics
		res.Categories, diagnostic = types.ListValueFrom(context.TODO(), types.StringType, pie.Map(n.Categories, func(cat platform.CategoryReference) types.String {
			return types.StringValue(cat.ID)
		}))
		if diagnostic.HasError() {
			panic(fmt.Sprintf("Failed to convert categories list: %s", diagnostic.Errors()))
		}
	}

	return res
}

func (p Product) draft(ctx context.Context) platform.ProductDraft {
	productData := p.MasterData[0].Current[0]

	return platform.ProductDraft{
		// TODO Is it OK to read the MasterData.Current ?
		ProductType: platform.ProductTypeResourceIdentifier{ID: p.ProductType.ValueStringPointer()},
		Name:        productData.Name.ValueLocalizedString(),
		Slug:        productData.Slug.ValueLocalizedString(),
		Key:         p.Key.ValueStringPointer(),
		Description: productData.Description.ValueLocalizedStringRef(),
		Categories: pie.Map(productData.Categories.Elements(), func(catString attr.Value) platform.CategoryResourceIdentifier {
			value, err := catString.ToTerraformValue(ctx)
			if err != nil {
				return platform.CategoryResourceIdentifier{}
			}
			var stringVal string
			err = value.As(&stringVal)
			if err != nil {
				return platform.CategoryResourceIdentifier{}
			}
			return platform.CategoryResourceIdentifier{ID: &stringVal}
		}),
		CategoryOrderHints: nil,
		MetaTitle:          productData.MetaTitle.ValueLocalizedStringRef(),
		MetaDescription:    productData.MetaDescription.ValueLocalizedStringRef(),
		MetaKeywords:       productData.MetaKeywords.ValueLocalizedStringRef(),
		MasterVariant:      utils.Ref(productData.MasterVariant[0].draft(ctx)),
		Variants: pie.Map(productData.Variants, func(variant ProductVariant) platform.ProductVariantDraft {
			return variant.draft(ctx)
		}),
		TaxCategory:    &platform.TaxCategoryResourceIdentifier{ID: p.TaxCategory.ValueStringPointer()},
		SearchKeywords: nil,
		State:          nil,
		Publish:        p.MasterData[0].Published.ValueBoolPointer(),
		PriceMode:      utils.Ref(platform.ProductPriceModeEnum(p.PriceMode.ValueString())),
	}

}

func (p Product) updateActions(plan Product) platform.ProductUpdate {
	result := platform.ProductUpdate{
		Version: int(p.Version.ValueInt64()),
		Actions: []platform.ProductUpdateAction{},
	}

	// TODO Check diff & add actions

	return result
}
