package ovh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Schema ---

func CloudProjectKubeNodePoolDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			// Required inputs
			"service_name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    os.Getenv("OVH_CLOUD_PROJECT_SERVICE") == "",
				Optional:    os.Getenv("OVH_CLOUD_PROJECT_SERVICE") != "",
				Description: "Service name",
			},
			"kube_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Kube ID",
			},
			"name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "NodePool resource name",
			},

			// Computed
			"id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "ID of the nodepool",
			},
			"autoscale": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Computed:    true,
				Description: "Enable auto-scaling for the pool",
			},
			"anti_affinity": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Computed:    true,
				Description: "Enable anti affinity groups for nodes in the pool",
			},
			"flavor_name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Flavor name",
			},
			"desired_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"max_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"min_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"monthly_billed": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Computed:    true,
				Description: "Enable monthly billing on all nodes in the pool",
			},
			"available_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes which are actually ready in the pool",
			},
			"created_at": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Creation date",
			},
			"current_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes present in the pool",
			},
			"flavor": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Flavor name",
			},
			"project_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Project id",
			},
			"size_status": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Status describing the state between number of nodes wanted and available ones",
			},
			"status": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Current status",
			},
			"up_to_date_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "Number of nodes with latest version installed in the pool",
			},
			"updated_at": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Computed:    true,
				Description: "Last update date",
			},
			"autoscaling_scale_down_unneeded_time_seconds": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "scaleDownUnneededTimeSeconds for autoscaling",
			},
			"autoscaling_scale_down_unready_time_seconds": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Computed:    true,
				Description: "scaleDownUnreadyTimeSeconds for autoscaling",
			},
			"autoscaling_scale_down_utilization_threshold": schema.NumberAttribute{
				CustomType:  ovhtypes.TfNumberType{},
				Computed:    true,
				Description: "scaleDownUtilizationThreshold for autoscaling",
			},

			"availability_zones": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Optional:    true,
				Computed:    true,
				Description: "Availability zones",
			},
		},
		Blocks: map[string]schema.Block{
			"template": schema.SingleNestedBlock{
				// NestedObject: schema.NestedBlockObject{
				Attributes: map[string]schema.Attribute{
					"metadata": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"finalizers": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Computed:    true,
								Description: "finalizers",
							},
							"labels": schema.MapAttribute{
								CustomType:  ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
								Computed:    true,
								Description: "labels",
							},
							"annotations": schema.MapAttribute{
								CustomType:  ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
								Computed:    true,
								Description: "annotations",
							},
						},
						CustomType: NodePoolTemplateMetadataType{
							ObjectType: types.ObjectType{
								AttrTypes: NodePoolTemplateMetadataValue{}.AttributeTypes(ctx),
							},
						},
						Computed:    true,
						Description: "metadata",
					},
					"spec": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"unschedulable": schema.BoolAttribute{
								CustomType:  ovhtypes.TfBoolType{},
								Computed:    true,
								Description: "unschedulable",
							},
							"taints": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"effect": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Computed:    true,
											Description: "effect",
										},
										"key": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Computed:    true,
											Description: "key",
										},
										"value": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Computed:    true,
											Description: "value",
										},
									},
									CustomType: NodePoolTaintType{
										ObjectType: types.ObjectType{
											AttrTypes: NodePoolTaintValue{}.AttributeTypes(ctx),
										},
									},
								},
								CustomType:  ovhtypes.NewTfListNestedType[NodePoolTaintValue](ctx),
								Computed:    true,
								Description: "taints",
							},
						},
						CustomType: NodePoolTemplateSpecType{
							ObjectType: types.ObjectType{
								AttrTypes: NodePoolTemplateSpecValue{}.AttributeTypes(ctx),
							},
						},
						Computed:    true,
						Description: "spec",
					},
				},
				// },
				Description: "Node pool template",
			},
		},
	}
}

// --- Model ---

type CloudProjectKubeNodePoolDataSourceModel struct {
	ServiceName                              ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId                                   ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	Name                                     ovhtypes.TfStringValue                             `tfsdk:"name" json:"name"`
	Id                                       ovhtypes.TfStringValue                             `tfsdk:"id" json:"id"`
	Autoscale                                ovhtypes.TfBoolValue                               `tfsdk:"autoscale" json:"autoscale"`
	AntiAffinity                             ovhtypes.TfBoolValue                               `tfsdk:"anti_affinity" json:"antiAffinity"`
	FlavorName                               ovhtypes.TfStringValue                             `tfsdk:"flavor_name"`
	DesiredNodes                             ovhtypes.TfInt64Value                              `tfsdk:"desired_nodes" json:"desiredNodes"`
	MaxNodes                                 ovhtypes.TfInt64Value                              `tfsdk:"max_nodes" json:"maxNodes"`
	MinNodes                                 ovhtypes.TfInt64Value                              `tfsdk:"min_nodes" json:"minNodes"`
	MonthlyBilled                            ovhtypes.TfBoolValue                               `tfsdk:"monthly_billed" json:"monthlyBilled"`
	AvailableNodes                           ovhtypes.TfInt64Value                              `tfsdk:"available_nodes" json:"availableNodes"`
	CreatedAt                                ovhtypes.TfStringValue                             `tfsdk:"created_at" json:"createdAt"`
	CurrentNodes                             ovhtypes.TfInt64Value                              `tfsdk:"current_nodes" json:"currentNodes"`
	Flavor                                   ovhtypes.TfStringValue                             `tfsdk:"flavor" json:"flavor"`
	ProjectId                                ovhtypes.TfStringValue                             `tfsdk:"project_id" json:"projectId"`
	SizeStatus                               ovhtypes.TfStringValue                             `tfsdk:"size_status" json:"sizeStatus"`
	Status                                   ovhtypes.TfStringValue                             `tfsdk:"status" json:"status"`
	UpToDateNodes                            ovhtypes.TfInt64Value                              `tfsdk:"up_to_date_nodes" json:"upToDateNodes"`
	UpdatedAt                                ovhtypes.TfStringValue                             `tfsdk:"updated_at" json:"updatedAt"`
	Autoscaling                              cloudProjectKubeNodePoolAutoscalingJSON            `tfsdk:"-" json:"autoscaling"`
	AutoscalingScaleDownUnneededTimeSeconds  ovhtypes.TfInt64Value                              `tfsdk:"autoscaling_scale_down_unneeded_time_seconds"`
	AutoscalingScaleDownUnreadyTimeSeconds   ovhtypes.TfInt64Value                              `tfsdk:"autoscaling_scale_down_unready_time_seconds"`
	AutoscalingScaleDownUtilizationThreshold ovhtypes.TfNumberValue                             `tfsdk:"autoscaling_scale_down_utilization_threshold"`
	Template                                 NodePoolTemplateValue                              `tfsdk:"template" json:"template"`
	AvailabilityZones                        ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"availability_zones" json:"availabilityZones"`
}

// cloudProjectKubeNodePoolAutoscalingJSON is used only for JSON deserialization
// of the nested autoscaling object from the API response, which is then
// flattened into top-level attributes.
type cloudProjectKubeNodePoolAutoscalingJSON struct {
	ScaleDownUtilizationThreshold *float64 `json:"scaleDownUtilizationThreshold,omitempty"`
	ScaleDownUnneededTimeSeconds  *int64   `json:"scaleDownUnneededTimeSeconds,omitempty"`
	ScaleDownUnreadyTimeSeconds   *int64   `json:"scaleDownUnreadyTimeSeconds,omitempty"`
}

// --- NodePoolTemplateType / NodePoolTemplateValue ---

var _ basetypes.ObjectTypable = NodePoolTemplateType{}

type NodePoolTemplateType struct {
	basetypes.ObjectType
}

func (t NodePoolTemplateType) Equal(o attr.Type) bool {
	other, ok := o.(NodePoolTemplateType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t NodePoolTemplateType) String() string {
	return "NodePoolTemplateType"
}

func (t NodePoolTemplateType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := in.Attributes()

	metadataAttribute, ok := attributes["metadata"]
	if !ok {
		diags.AddError("Attribute Missing", `metadata is missing from object`)
		return nil, diags
	}
	metadataVal, ok := metadataAttribute.(NodePoolTemplateMetadataValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`metadata expected to be NodePoolTemplateMetadataValue, was: %T`, metadataAttribute))
	}

	specAttribute, ok := attributes["spec"]
	if !ok {
		diags.AddError("Attribute Missing", `spec is missing from object`)
		return nil, diags
	}
	specVal, ok := specAttribute.(NodePoolTemplateSpecValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`spec expected to be NodePoolTemplateSpecValue, was: %T`, specAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return NodePoolTemplateValue{
		Metadata: metadataVal,
		Spec:     specVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateValueNull() NodePoolTemplateValue {
	return NodePoolTemplateValue{
		state: attr.ValueStateNull,
	}
}

func NewNodePoolTemplateValueUnknown() NodePoolTemplateValue {
	return NodePoolTemplateValue{
		state: attr.ValueStateUnknown,
	}
}

func NewNodePoolTemplateValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (NodePoolTemplateValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]
		if !ok {
			diags.AddError(
				"Missing NodePoolTemplateValue Attribute Value",
				"While creating a NodePoolTemplateValue value, a missing attribute value was detected. "+
					"A NodePoolTemplateValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)
			continue
		}
		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid NodePoolTemplateValue Attribute Type",
				"While creating a NodePoolTemplateValue value, an invalid attribute value was detected. "+
					"A NodePoolTemplateValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("NodePoolTemplateValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]
		if !ok {
			diags.AddError(
				"Extra NodePoolTemplateValue Attribute Value",
				"While creating a NodePoolTemplateValue value, an extra attribute value was detected. "+
					"A NodePoolTemplateValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra NodePoolTemplateValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewNodePoolTemplateValueUnknown(), diags
	}

	metadataAttribute, ok := attributes["metadata"]
	if !ok {
		diags.AddError("Attribute Missing", `metadata is missing from object`)
		return NewNodePoolTemplateValueUnknown(), diags
	}
	metadataVal, ok := metadataAttribute.(NodePoolTemplateMetadataValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`metadata expected to be NodePoolTemplateMetadataValue, was: %T`, metadataAttribute))
	}

	specAttribute, ok := attributes["spec"]
	if !ok {
		diags.AddError("Attribute Missing", `spec is missing from object`)
		return NewNodePoolTemplateValueUnknown(), diags
	}
	specVal, ok := specAttribute.(NodePoolTemplateSpecValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`spec expected to be NodePoolTemplateSpecValue, was: %T`, specAttribute))
	}

	if diags.HasError() {
		return NewNodePoolTemplateValueUnknown(), diags
	}

	return NodePoolTemplateValue{
		Metadata: metadataVal,
		Spec:     specVal,
		state:    attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) NodePoolTemplateValue {
	object, diags := NewNodePoolTemplateValue(attributeTypes, attributes)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}
		panic("NewNodePoolTemplateValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return object
}

func (t NodePoolTemplateType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewNodePoolTemplateValueNull(), nil
	}
	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		return NewNodePoolTemplateValueUnknown(), nil
	}
	if in.IsNull() {
		return NewNodePoolTemplateValueNull(), nil
	}

	attributes := map[string]attr.Value{}
	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}

	return NewNodePoolTemplateValueMust(NodePoolTemplateValue{}.AttributeTypes(ctx), attributes), nil
}

func (t NodePoolTemplateType) ValueType(ctx context.Context) attr.Value {
	return NodePoolTemplateValue{}
}

var _ basetypes.ObjectValuable = NodePoolTemplateValue{}

type NodePoolTemplateValue struct {
	Metadata NodePoolTemplateMetadataValue `tfsdk:"metadata" json:"metadata"`
	Spec     NodePoolTemplateSpecValue     `tfsdk:"spec" json:"spec"`
	state    attr.ValueState
}

func (v *NodePoolTemplateValue) UnmarshalJSON(data []byte) error {
	type JsonNodePoolTemplateValue NodePoolTemplateValue

	var tmp JsonNodePoolTemplateValue
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Metadata = tmp.Metadata
	v.Spec = tmp.Spec
	v.state = attr.ValueStateKnown

	return nil
}

func (v NodePoolTemplateValue) Attributes() map[string]attr.Value {
	return map[string]attr.Value{
		"metadata": v.Metadata,
		"spec":     v.Spec,
	}
}

func (v NodePoolTemplateValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)
	attrTypes["metadata"] = NodePoolTemplateMetadataValue{}.Type(ctx).TerraformType(ctx)
	attrTypes["spec"] = NodePoolTemplateSpecValue{}.Type(ctx).TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err := v.Metadata.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["metadata"] = val

		val, err = v.Spec.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["spec"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v NodePoolTemplateValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v NodePoolTemplateValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v NodePoolTemplateValue) String() string {
	return "NodePoolTemplateValue"
}

func (v NodePoolTemplateValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		map[string]attr.Type{
			"metadata": NodePoolTemplateMetadataType{
				basetypes.ObjectType{AttrTypes: NodePoolTemplateMetadataValue{}.AttributeTypes(ctx)},
			},
			"spec": NodePoolTemplateSpecType{
				basetypes.ObjectType{AttrTypes: NodePoolTemplateSpecValue{}.AttributeTypes(ctx)},
			},
		},
		map[string]attr.Value{
			"metadata": v.Metadata,
			"spec":     v.Spec,
		})
}

func (v NodePoolTemplateValue) Equal(o attr.Value) bool {
	other, ok := o.(NodePoolTemplateValue)
	if !ok {
		return false
	}
	if v.state != other.state {
		return false
	}
	if v.state != attr.ValueStateKnown {
		return true
	}
	if !v.Metadata.Equal(other.Metadata) {
		return false
	}
	if !v.Spec.Equal(other.Spec) {
		return false
	}
	return true
}

func (v NodePoolTemplateValue) Type(ctx context.Context) attr.Type {
	return NodePoolTemplateType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v NodePoolTemplateValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"metadata": NodePoolTemplateMetadataType{
			basetypes.ObjectType{AttrTypes: NodePoolTemplateMetadataValue{}.AttributeTypes(ctx)},
		},
		"spec": NodePoolTemplateSpecType{
			basetypes.ObjectType{AttrTypes: NodePoolTemplateSpecValue{}.AttributeTypes(ctx)},
		},
	}
}

// --- NodePoolTemplateMetadataType / NodePoolTemplateMetadataValue ---

var _ basetypes.ObjectTypable = NodePoolTemplateMetadataType{}

type NodePoolTemplateMetadataType struct {
	basetypes.ObjectType
}

func (t NodePoolTemplateMetadataType) Equal(o attr.Type) bool {
	other, ok := o.(NodePoolTemplateMetadataType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t NodePoolTemplateMetadataType) String() string {
	return "NodePoolTemplateMetadataType"
}

func (t NodePoolTemplateMetadataType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := in.Attributes()

	finalizersAttribute, ok := attributes["finalizers"]
	if !ok {
		diags.AddError("Attribute Missing", `finalizers is missing from object`)
		return nil, diags
	}
	finalizersVal, ok := finalizersAttribute.(ovhtypes.TfListNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`finalizers expected to be ovhtypes.TfListNestedValue[ovhtypes.TfStringValue], was: %T`, finalizersAttribute))
	}

	labelsAttribute, ok := attributes["labels"]
	if !ok {
		diags.AddError("Attribute Missing", `labels is missing from object`)
		return nil, diags
	}
	labelsVal, ok := labelsAttribute.(ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`labels expected to be ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue], was: %T`, labelsAttribute))
	}

	annotationsAttribute, ok := attributes["annotations"]
	if !ok {
		diags.AddError("Attribute Missing", `annotations is missing from object`)
		return nil, diags
	}
	annotationsVal, ok := annotationsAttribute.(ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`annotations expected to be ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue], was: %T`, annotationsAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return NodePoolTemplateMetadataValue{
		Finalizers:  finalizersVal,
		Labels:      labelsVal,
		Annotations: annotationsVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateMetadataValueNull() NodePoolTemplateMetadataValue {
	return NodePoolTemplateMetadataValue{
		state: attr.ValueStateNull,
	}
}

func NewNodePoolTemplateMetadataValueUnknown() NodePoolTemplateMetadataValue {
	return NodePoolTemplateMetadataValue{
		state: attr.ValueStateUnknown,
	}
}

func NewNodePoolTemplateMetadataValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (NodePoolTemplateMetadataValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]
		if !ok {
			diags.AddError(
				"Missing NodePoolTemplateMetadataValue Attribute Value",
				"While creating a NodePoolTemplateMetadataValue value, a missing attribute value was detected. "+
					"A NodePoolTemplateMetadataValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateMetadataValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)
			continue
		}
		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid NodePoolTemplateMetadataValue Attribute Type",
				"While creating a NodePoolTemplateMetadataValue value, an invalid attribute value was detected. "+
					"A NodePoolTemplateMetadataValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateMetadataValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("NodePoolTemplateMetadataValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]
		if !ok {
			diags.AddError(
				"Extra NodePoolTemplateMetadataValue Attribute Value",
				"While creating a NodePoolTemplateMetadataValue value, an extra attribute value was detected. "+
					"A NodePoolTemplateMetadataValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra NodePoolTemplateMetadataValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewNodePoolTemplateMetadataValueUnknown(), diags
	}

	finalizersAttribute, ok := attributes["finalizers"]
	if !ok {
		diags.AddError("Attribute Missing", `finalizers is missing from object`)
		return NewNodePoolTemplateMetadataValueUnknown(), diags
	}
	finalizersVal, ok := finalizersAttribute.(ovhtypes.TfListNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`finalizers expected to be ovhtypes.TfListNestedValue[ovhtypes.TfStringValue], was: %T`, finalizersAttribute))
	}

	labelsAttribute, ok := attributes["labels"]
	if !ok {
		diags.AddError("Attribute Missing", `labels is missing from object`)
		return NewNodePoolTemplateMetadataValueUnknown(), diags
	}
	labelsVal, ok := labelsAttribute.(ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`labels expected to be ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue], was: %T`, labelsAttribute))
	}

	annotationsAttribute, ok := attributes["annotations"]
	if !ok {
		diags.AddError("Attribute Missing", `annotations is missing from object`)
		return NewNodePoolTemplateMetadataValueUnknown(), diags
	}
	annotationsVal, ok := annotationsAttribute.(ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`annotations expected to be ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue], was: %T`, annotationsAttribute))
	}

	if diags.HasError() {
		return NewNodePoolTemplateMetadataValueUnknown(), diags
	}

	return NodePoolTemplateMetadataValue{
		Finalizers:  finalizersVal,
		Labels:      labelsVal,
		Annotations: annotationsVal,
		state:       attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateMetadataValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) NodePoolTemplateMetadataValue {
	object, diags := NewNodePoolTemplateMetadataValue(attributeTypes, attributes)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}
		panic("NewNodePoolTemplateMetadataValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return object
}

func (t NodePoolTemplateMetadataType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewNodePoolTemplateMetadataValueNull(), nil
	}
	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		return NewNodePoolTemplateMetadataValueUnknown(), nil
	}
	if in.IsNull() {
		return NewNodePoolTemplateMetadataValueNull(), nil
	}

	attributes := map[string]attr.Value{}
	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}

	return NewNodePoolTemplateMetadataValueMust(NodePoolTemplateMetadataValue{}.AttributeTypes(ctx), attributes), nil
}

func (t NodePoolTemplateMetadataType) ValueType(ctx context.Context) attr.Value {
	return NodePoolTemplateMetadataValue{}
}

var _ basetypes.ObjectValuable = NodePoolTemplateMetadataValue{}

type NodePoolTemplateMetadataValue struct {
	Annotations ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue]  `tfsdk:"annotations" json:"annotations"`
	Finalizers  ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"finalizers" json:"finalizers"`
	Labels      ovhtypes.TfMapNestedValue[ovhtypes.TfStringValue]  `tfsdk:"labels" json:"labels"`
	state       attr.ValueState
}

func (v *NodePoolTemplateMetadataValue) UnmarshalJSON(data []byte) error {
	type JsonNodePoolTemplateMetadataValue NodePoolTemplateMetadataValue

	var tmp JsonNodePoolTemplateMetadataValue
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Annotations = tmp.Annotations
	v.Finalizers = tmp.Finalizers
	v.Labels = tmp.Labels
	v.state = attr.ValueStateKnown

	return nil
}

func (v NodePoolTemplateMetadataValue) Attributes() map[string]attr.Value {
	return map[string]attr.Value{
		"annotations": v.Annotations,
		"finalizers":  v.Finalizers,
		"labels":      v.Labels,
	}
}

func (v NodePoolTemplateMetadataValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)
	attrTypes["annotations"] = basetypes.MapType{ElemType: ovhtypes.TfStringType{}}.TerraformType(ctx)
	attrTypes["finalizers"] = basetypes.ListType{ElemType: ovhtypes.TfStringType{}}.TerraformType(ctx)
	attrTypes["labels"] = basetypes.MapType{ElemType: ovhtypes.TfStringType{}}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err := v.Annotations.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["annotations"] = val

		val, err = v.Finalizers.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["finalizers"] = val

		val, err = v.Labels.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["labels"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v NodePoolTemplateMetadataValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v NodePoolTemplateMetadataValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v NodePoolTemplateMetadataValue) String() string {
	return "NodePoolTemplateMetadataValue"
}

func (v NodePoolTemplateMetadataValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		map[string]attr.Type{
			"annotations": ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
			"finalizers":  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
			"labels":      ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
		},
		map[string]attr.Value{
			"annotations": v.Annotations,
			"finalizers":  v.Finalizers,
			"labels":      v.Labels,
		})
}

func (v NodePoolTemplateMetadataValue) Equal(o attr.Value) bool {
	other, ok := o.(NodePoolTemplateMetadataValue)
	if !ok {
		return false
	}
	if v.state != other.state {
		return false
	}
	if v.state != attr.ValueStateKnown {
		return true
	}
	if !v.Annotations.Equal(other.Annotations) {
		return false
	}
	if !v.Finalizers.Equal(other.Finalizers) {
		return false
	}
	if !v.Labels.Equal(other.Labels) {
		return false
	}
	return true
}

func (v NodePoolTemplateMetadataValue) Type(ctx context.Context) attr.Type {
	return NodePoolTemplateMetadataType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v NodePoolTemplateMetadataValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"annotations": ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
		"finalizers":  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
		"labels":      ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
	}
}

// --- NodePoolTemplateSpecType / NodePoolTemplateSpecValue ---

var _ basetypes.ObjectTypable = NodePoolTemplateSpecType{}

type NodePoolTemplateSpecType struct {
	basetypes.ObjectType
}

func (t NodePoolTemplateSpecType) Equal(o attr.Type) bool {
	other, ok := o.(NodePoolTemplateSpecType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t NodePoolTemplateSpecType) String() string {
	return "NodePoolTemplateSpecType"
}

func (t NodePoolTemplateSpecType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := in.Attributes()

	unschedulableAttribute, ok := attributes["unschedulable"]
	if !ok {
		diags.AddError("Attribute Missing", `unschedulable is missing from object`)
		return nil, diags
	}
	unschedulableVal, ok := unschedulableAttribute.(ovhtypes.TfBoolValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`unschedulable expected to be ovhtypes.TfBoolValue, was: %T`, unschedulableAttribute))
	}

	taintsAttribute, ok := attributes["taints"]
	if !ok {
		diags.AddError("Attribute Missing", `taints is missing from object`)
		return nil, diags
	}
	taintsVal, ok := taintsAttribute.(ovhtypes.TfListNestedValue[NodePoolTaintValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`taints expected to be ovhtypes.TfListNestedValue[NodePoolTaintValue], was: %T`, taintsAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return NodePoolTemplateSpecValue{
		Unschedulable: unschedulableVal,
		Taints:        taintsVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateSpecValueNull() NodePoolTemplateSpecValue {
	return NodePoolTemplateSpecValue{
		state: attr.ValueStateNull,
	}
}

func NewNodePoolTemplateSpecValueUnknown() NodePoolTemplateSpecValue {
	return NodePoolTemplateSpecValue{
		state: attr.ValueStateUnknown,
	}
}

func NewNodePoolTemplateSpecValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (NodePoolTemplateSpecValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]
		if !ok {
			diags.AddError(
				"Missing NodePoolTemplateSpecValue Attribute Value",
				"While creating a NodePoolTemplateSpecValue value, a missing attribute value was detected. "+
					"A NodePoolTemplateSpecValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateSpecValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)
			continue
		}
		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid NodePoolTemplateSpecValue Attribute Type",
				"While creating a NodePoolTemplateSpecValue value, an invalid attribute value was detected. "+
					"A NodePoolTemplateSpecValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTemplateSpecValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("NodePoolTemplateSpecValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]
		if !ok {
			diags.AddError(
				"Extra NodePoolTemplateSpecValue Attribute Value",
				"While creating a NodePoolTemplateSpecValue value, an extra attribute value was detected. "+
					"A NodePoolTemplateSpecValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra NodePoolTemplateSpecValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewNodePoolTemplateSpecValueUnknown(), diags
	}

	unschedulableAttribute, ok := attributes["unschedulable"]
	if !ok {
		diags.AddError("Attribute Missing", `unschedulable is missing from object`)
		return NewNodePoolTemplateSpecValueUnknown(), diags
	}
	unschedulableVal, ok := unschedulableAttribute.(ovhtypes.TfBoolValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`unschedulable expected to be ovhtypes.TfBoolValue, was: %T`, unschedulableAttribute))
	}

	taintsAttribute, ok := attributes["taints"]
	if !ok {
		diags.AddError("Attribute Missing", `taints is missing from object`)
		return NewNodePoolTemplateSpecValueUnknown(), diags
	}
	taintsVal, ok := taintsAttribute.(ovhtypes.TfListNestedValue[NodePoolTaintValue])
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`taints expected to be ovhtypes.TfListNestedValue[NodePoolTaintValue], was: %T`, taintsAttribute))
	}

	if diags.HasError() {
		return NewNodePoolTemplateSpecValueUnknown(), diags
	}

	return NodePoolTemplateSpecValue{
		Unschedulable: unschedulableVal,
		Taints:        taintsVal,
		state:         attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTemplateSpecValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) NodePoolTemplateSpecValue {
	object, diags := NewNodePoolTemplateSpecValue(attributeTypes, attributes)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}
		panic("NewNodePoolTemplateSpecValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return object
}

func (t NodePoolTemplateSpecType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewNodePoolTemplateSpecValueNull(), nil
	}
	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		return NewNodePoolTemplateSpecValueUnknown(), nil
	}
	if in.IsNull() {
		return NewNodePoolTemplateSpecValueNull(), nil
	}

	attributes := map[string]attr.Value{}
	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}

	return NewNodePoolTemplateSpecValueMust(NodePoolTemplateSpecValue{}.AttributeTypes(ctx), attributes), nil
}

func (t NodePoolTemplateSpecType) ValueType(ctx context.Context) attr.Value {
	return NodePoolTemplateSpecValue{}
}

var _ basetypes.ObjectValuable = NodePoolTemplateSpecValue{}

type NodePoolTemplateSpecValue struct {
	Unschedulable ovhtypes.TfBoolValue                           `tfsdk:"unschedulable" json:"unschedulable"`
	Taints        ovhtypes.TfListNestedValue[NodePoolTaintValue] `tfsdk:"taints" json:"taints"`
	state         attr.ValueState
}

func (v *NodePoolTemplateSpecValue) UnmarshalJSON(data []byte) error {
	type JsonNodePoolTemplateSpecValue NodePoolTemplateSpecValue

	var tmp JsonNodePoolTemplateSpecValue
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Unschedulable = tmp.Unschedulable
	v.Taints = tmp.Taints
	v.state = attr.ValueStateKnown

	return nil
}

func (v NodePoolTemplateSpecValue) Attributes() map[string]attr.Value {
	return map[string]attr.Value{
		"unschedulable": v.Unschedulable,
		"taints":        v.Taints,
	}
}

func (v NodePoolTemplateSpecValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)
	attrTypes["unschedulable"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["taints"] = basetypes.ListType{
		ElemType: NodePoolTaintValue{}.Type(ctx),
	}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err := v.Unschedulable.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["unschedulable"] = val

		val, err = v.Taints.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["taints"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v NodePoolTemplateSpecValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v NodePoolTemplateSpecValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v NodePoolTemplateSpecValue) String() string {
	return "NodePoolTemplateSpecValue"
}

func (v NodePoolTemplateSpecValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		map[string]attr.Type{
			"unschedulable": ovhtypes.TfBoolType{},
			"taints":        ovhtypes.NewTfListNestedType[NodePoolTaintValue](ctx),
		},
		map[string]attr.Value{
			"unschedulable": v.Unschedulable,
			"taints":        v.Taints,
		})
}

func (v NodePoolTemplateSpecValue) Equal(o attr.Value) bool {
	other, ok := o.(NodePoolTemplateSpecValue)
	if !ok {
		return false
	}
	if v.state != other.state {
		return false
	}
	if v.state != attr.ValueStateKnown {
		return true
	}
	if !v.Unschedulable.Equal(other.Unschedulable) {
		return false
	}
	if !v.Taints.Equal(other.Taints) {
		return false
	}
	return true
}

func (v NodePoolTemplateSpecValue) Type(ctx context.Context) attr.Type {
	return NodePoolTemplateSpecType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v NodePoolTemplateSpecValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"unschedulable": ovhtypes.TfBoolType{},
		"taints":        ovhtypes.NewTfListNestedType[NodePoolTaintValue](ctx),
	}
}

// --- NodePoolTaintType / NodePoolTaintValue ---

var _ basetypes.ObjectTypable = NodePoolTaintType{}

type NodePoolTaintType struct {
	basetypes.ObjectType
}

func (t NodePoolTaintType) Equal(o attr.Type) bool {
	other, ok := o.(NodePoolTaintType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t NodePoolTaintType) String() string {
	return "NodePoolTaintType"
}

func (t NodePoolTaintType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := in.Attributes()

	effectAttribute, ok := attributes["effect"]
	if !ok {
		diags.AddError("Attribute Missing", `effect is missing from object`)
		return nil, diags
	}
	effectVal, ok := effectAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`effect expected to be ovhtypes.TfStringValue, was: %T`, effectAttribute))
	}

	keyAttribute, ok := attributes["key"]
	if !ok {
		diags.AddError("Attribute Missing", `key is missing from object`)
		return nil, diags
	}
	keyVal, ok := keyAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`key expected to be ovhtypes.TfStringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]
	if !ok {
		diags.AddError("Attribute Missing", `value is missing from object`)
		return nil, diags
	}
	valueVal, ok := valueAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`value expected to be ovhtypes.TfStringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return NodePoolTaintValue{
		Effect: effectVal,
		Key:    keyVal,
		Value:  valueVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTaintValueNull() NodePoolTaintValue {
	return NodePoolTaintValue{
		state: attr.ValueStateNull,
	}
}

func NewNodePoolTaintValueUnknown() NodePoolTaintValue {
	return NodePoolTaintValue{
		state: attr.ValueStateUnknown,
	}
}

func NewNodePoolTaintValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (NodePoolTaintValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]
		if !ok {
			diags.AddError(
				"Missing NodePoolTaintValue Attribute Value",
				"While creating a NodePoolTaintValue value, a missing attribute value was detected. "+
					"A NodePoolTaintValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTaintValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)
			continue
		}
		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid NodePoolTaintValue Attribute Type",
				"While creating a NodePoolTaintValue value, an invalid attribute value was detected. "+
					"A NodePoolTaintValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("NodePoolTaintValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("NodePoolTaintValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]
		if !ok {
			diags.AddError(
				"Extra NodePoolTaintValue Attribute Value",
				"While creating a NodePoolTaintValue value, an extra attribute value was detected. "+
					"A NodePoolTaintValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra NodePoolTaintValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewNodePoolTaintValueUnknown(), diags
	}

	effectAttribute, ok := attributes["effect"]
	if !ok {
		diags.AddError("Attribute Missing", `effect is missing from object`)
		return NewNodePoolTaintValueUnknown(), diags
	}
	effectVal, ok := effectAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`effect expected to be ovhtypes.TfStringValue, was: %T`, effectAttribute))
	}

	keyAttribute, ok := attributes["key"]
	if !ok {
		diags.AddError("Attribute Missing", `key is missing from object`)
		return NewNodePoolTaintValueUnknown(), diags
	}
	keyVal, ok := keyAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`key expected to be ovhtypes.TfStringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]
	if !ok {
		diags.AddError("Attribute Missing", `value is missing from object`)
		return NewNodePoolTaintValueUnknown(), diags
	}
	valueVal, ok := valueAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`value expected to be ovhtypes.TfStringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return NewNodePoolTaintValueUnknown(), diags
	}

	return NodePoolTaintValue{
		Effect: effectVal,
		Key:    keyVal,
		Value:  valueVal,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewNodePoolTaintValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) NodePoolTaintValue {
	object, diags := NewNodePoolTaintValue(attributeTypes, attributes)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}
		panic("NewNodePoolTaintValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return object
}

func (t NodePoolTaintType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewNodePoolTaintValueNull(), nil
	}
	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		return NewNodePoolTaintValueUnknown(), nil
	}
	if in.IsNull() {
		return NewNodePoolTaintValueNull(), nil
	}

	attributes := map[string]attr.Value{}
	val := map[string]tftypes.Value{}
	err := in.As(&val)
	if err != nil {
		return nil, err
	}
	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)
		if err != nil {
			return nil, err
		}
		attributes[k] = a
	}

	return NewNodePoolTaintValueMust(NodePoolTaintValue{}.AttributeTypes(ctx), attributes), nil
}

func (t NodePoolTaintType) ValueType(ctx context.Context) attr.Value {
	return NodePoolTaintValue{}
}

var _ basetypes.ObjectValuable = NodePoolTaintValue{}

type NodePoolTaintValue struct {
	Effect ovhtypes.TfStringValue `tfsdk:"effect" json:"effect"`
	Key    ovhtypes.TfStringValue `tfsdk:"key" json:"key"`
	Value  ovhtypes.TfStringValue `tfsdk:"value" json:"value"`
	state  attr.ValueState
}

func (v *NodePoolTaintValue) UnmarshalJSON(data []byte) error {
	type JsonNodePoolTaintValue NodePoolTaintValue

	var tmp JsonNodePoolTaintValue
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.Effect = tmp.Effect
	v.Key = tmp.Key
	v.Value = tmp.Value
	v.state = attr.ValueStateKnown

	return nil
}

func (v NodePoolTaintValue) Attributes() map[string]attr.Value {
	return map[string]attr.Value{
		"effect": v.Effect,
		"key":    v.Key,
		"value":  v.Value,
	}
}

func (v NodePoolTaintValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)
	attrTypes["effect"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["value"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 3)

		val, err := v.Effect.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["effect"] = val

		val, err = v.Key.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["key"] = val

		val, err = v.Value.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["value"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v NodePoolTaintValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v NodePoolTaintValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v NodePoolTaintValue) String() string {
	return "NodePoolTaintValue"
}

func (v NodePoolTaintValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		map[string]attr.Type{
			"effect": ovhtypes.TfStringType{},
			"key":    ovhtypes.TfStringType{},
			"value":  ovhtypes.TfStringType{},
		},
		map[string]attr.Value{
			"effect": v.Effect,
			"key":    v.Key,
			"value":  v.Value,
		})
}

func (v NodePoolTaintValue) Equal(o attr.Value) bool {
	other, ok := o.(NodePoolTaintValue)
	if !ok {
		return false
	}
	if v.state != other.state {
		return false
	}
	if v.state != attr.ValueStateKnown {
		return true
	}
	if !v.Effect.Equal(other.Effect) {
		return false
	}
	if !v.Key.Equal(other.Key) {
		return false
	}
	if !v.Value.Equal(other.Value) {
		return false
	}
	return true
}

func (v NodePoolTaintValue) Type(ctx context.Context) attr.Type {
	return NodePoolTaintType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v NodePoolTaintValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"effect": ovhtypes.TfStringType{},
		"key":    ovhtypes.TfStringType{},
		"value":  ovhtypes.TfStringType{},
	}
}
