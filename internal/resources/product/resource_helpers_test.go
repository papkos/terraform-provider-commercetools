package product_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type variantConfig struct {
	Sku        string
	Prices     []priceConfig
	Attributes []product.ProductVariantAttribute
}

type priceConfig struct {
	Curr      string
	Val       int
	Country   string
	ValidFrom string
}

func testAccProductVariantAttributeConfig(pva product.ProductVariantAttribute) string {
	return utils.HCLTemplate(`
		name = "{{ .pva.Name.ValueString }}"
	{{ if not .pva.BoolValue.IsNull }}
		bool_value = {{ .pva.BoolValue.ValueBool }}
	{{ else if not .pva.TextValue.IsNull }}
		text_value = "{{ .pva.TextValue.ValueString }}"
	{{ else if not .pva.PTReferenceValue.IsNull }}
		product_type_reference_value = {{ .pva.PTReferenceValue.ValueString }}
	{{ end }}
	`, map[string]any{"pva": pva})
}

func testAccProductVariantConfig(variant variantConfig) string {
	return utils.HCLTemplate(`
				sku = "{{ .v.Sku }}"
				
{{ range $_, $price := .v.Prices }}
				price {
					value {
						cent_amount = "{{ $price.Val }}"
						currency_code = "{{ $price.Curr }}"
					}
	{{ if gt (len ($price.Country)) 0 }}
						country = "{{ $price.Country }}"
	{{ end }}
	{{ if gt (len ($price.ValidFrom)) 0 }}
						valid_from = "{{ $price.ValidFrom }}"
	{{ end }}
				}
{{ end }}

{{ range $_, $attr := .attributeConfigs }}
				attribute {
					{{ $attr }}
				}
{{ end }}
	`,
		map[string]any{
			"v":                variant,
			"attributeConfigs": pie.Map(variant.Attributes, testAccProductVariantAttributeConfig),
		})
}

type productType struct {
	Config string
	Name   string
}

func testAccProductTypeConfigSimple(t *testing.T) productType {
	return productType{
		Config: utils.HCLTemplate(`
			resource "commercetools_product_type" "tf-acctest-product-type-simple" {
				key         = "tf-acctest-product-type-simple"
				name        = "Some generic product properties"
				description = "All the generic product properties"
			}
		`, nil),
		Name: "tf-acctest-product-type-simple",
	}
}

func testAccProductTypeConfigWithAttributes(t *testing.T) productType {
	return productType{
		Config: utils.HCLTemplate(`
			resource "commercetools_product_type" "tf-acctest-product-type-with-attributes" {
				key         = "tf-acctest-product-type-with-attributes"
				name        = "Some generic product properties"
				description = "All the generic product properties"

				attribute {
					name = "bool_attr_name"
				
					constraint = "None"
					label = {
						en-US = "A TF test Bool attribute"
					}
					required   = false
					searchable = false
					type {
						name = "boolean"
					}
				}

				attribute {
					name = "text_attr_name"
				
					constraint = "None"
					label = {
						en-US = "A TF test Text attribute"
					}
					required   = false
					searchable = false
					type {
						name = "text"
					}
				}

				attribute {
					name = "pt_ref_attr_name"
				
					constraint = "None"
					label = {
						en-US = "A TF test Product Type Reference attribute"
					}
					required   = false
					searchable = false
					type {
						name = "reference"
						reference_type_id = "product-type"
					}
				}
			}
		`, nil),
		Name: "tf-acctest-product-type-with-attributes",
	}
}

func testAccProductConfig(t *testing.T, productType productType, name string, key string, variants []variantConfig) string {
	return utils.HCLTemplate(`
      {{ .productType.Config }}
      
      resource "commercetools_tax_category" "tf-acctest-tax-category" {
			key         = "tf-acctest-tax-category"
			name        = "Standard tax category"
			description = "Example category"
		}

      resource "commercetools_product" "tf-acctest-product" {
        key          = "{{ .key }}"
        product_type = commercetools_product_type.{{ .productType.Name }}.id
        tax_category = commercetools_tax_category.tf-acctest-tax-category.id
      
        master_data {
          published = false
          current {
            name = {
              en = "{{ .name }}"
            }
            slug = {
              en = "{{ .key }}"
            }

			master_variant {
				{{ index .variants 0 }}
			}

	{{ range $index, $element := slice .variants 1 }}
			variant {
				{{ $element }}
			}
	{{ end }}
          }
          staged {
            name = {
              en = "{{ .name }}"
            }
            slug = {
              en = "{{ .key }}"
            }

			master_variant {
				{{ index .variants 0 }}
			}

	{{ range $index, $element := slice .variants 1 }}
			variant {
				{{ $element }}
			}
	{{ end }}
          }
        }
      }
      `,
		map[string]any{
			"key":         key,
			"name":        name,
			"variants":    pie.Map(variants, testAccProductVariantConfig),
			"productType": productType,
		})
}

func testAccCheckProductDestroy(s *terraform.State) error {
	client, err := acctest.GetClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_product" {
			continue
		}
		response, err := client.Products().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("state (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := acctest.CheckApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}
