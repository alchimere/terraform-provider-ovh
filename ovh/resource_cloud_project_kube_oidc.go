package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/helpers"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/ovhwrap"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Resource struct ---

type cloudProjectKubeOIDCResource struct {
	config *Config
}

func NewCloudProjectKubeOIDCResource() resource.Resource {
	return &cloudProjectKubeOIDCResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &cloudProjectKubeOIDCResource{}
	_ resource.ResourceWithConfigure   = &cloudProjectKubeOIDCResource{}
	_ resource.ResourceWithImportState = &cloudProjectKubeOIDCResource{}
)

// --- Model ---

type cloudProjectKubeOIDCResourceModel struct {
	ID                 ovhtypes.TfStringValue                             `tfsdk:"id"`
	ServiceName        ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId             ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	ClientId           ovhtypes.TfStringValue                             `tfsdk:"client_id"`
	IssuerUrl          ovhtypes.TfStringValue                             `tfsdk:"issuer_url"`
	OidcUsernameClaim  ovhtypes.TfStringValue                             `tfsdk:"oidc_username_claim"`
	OidcUsernamePrefix ovhtypes.TfStringValue                             `tfsdk:"oidc_username_prefix"`
	OidcGroupsClaim    ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"oidc_groups_claim"`
	OidcGroupsPrefix   ovhtypes.TfStringValue                             `tfsdk:"oidc_groups_prefix"`
	OidcRequiredClaim  ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"oidc_required_claim"`
	OidcSigningAlgs    ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"oidc_signing_algs"`
	OidcCaContent      ovhtypes.TfStringValue                             `tfsdk:"oidc_ca_content"`
}

// --- Metadata / Configure / Schema ---

func (r *cloudProjectKubeOIDCResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_oidc"
}

func (r *cloudProjectKubeOIDCResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudProjectKubeOIDCResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"client_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "",
			},
			"issuer_url": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "",
			},
			"oidc_username_claim": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "",
			},
			"oidc_username_prefix": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "",
			},
			"oidc_groups_claim": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Optional:    true,
				Description: "",
			},
			"oidc_groups_prefix": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "",
			},
			"oidc_required_claim": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Optional:    true,
				Description: "",
			},
			"oidc_signing_algs": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Optional:    true,
				Description: "",
			},
			"oidc_ca_content": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "",
			},
		},
	}
}

// --- CRUD Operations ---

func (r *cloudProjectKubeOIDCResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data cloudProjectKubeOIDCResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	params := oidcCreateOptsFromModel(&data)
	res := &CloudProjectKubeOIDCResponse{}

	log.Printf("[DEBUG] Will create OIDC: %+v", params)
	if err := r.config.OVHClient.PostWithContext(ctx, endpoint, params, res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kube OIDC",
			fmt.Sprintf("error calling POST %s: %s", endpoint, err),
		)
		return
	}

	// Set the ID
	data.ID = ovhtypes.NewTfStringValue(serviceName + "/" + kubeId)
	data.ServiceName = ovhtypes.NewTfStringValue(serviceName)
	data.KubeId = ovhtypes.NewTfStringValue(kubeId)

	log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
	if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for kube to be READY",
			fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
		)
		return
	}
	log.Printf("[DEBUG] kube %s is READY", kubeId)

	// Read back the resource
	if err := r.readOIDC(ctx, serviceName, kubeId, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube OIDC after create",
			err.Error(),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeOIDCResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cloudProjectKubeOIDCResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	log.Printf("[DEBUG] Will read OIDC from kube %s and project: %s", kubeId, serviceName)

	if err := r.readOIDC(ctx, serviceName, kubeId, &data); err != nil {
		helpers.CheckDeletedWithContext(ctx, resp, err, fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect", serviceName, kubeId))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeOIDCResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data cloudProjectKubeOIDCResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	params := oidcUpdateOptsFromModel(&data)

	log.Printf("[DEBUG] Will update OIDC: %+v", params)
	if err := r.config.OVHClient.PutWithContext(ctx, endpoint, params, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to update kube OIDC",
			fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
	if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for kube to be READY",
			fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
		)
		return
	}
	log.Printf("[DEBUG] kube %s is READY", kubeId)

	// Read back the resource
	if err := r.readOIDC(ctx, serviceName, kubeId, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube OIDC after update",
			err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeOIDCResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cloudProjectKubeOIDCResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	log.Printf("[DEBUG] Will delete OIDC")
	if err := r.config.OVHClient.DeleteWithContext(ctx, endpoint, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete kube OIDC",
			fmt.Sprintf("error calling DELETE %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
	if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for kube to be READY",
			fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
		)
		return
	}
	log.Printf("[DEBUG] kube %s is READY", kubeId)
}

func (r *cloudProjectKubeOIDCResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	splits := strings.SplitN(req.ID, "/", 2)
	if len(splits) != 2 {
		resp.Diagnostics.AddError(
			"Given ID is malformed",
			"ID must be formatted as: service_name/kube_id",
		)
		return
	}

	serviceName := splits[0]
	kubeId := splits[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), serviceName+"/"+kubeId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), serviceName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("kube_id"), kubeId)...)
}

// --- Helpers ---

// readOIDC reads OIDC configuration from the API and populates the model.
func (r *cloudProjectKubeOIDCResource) readOIDC(ctx context.Context, serviceName, kubeId string, data *cloudProjectKubeOIDCResourceModel) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	res := &CloudProjectKubeOIDCResponse{}

	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, res); err != nil {
		return fmt.Errorf("error calling GET %s: %w", endpoint, err)
	}

	// Map response to model
	data.ID = ovhtypes.NewTfStringValue(serviceName + "/" + kubeId)
	data.ServiceName = ovhtypes.NewTfStringValue(serviceName)
	data.KubeId = ovhtypes.NewTfStringValue(kubeId)
	data.ClientId = ovhtypes.NewTfStringValue(res.ClientID)
	data.IssuerUrl = ovhtypes.NewTfStringValue(res.IssuerUrl)
	data.OidcUsernameClaim = ovhtypes.NewTfStringValue(res.UsernameClaim)
	data.OidcUsernamePrefix = ovhtypes.NewTfStringValue(res.UsernamePrefix)
	data.OidcGroupsClaim = tfListFromStringSlice(ctx, res.GroupsClaim)
	data.OidcGroupsPrefix = ovhtypes.NewTfStringValue(res.GroupsPrefix)
	data.OidcRequiredClaim = tfListFromStringSlice(ctx, res.RequiredClaim)
	data.OidcSigningAlgs = tfListFromStringSlice(ctx, res.SigningAlgs)
	data.OidcCaContent = ovhtypes.NewTfStringValue(res.CaContent)

	log.Printf("[DEBUG] Read OIDC: %+v", res)
	return nil
}

// oidcCreateOptsFromModel builds create opts from the framework model.
func oidcCreateOptsFromModel(data *cloudProjectKubeOIDCResourceModel) *CloudProjectKubeOIDCCreateOpts {
	opts := &CloudProjectKubeOIDCCreateOpts{
		ClientID:  data.ClientId.ValueString(),
		IssuerUrl: data.IssuerUrl.ValueString(),
	}

	if !data.OidcUsernameClaim.IsNull() && !data.OidcUsernameClaim.IsUnknown() {
		opts.UsernameClaim = data.OidcUsernameClaim.ValueString()
	}
	if !data.OidcUsernamePrefix.IsNull() && !data.OidcUsernamePrefix.IsUnknown() {
		opts.UsernamePrefix = data.OidcUsernamePrefix.ValueString()
	}
	if !data.OidcGroupsClaim.IsNull() && !data.OidcGroupsClaim.IsUnknown() {
		opts.GroupsClaim = stringSliceFromTfList(data.OidcGroupsClaim)
	}
	if !data.OidcGroupsPrefix.IsNull() && !data.OidcGroupsPrefix.IsUnknown() {
		opts.GroupsPrefix = data.OidcGroupsPrefix.ValueString()
	}
	if !data.OidcRequiredClaim.IsNull() && !data.OidcRequiredClaim.IsUnknown() {
		opts.RequiredClaim = stringSliceFromTfList(data.OidcRequiredClaim)
	}
	if !data.OidcSigningAlgs.IsNull() && !data.OidcSigningAlgs.IsUnknown() {
		opts.SigningAlgs = stringSliceFromTfList(data.OidcSigningAlgs)
	}
	if !data.OidcCaContent.IsNull() && !data.OidcCaContent.IsUnknown() {
		opts.CaContent = data.OidcCaContent.ValueString()
	}

	return opts
}

// oidcUpdateOptsFromModel builds update opts from the framework model.
func oidcUpdateOptsFromModel(data *cloudProjectKubeOIDCResourceModel) *CloudProjectKubeOIDCUpdateOpts {
	opts := &CloudProjectKubeOIDCUpdateOpts{
		ClientID:  data.ClientId.ValueString(),
		IssuerUrl: data.IssuerUrl.ValueString(),
	}

	if !data.OidcUsernameClaim.IsNull() && !data.OidcUsernameClaim.IsUnknown() {
		opts.UsernameClaim = data.OidcUsernameClaim.ValueString()
	}
	if !data.OidcUsernamePrefix.IsNull() && !data.OidcUsernamePrefix.IsUnknown() {
		opts.UsernamePrefix = data.OidcUsernamePrefix.ValueString()
	}
	if !data.OidcGroupsClaim.IsNull() && !data.OidcGroupsClaim.IsUnknown() {
		opts.GroupsClaim = stringSliceFromTfList(data.OidcGroupsClaim)
	}
	if !data.OidcGroupsPrefix.IsNull() && !data.OidcGroupsPrefix.IsUnknown() {
		opts.GroupsPrefix = data.OidcGroupsPrefix.ValueString()
	}
	if !data.OidcRequiredClaim.IsNull() && !data.OidcRequiredClaim.IsUnknown() {
		opts.RequiredClaim = stringSliceFromTfList(data.OidcRequiredClaim)
	}
	if !data.OidcSigningAlgs.IsNull() && !data.OidcSigningAlgs.IsUnknown() {
		opts.SigningAlgs = stringSliceFromTfList(data.OidcSigningAlgs)
	}
	if !data.OidcCaContent.IsNull() && !data.OidcCaContent.IsUnknown() {
		opts.CaContent = data.OidcCaContent.ValueString()
	}

	return opts
}

// tfListFromStringSlice converts a []string to TfListNestedValue[TfStringValue].
func tfListFromStringSlice(ctx context.Context, slice []string) ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] {
	if slice == nil {
		return ovhtypes.NewListNestedObjectValueOfNull[ovhtypes.TfStringValue](ctx)
	}

	elements := make([]attr.Value, len(slice))
	for i, v := range slice {
		elements[i] = ovhtypes.NewTfStringValue(v)
	}

	return ovhtypes.TfListNestedValue[ovhtypes.TfStringValue]{
		ListValue: basetypes.NewListValueMust(ovhtypes.TfStringType{}, elements),
	}
}

// waitForCloudProjectKubeReadyCtx waits for a kube cluster to be READY.
func waitForCloudProjectKubeReadyCtx(ctx context.Context, client *ovhwrap.Client, serviceName, kubeId string, timeout time.Duration) error {
	err := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		res := &CloudProjectKubeResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", serviceName, kubeId)
		if err := client.GetWithContext(ctx, endpoint, res); err != nil {
			return retry.NonRetryableError(fmt.Errorf("error reading kube %s: %w", kubeId, err))
		}

		if res.Status == "READY" {
			return nil
		}
		if res.Status == "ERROR" {
			return retry.NonRetryableError(fmt.Errorf("kube %s is in ERROR state", kubeId))
		}

		return retry.RetryableError(fmt.Errorf("kube %s is in state %s, waiting for READY", kubeId, res.Status))
	})
	return err
}
