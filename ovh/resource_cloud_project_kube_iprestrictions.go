package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/helpers"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeIPRestrictionsResource struct {
	config *Config
}

type cloudProjectKubeIPRestrictionsResourceModel struct {
	ID          ovhtypes.TfStringValue                             `tfsdk:"id"`
	ServiceName ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId      ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	IPs         ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"ips"`
}

func NewCloudProjectKubeIPRestrictionsResource() resource.Resource {
	return &cloudProjectKubeIPRestrictionsResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &cloudProjectKubeIPRestrictionsResource{}
	_ resource.ResourceWithConfigure   = &cloudProjectKubeIPRestrictionsResource{}
	_ resource.ResourceWithImportState = &cloudProjectKubeIPRestrictionsResource{}
)

func (r *cloudProjectKubeIPRestrictionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_iprestrictions"
}

func (r *cloudProjectKubeIPRestrictionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudProjectKubeIPRestrictionsResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"ips": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Required:    true,
				Description: "List of IP restrictions for the cluster",
			},
		},
	}
}

func (r *cloudProjectKubeIPRestrictionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data cloudProjectKubeIPRestrictionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	params := &CloudProjectKubeIpRestrictionsCreateOrUpdateOpts{
		Ips: ipsFromModel(data),
	}

	if err := r.updateIPRestrictions(ctx, serviceName, kubeId, params, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kube IP restrictions",
			err.Error(),
		)
		return
	}

	// Read back the resource
	if err := r.readIPRestrictions(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube IP restrictions after create",
			err.Error(),
		)
		return
	}

	data.ID = ovhtypes.NewTfStringValue(kubeId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeIPRestrictionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cloudProjectKubeIPRestrictionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions",
		url.PathEscape(data.ServiceName.ValueString()),
		url.PathEscape(data.KubeId.ValueString()),
	)

	log.Printf("[DEBUG] Will read iprestrictions from cluster %s in project %s",
		data.KubeId.ValueString(), data.ServiceName.ValueString())

	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, &data.IPs); err != nil {
		helpers.CheckDeletedWithContext(ctx, resp, err, endpoint)
		return
	}

	data.ID = ovhtypes.NewTfStringValue(data.KubeId.ValueString())

	log.Printf("[DEBUG] Read iprestrictions: %+v", data.IPs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeIPRestrictionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data cloudProjectKubeIPRestrictionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	params := &CloudProjectKubeIpRestrictionsCreateOrUpdateOpts{
		Ips: ipsFromModel(data),
	}

	if err := r.updateIPRestrictions(ctx, serviceName, kubeId, params, 5*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update kube IP restrictions",
			err.Error(),
		)
		return
	}

	// Read back the resource
	if err := r.readIPRestrictions(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube IP restrictions after update",
			err.Error(),
		)
		return
	}

	data.ID = ovhtypes.NewTfStringValue(kubeId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeIPRestrictionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cloudProjectKubeIPRestrictionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	params := &CloudProjectKubeIpRestrictionsCreateOrUpdateOpts{
		Ips: []string{},
	}

	if err := r.updateIPRestrictions(ctx, serviceName, kubeId, params, 5*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete kube IP restrictions",
			err.Error(),
		)
	}
}

func (r *cloudProjectKubeIPRestrictionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	splits := strings.SplitN(req.ID, "/", 3)
	if len(splits) != 2 {
		resp.Diagnostics.AddError(
			"Given ID is malformed",
			"ID must be formatted as: service_name/kube_id",
		)
		return
	}

	serviceName := splits[0]
	kubeId := splits[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), kubeId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), serviceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kube_id"), kubeId)...)
}

// updateIPRestrictions PUTs the IP restrictions and waits for the cluster to be READY.
func (r *cloudProjectKubeIPRestrictionsResource) updateIPRestrictions(ctx context.Context, serviceName, kubeId string, params *CloudProjectKubeIpRestrictionsCreateOrUpdateOpts, timeout time.Duration) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	log.Printf("[DEBUG] Will update iprestrictions: %+v", params)
	if err := r.config.OVHClient.PutWithContext(ctx, endpoint, params, nil); err != nil {
		return fmt.Errorf("calling PUT %s with params %s:\n\t %w", endpoint, params, err)
	}

	log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
	if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, timeout); err != nil {
		return fmt.Errorf("timeout while waiting kube %s to be READY: %w", kubeId, err)
	}
	log.Printf("[DEBUG] kube %s is READY", kubeId)

	return nil
}

// readIPRestrictions GETs the IP restrictions from the API and populates the model's IPs field.
func (r *cloudProjectKubeIPRestrictionsResource) readIPRestrictions(ctx context.Context, data *cloudProjectKubeIPRestrictionsResourceModel) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions",
		url.PathEscape(data.ServiceName.ValueString()),
		url.PathEscape(data.KubeId.ValueString()),
	)

	log.Printf("[DEBUG] Will read iprestrictions from cluster %s in project %s",
		data.KubeId.ValueString(), data.ServiceName.ValueString())

	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, &data.IPs); err != nil {
		return fmt.Errorf("error calling GET %s: %w", endpoint, err)
	}

	log.Printf("[DEBUG] Read iprestrictions: %+v", data.IPs)
	return nil
}

// ipsFromModel extracts a []string from the model's IPs field.
func ipsFromModel(data cloudProjectKubeIPRestrictionsResourceModel) []string {
	elems := data.IPs.Elements()
	ips := make([]string, 0, len(elems))
	for _, elem := range elems {
		ips = append(ips, elem.(ovhtypes.TfStringValue).ValueString())
	}
	return ips
}
