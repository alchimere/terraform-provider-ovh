package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = requireAttributesValidator{}

type requireAttributesValidator struct {
	childNames []string
}

func (v requireAttributesValidator) Description(_ context.Context) string {
	return fmt.Sprintf("all attributes %v must be specified when the block is present", v.childNames)
}

func (v requireAttributesValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requireAttributesValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	attrs := req.ConfigValue.Attributes()
	for _, name := range v.childNames {
		val, ok := attrs[name]
		if !ok || val.IsNull() || val.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				req.Path.AtName(name),
				"Missing Required Attribute",
				fmt.Sprintf("The attribute %q must be specified when the block is present.", name),
			)
		}
	}
}

// RequireAttributesWhenBlockPresent returns an Object validator that ensures
// all the named child attributes are set whenever the enclosing block is
// present in the configuration. This is useful for SingleNestedBlock
// attributes whose children should be required when the block is written
// but where the block itself is optional.
//
// Null (unconfigured) and unknown (known after apply) blocks are skipped.
func RequireAttributesWhenBlockPresent(childNames ...string) validator.Object {
	return requireAttributesValidator{childNames: childNames}
}
