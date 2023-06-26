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

func NewProductFromNative(n platform.Product, productType platform.ProductType) Product {
	product := Product{
		ID:          types.StringValue(n.ID),
		Key:         utils.FromOptionalString(n.Key),
		Version:     types.Int64Value(int64(n.Version)),
		ProductType: types.StringValue(n.ProductType.ID),
		TaxCategory: types.StringNull(),
		PriceMode:   utils.FromOptionalString((*string)(n.PriceMode)),
		MasterData:  []ProductCatalogData{NewProductCatalogDataFromNative(n.MasterData, productType)},
	}

	if n.TaxCategory != nil {
		product.TaxCategory = types.StringValue(n.TaxCategory.ID)
	}

	return product
}

//goland:noinspection GoNameStartsWithPackageName
type ProductCatalogData struct {
	Published types.Bool `tfsdk:"published"`
	// These items all have maximal one item. We don't use SingleNestedBlock
	// here since it isn't quite robust currently.
	// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
	Current []ProductData `tfsdk:"current"`

	// We don't care about Staged data
}

func NewProductCatalogDataFromNative(n platform.ProductCatalogData, productType platform.ProductType) ProductCatalogData {
	return ProductCatalogData{
		Published: types.BoolValue(n.Published),
		Current:   []ProductData{NewProductDataFromNative(n.Current, productType)},
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ProductData struct {
	Name        customtypes.LocalizedStringValue `tfsdk:"name"`
	Categories  types.List                       `tfsdk:"categories"`
	Description customtypes.LocalizedStringValue `tfsdk:"description"`
	Slug        customtypes.LocalizedStringValue `tfsdk:"slug"`
	// Ignore these fields for now, could be implemented later
	// MetaTitle       customtypes.LocalizedStringValue `tfsdk:"meta_title"`
	// MetaDescription customtypes.LocalizedStringValue `tfsdk:"meta_description"`
	// MetaKeywords    customtypes.LocalizedStringValue `tfsdk:"meta_keywords"`
	MasterVariant []ProductVariant `tfsdk:"master_variant"`
	Variants      []ProductVariant `tfsdk:"variant"`
	// TODO CategoryOrderHints
	// TODO searchKeywords
}

func NewProductDataFromNative(n platform.ProductData, productType platform.ProductType) ProductData {
	res := ProductData{
		Name:        utils.FromLocalizedString(n.Name),
		Categories:  types.ListNull(types.StringType),
		Description: utils.FromOptionalLocalizedString(n.Description),
		Slug:        utils.FromLocalizedString(n.Slug),
		// MetaTitle:       utils.FromOptionalLocalizedString(n.MetaTitle),
		// MetaDescription: utils.FromOptionalLocalizedString(n.MetaDescription),
		// MetaKeywords:    utils.FromOptionalLocalizedString(n.MetaKeywords),
		MasterVariant: []ProductVariant{NewProductVariantFromNative(n.MasterVariant, productType)},
		Variants: pie.Map(n.Variants, func(t platform.ProductVariant) ProductVariant {
			return NewProductVariantFromNative(t, productType)
		}),
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

	draft := platform.ProductDraft{
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
		// MetaTitle:          productData.MetaTitle.ValueLocalizedStringRef(),
		// MetaDescription:    productData.MetaDescription.ValueLocalizedStringRef(),
		// MetaKeywords:       productData.MetaKeywords.ValueLocalizedStringRef(),
		MasterVariant: utils.Ref(productData.MasterVariant[0].draftCreate(ctx)),
		Variants: pie.Map(productData.Variants, func(variant ProductVariant) platform.ProductVariantDraft {
			return variant.draftCreate(ctx)
		}),
		TaxCategory:    nil,
		SearchKeywords: nil,
		State:          nil,
		Publish:        p.MasterData[0].Published.ValueBoolPointer(),
		PriceMode:      utils.Ref(platform.ProductPriceModeEnum(p.PriceMode.ValueString())),
	}

	if !p.TaxCategory.IsNull() {
		draft.TaxCategory = &platform.TaxCategoryResourceIdentifier{ID: p.TaxCategory.ValueStringPointer()}
	}

	return draft
}

func (p Product) calculateUpdateActions(ctx context.Context, plan Product) platform.ProductUpdate {
	result := platform.ProductUpdate{
		Version: int(p.Version.ValueInt64()),
		Actions: []platform.ProductUpdateAction{},
	}

	// Top-level property changes first, variants at the end.
	// Ref: https://docs.commercetools.com/api/projects/products#update-actions

	// setKey
	// TODO shouldn't this be p.Key.Equal(plan.Key) ??
	// 	It actually works as expected, i.e. if the underlying values are identical, we don't enter the if body. Why?
	if p.Key != plan.Key {
		var value *string
		if !plan.Key.IsNull() && !plan.Key.IsUnknown() {
			value = plan.Key.ValueStringPointer()
		}
		result.Actions = append(result.Actions, platform.ProductSetKeyAction{Key: value})
	}

	// changeName
	if !p.MasterData[0].Current[0].Name.Equal(plan.MasterData[0].Current[0].Name) {
		var planValue = plan.MasterData[0].Current[0].Name
		result.Actions = append(result.Actions, platform.ProductChangeNameAction{
			Name: planValue.ValueLocalizedString(), // Unknown & Null is nil, but that is invalid and the schema should have prevented it
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	// setDescription
	if !p.MasterData[0].Current[0].Description.Equal(plan.MasterData[0].Current[0].Description) {
		var planValue = plan.MasterData[0].Current[0].Description
		result.Actions = append(result.Actions, platform.ProductSetDescriptionAction{
			Description: planValue.ValueLocalizedStringRef(), // Unknown & Null is nil
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	// changeSlug
	if !p.MasterData[0].Current[0].Slug.Equal(plan.MasterData[0].Current[0].Slug) {
		var planValue = plan.MasterData[0].Current[0].Slug
		result.Actions = append(result.Actions, platform.ProductChangeSlugAction{
			Slug: planValue.ValueLocalizedString(), // Unknown & Null is nil, but that is invalid and the schema should have prevented it
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	// setPriceMode
	if p.PriceMode != plan.PriceMode {
		var value *platform.ProductPriceModeEnum
		if !plan.PriceMode.IsNull() && !plan.PriceMode.IsUnknown() {
			value = (*platform.ProductPriceModeEnum)(plan.PriceMode.ValueStringPointer())
		}
		result.Actions = append(result.Actions, platform.ProductSetPriceModeAction{PriceMode: value})
	}

	// setTaxCategory
	if p.TaxCategory != plan.TaxCategory {
		var value *string
		if !plan.TaxCategory.IsNull() && !plan.TaxCategory.IsUnknown() {
			value = plan.TaxCategory.ValueStringPointer()
		}
		result.Actions = append(result.Actions, platform.ProductSetTaxCategoryAction{
			TaxCategory: &platform.TaxCategoryResourceIdentifier{ID: value},
		})
	}

	// publish/unpublish
	if !p.MasterData[0].Published.Equal(plan.MasterData[0].Published) {
		value := plan.MasterData[0].Published.ValueBool()

		if value {
			result.Actions = append(result.Actions, platform.ProductPublishAction{
				Scope: utils.Ref(platform.ProductPublishScopeAll),
			})
		} else {
			result.Actions = append(result.Actions, platform.ProductUnpublishAction{})
		}
	}

	// Check the master variant

	stateMasterVariant := p.MasterData[0].Current[0].MasterVariant[0]
	planMasterVariant := plan.MasterData[0].Current[0].MasterVariant[0]
	result.Actions = append(result.Actions, stateMasterVariant.calculateUpdateActions(planMasterVariant)...)

	// Extra variants
	// Variants can be added, removed or changed.
	stateVariantsById := map[int64]ProductVariant{}
	for _, sv := range p.MasterData[0].Current[0].Variants {
		stateVariantsById[sv.ID.ValueInt64()] = sv
	}

	for _, pv := range plan.MasterData[0].Current[0].Variants {
		sv, ok := stateVariantsById[pv.ID.ValueInt64()]
		if !ok {
			result.Actions = append(result.Actions, pv.draftAddNew(ctx))
		} else {
			result.Actions = append(result.Actions, pv.calculateUpdateActions(sv)...)
			delete(stateVariantsById, sv.ID.ValueInt64())
		}
	}

	// The ones left in this map are not part of the plan, thus need to be removed
	for _, sv := range stateVariantsById {
		result.Actions = append(result.Actions, platform.ProductRemoveVariantAction{
			ID: utils.Ref(int(sv.ID.ValueInt64())),
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	return result
}
