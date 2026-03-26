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

func CloudProjectKubeNodesDataSourceSchema(ctx context.Context) schema.Schema {
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

			// Computed
			"nodes": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[CloudProjectKubeNodeValue](ctx),
				Computed:    true,
				Description: "Nodes composing the cluster",
			},
		},
	}
}

// --- Model ---

type CloudProjectKubeNodesDataSourceModel struct {
	ServiceName ovhtypes.TfStringValue                                `tfsdk:"service_name"`
	KubeId      ovhtypes.TfStringValue                                `tfsdk:"kube_id"`
	Nodes       ovhtypes.TfListNestedValue[CloudProjectKubeNodeValue] `tfsdk:"nodes"`
}

// --- CloudProjectKubeNodeValue / CloudProjectKubeNodeType ---

var _ basetypes.ObjectTypable = CloudProjectKubeNodeType{}

type CloudProjectKubeNodeType struct {
	basetypes.ObjectType
}

func (t CloudProjectKubeNodeType) Equal(o attr.Type) bool {
	other, ok := o.(CloudProjectKubeNodeType)
	if !ok {
		return false
	}
	return t.ObjectType.Equal(other.ObjectType)
}

func (t CloudProjectKubeNodeType) String() string {
	return "CloudProjectKubeNodeType"
}

func (t CloudProjectKubeNodeType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	attributes := in.Attributes()

	createdAtAttribute, ok := attributes["created_at"]
	if !ok {
		diags.AddError("Attribute Missing", `created_at is missing from object`)
		return nil, diags
	}
	createdAtVal, ok := createdAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`created_at expected to be ovhtypes.TfStringValue, was: %T`, createdAtAttribute))
	}

	updatedAtAttribute, ok := attributes["updated_at"]
	if !ok {
		diags.AddError("Attribute Missing", `updated_at is missing from object`)
		return nil, diags
	}
	updatedAtVal, ok := updatedAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`updated_at expected to be ovhtypes.TfStringValue, was: %T`, updatedAtAttribute))
	}

	deployedAtAttribute, ok := attributes["deployed_at"]
	if !ok {
		diags.AddError("Attribute Missing", `deployed_at is missing from object`)
		return nil, diags
	}
	deployedAtVal, ok := deployedAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`deployed_at expected to be ovhtypes.TfStringValue, was: %T`, deployedAtAttribute))
	}

	flavorAttribute, ok := attributes["flavor"]
	if !ok {
		diags.AddError("Attribute Missing", `flavor is missing from object`)
		return nil, diags
	}
	flavorVal, ok := flavorAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`flavor expected to be ovhtypes.TfStringValue, was: %T`, flavorAttribute))
	}

	idAttribute, ok := attributes["id"]
	if !ok {
		diags.AddError("Attribute Missing", `id is missing from object`)
		return nil, diags
	}
	idVal, ok := idAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`id expected to be ovhtypes.TfStringValue, was: %T`, idAttribute))
	}

	instanceIdAttribute, ok := attributes["instance_id"]
	if !ok {
		diags.AddError("Attribute Missing", `instance_id is missing from object`)
		return nil, diags
	}
	instanceIdVal, ok := instanceIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`instance_id expected to be ovhtypes.TfStringValue, was: %T`, instanceIdAttribute))
	}

	isUpToDateAttribute, ok := attributes["is_up_to_date"]
	if !ok {
		diags.AddError("Attribute Missing", `is_up_to_date is missing from object`)
		return nil, diags
	}
	isUpToDateVal, ok := isUpToDateAttribute.(ovhtypes.TfBoolValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`is_up_to_date expected to be ovhtypes.TfBoolValue, was: %T`, isUpToDateAttribute))
	}

	nameAttribute, ok := attributes["name"]
	if !ok {
		diags.AddError("Attribute Missing", `name is missing from object`)
		return nil, diags
	}
	nameVal, ok := nameAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`name expected to be ovhtypes.TfStringValue, was: %T`, nameAttribute))
	}

	nodePoolIdAttribute, ok := attributes["node_pool_id"]
	if !ok {
		diags.AddError("Attribute Missing", `node_pool_id is missing from object`)
		return nil, diags
	}
	nodePoolIdVal, ok := nodePoolIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`node_pool_id expected to be ovhtypes.TfStringValue, was: %T`, nodePoolIdAttribute))
	}

	projectIdAttribute, ok := attributes["project_id"]
	if !ok {
		diags.AddError("Attribute Missing", `project_id is missing from object`)
		return nil, diags
	}
	projectIdVal, ok := projectIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`project_id expected to be ovhtypes.TfStringValue, was: %T`, projectIdAttribute))
	}

	statusAttribute, ok := attributes["status"]
	if !ok {
		diags.AddError("Attribute Missing", `status is missing from object`)
		return nil, diags
	}
	statusVal, ok := statusAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`status expected to be ovhtypes.TfStringValue, was: %T`, statusAttribute))
	}

	versionAttribute, ok := attributes["version"]
	if !ok {
		diags.AddError("Attribute Missing", `version is missing from object`)
		return nil, diags
	}
	versionVal, ok := versionAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`version expected to be ovhtypes.TfStringValue, was: %T`, versionAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return CloudProjectKubeNodeValue{
		CreatedAt:  createdAtVal,
		UpdatedAt:  updatedAtVal,
		DeployedAt: deployedAtVal,
		Flavor:     flavorVal,
		Id:         idVal,
		InstanceId: instanceIdVal,
		IsUpToDate: isUpToDateVal,
		Name:       nameVal,
		NodePoolId: nodePoolIdVal,
		ProjectId:  projectIdVal,
		Status:     statusVal,
		Version:    versionVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewCloudProjectKubeNodeValueNull() CloudProjectKubeNodeValue {
	return CloudProjectKubeNodeValue{
		state: attr.ValueStateNull,
	}
}

func NewCloudProjectKubeNodeValueUnknown() CloudProjectKubeNodeValue {
	return CloudProjectKubeNodeValue{
		state: attr.ValueStateUnknown,
	}
}

func NewCloudProjectKubeNodeValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (CloudProjectKubeNodeValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]
		if !ok {
			diags.AddError(
				"Missing CloudProjectKubeNodeValue Attribute Value",
				"While creating a CloudProjectKubeNodeValue value, a missing attribute value was detected. "+
					"A CloudProjectKubeNodeValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("CloudProjectKubeNodeValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)
			continue
		}
		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid CloudProjectKubeNodeValue Attribute Type",
				"While creating a CloudProjectKubeNodeValue value, an invalid attribute value was detected. "+
					"A CloudProjectKubeNodeValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("CloudProjectKubeNodeValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("CloudProjectKubeNodeValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]
		if !ok {
			diags.AddError(
				"Extra CloudProjectKubeNodeValue Attribute Value",
				"While creating a CloudProjectKubeNodeValue value, an extra attribute value was detected. "+
					"A CloudProjectKubeNodeValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra CloudProjectKubeNodeValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}

	createdAtAttribute, ok := attributes["created_at"]
	if !ok {
		diags.AddError("Attribute Missing", `created_at is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	createdAtVal, ok := createdAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`created_at expected to be ovhtypes.TfStringValue, was: %T`, createdAtAttribute))
	}

	updatedAtAttribute, ok := attributes["updated_at"]
	if !ok {
		diags.AddError("Attribute Missing", `updated_at is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	updatedAtVal, ok := updatedAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`updated_at expected to be ovhtypes.TfStringValue, was: %T`, updatedAtAttribute))
	}

	deployedAtAttribute, ok := attributes["deployed_at"]
	if !ok {
		diags.AddError("Attribute Missing", `deployed_at is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	deployedAtVal, ok := deployedAtAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`deployed_at expected to be ovhtypes.TfStringValue, was: %T`, deployedAtAttribute))
	}

	flavorAttribute, ok := attributes["flavor"]
	if !ok {
		diags.AddError("Attribute Missing", `flavor is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	flavorVal, ok := flavorAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`flavor expected to be ovhtypes.TfStringValue, was: %T`, flavorAttribute))
	}

	idAttribute, ok := attributes["id"]
	if !ok {
		diags.AddError("Attribute Missing", `id is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	idVal, ok := idAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`id expected to be ovhtypes.TfStringValue, was: %T`, idAttribute))
	}

	instanceIdAttribute, ok := attributes["instance_id"]
	if !ok {
		diags.AddError("Attribute Missing", `instance_id is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	instanceIdVal, ok := instanceIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`instance_id expected to be ovhtypes.TfStringValue, was: %T`, instanceIdAttribute))
	}

	isUpToDateAttribute, ok := attributes["is_up_to_date"]
	if !ok {
		diags.AddError("Attribute Missing", `is_up_to_date is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	isUpToDateVal, ok := isUpToDateAttribute.(ovhtypes.TfBoolValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`is_up_to_date expected to be ovhtypes.TfBoolValue, was: %T`, isUpToDateAttribute))
	}

	nameAttribute, ok := attributes["name"]
	if !ok {
		diags.AddError("Attribute Missing", `name is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	nameVal, ok := nameAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`name expected to be ovhtypes.TfStringValue, was: %T`, nameAttribute))
	}

	nodePoolIdAttribute, ok := attributes["node_pool_id"]
	if !ok {
		diags.AddError("Attribute Missing", `node_pool_id is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	nodePoolIdVal, ok := nodePoolIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`node_pool_id expected to be ovhtypes.TfStringValue, was: %T`, nodePoolIdAttribute))
	}

	projectIdAttribute, ok := attributes["project_id"]
	if !ok {
		diags.AddError("Attribute Missing", `project_id is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	projectIdVal, ok := projectIdAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`project_id expected to be ovhtypes.TfStringValue, was: %T`, projectIdAttribute))
	}

	statusAttribute, ok := attributes["status"]
	if !ok {
		diags.AddError("Attribute Missing", `status is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	statusVal, ok := statusAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`status expected to be ovhtypes.TfStringValue, was: %T`, statusAttribute))
	}

	versionAttribute, ok := attributes["version"]
	if !ok {
		diags.AddError("Attribute Missing", `version is missing from object`)
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}
	versionVal, ok := versionAttribute.(ovhtypes.TfStringValue)
	if !ok {
		diags.AddError("Attribute Wrong Type", fmt.Sprintf(`version expected to be ovhtypes.TfStringValue, was: %T`, versionAttribute))
	}

	if diags.HasError() {
		return NewCloudProjectKubeNodeValueUnknown(), diags
	}

	return CloudProjectKubeNodeValue{
		CreatedAt:  createdAtVal,
		UpdatedAt:  updatedAtVal,
		DeployedAt: deployedAtVal,
		Flavor:     flavorVal,
		Id:         idVal,
		InstanceId: instanceIdVal,
		IsUpToDate: isUpToDateVal,
		Name:       nameVal,
		NodePoolId: nodePoolIdVal,
		ProjectId:  projectIdVal,
		Status:     statusVal,
		Version:    versionVal,
		state:      attr.ValueStateKnown,
	}, diags
}

func NewCloudProjectKubeNodeValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) CloudProjectKubeNodeValue {
	object, diags := NewCloudProjectKubeNodeValue(attributeTypes, attributes)
	if diags.HasError() {
		diagsStrings := make([]string, 0, len(diags))
		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}
		panic("NewCloudProjectKubeNodeValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}
	return object
}

func (t CloudProjectKubeNodeType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewCloudProjectKubeNodeValueNull(), nil
	}
	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}
	if !in.IsKnown() {
		return NewCloudProjectKubeNodeValueUnknown(), nil
	}
	if in.IsNull() {
		return NewCloudProjectKubeNodeValueNull(), nil
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

	return NewCloudProjectKubeNodeValueMust(CloudProjectKubeNodeValue{}.AttributeTypes(ctx), attributes), nil
}

func (t CloudProjectKubeNodeType) ValueType(ctx context.Context) attr.Value {
	return CloudProjectKubeNodeValue{}
}

var _ basetypes.ObjectValuable = CloudProjectKubeNodeValue{}

type CloudProjectKubeNodeValue struct {
	CreatedAt  ovhtypes.TfStringValue `tfsdk:"created_at" json:"createdAt"`
	UpdatedAt  ovhtypes.TfStringValue `tfsdk:"updated_at" json:"updatedAt"`
	DeployedAt ovhtypes.TfStringValue `tfsdk:"deployed_at" json:"deployedAt"`
	Flavor     ovhtypes.TfStringValue `tfsdk:"flavor" json:"flavor"`
	Id         ovhtypes.TfStringValue `tfsdk:"id" json:"id"`
	InstanceId ovhtypes.TfStringValue `tfsdk:"instance_id" json:"instanceId"`
	IsUpToDate ovhtypes.TfBoolValue   `tfsdk:"is_up_to_date" json:"isUpToDate"`
	Name       ovhtypes.TfStringValue `tfsdk:"name" json:"name"`
	NodePoolId ovhtypes.TfStringValue `tfsdk:"node_pool_id" json:"nodePoolId"`
	ProjectId  ovhtypes.TfStringValue `tfsdk:"project_id" json:"projectId"`
	Status     ovhtypes.TfStringValue `tfsdk:"status" json:"status"`
	Version    ovhtypes.TfStringValue `tfsdk:"version" json:"version"`
	state      attr.ValueState
}

func (v *CloudProjectKubeNodeValue) UnmarshalJSON(data []byte) error {
	type JsonCloudProjectKubeNodeValue CloudProjectKubeNodeValue

	var tmp JsonCloudProjectKubeNodeValue
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	v.CreatedAt = tmp.CreatedAt
	v.UpdatedAt = tmp.UpdatedAt
	v.DeployedAt = tmp.DeployedAt
	v.Flavor = tmp.Flavor
	v.Id = tmp.Id
	v.InstanceId = tmp.InstanceId
	v.IsUpToDate = tmp.IsUpToDate
	v.Name = tmp.Name
	v.NodePoolId = tmp.NodePoolId
	v.ProjectId = tmp.ProjectId
	v.Status = tmp.Status
	v.Version = tmp.Version
	v.state = attr.ValueStateKnown

	return nil
}

func (v CloudProjectKubeNodeValue) Attributes() map[string]attr.Value {
	return map[string]attr.Value{
		"created_at":    v.CreatedAt,
		"updated_at":    v.UpdatedAt,
		"deployed_at":   v.DeployedAt,
		"flavor":        v.Flavor,
		"id":            v.Id,
		"instance_id":   v.InstanceId,
		"is_up_to_date": v.IsUpToDate,
		"name":          v.Name,
		"node_pool_id":  v.NodePoolId,
		"project_id":    v.ProjectId,
		"status":        v.Status,
		"version":       v.Version,
	}
}

func (v CloudProjectKubeNodeValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 12)
	attrTypes["created_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["updated_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["deployed_at"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["flavor"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["instance_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["is_up_to_date"] = basetypes.BoolType{}.TerraformType(ctx)
	attrTypes["name"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["node_pool_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["project_id"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["status"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["version"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 12)

		val, err := v.CreatedAt.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["created_at"] = val

		val, err = v.UpdatedAt.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["updated_at"] = val

		val, err = v.DeployedAt.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["deployed_at"] = val

		val, err = v.Flavor.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["flavor"] = val

		val, err = v.Id.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["id"] = val

		val, err = v.InstanceId.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["instance_id"] = val

		val, err = v.IsUpToDate.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["is_up_to_date"] = val

		val, err = v.Name.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["name"] = val

		val, err = v.NodePoolId.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["node_pool_id"] = val

		val, err = v.ProjectId.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["project_id"] = val

		val, err = v.Status.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["status"] = val

		val, err = v.Version.ToTerraformValue(ctx)
		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}
		vals["version"] = val

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

func (v CloudProjectKubeNodeValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v CloudProjectKubeNodeValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v CloudProjectKubeNodeValue) String() string {
	return "CloudProjectKubeNodeValue"
}

func (v CloudProjectKubeNodeValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	return types.ObjectValue(
		map[string]attr.Type{
			"created_at":    ovhtypes.TfStringType{},
			"updated_at":    ovhtypes.TfStringType{},
			"deployed_at":   ovhtypes.TfStringType{},
			"flavor":        ovhtypes.TfStringType{},
			"id":            ovhtypes.TfStringType{},
			"instance_id":   ovhtypes.TfStringType{},
			"is_up_to_date": ovhtypes.TfBoolType{},
			"name":          ovhtypes.TfStringType{},
			"node_pool_id":  ovhtypes.TfStringType{},
			"project_id":    ovhtypes.TfStringType{},
			"status":        ovhtypes.TfStringType{},
			"version":       ovhtypes.TfStringType{},
		},
		map[string]attr.Value{
			"created_at":    v.CreatedAt,
			"updated_at":    v.UpdatedAt,
			"deployed_at":   v.DeployedAt,
			"flavor":        v.Flavor,
			"id":            v.Id,
			"instance_id":   v.InstanceId,
			"is_up_to_date": v.IsUpToDate,
			"name":          v.Name,
			"node_pool_id":  v.NodePoolId,
			"project_id":    v.ProjectId,
			"status":        v.Status,
			"version":       v.Version,
		})
}

func (v CloudProjectKubeNodeValue) Equal(o attr.Value) bool {
	other, ok := o.(CloudProjectKubeNodeValue)
	if !ok {
		return false
	}
	if v.state != other.state {
		return false
	}
	if v.state != attr.ValueStateKnown {
		return true
	}
	if !v.CreatedAt.Equal(other.CreatedAt) {
		return false
	}
	if !v.UpdatedAt.Equal(other.UpdatedAt) {
		return false
	}
	if !v.DeployedAt.Equal(other.DeployedAt) {
		return false
	}
	if !v.Flavor.Equal(other.Flavor) {
		return false
	}
	if !v.Id.Equal(other.Id) {
		return false
	}
	if !v.InstanceId.Equal(other.InstanceId) {
		return false
	}
	if !v.IsUpToDate.Equal(other.IsUpToDate) {
		return false
	}
	if !v.Name.Equal(other.Name) {
		return false
	}
	if !v.NodePoolId.Equal(other.NodePoolId) {
		return false
	}
	if !v.ProjectId.Equal(other.ProjectId) {
		return false
	}
	if !v.Status.Equal(other.Status) {
		return false
	}
	if !v.Version.Equal(other.Version) {
		return false
	}
	return true
}

func (v CloudProjectKubeNodeValue) Type(ctx context.Context) attr.Type {
	return CloudProjectKubeNodeType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v CloudProjectKubeNodeValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"created_at":    ovhtypes.TfStringType{},
		"updated_at":    ovhtypes.TfStringType{},
		"deployed_at":   ovhtypes.TfStringType{},
		"flavor":        ovhtypes.TfStringType{},
		"id":            ovhtypes.TfStringType{},
		"instance_id":   ovhtypes.TfStringType{},
		"is_up_to_date": ovhtypes.TfBoolType{},
		"name":          ovhtypes.TfStringType{},
		"node_pool_id":  ovhtypes.TfStringType{},
		"project_id":    ovhtypes.TfStringType{},
		"status":        ovhtypes.TfStringType{},
		"version":       ovhtypes.TfStringType{},
	}
}
