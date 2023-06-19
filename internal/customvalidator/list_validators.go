package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.List = uniqueFuncValuesValidator{}

// uniqueFuncValuesValidator implements the validator.
type uniqueFuncValuesValidator struct {
	keyFunc func(elem attr.Value) attr.Value
}

// Description returns the plaintext description of the validator.
func (v uniqueFuncValuesValidator) Description(_ context.Context) string {
	return "all values must be unique"
}

// MarkdownDescription returns the Markdown description of the validator.
func (v uniqueFuncValuesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList implements the validation logic.
func (v uniqueFuncValuesValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()

	for indexOuter, elementOuter := range elements {
		// Only evaluate known values for duplicates.
		if elementOuter.IsUnknown() {
			continue
		}

		outerKey := v.keyFunc(elementOuter)

		for indexInner := indexOuter + 1; indexInner < len(elements); indexInner++ {
			elementInner := elements[indexInner]

			if elementInner.IsUnknown() {
				continue
			}

			innerKey := v.keyFunc(elementInner)

			if !innerKey.Equal(outerKey) {
				continue
			}

			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Duplicate List Value",
				fmt.Sprintf("This attribute contains duplicate values of: %s", elementInner),
			)
		}
	}
}

// UniqueValuesFunc returns a validator which ensures that any configured list
// only contains unique values. This is similar to using a set attribute type
// which inherently validates unique values, but with list ordering semantics.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func UniqueValuesFunc(keyFunc func(elem attr.Value) attr.Value) validator.List {
	return uniqueFuncValuesValidator{keyFunc: keyFunc}
}
