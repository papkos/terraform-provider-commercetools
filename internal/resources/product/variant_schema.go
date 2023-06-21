package product

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		},
		Blocks: map[string]schema.Block{
			"attribute": schema.SetNestedBlock{
				Description: "Attributes of the Product Variant.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the Attribute.",
							Required:    true,
						},
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
						//objectvalidator.ExactlyOneOf(
						//	path.MatchRoot("bool_value"),
						//	path.MatchRoot("text_value"),
						//	path.MatchRoot("localized_text_value"),
						//	path.MatchRoot("product_type_reference_value"),
						//),
					},
					PlanModifiers: nil,
				},
				Validators:    nil,
				PlanModifiers: nil,
			},
			"price": schema.ListNestedBlock{
				Description: "The Embedded Prices for the Product Variant. " +
					"Each Price must have its unique Price scope " +
					"(with same currency, country, Customer Group, Channel, validFrom and validUntil).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier of this Price.",
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
						},
					},
				},
				Validators: []validator.List{
					customvalidator.UniqueValuesFunc(func(price attr.Value) attr.Value {
						// 🤮
						return price.(basetypes.ObjectValue).Attributes()["value"].(basetypes.ListValue).Elements()[0].(basetypes.ObjectValue).Attributes()["currency_code"]
					}),
				},
			},
		},
	}
}
