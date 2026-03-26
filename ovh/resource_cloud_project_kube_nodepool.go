package ovh

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/helpers"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/ovhwrap"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Resource struct ---

type cloudProjectKubeNodePoolResource struct {
	config *Config
}

func NewCloudProjectKubeNodePoolResource() resource.Resource {
	return &cloudProjectKubeNodePoolResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &cloudProjectKubeNodePoolResource{}
	_ resource.ResourceWithConfigure   = &cloudProjectKubeNodePoolResource{}
	_ resource.ResourceWithImportState = &cloudProjectKubeNodePoolResource{}
)

// --- Model ---

type cloudProjectKubeNodePoolResourceModel struct {
	ID                                       ovhtypes.TfStringValue                             `tfsdk:"id"`
	ServiceName                              ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId                                   ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	Name                                     ovhtypes.TfStringValue                             `tfsdk:"name" json:"name"`
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

// --- Metadata / Configure / Schema ---

func (r *cloudProjectKubeNodePoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_nodepool"
}

func (r *cloudProjectKubeNodePoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.config = config
}

func (r *cloudProjectKubeNodePoolResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"service_name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    os.Getenv("OVH_CLOUD_PROJECT_SERVICE") == "",
				Optional:    os.Getenv("OVH_CLOUD_PROJECT_SERVICE") != "",
				Description: "Service name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kube_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Kube ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "NodePool resource name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flavor_name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Flavor name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"autoscale": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Optional:    true,
				Computed:    true,
				Description: "Enable auto-scaling for the pool",
			},
			"anti_affinity": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Optional:    true,
				Computed:    true,
				Description: "Enable anti affinity groups for nodes in the pool",
				PlanModifiers: []planmodifier.Bool{
					// boolplanmodifier.RequiresReplace(),
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							// If not specified in config, allow server to set without replacement
							if req.ConfigValue.IsNull() {
								return
							}
							// If plan is unknown, defer replacement decision
							if req.PlanValue.IsUnknown() {
								return
							}
							// If specified, known, and changed from state, require replacement
							if !req.PlanValue.Equal(req.StateValue) {
								resp.RequiresReplace = true
							}
						},
						"Only replace when user specifies anti_affinity and it differs from state",
						"Only replace when user specifies anti_affinity and it differs from state",
					),
				},
			},
			"desired_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Optional:    true,
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"max_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Optional:    true,
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"min_nodes": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Optional:    true,
				Computed:    true,
				Description: "Number of nodes you desire in the pool",
			},
			"monthly_billed": schema.BoolAttribute{
				CustomType:  ovhtypes.TfBoolType{},
				Optional:    true,
				Computed:    true,
				Description: "Enable monthly billing on all nodes in the pool",
				PlanModifiers: []planmodifier.Bool{
					// boolplanmodifier.RequiresReplace(),
					boolplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.BoolRequest, resp *boolplanmodifier.RequiresReplaceIfFuncResponse) {
							// If not specified in config, allow server to set without replacement
							if req.ConfigValue.IsNull() {
								return
							}
							// If plan is unknown, defer replacement decision
							if req.PlanValue.IsUnknown() {
								return
							}
							// If specified, known, and changed from state, require replacement
							if !req.PlanValue.Equal(req.StateValue) {
								resp.RequiresReplace = true
							}
						},
						"Only replace when user specifies monthly_billed and it differs from state",
						"Only replace when user specifies monthly_billed and it differs from state",
					),
				},
			},
			// Computed-only
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
				Optional:    true,
				Computed:    true,
				Description: "scaleDownUnneededTimeSeconds for autoscaling",
			},
			"autoscaling_scale_down_unready_time_seconds": schema.Int64Attribute{
				CustomType:  ovhtypes.TfInt64Type{},
				Optional:    true,
				Computed:    true,
				Description: "scaleDownUnreadyTimeSeconds for autoscaling",
			},
			"autoscaling_scale_down_utilization_threshold": schema.NumberAttribute{
				CustomType:  ovhtypes.TfNumberType{},
				Optional:    true,
				Computed:    true,
				Description: "scaleDownUtilizationThreshold for autoscaling",
			},
			"availability_zones": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Optional:    true,
				Computed:    true,
				Description: "Availability zones",
				PlanModifiers: []planmodifier.List{
					// listplanmodifier.RequiresReplace(),
					listplanmodifier.RequiresReplaceIf(
						func(ctx context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
							// If not specified in config, allow server to set without replacement
							if req.ConfigValue.IsNull() {
								return
							}
							// If plan is unknown, defer replacement decision
							if req.PlanValue.IsUnknown() {
								return
							}
							// If specified, known, and changed from state, require replacement
							if !req.PlanValue.Equal(req.StateValue) {
								resp.RequiresReplace = true
							}
						},
						"Only replace when user specifies availability_zones and it differs from state",
						"Only replace when user specifies availability_zones and it differs from state",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"template": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{ //},
					"metadata": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"finalizers": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Optional:    true,
								Description: "finalizers",
							},
							"labels": schema.MapAttribute{
								CustomType:  ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
								Optional:    true,
								Description: "labels",
							},
							"annotations": schema.MapAttribute{
								CustomType:  ovhtypes.NewTfMapNestedType[ovhtypes.TfStringValue](ctx),
								Optional:    true,
								Description: "annotations",
							},
						},
						CustomType: NodePoolTemplateMetadataType{
							ObjectType: types.ObjectType{
								AttrTypes: NodePoolTemplateMetadataValue{}.AttributeTypes(ctx),
							},
						},
						Description: "metadata",
					},
					"spec": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"unschedulable": schema.BoolAttribute{
								CustomType:  ovhtypes.TfBoolType{},
								Optional:    true,
								Description: "unschedulable",
							},
							"taints": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"effect": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Required:    true,
											Description: "effect",
										},
										"key": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Required:    true,
											Description: "key",
										},
										"value": schema.StringAttribute{
											CustomType:  ovhtypes.TfStringType{},
											Optional:    true,
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
								Optional:    true,
								Description: "taints",
							},
						},
						CustomType: NodePoolTemplateSpecType{
							ObjectType: types.ObjectType{
								AttrTypes: NodePoolTemplateSpecValue{}.AttributeTypes(ctx),
							},
						},
						Description: "spec",
					},
				},
				CustomType: NodePoolTemplateType{
					ObjectType: types.ObjectType{
						AttrTypes: NodePoolTemplateValue{}.AttributeTypes(ctx),
					},
				},
				Description: "Node pool template",
			},
		},
	}
}

// --- CRUD ---

func (r *cloudProjectKubeNodePoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data cloudProjectKubeNodePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	params := nodePoolCreateOptsFromModel(&data)
	res := &CloudProjectKubeNodePoolResponse{}

	log.Printf("[DEBUG] Will create nodepool: %+v", params)
	if err := r.config.OVHClient.PostWithContext(ctx, endpoint, params, res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kube nodepool",
			fmt.Sprintf("error calling POST %s: %s", endpoint, err),
		)
		return
	}

	// This is a fix for a weird bug where the nodepool is not immediately available on API
	log.Printf("[DEBUG] Waiting for nodepool %s to be available", res.Id)
	poolEndpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(res.Id),
	)
	if err := helpers.WaitAvailable(r.config.OVHClient, poolEndpoint, 2*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Failed waiting for nodepool to be available",
			err.Error(),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for nodepool %s to be READY", res.Id)
	if err := waitForCloudProjectKubeNodePoolReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, res.Id, time.Hour); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for nodepool to be READY",
			fmt.Sprintf("timeout while waiting nodepool %s to be READY: %s", res.Id, err),
		)
		return
	}
	log.Printf("[DEBUG] nodepool %s is READY", res.Id)

	// Read back the resource
	if err := r.readNodePool(ctx, serviceName, kubeId, res.Id, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube nodepool after create",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeNodePoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cloudProjectKubeNodePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()
	poolId := data.ID.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(poolId),
	)

	log.Printf("[DEBUG] Will read nodepool %s from cluster %s in project %s", poolId, kubeId, serviceName)

	// Save template state before read so we can restore null if user didn't specify one
	templateWasNull := data.Template.IsNull()

	var res cloudProjectKubeNodePoolResourceModel
	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, &res); err != nil {
		helpers.CheckDeletedWithContext(ctx, resp, err, endpoint)
		return
	}

	// Preserve service_name and kube_id from state (not returned by API)
	res.ServiceName = data.ServiceName
	res.KubeId = data.KubeId
	res.ID = data.ID
	res.FlavorName = res.Flavor

	// Flatten autoscaling
	flattenAutoscaling(&res)

	// Handle template: if user didn't specify one and API returns empty, set null
	handleTemplateOnRead(&res, templateWasNull)

	log.Printf("[DEBUG] Read nodepool: %+v", res)

	resp.Diagnostics.Append(resp.State.Set(ctx, &res)...)
}

func (r *cloudProjectKubeNodePoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data cloudProjectKubeNodePoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get pool ID from state
	var state cloudProjectKubeNodePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()
	poolId := state.ID.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(poolId),
	)

	params := nodePoolUpdateOptsFromModel(&data)

	log.Printf("[DEBUG] Will update nodepool: %+v", params)
	if err := r.config.OVHClient.PutWithContext(ctx, endpoint, params, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update kube nodepool",
			fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for nodepool %s to be READY", poolId)
	if err := waitForCloudProjectKubeNodePoolReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, poolId, time.Hour); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for nodepool to be READY",
			fmt.Sprintf("timeout while waiting nodepool %s to be READY: %s", poolId, err),
		)
		return
	}
	log.Printf("[DEBUG] nodepool %s is READY", poolId)

	// Read back the resource
	if err := r.readNodePool(ctx, serviceName, kubeId, poolId, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube nodepool after update",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeNodePoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cloudProjectKubeNodePoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()
	poolId := data.ID.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(poolId),
	)

	log.Printf("[DEBUG] Will delete nodepool %s from cluster %s in project %s", poolId, kubeId, serviceName)
	if err := r.config.OVHClient.DeleteWithContext(ctx, endpoint, nil); err != nil {
		if errOvh, ok := err.(*ovh.APIError); ok && errOvh.Code == 404 {
			// Already deleted
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete kube nodepool",
			fmt.Sprintf("error calling DELETE %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for nodepool %s to be DELETED", poolId)
	if err := waitForCloudProjectKubeNodePoolDeletedCtx(ctx, r.config.OVHClient, serviceName, kubeId, poolId, time.Hour); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for nodepool to be DELETED",
			fmt.Sprintf("timeout while waiting nodepool %s to be DELETED: %s", poolId, err),
		)
		return
	}
	log.Printf("[DEBUG] nodepool %s is DELETED", poolId)
}

func (r *cloudProjectKubeNodePoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	splits := strings.SplitN(req.ID, "/", 3)
	if len(splits) != 3 {
		resp.Diagnostics.AddError(
			"Given ID is malformed",
			"ID must be formatted as: service_name/kube_id/pool_id",
		)
		return
	}

	serviceName := splits[0]
	kubeId := splits[1]
	poolId := splits[2]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), poolId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), serviceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kube_id"), kubeId)...)
}

// --- Helpers ---

// readNodePool reads a nodepool from the API and populates the model.
func (r *cloudProjectKubeNodePoolResource) readNodePool(ctx context.Context, serviceName, kubeId, poolId string, data *cloudProjectKubeNodePoolResourceModel) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(poolId),
	)

	// Save template state before read
	templateWasNull := data.Template.IsNull()

	var res cloudProjectKubeNodePoolResourceModel
	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, &res); err != nil {
		return fmt.Errorf("error calling GET %s: %w", endpoint, err)
	}

	// Preserve identifiers
	res.ServiceName = data.ServiceName
	res.KubeId = data.KubeId
	res.ID = ovhtypes.NewTfStringValue(poolId)
	res.FlavorName = res.Flavor

	// Flatten autoscaling
	flattenAutoscaling(&res)

	// Handle template
	handleTemplateOnRead(&res, templateWasNull)

	*data = res
	return nil
}

// flattenAutoscaling flattens the nested autoscaling JSON into top-level attributes.
func flattenAutoscaling(data *cloudProjectKubeNodePoolResourceModel) {
	autoscaling := data.Autoscaling
	if autoscaling.ScaleDownUnneededTimeSeconds != nil {
		data.AutoscalingScaleDownUnneededTimeSeconds = ovhtypes.TfInt64Value{
			Int64Value: basetypes.NewInt64Value(*autoscaling.ScaleDownUnneededTimeSeconds),
		}
	}
	if autoscaling.ScaleDownUnreadyTimeSeconds != nil {
		data.AutoscalingScaleDownUnreadyTimeSeconds = ovhtypes.TfInt64Value{
			Int64Value: basetypes.NewInt64Value(*autoscaling.ScaleDownUnreadyTimeSeconds),
		}
	}
	if autoscaling.ScaleDownUtilizationThreshold != nil {
		data.AutoscalingScaleDownUtilizationThreshold = ovhtypes.TfNumberValue{
			NumberValue: basetypes.NewNumberValue(big.NewFloat(*autoscaling.ScaleDownUtilizationThreshold)),
		}
	}
}

// handleTemplateOnRead sets template to null if the user didn't specify one
// and the API returned an empty template.
func handleTemplateOnRead(data *cloudProjectKubeNodePoolResourceModel, templateWasNull bool) {
	if templateWasNull && isEmptyTemplate(data.Template) {
		data.Template = NewNodePoolTemplateValueNull()
	}
}

// isEmptyTemplate checks if a template value is semantically empty.
func isEmptyTemplate(tmpl NodePoolTemplateValue) bool {
	if tmpl.IsNull() || tmpl.IsUnknown() {
		return true
	}

	metadata := tmpl.Metadata
	spec := tmpl.Spec

	if metadata.IsNull() || metadata.IsUnknown() {
		// metadata missing means empty
	} else {
		if len(metadata.Annotations.Elements()) > 0 {
			return false
		}
		if len(metadata.Finalizers.Elements()) > 0 {
			return false
		}
		if len(metadata.Labels.Elements()) > 0 {
			return false
		}
	}

	if spec.IsNull() || spec.IsUnknown() {
		// spec missing means empty
	} else {
		if len(spec.Taints.Elements()) > 0 {
			return false
		}
		if !spec.Unschedulable.IsNull() && !spec.Unschedulable.IsUnknown() && spec.Unschedulable.ValueBool() {
			return false
		}
	}

	return true
}

// nodePoolCreateOptsFromModel builds create opts from the framework model.
func nodePoolCreateOptsFromModel(data *cloudProjectKubeNodePoolResourceModel) *CloudProjectKubeNodePoolCreateOpts {
	opts := &CloudProjectKubeNodePoolCreateOpts{
		FlavorName: data.FlavorName.ValueString(),
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		opts.Name = &name
	}

	if !data.AntiAffinity.IsNull() && !data.AntiAffinity.IsUnknown() {
		v := data.AntiAffinity.ValueBool()
		opts.AntiAffinity = &v
	}

	if !data.Autoscale.IsNull() && !data.Autoscale.IsUnknown() {
		v := data.Autoscale.ValueBool()
		opts.Autoscale = &v
	}

	if !data.MonthlyBilled.IsNull() && !data.MonthlyBilled.IsUnknown() {
		v := data.MonthlyBilled.ValueBool()
		opts.MonthlyBilled = &v
	}

	if !data.DesiredNodes.IsNull() && !data.DesiredNodes.IsUnknown() {
		v := int(data.DesiredNodes.ValueInt64())
		opts.DesiredNodes = &v
	}

	if !data.MaxNodes.IsNull() && !data.MaxNodes.IsUnknown() {
		v := int(data.MaxNodes.ValueInt64())
		opts.MaxNodes = &v
	}

	if !data.MinNodes.IsNull() && !data.MinNodes.IsUnknown() {
		v := int(data.MinNodes.ValueInt64())
		opts.MinNodes = &v
	}

	if !data.AvailabilityZones.IsNull() && !data.AvailabilityZones.IsUnknown() {
		azs := stringSliceFromTfList(data.AvailabilityZones)
		opts.AvailabilityZones = &azs
	}

	opts.Autoscaling = autoscalingOptsFromModel(data)
	opts.Template = templateFromModel(data.Template)

	return opts
}

// nodePoolUpdateOptsFromModel builds update opts from the framework model.
func nodePoolUpdateOptsFromModel(data *cloudProjectKubeNodePoolResourceModel) *CloudProjectKubeNodePoolUpdateOpts {
	opts := &CloudProjectKubeNodePoolUpdateOpts{}

	if !data.Autoscale.IsNull() && !data.Autoscale.IsUnknown() {
		v := data.Autoscale.ValueBool()
		opts.Autoscale = &v
	}

	if !data.DesiredNodes.IsNull() && !data.DesiredNodes.IsUnknown() {
		v := int(data.DesiredNodes.ValueInt64())
		opts.DesiredNodes = &v
	}

	if !data.MaxNodes.IsNull() && !data.MaxNodes.IsUnknown() {
		v := int(data.MaxNodes.ValueInt64())
		opts.MaxNodes = &v
	}

	if !data.MinNodes.IsNull() && !data.MinNodes.IsUnknown() {
		v := int(data.MinNodes.ValueInt64())
		opts.MinNodes = &v
	}

	opts.Autoscaling = autoscalingOptsFromModel(data)
	opts.Template = templateFromModel(data.Template)

	return opts
}

// autoscalingOptsFromModel builds autoscaling opts from the model.
func autoscalingOptsFromModel(data *cloudProjectKubeNodePoolResourceModel) *CloudProjectKubeNodePoolAutoscaling {
	var autoscaling CloudProjectKubeNodePoolAutoscaling
	hasValue := false

	if !data.AutoscalingScaleDownUnneededTimeSeconds.IsNull() && !data.AutoscalingScaleDownUnneededTimeSeconds.IsUnknown() {
		v := int(data.AutoscalingScaleDownUnneededTimeSeconds.ValueInt64())
		autoscaling.ScaleDownUnneededTimeSeconds = &v
		hasValue = true
	}

	if !data.AutoscalingScaleDownUnreadyTimeSeconds.IsNull() && !data.AutoscalingScaleDownUnreadyTimeSeconds.IsUnknown() {
		v := int(data.AutoscalingScaleDownUnreadyTimeSeconds.ValueInt64())
		autoscaling.ScaleDownUnreadyTimeSeconds = &v
		hasValue = true
	}

	if !data.AutoscalingScaleDownUtilizationThreshold.IsNull() && !data.AutoscalingScaleDownUtilizationThreshold.IsUnknown() {
		bf := data.AutoscalingScaleDownUtilizationThreshold.ValueBigFloat()
		if bf != nil {
			f, _ := bf.Float64()
			autoscaling.ScaleDownUtilizationThreshold = &f
			hasValue = true
		}
	}

	if !hasValue {
		return nil
	}
	return &autoscaling
}

// templateFromModel converts a NodePoolTemplateValue to the API struct.
func templateFromModel(tmpl NodePoolTemplateValue) *CloudProjectKubeNodePoolTemplate {
	if tmpl.IsNull() || tmpl.IsUnknown() {
		return nil
	}

	template := &CloudProjectKubeNodePoolTemplate{
		Metadata: CloudProjectKubeNodePoolTemplateMetadata{
			Annotations: make(map[string]string),
			Finalizers:  make([]string, 0),
			Labels:      make(map[string]string),
		},
		Spec: CloudProjectKubeNodePoolTemplateSpec{
			Taints:        make([]Taint, 0),
			Unschedulable: false,
		},
	}

	// Metadata
	if !tmpl.Metadata.IsNull() && !tmpl.Metadata.IsUnknown() {
		// Annotations
		for k, v := range tmpl.Metadata.Annotations.Elements() {
			template.Metadata.Annotations[k] = v.(ovhtypes.TfStringValue).ValueString()
		}
		// Finalizers
		for _, v := range tmpl.Metadata.Finalizers.Elements() {
			template.Metadata.Finalizers = append(template.Metadata.Finalizers, v.(ovhtypes.TfStringValue).ValueString())
		}
		// Labels
		for k, v := range tmpl.Metadata.Labels.Elements() {
			template.Metadata.Labels[k] = v.(ovhtypes.TfStringValue).ValueString()
		}
	}

	// Spec
	if !tmpl.Spec.IsNull() && !tmpl.Spec.IsUnknown() {
		if !tmpl.Spec.Unschedulable.IsNull() && !tmpl.Spec.Unschedulable.IsUnknown() {
			template.Spec.Unschedulable = tmpl.Spec.Unschedulable.ValueBool()
		}

		for _, elem := range tmpl.Spec.Taints.Elements() {
			taintVal := elem.(NodePoolTaintValue)
			effectStr := taintVal.Effect.ValueString()
			effect := TaintEffecTypeToID[effectStr]

			taint := Taint{
				Effect: effect,
				Key:    taintVal.Key.ValueString(),
			}
			if !taintVal.Value.IsNull() && !taintVal.Value.IsUnknown() {
				taint.Value = taintVal.Value.ValueString()
			}

			template.Spec.Taints = append(template.Spec.Taints, taint)
		}
	}

	return template
}

// stringSliceFromTfList extracts a []string from a TfListNestedValue[TfStringValue].
func stringSliceFromTfList(list ovhtypes.TfListNestedValue[ovhtypes.TfStringValue]) []string {
	elems := list.Elements()
	result := make([]string, 0, len(elems))
	for _, elem := range elems {
		result = append(result, elem.(ovhtypes.TfStringValue).ValueString())
	}
	return result
}

// --- Waiters ---

func waitForCloudProjectKubeNodePoolReadyCtx(ctx context.Context, client *ovhwrap.Client, serviceName, kubeId, id string, timeout time.Duration) error {
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		res := &CloudProjectKubeNodePoolResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
			url.PathEscape(id),
		)
		if err := client.GetWithContext(ctx, endpoint, res); err != nil {
			return retry.NonRetryableError(fmt.Errorf("error reading nodepool %s: %w", id, err))
		}

		if res.Status == "READY" {
			return nil
		}
		if res.Status == "ERROR" {
			return retry.NonRetryableError(fmt.Errorf("nodepool %s is in ERROR state", id))
		}

		return retry.RetryableError(fmt.Errorf("nodepool %s is in state %s, waiting for READY", id, res.Status))
	})
	return err
}

func waitForCloudProjectKubeNodePoolDeletedCtx(ctx context.Context, client *ovhwrap.Client, serviceName, kubeId, id string, timeout time.Duration) error {
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		res := &CloudProjectKubeNodePoolResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
			url.PathEscape(id),
		)
		if err := client.GetWithContext(ctx, endpoint, res); err != nil {
			if errOvh, ok := err.(*ovh.APIError); ok && errOvh.Code == 404 {
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error reading nodepool %s: %w", id, err))
		}

		return retry.RetryableError(fmt.Errorf("nodepool %s is in state %s, waiting for deletion", id, res.Status))
	})
	return err
}
