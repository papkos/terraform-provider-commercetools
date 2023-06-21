package product

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

//goland:noinspection GoNameStartsWithPackageName
type ProductVariant struct {
	ID         types.Int64               `tfsdk:"id"`
	Key        types.String              `tfsdk:"key"`
	SKU        types.String              `tfsdk:"sku"`
	Prices     []ProductVariantPrice     `tfsdk:"price"`     // of ProductVariantPrice
	Attributes []ProductVariantAttribute `tfsdk:"attribute"` // of ProductVariantAttribute
}

func NewProductVariantFromNative(n platform.ProductVariant) ProductVariant {
	return ProductVariant{
		ID:         types.Int64Value(int64(n.ID)),
		Key:        types.StringPointerValue(n.Key),
		SKU:        types.StringPointerValue(n.Sku),
		Prices:     pie.Map(n.Prices, NewProductVariantPriceFromNative),
		Attributes: pie.Map(n.Attributes, NewProductVariantAttributeFromNative),
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ProductVariantPrice struct {
	ID         types.String              `tfsdk:"id"`
	Key        types.String              `tfsdk:"key"`
	Value      []customtypes.SimpleMoney `tfsdk:"value"`
	Country    types.String              `tfsdk:"country"`
	ValidFrom  types.String              `tfsdk:"valid_from"`
	ValidUntil types.String              `tfsdk:"valid_until"`
}

func NewProductVariantPriceFromNative(n platform.Price) ProductVariantPrice {
	return ProductVariantPrice{
		ID:         types.StringValue(n.ID),
		Key:        types.StringPointerValue(n.Key),
		Value:      []customtypes.SimpleMoney{customtypes.SimpleMoneyFromTypedMoney(n.Value)},
		Country:    types.StringPointerValue(n.Country),
		ValidFrom:  types.StringPointerValue(flattenDateTime(n.ValidFrom)),
		ValidUntil: types.StringPointerValue(flattenDateTime(n.ValidUntil)),
	}
}

func (price ProductVariantPrice) ToNative() platform.PriceDraft {
	validFrom, err := parseDateTime(price.ValidFrom.ValueStringPointer())
	if err != nil {
		// TODO How to signal error?
		return platform.PriceDraft{}
	}

	validUntil, err := parseDateTime(price.ValidUntil.ValueStringPointer())
	if err != nil {
		// TODO How to signal error?
		return platform.PriceDraft{}
	}
	return platform.PriceDraft{
		Key:        price.Key.ValueStringPointer(),
		Value:      price.Value[0].ToNative(),
		Country:    price.Country.ValueStringPointer(),
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ProductVariantAttribute struct {
	Name types.String `tfsdk:"name"`

	BoolValue          types.Bool                       `tfsdk:"bool_value"`
	TextValue          types.String                     `tfsdk:"text_value"`
	LocalizedTextValue customtypes.LocalizedStringValue `tfsdk:"localized_text_value"`
	PTReferenceValue   types.String                     `tfsdk:"product_type_reference_value"`
}

func NewProductVariantAttributeFromNative(n platform.Attribute) ProductVariantAttribute {
	pva := ProductVariantAttribute{
		Name: types.StringValue(n.Name),

		// Initialize to respective null values
		BoolValue:          types.BoolNull(),
		TextValue:          types.StringNull(),
		LocalizedTextValue: customtypes.NewLocalizedStringNull(),
		PTReferenceValue:   types.StringNull(),
	}

	switch val := n.Value.(type) {
	case bool:
		pva.BoolValue = types.BoolValue(val)
	case string:
		pva.TextValue = types.StringValue(val)
	case platform.LocalizedString:
		pva.LocalizedTextValue = utils.FromLocalizedString(val)

	// Complex value, dig deeper
	case map[string]any:
		if typeId, ok := val["typeId"]; ok {
			// Probably a "reference" value, check what type it is referencing & set the relevant field
			switch typeId {
			case "product-type":
				pva.PTReferenceValue = types.StringValue(val["id"].(string))
			}
		}
	}
	return pva
}

func (pva ProductVariantAttribute) ToNative() platform.Attribute {
	attribute := platform.Attribute{
		Name: pva.Name.ValueString(),
	}

	switch {
	case !pva.BoolValue.IsNull():
		attribute.Value = pva.BoolValue.ValueBool()
	case !pva.TextValue.IsNull():
		attribute.Value = pva.TextValue.ValueString()
	case !pva.LocalizedTextValue.IsNull():
		attribute.Value = pva.LocalizedTextValue.ValueLocalizedString()
	case !pva.PTReferenceValue.IsNull():
		attribute.Value = platform.ProductTypeReference{ID: pva.PTReferenceValue.ValueString()}
	}

	return attribute
}

// Same as draftAddNew, but uses a different type, and lacks a field: Staged
func (pv ProductVariant) draftCreate(ctx context.Context) platform.ProductVariantDraft {
	ret := platform.ProductVariantDraft{
		Sku:        pv.SKU.ValueStringPointer(),
		Key:        pv.Key.ValueStringPointer(),
		Prices:     nil,
		Attributes: nil,
	}

	if len(pv.Attributes) > 0 {
		ret.Attributes = pie.Map(pv.Attributes, ProductVariantAttribute.ToNative)
	}

	if len(pv.Prices) > 0 {
		ret.Prices = pie.Map(pv.Prices, ProductVariantPrice.ToNative)
	}

	return ret
}

// Same as draftCreate, but uses a different type, and has an extra field: Staged
func (pv ProductVariant) draftAddNew(ctx context.Context) platform.ProductAddVariantAction {
	ret := platform.ProductAddVariantAction{
		Sku:        pv.SKU.ValueStringPointer(),
		Key:        pv.Key.ValueStringPointer(),
		Prices:     nil,
		Attributes: nil,
		// If true, only the staged description is updated. If false, both the current and staged description are updated.
		// Default: true
		Staged: utils.Ref(false),
	}

	if len(pv.Attributes) > 0 {
		ret.Attributes = pie.Map(pv.Attributes, ProductVariantAttribute.ToNative)
	}

	if len(pv.Prices) > 0 {
		ret.Prices = pie.Map(pv.Prices, ProductVariantPrice.ToNative)
	}

	return ret
}

func (pv ProductVariant) calculateUpdateActions(plan ProductVariant) []platform.ProductUpdateAction {
	var ret []platform.ProductUpdateAction

	// Variant-related actions
	// Ref: https://docs.commercetools.com/api/projects/products#update-actions

	// setSku
	if !pv.SKU.Equal(plan.SKU) {
		var value *string
		if !plan.SKU.IsNull() && !plan.SKU.IsUnknown() {
			value = plan.SKU.ValueStringPointer()
		}
		ret = append(ret, platform.ProductSetSkuAction{
			VariantId: int(pv.ID.ValueInt64()),
			Sku:       value,
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	// setPrices
	// Trying to get away cheap, instead of only changing what actually changed (comparing prices by their ID),
	// see if there's any difference, and completely replace all the prices
	if !reflect.DeepEqual(pv.Prices, plan.Prices) {
		ret = append(ret, platform.ProductSetPricesAction{
			VariantId: utils.Ref(int(pv.ID.ValueInt64())),
			Prices:    pie.Map(plan.Prices, ProductVariantPrice.ToNative),
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	// setAttribute
	// Attributes can be added, removed or changed. There is no "setAttributes" action that would replace all at once.
	stateAttributesByName := map[string]ProductVariantAttribute{}
	for _, sa := range pv.Attributes {
		stateAttributesByName[sa.Name.ValueString()] = sa
	}

	for _, pa := range plan.Attributes {
		ret = append(ret, platform.ProductSetAttributeAction{
			VariantId: utils.Ref(int(pv.ID.ValueInt64())),
			Name:      pa.Name.ValueString(),
			Value:     pa.ToNative(),
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
		delete(stateAttributesByName, pa.Name.ValueString())
	}

	// The ones left in this map are not part of the plan, thus need to be removed
	for _, sa := range stateAttributesByName {
		ret = append(ret, platform.ProductSetAttributeAction{
			VariantId: utils.Ref(int(pv.ID.ValueInt64())),
			Name:      sa.Name.ValueString(),
			Value:     nil, // Nil removes the value
			// If true, only the staged description is updated. If false, both the current and staged description are updated.
			// Default: true
			Staged: utils.Ref(false),
		})
	}

	return ret
}

func parseDateTime(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}

	dt, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse as RFC3339: %w", err)
	}

	return &dt, nil
}

func flattenDateTime(dt *time.Time) *string {
	if dt == nil {
		return nil
	}

	formatted := (*dt).Format(time.RFC3339)

	return &formatted
}
