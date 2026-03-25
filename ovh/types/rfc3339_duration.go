package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/ybriffa/rfc3339"
)

// --- Value ---

// TfRFC3339DurationValue is a custom string value that implements semantic
// equality for RFC 3339 duration strings.  Two durations are considered equal
// when they represent the same time.Duration, even if their textual
// representations differ (e.g. "PT0S" vs "P0D").
type TfRFC3339DurationValue struct {
	basetypes.StringValue
}

var (
	_ basetypes.StringValuable                   = TfRFC3339DurationValue{}
	_ basetypes.StringValuableWithSemanticEquals = TfRFC3339DurationValue{}
)

func (t *TfRFC3339DurationValue) UnmarshalJSON(data []byte) error {
	var v *string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if v == nil {
		t.StringValue = basetypes.NewStringNull()
	} else {
		t.StringValue = basetypes.NewStringValue(*v)
	}

	return nil
}

func (t TfRFC3339DurationValue) MarshalJSON() ([]byte, error) {
	if t.IsNull() || t.IsUnknown() {
		return []byte("null"), nil
	}
	return json.Marshal(t.StringValue.ValueString())
}

func (v TfRFC3339DurationValue) Equal(o attr.Value) bool {
	other, ok := o.(TfRFC3339DurationValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v TfRFC3339DurationValue) Type(_ context.Context) attr.Type {
	return TfRFC3339DurationType{}
}

// StringSemanticEquals returns true when both values parse to the same
// time.Duration.  If either value fails to parse, it falls back to a
// plain string comparison.
func (v TfRFC3339DurationValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TfRFC3339DurationValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected TfRFC3339DurationValue, got: %T", newValuable),
		)
		return false, diags
	}

	oldStr := v.ValueString()
	newStr := newValue.ValueString()

	// Fast path: identical strings
	if oldStr == newStr {
		return true, diags
	}

	oldDur, err1 := rfc3339.ParseDuration(oldStr)
	newDur, err2 := rfc3339.ParseDuration(newStr)
	if err1 != nil || err2 != nil {
		// Cannot parse one of them; not semantically equal
		return false, diags
	}

	return oldDur == newDur, diags
}

// NewTfRFC3339DurationValue creates a known, non-null TfRFC3339DurationValue.
func NewTfRFC3339DurationValue(value string) TfRFC3339DurationValue {
	return TfRFC3339DurationValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewTfRFC3339DurationNull creates a null TfRFC3339DurationValue.
func NewTfRFC3339DurationNull() TfRFC3339DurationValue {
	return TfRFC3339DurationValue{
		StringValue: basetypes.NewStringNull(),
	}
}

// --- Type ---

// TfRFC3339DurationType is the attr.Type for TfRFC3339DurationValue.
type TfRFC3339DurationType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = TfRFC3339DurationType{}

func (t TfRFC3339DurationType) Equal(o attr.Type) bool {
	_, ok := o.(TfRFC3339DurationType)
	return ok
}

func (t TfRFC3339DurationType) String() string {
	return "TfRFC3339DurationType"
}

func (t TfRFC3339DurationType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return TfRFC3339DurationValue{StringValue: in}, nil
}

func (t TfRFC3339DurationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t TfRFC3339DurationType) ValueType(_ context.Context) attr.Value {
	return TfRFC3339DurationValue{}
}
