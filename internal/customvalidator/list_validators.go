package customvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// UniqueValuesKeyFunc returns a validator which ensures that any configured list
// only contains unique values. This is similar to using a set attribute type
// which inherently validates unique values, but with list ordering semantics.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func UniqueValuesKeyFunc(keyFunc func(elem attr.Value) attr.Value) validator.List {
	return uniqueValuesKeyFuncValidator{keyFunc: keyFunc}
}

// UniqueValuesCompareFunc returns a validator which ensures that any configured list
// only contains unique values. This is similar to using a set attribute type
// which inherently validates unique values, but with list ordering semantics.
// Null (unconfigured) and unknown (known after apply) values are skipped.
func UniqueValuesCompareFunc(equalFunc func(a, b attr.Value) bool) validator.List {
	return uniqueValuesCompareFuncValidator{equalFunc: equalFunc}
}

var _ validator.List = uniqueValuesKeyFuncValidator{}

// uniqueValuesKeyFuncValidator implements the validator.
type uniqueValuesKeyFuncValidator struct {
	keyFunc func(elem attr.Value) attr.Value
}

// Description returns the plaintext description of the validator.
func (v uniqueValuesKeyFuncValidator) Description(_ context.Context) string {
	return "all values must be unique"
}

// MarkdownDescription returns the Markdown description of the validator.
func (v uniqueValuesKeyFuncValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList implements the validation logic.
func (v uniqueValuesKeyFuncValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
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

var _ validator.List = uniqueValuesCompareFuncValidator{}

// uniqueValuesKeyFuncValidator implements the validator.
type uniqueValuesCompareFuncValidator struct {
	equalFunc func(a, b attr.Value) bool
}

// Description returns the plaintext description of the validator.
func (v uniqueValuesCompareFuncValidator) Description(_ context.Context) string {
	return "all values must be unique"
}

// MarkdownDescription returns the Markdown description of the validator.
func (v uniqueValuesCompareFuncValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateList implements the validation logic.
func (v uniqueValuesCompareFuncValidator) ValidateList(_ context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	elements := req.ConfigValue.Elements()

	for indexOuter, elementOuter := range elements {
		// Only evaluate known values for duplicates.
		if elementOuter.IsUnknown() {
			continue
		}

		for indexInner := indexOuter + 1; indexInner < len(elements); indexInner++ {
			elementInner := elements[indexInner]

			if elementInner.IsUnknown() {
				continue
			}

			if !v.equalFunc(elementOuter, elementInner) {
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
