package product_test

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
	"github.com/labd/terraform-provider-commercetools/internal/resources/product"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type productTypeConfig struct {
	Config string
	Key    string
}

type taxCategoryConfig struct {
	Config string
	Key    string
}

type productConfig struct {
	ProductType    productTypeConfig
	TaxCategory    taxCategoryConfig
	Key            string
	Name           string
	Variants       []variantConfig
	VariantConfigs []string // Populated automatically
}

type variantConfig struct {
	Sku              string
	Prices           []priceConfig
	Attributes       product.ProductVariantAttributes
	AttributesConfig string // Populated automatically
}

type priceConfig struct {
	Curr      string
	Val       int
	Country   string
	ValidFrom string
}

func testAccProductTypeConfigSimple() productTypeConfig {
	key := "tf-acctest-product-type-with-simple"
	return productTypeConfig{
		Config: utils.HCLTemplateFiles("test_resources/product-type-with-simple.go.tf")(map[string]any{"key": key}),
		Key:    key,
	}
}

func testAccProductTypeConfigWithAttributes() productTypeConfig {
	key := "tf-acctest-product-type-with-attributes"
	return productTypeConfig{
		Config: utils.HCLTemplateFiles("test_resources/product-type-with-attributes.go.tf")(map[string]any{"key": key}),
		Key:    key,
	}
}

func testAccTaxCategoryConfig() taxCategoryConfig {
	key := "tf-acctest-product-type-with-attributes"
	return taxCategoryConfig{
		Config: utils.HCLTemplateFiles("test_resources/tax-category.go.tf")(map[string]any{"key": key}),
		Key:    key,
	}
}

func testAccProductVariantAttributesConfig(pva product.ProductVariantAttributes) string {
	return utils.HCLTemplateFiles("test_resources/product-variant-attributes.go.tf")(pva)
}

func testAccProductVariantConfig(variant variantConfig) string {
	variant.AttributesConfig = testAccProductVariantAttributesConfig(variant.Attributes)

	return utils.HCLTemplateFiles("test_resources/product-variant-body.go.tf")(variant)
}

func testAccProductConfig(config productConfig) string {
	config.VariantConfigs = pie.Map(config.Variants, testAccProductVariantConfig)

	return utils.HCLTemplate(`
      {{ .productType.Config }}

      {{ .taxCategory.Config }}
      
      {{ .productConfig }}
      `,
		map[string]any{
			"productType":   config.ProductType,
			"taxCategory":   config.TaxCategory,
			"productConfig": utils.HCLTemplateFiles("test_resources/product.go.tf")(config),
		},
	)
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
