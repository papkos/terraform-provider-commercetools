package product_test

import (
	"context"
	"fmt"
	"testing"

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
				Config: testAccProductConfig(t, name, key),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.sku", key+"_sku"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.value.0.cent_amount", "1000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.value.0.currency_code", "USD"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.country", "US"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.0.valid_from", "2023-09-15T12:34:56Z"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.value.0.cent_amount", "2000"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.value.0.currency_code", "NOK"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.price.1.country", "NO"),
				),
			},
		},
	})
}

func testAccProductConfig(t *testing.T, name string, key string) string {
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
				key = "{{ .key }}_master_variant"
				sku = "{{ .key }}_sku"
				
				price {
					value {
						cent_amount = 1000
						currency_code = "USD"
					}
					country = "US"
					valid_from = "2023-09-15T12:34:56Z"
				}
				
				price {
					value {
						cent_amount = 2000
						currency_code = "NOK"
					}
					country = "NO"
					valid_from = "2023-09-15T12:34:56Z"
				}
			}
          }
          staged {
            name = {
              en = "{{ .name }}"
            }
            slug = {
              en = "{{ .key }}"
            }
			
			master_variant {
				key = "{{ .key }}_master_variant"
				sku = "{{ .key }}_sku"
				
				price {
					value {
						cent_amount = 1000
						currency_code = "USD"
					}
					country = "US"
					valid_from = "2023-09-15T12:34:56Z"
				}
				
				price {
					value {
						cent_amount = 2000
						currency_code = "NOK"
					}
					country = "NO"
					valid_from = "2023-09-15T12:34:56Z"
				}
			}
          }
        }
      }
      `,
		map[string]any{
			"key":  key,
			"name": name,
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
