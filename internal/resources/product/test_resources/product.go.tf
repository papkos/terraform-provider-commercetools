{{- /*gotype: github.com/labd/terraform-provider-commercetools/internal/resources/product.productConfig*/ -}}
resource "commercetools_product" "{{ .Key }}" {
    key          = "{{ .Key }}"
    product_type = commercetools_product_type.{{ .ProductType.Key }}.id
    tax_category = commercetools_tax_category.{{ .TaxCategory.Key }}.id

    master_data {
        published = false
        current {
            name = {
                en = "{{ .Name }}"
            }
            slug = {
                en = "{{ .Key }}"
            }

            master_variant {
                {{ index .VariantConfigs 0 }}
            }

            {{- range $index, $element := slice .VariantConfigs 1 }}

            variant {
                {{ $element }}
            }
            {{ end }}
        }
        staged {
            name = {
                en = "{{ .Name }}"
            }
            slug = {
                en = "{{ .Key }}"
            }

            master_variant {
                {{ index .VariantConfigs 0 }}
            }

            {{- range $index, $element := slice .VariantConfigs 1 }}

            variant {
                {{ $element }}
            }
            {{ end }}
        }
    }
}
