package product_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func TestAccProduct_basic(t *testing.T) {

	name := "TF ACC test product"
	key := "tf-acctest-product"
	resourceName := "commercetools_product.tf-acctest-product"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProductConfig(t, name, key, []variantConfig{
					{
						Sku: "tf-testacc-sku1",
						Prices: []priceConfig{
							{Curr: "USD", Val: 1000, Country: "US", ValidFrom: "2023-09-15T12:34:56Z"},
							{Curr: "NOK", Val: 2000, Country: "NO"},
						},
					},
					{
						Sku: "tf-testacc-sku2",
						Prices: []priceConfig{
							{Curr: "USD", Val: 4000, Country: "US", ValidFrom: "2024-09-15T12:34:56Z"},
							{Curr: "NOK", Val: 8000, Country: "NO"},
						},
					},
					{
						Sku: "tf-testacc-sku3",
						Prices: []priceConfig{
							{Curr: "USD", Val: 16000, Country: "US", ValidFrom: "2025-09-15T12:34:56Z"},
							{Curr: "NOK", Val: 32000, Country: "NO"},
						},
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.sku", "tf-testacc-sku1"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.value.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.value.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.valid_from", "2023-09-15T12:34:56Z"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.value.0.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.value.0.currency_code", "NOK"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.country", "NO"),

					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.sku", "tf-testacc-sku2"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.0.value.0.cent_amount", "4000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.0.value.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.0.valid_from", "2024-09-15T12:34:56Z"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.1.value.0.cent_amount", "8000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.1.value.0.currency_code", "NOK"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.0.price.1.country", "NO"),
				),
			},
		},
	})
}

type variantConfig struct {
	Sku    string
	Prices []priceConfig
}

type priceConfig struct {
	Curr      string
	Val       int
	Country   string
	ValidFrom string
}

func testAccProductVariantConfig(t *testing.T, variant variantConfig) string {
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
	`,
		map[string]any{
			"v": variant,
		})
}

func testAccProductConfig(t *testing.T, name string, key string, variants []variantConfig) string {
	return utils.HCLTemplate(`
      resource "commercetools_product_type" "tf-acctest-product-type" {
        key         = "tf-acctest-product-type"
        name        = "Some generic product properties"
        description = "All the generic product properties"
      }
      
      resource "commercetools_tax_category" "tf-acctest-tax-category" {
			key         = "tf-acctest-tax-category"
			name        = "Standard tax category"
			description = "Example category"
		}

      resource "commercetools_product" "tf-acctest-product" {
        key          = "{{ .key }}"
        product_type = commercetools_product_type.tf-acctest-product-type.id
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
			"key":  key,
			"name": name,
			"variants": pie.Map(variants, func(vc variantConfig) string {
				return testAccProductVariantConfig(t, vc)
			}),
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
