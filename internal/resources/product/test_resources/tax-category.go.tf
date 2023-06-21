resource "commercetools_tax_category" "{{ .key }}" {
    key         = "{{ .key }}"
    name        = "Standard tax category"
    description = "Example category"
}
