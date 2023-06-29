{{- /*gotype: github.com/labd/terraform-provider-commercetools/internal/resources/product.variantConfig*/ -}}
{{- /* Will be used with either 'master_variant' or 'variant' block name */ -}}
sku = "{{ .Sku }}"

{{ range $_, $price := .Prices }}
price {
    value {
        cent_amount   = "{{ $price.Val }}"
        currency_code = "{{ $price.Curr }}"
    }
    {{ if gt (len ($price.Country)) 0 }}
    country    = "{{ $price.Country }}"
    {{ end }}
    {{ if gt (len ($price.ValidFrom)) 0 }}
    valid_from = "{{ $price.ValidFrom }}"
    {{ end }}
}
{{ end }}

    attributes = {
        {{ .AttributesConfig }}
    }
