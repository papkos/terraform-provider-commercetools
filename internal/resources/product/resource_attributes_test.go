package product_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/customtypes"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product"
)

func TestAccProduct_attributes_create(t *testing.T) {

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
				Attributes: product.ProductVariantAttributes{
					"bool_attr_name":   {BoolValue: types.BoolValue(true)},
					"text_attr_name":   {TextValue: types.StringValue("text attribute value")},
					"pt_ref_attr_name": {PTReferenceValue: types.StringValue(fmt.Sprintf(`commercetools_product_type.%s.id`, productTypeConfigWithAttributes.Key))},
				},
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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
func TestAccProduct_attributes_change(t *testing.T) {

	name := "TF ACC test product"
	key := "tf-acctest-product"
	resourceName := fmt.Sprintf("commercetools_product.%s", key)
	productTypeConfigWithAttributes := testAccProductTypeConfigWithAttributes()

	step1 := productConfig{
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
				Attributes: product.ProductVariantAttributes{
					"bool_attr_name":   {BoolValue: types.BoolValue(true)},
					"text_attr_name":   {TextValue: types.StringValue("text attribute value")},
					"pt_ref_attr_name": {PTReferenceValue: types.StringValue(fmt.Sprintf(`commercetools_product_type.%s.id`, productTypeConfigWithAttributes.Key))},
				},
			},
		},
	}
	step1Config := testAccProductConfig(step1)

	step2 := productConfig{
		ProductType: step1.ProductType,
		TaxCategory: step1.TaxCategory,
		Key:         step1.Key,
		Name:        step1.Name,
		Variants: []variantConfig{
			{
				Sku: "tf-testacc-sku1",
				Prices: []priceConfig{
					{Curr: "USD", Val: 1000, Country: "US", ValidFrom: "2023-09-15T12:34:56Z"},
					{Curr: "NOK", Val: 2000, Country: "NO"},
				},
				Attributes: product.ProductVariantAttributes{
					"bool_attr_name": {BoolValue: types.BoolValue(true)},
					"text_attr_name": {TextValue: types.StringValue("text attribute value")},
					// product_type_attribute removed
				},
			},
		},
	}
	step2Config := testAccProductConfig(step2)

	step3 := productConfig{
		ProductType: step2.ProductType,
		TaxCategory: step2.TaxCategory,
		Key:         step2.Key,
		Name:        step2.Name,
		Variants: []variantConfig{
			{
				Sku:    step2.Variants[0].Sku,
				Prices: step2.Variants[0].Prices,
				Attributes: product.ProductVariantAttributes{
					"bool_attr_name":           {BoolValue: types.BoolValue(true)},
					"text_attr_name":           {TextValue: types.StringValue("text attribute value")},
					"localized_text_attr_name": {LocalizedTextValue: customtypes.NewLocalizedStringValue(map[string]attr.Value{"en-US": types.StringValue("ltext value")})},
				},
			},
		},
	}
	step3Config := testAccProductConfig(step3)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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
			{
				Config: step2Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.#", "0"),
				),
			},
			{
				Config: step3Config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.name.en", name),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.variant.#", "0"),

					resource.TestCheckResourceAttr(resourceName, "master_data.0.current.0.master_variant.0.attributes.%", "3"),
				),
			},
		},
	})
}
