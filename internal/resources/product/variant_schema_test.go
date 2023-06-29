package product

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestArePricesConflicting(t *testing.T) {
	t.Parallel()

	priceBlockAttributeTypes := productVariantSchema().Blocks["price"].Type().(types.ListType).ElementType().(types.ObjectType).AttributeTypes()
	priceValueBlockAttributeTypes := priceBlockAttributeTypes["value"].(types.ListType).ElementType().(types.ObjectType)
	priceValueActualType := priceValueBlockAttributeTypes.AttributeTypes()

	buildPrice := func(centAmount int64, currencyCode string, opts map[string]any) attr.Value {
		attributes := map[string]attr.Value{
			"value": types.ListValueMust(
				priceValueBlockAttributeTypes,
				[]attr.Value{types.ObjectValueMust(
					priceValueActualType,
					map[string]attr.Value{
						"cent_amount":   types.Int64Value(centAmount),
						"currency_code": types.StringValue(currencyCode),
					},
				)},
			),
			"id":          types.StringNull(),
			"key":         types.StringNull(),
			"country":     types.StringNull(),
			"valid_from":  types.StringNull(),
			"valid_until": types.StringNull(),
		}

		if val, ok := opts["country"]; ok {
			attributes["country"] = types.StringValue(val.(string))
		}

		if val, ok := opts["valid_from"]; ok {
			attributes["valid_from"] = types.StringValue(val.(string))
		}

		if val, ok := opts["valid_until"]; ok {
			attributes["valid_until"] = types.StringValue(val.(string))
		}
		return types.ObjectValueMust(priceBlockAttributeTypes, attributes)
	}

	testCases := map[string]struct {
		a, b     attr.Value
		conflict bool
	}{
		"conflict because same currency, no country, no datetime": {
			a:        buildPrice(1000, "USD", nil),
			b:        buildPrice(2000, "USD", nil),
			conflict: true,
		},
		"conflict because same currency, same country, no datetime": {
			a:        buildPrice(1000, "USD", map[string]any{"country": "US"}),
			b:        buildPrice(2000, "USD", map[string]any{"country": "US"}),
			conflict: true,
		},
		"conflict because same currency, same country, overlapping datetime - open": {
			a:        buildPrice(1000, "USD", map[string]any{"country": "US", "valid_from": "2023-09-15T12:34:56Z"}),
			b:        buildPrice(2000, "USD", map[string]any{"country": "US", "valid_from": "2024-09-15T12:34:56Z"}),
			conflict: true,
		},
		"no conflict: same currency, same country, back-to-back 1": {
			a:        buildPrice(1000, "USD", map[string]any{"country": "US", "valid_until": "2023-09-15T12:00:00Z"}),
			b:        buildPrice(2000, "USD", map[string]any{"country": "US", "valid_from": "2023-09-15T12:00:00Z"}),
			conflict: false,
		},
		"no conflict: same currency, same country, back-to-back 2": {
			a:        buildPrice(2000, "USD", map[string]any{"country": "US", "valid_from": "2023-09-15T12:00:00Z"}),
			b:        buildPrice(1000, "USD", map[string]any{"country": "US", "valid_until": "2023-09-15T12:00:00Z"}),
			conflict: false,
		},
		"conflict: same currency, same country, back-to-back mistake": {
			a:        buildPrice(1000, "USD", map[string]any{"country": "US", "valid_until": "2023-09-15T12:00:01Z"}),
			b:        buildPrice(2000, "USD", map[string]any{"country": "US", "valid_from": "2023-09-15T12:00:00Z"}),
			conflict: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if arePricesConflicting(testCase.a, testCase.b) != testCase.conflict {
				t.Errorf("Expected conflicting=%v got %v", testCase.conflict, !testCase.conflict)
			}
		})
	}

}
