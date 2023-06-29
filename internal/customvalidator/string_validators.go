package customvalidator

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func DateTimeValidator() validator.String {
	return dateTimeValidator{}
}

var _ validator.String = dateTimeValidator{}

type dateTimeValidator struct {
}

func (v dateTimeValidator) Description(ctx context.Context) string {
	return "Value must be a valid timestamp in ISO 8601 format (YYYY-MM-DDThh:mm:ss.sssZ)"
}

func (v dateTimeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v dateTimeValidator) ValidateString(ctx context.Context, request validator.StringRequest, resp *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	_, err := time.Parse(time.RFC3339, value)
	if err != nil {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}
