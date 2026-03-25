package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/ybriffa/rfc3339"
)

type String interface {
	validator.Describer

	// ValidateString should perform the validation.
	ValidateString(context.Context, validator.StringRequest, *validator.StringResponse)
}

var _ validator.String = &rfc3339DurationValidator{}

type rfc3339DurationValidator struct{}

func (v rfc3339DurationValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v rfc3339DurationValidator) MarkdownDescription(_ context.Context) string {
	return "value must respect RFC3339 duration format"
}

func (v rfc3339DurationValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	value := request.ConfigValue.ValueString()

	_, err := rfc3339.ParseDuration(value)
	if err != nil {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			value,
		))
	}
}

func (v rfc3339DurationValidator) ValidateParameterString(ctx context.Context, request function.StringParameterValidatorRequest, response *function.StringParameterValidatorResponse) {
	if request.Value.IsNull() || request.Value.IsUnknown() {
		return
	}

	value := request.Value.ValueString()

	_, err := rfc3339.ParseDuration(value)
	if err != nil {
		response.Error = validatorfuncerr.InvalidParameterValueMatchFuncError(
			request.ArgumentPosition,
			v.Description(ctx),
			value,
		)
	}
}

// RFC3339Duration returns an AttributeValidator which ensures that any configured
// attribute or function parameter value:
//
//   - Is a string.
//   - Is conform with RFC3339 duration.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func RFC3339Duration() rfc3339DurationValidator {
	return rfc3339DurationValidator{}
}
