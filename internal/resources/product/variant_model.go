package product

import (
	"context"
	"fmt"
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

func NewProductVariant(n platform.ProductVariant) ProductVariant {
	return ProductVariant{
		ID:         types.Int64Value(int64(n.ID)),
		Key:        types.StringPointerValue(n.Key),
		SKU:        types.StringPointerValue(n.Sku),
		Prices:     pie.Map(n.Prices, NewProductVariantPrice),
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

func NewProductVariantPrice(n platform.Price) ProductVariantPrice {
	return ProductVariantPrice{
		ID:         types.StringValue(n.ID),
		Key:        types.StringPointerValue(n.Key),
		Value:      []customtypes.SimpleMoney{customtypes.SimpleMoneyFromTypedMoney(n.Value)},
		Country:    types.StringPointerValue(n.Country),
		ValidFrom:  types.StringPointerValue(flattenDateTime(n.ValidFrom)),
		ValidUntil: types.StringPointerValue(flattenDateTime(n.ValidUntil)),
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

func (pv ProductVariant) draft(ctx context.Context) platform.ProductVariantDraft {
	ret := platform.ProductVariantDraft{
		Sku:        pv.SKU.ValueStringPointer(),
		Key:        pv.Key.ValueStringPointer(),
		Prices:     nil,
		Attributes: nil,
	}

	if len(pv.Attributes) > 0 {
		ret.Attributes = pie.Map(pv.Attributes, func(pva ProductVariantAttribute) platform.Attribute {
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
		})
	}

	if len(pv.Prices) > 0 {
		ret.Prices = pie.Map(pv.Prices, func(price ProductVariantPrice) platform.PriceDraft {
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
