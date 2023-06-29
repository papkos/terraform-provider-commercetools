{{- /*gotype: github.com/labd/terraform-provider-commercetools/internal/resources/product.ProductVariantAttribute*/ -}}

{{range $name, $val := . }}
"{{$name}}" = {
    {{ if not $val.BoolValue.IsNull }}
    bool_value = {{ $val.BoolValue.ValueBool }}{{- /* unquoted */ -}}
    {{ else if not $val.TextValue.IsNull }}
    text_value = "{{ $val.TextValue.ValueString }}"
    {{ else if not $val.PTReferenceValue.IsNull }}
    product_type_reference_value = {{ $val.PTReferenceValue.ValueString }}{{- /* unquoted */ -}}
    {{ else if not $val.LocalizedTextValue.IsNull }}
    localized_text_value = {
        {{ range $key, $val := $val.LocalizedTextValue.ValueLocalizedString }}
        "{{ $key }}" = "{{ $val }}"
        {{ end }}
    }
    {{ end }}
}
{{ end }}
