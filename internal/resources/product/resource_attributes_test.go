package product_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product"
)

func TestAccProduct_attributes(t *testing.T) {

	name := "TF ACC test product"
	key := "tf-acctest-product"
	resourceName := fmt.Sprintf("commercetools_product.%s", key)
	productTypeConfigWithAttributes := testAccProductTypeConfigWithAttributes()

	step1Config := testAccProductConfig(productConfig{
		ProductType: productTypeConfigWithAttributes,
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
				Attributes: []product.ProductVariantAttribute{
					{Name: types.StringValue("bool_attr_name"), BoolValue: types.BoolValue(true)},
					{Name: types.StringValue("text_attr_name"), TextValue: types.StringValue("text attribute value")},
					{Name: types.StringValue("pt_ref_attr_name"), PTReferenceValue: types.StringValue(fmt.Sprintf(`commercetools_product_type.%s.id`, productTypeConfigWithAttributes.Key))}, // These probably should be UUIDs...
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
