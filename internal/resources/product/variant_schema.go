package product

import (
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
)

func productVariantSchema() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "A unique, sequential identifier of the Product Variant within the Product.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the ProductVariant.\n\n" +
					"This is different from Product key.",
				Optional: true,
			},
			"sku": schema.StringAttribute{
				Description: "User-defined unique SKU of the Product Variant.",
				Optional:    true,
			},
			"attributes": schema.MapNestedAttribute{
				Description: "Attributes of the Product Variant. The map key is the name of the attribute",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bool_value": schema.BoolAttribute{
							Optional:    true,
							Description: "Use this to provide value to bool type attribute",
						},
						"text_value": schema.StringAttribute{
							Optional:    true,
							Description: "Use this to provide value to text type attribute",
						},
						"localized_text_value": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
							Optional:    true,
							Description: "Use this to provide value to localized text type attribute",
						}),
						"product_type_reference_value": schema.StringAttribute{
							Optional:    true,
							Description: "Use this to provide value to Product Type reference type attribute",
						},
						// There are more, to be implemented: https://docs.commercetools.com/api/projects/products#attribute
					},
					Validators: []validator.Object{
						// TODO This should be on each attribute, but that's insane... Is there a validator that validates an object's attributes?
						// objectvalidator.ExactlyOneOf(
						//	path.MatchRoot("bool_value"),
						//	path.MatchRoot("text_value"),
						//	path.MatchRoot("localized_text_value"),
						//	path.MatchRoot("product_type_reference_value"),
						// ),
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"price": schema.ListNestedBlock{
				Description: "The Embedded Prices for the Product Variant. " +
					"Each Price must have its unique Price scope " +
					"(with same currency, country, Customer Group, Channel, validFrom and validUntil).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:      true,
							Description:   "Unique identifier of this Price.",
							PlanModifiers: []planmodifier.String{
								// stringplanmodifier.UseStateForUnknown(), // Use once we have proper diff for prices
							},
						},
						"key": schema.StringAttribute{
							Description: "User-defined unique identifier of the ProductVariant.\n\n" +
								"This is different from Product key.",
							Optional: true,
						},
						"country": schema.StringAttribute{
							Optional:    true,
							Description: "Country for which this Price is valid.",
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]{2}$`),
									"Must be a 2-letter country code as per ISO 3166-1 alpha-2"),
							},
							PlanModifiers: nil,
						},
						"valid_from": schema.StringAttribute{
							Optional:    true,
							Description: "Date and time from which this Price is valid.",
							Validators: []validator.String{
								customvalidator.DateTimeValidator(),
							},
						},
						"valid_until": schema.StringAttribute{
							Optional: true,
							Description: "Date and time until this Price is valid. " +
								"Prices that are no longer valid are not automatically removed.",
							Validators: []validator.String{
								customvalidator.DateTimeValidator(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"value": schema.ListNestedBlock{
							Description: "Money value of this Price.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"cent_amount": schema.Int64Attribute{
										Description: "Amount in the smallest indivisible unit of a currency",
										Required:    true,
									},
									"currency_code": schema.StringAttribute{
										Description: "Currency code compliant to ISO 4217.",
										Required:    true,
									},
								},
							},
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
						},
					},
				},
				Validators: []validator.List{
					customvalidator.UniqueValuesCompareFunc(arePricesConflicting),
				},
			},
		},
	}
}

// As per the docs, all prices must have different scopes, where "scope" is defined as
// currency, country, customer group, channel, valid from and valid until.
// See more details at https://docs.commercetools.com/api/projects/products#add-price
func arePricesConflicting(a, b attr.Value) bool {
	if !getCurrencyCode(a).Equal(getCurrencyCode(b)) {
		// Different currencies can't be equal
		return false
	}

	if !getCountry(a).Equal(getCountry(b)) && !(getCountry(a).IsNull() && getCountry(b).IsNull()) {
		// Different NON-NULL countries can't be equal
		return false
	}

	validFromA := getValidFrom(a)
	validUntilA := getValidUntil(a)
	validFromB := getValidFrom(b)
	validUntilB := getValidUntil(b)

	// Common case: all are null --> prices are equal
	if validFromA.IsNull() && validUntilA.IsNull() && validFromB.IsNull() && validUntilB.IsNull() {
		return true
	}

	// Else we need to do an interval overlap check

	minTime := time.Time{}
	maxTime := time.Date(9999, 0, 0, 0, 0, 0, 0, time.UTC)

	validFromATime, err := parseDateTime(validFromA.ValueStringPointer(), minTime)
	if err != nil {
		return false // Not a correct datetime
	}

	validUntilATime, err := parseDateTime(validUntilA.ValueStringPointer(), maxTime)
	if err != nil {
		return false // Not a correct datetime
	}

	validFromBTime, err := parseDateTime(validFromB.ValueStringPointer(), minTime)
	if err != nil {
		return false // Not a correct datetime
	}

	validUntilBTime, err := parseDateTime(validUntilB.ValueStringPointer(), maxTime)
	if err != nil {
		return false // Not a correct datetime
	}

	// "order" them
	if (*validFromATime).Before(*validFromBTime) {
		if (*validFromBTime).Before(*validUntilATime) {
			// There is an overlap, signal equivalency
			return true
		}
	} else {
		if (*validFromATime).Before(*validUntilBTime) {
			// There is an overlap, signal equivalency
			return true
		}
	}

	return false
}

func getCurrencyCode(price attr.Value) types.String {
	return price.(basetypes.ObjectValue).Attributes()["value"].(basetypes.ListValue).Elements()[0].(basetypes.ObjectValue).Attributes()["currency_code"].(types.String)
}
func getCountry(price attr.Value) types.String {
	return price.(basetypes.ObjectValue).Attributes()["country"].(types.String)
}

func getValidFrom(price attr.Value) types.String {
	return price.(basetypes.ObjectValue).Attributes()["valid_from"].(types.String)
}

func getValidUntil(price attr.Value) types.String {
	return price.(basetypes.ObjectValue).Attributes()["valid_until"].(types.String)
}
