package product

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
)

func productDataSchema(description string) schema.Block {
	// These items all have maximal one item. We don't use SingleNestedBlock
	// here since it isn't quite robust currently.
	// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
	return schema.ListNestedBlock{
		Description: description,
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"name": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:            false,
					Description:         "Name of the Product.",
					MarkdownDescription: "",
				}),
				"categories": schema.ListAttribute{
					ElementType:         types.StringType,
					Description:         "IDs of Categories assigned to the Product.",
					MarkdownDescription: "",
					Validators:          nil,
					Optional:            true,
					Computed:            true,
					PlanModifiers: []planmodifier.List{
						listplanmodifier.UseStateForUnknown(),
					},
				},
				// TODO CategoryOrderHints https://docs.commercetools.com/api/projects/products#productdata
				"description": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:    true,
					Description: "Description of the Product.",
				}),
				"slug": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:            false,
					Description:         "User-defined identifier used in a deep-link URL for the Product. Must be unique across a Project, but can be the same for Products in different Locales.",
					MarkdownDescription: "",
					// TODO make it possible to add (value) validators
					// Validators: []validator.String{
					//	stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z0-9_-]{2,256}`), ""),
					// },
				}),
				"meta_title": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:    true,
					Description: "Title of the Product displayed in search results.",
				}),
				"meta_description": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:    true,
					Description: "Description of the Product displayed in search results below the meta title.",
				}),
				"meta_keywords": customtypes.LocalizedString(customtypes.LocalizedStringOpts{
					Optional:    true,
					Description: "Keywords that give additional information about the Product to search engines.",
				}),
				// TODO masterVariant is a hard link to a productVariant -- which may or may not be dependent on this product
				// TODO variants is a list of productVariants -- use references?
				// TODO searchKeywords is required, but it is just a JSON (with schema) https://docs.commercetools.com/api/projects/products#productdata
			},
			Blocks: map[string]schema.Block{
				"master_variant": schema.ListNestedBlock{
					// These items all have maximal one item. We don't use SingleNestedBlock
					// here since it isn't quite robust currently.
					// See https://github.com/hashicorp/terraform-plugin-framework/issues/603
					Description:  "The Master Variant of the Product.",
					NestedObject: productVariantSchema(),
					Validators: []validator.List{
						listvalidator.SizeAtMost(1),
					},
				},
				"variant": schema.ListNestedBlock{
					Description:  "Additional Product Variants.",
					NestedObject: productVariantSchema(),
					Validators: []validator.List{
						customvalidator.UniqueValuesKeyFunc(func(variant attr.Value) attr.Value {
							// 🤮
							return variant.(basetypes.ObjectValue).Attributes()["sku"]
						}),
					},
				},
			},
			CustomType:    nil,
			Validators:    nil,
			PlanModifiers: nil,
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}
