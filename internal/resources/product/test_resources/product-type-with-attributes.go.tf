resource "commercetools_product_type" "{{ .key }}" {
    key         = "{{ .key }}"
    name        = "TF ACCTEST Some generic product properties"

    attribute {
        name = "bool_attr_name"

        constraint = "None"
        label      = {
            en-US = "A TF test Bool attribute"
        }
        required   = false
        searchable = false
        type {
            name = "boolean"
        }
    }

    attribute {
        name = "text_attr_name"

        constraint = "None"
        label      = {
            en-US = "A TF test Text attribute"
        }
        required   = false
        searchable = false
        type {
            name = "text"
        }
    }

    attribute {
        name = "pt_ref_attr_name"

        constraint = "None"
        label      = {
            en-US = "A TF test Product Type Reference attribute"
        }
        required   = false
        searchable = false
        type {
            name              = "reference"
            reference_type_id = "product-type"
        }
    }
}
