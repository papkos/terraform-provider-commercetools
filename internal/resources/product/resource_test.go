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
	key := "tf-acc-test-product"
	resourceName := "commercetools_product.tf-acctest-test-product"
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
				),
			},
		},
	})
}

func testAccProductConfig(t *testing.T, name string, key string) string {
	return utils.HCLTemplate(`
      resource "commercetools_product_type" "some-generic-properties-product-type" {
        key         = "some-key"
        name        = "Some generic product properties"
        description = "All the generic product properties"
      }
      
      resource "commercetools_tax_category" "my-tax-category" {
			key         = "my-tax-category-key"
			name        = "Standard tax category"
			description = "Example category"
		}

      resource "commercetools_product" "tf-acctest-test-product" {
        key          = "{{ .key }}"
        product_type = commercetools_product_type.some-generic-properties-product-type.id
        tax_category = commercetools_tax_category.my-tax-category.id
      
        master_data {
          published = false
          current {
            name = {
              en = "{{ .name }}"
            }
            slug = {
              en = "{{ .key }}"
            }
          }
          staged {
            name = {
              en = "{{ .name }}"
            }
            slug = {
              en = "{{ .key }}"
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
