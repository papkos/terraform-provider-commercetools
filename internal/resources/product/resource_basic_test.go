package product_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

func TestAccProduct_basic_create(t *testing.T) {
	name := "TF ACC test product"
	key := "tf-acctest-product"
	resourceName := "commercetools_product.tf-acctest-product"

	step1Config := testAccProductConfig(productConfig{
		ProductType: testAccProductTypeConfigSimple(),
		TaxCategory: testAccTaxCategoryConfig(),
		Key:         key,
		Name:        name,
		Variants: []variantConfig{
			{
				Sku: "tf-testacc-sku1",
				Prices: []priceConfig{
					{Curr: "USD", Val: 1000, Country: "US", ValidFrom: "2023-09-15T12:34:56Z"},
					{Curr: "NOK", Val: 2000, Country: "NO"},
				},
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductDestroy,

		Steps: []resource.TestStep{
			{
				Config: step1Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.#", "0"),

					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.sku", "tf-testacc-sku1"),
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
func TestAccProduct_basic_change(t *testing.T) {
	name := "TF ACC test product"
	key := "tf-acctest-product"
	resourceName := "commercetools_product.tf-acctest-product"

	step1 := productConfig{
		ProductType: testAccProductTypeConfigSimple(),
		TaxCategory: testAccTaxCategoryConfig(),
		Key:         key,
		Name:        name,
		Variants: []variantConfig{
			{
				Sku: "tf-testacc-sku1",
				Prices: []priceConfig{
					{Curr: "USD", Val: 1000, Country: "US", ValidFrom: "2023-09-15T12:34:56Z"},
					{Curr: "NOK", Val: 2000, Country: "NO"},
				},
			},
		},
	}
	step1Config := testAccProductConfig(step1)

	step2 := productConfig{
		ProductType: testAccProductTypeConfigSimple(),
		TaxCategory: testAccTaxCategoryConfig(),
		Key:         key,
		Name:        name + " CHANGED", // Changed
		// TODO It would be nice to introduce more changes, covering all the top-level diffs in Product.calculateUpdateActions()
		Variants: step1.Variants,
	}
	step2Config := testAccProductConfig(step2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductDestroy,

		Steps: []resource.TestStep{
			{
				Config: step1Config,
				Check: resource.ComposeTestCheckFunc(
					// Just some basic checks, more detailed testing in TestAccProduct_basic_create
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
				),
			},
			{
				Config: step2Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name+" CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
				),
			},
		},
	})
}
