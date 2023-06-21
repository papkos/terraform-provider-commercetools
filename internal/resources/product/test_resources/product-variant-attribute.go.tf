{{- /*gotype: github.com/labd/terraform-provider-commercetools/internal/resources/product.ProductVariantAttribute*/ -}}
attribute {
    name       = "{{ .Name.ValueString }}"

    {{ if not .BoolValue.IsNull }}
    bool_value = {{ .BoolValue.ValueBool }}{{- /* unquoted */ -}}
    {{ else if not .TextValue.IsNull }}
    text_value = "{{ .TextValue.ValueString }}"
    {{ else if not .PTReferenceValue.IsNull }}
    product_type_reference_value = {{ .PTReferenceValue.ValueString }}{{- /* unquoted */ -}}
    {{ end }}
}
