package product

import (
	"context"
	"fmt"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
)

//goland:noinspection GoNameStartsWithPackageName
type ProductVariant struct {
	ID     types.Int64           `tfsdk:"id"`
	Key    types.String          `tfsdk:"key"`
	SKU    types.String          `tfsdk:"sku"`
	Prices []ProductVariantPrice `tfsdk:"price"` // of ProductVariantPrice
	//Attributes types.Set             `tfsdk:"attribute"` // of ProductVariantAttribute
}

func NewProductVariant(n platform.ProductVariant) ProductVariant {
	return ProductVariant{
		ID:     types.Int64Value(int64(n.ID)),
		Key:    types.StringPointerValue(n.Key),
		SKU:    types.StringPointerValue(n.Sku),
		Prices: pie.Map(n.Prices, NewProductVariantPrice),
		// Attributes: , // TODO
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
	Name  types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"` // TODO this should be a complex type https://docs.commercetools.com/api/projects/products#attribute
}

func NewProductVariantAttributeFromNative(n platform.AttributeNestedType) ProductVariantAttribute {
	return ProductVariantAttribute{
		Name:  types.StringValue("TODO this is a hardcoded product variant attribute name"),  // TODO
		Value: types.StringValue("TODO this is a hardcoded product variant attribute value"), // TODO
	}
}

func (pv ProductVariant) draft(ctx context.Context) platform.ProductVariantDraft {
	ret := platform.ProductVariantDraft{
		Sku:        pv.SKU.ValueStringPointer(),
		Key:        pv.Key.ValueStringPointer(),
		Prices:     nil,
		Attributes: nil,
	}

	//if len(pv.Attributes.Elements()) > 0 {
	//	ret.Attributes = pie.Map(pv.Attributes.Elements(), func(attr attr.Value) platform.Attribute {
	//		var pva ProductVariantAttribute
	//		err := tfAttrAs(attr, &pva)
	//		if err != nil {
	//			return platform.Attribute{}
	//		}
	//		return platform.Attribute{
	//			Name:  pva.Name.ValueString(),
	//			Value: pva.Value.ValueString(),
	//		}
	//	})
	//}

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

func tfAttrAs(attr attr.Value, dst any) error {
	tfVal, err := attr.ToTerraformValue(context.TODO())
	if err != nil {
		return err
	}

	return tfVal.As(dst)
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
