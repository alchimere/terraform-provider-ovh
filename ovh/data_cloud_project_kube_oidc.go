package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeOIDCDataSource struct {
	config *Config
}

type cloudProjectKubeOIDCModel struct {
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

func NewCloudProjectKubeOIDCDataSource() datasource.DataSource {
	return &cloudProjectKubeOIDCDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeOIDCDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeOIDCDataSource{}
)

func (d *cloudProjectKubeOIDCDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_oidc"
}

func (d *cloudProjectKubeOIDCDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.config = config
}

func (d *cloudProjectKubeOIDCDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Service name",
			},
			"kube_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Kube ID",
			},
			"client_id": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"issuer_url": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"oidc_username_claim": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"oidc_username_prefix": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"oidc_groups_claim": schema.ListAttribute{
				CustomType: ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Computed:   true,
			},
			"oidc_groups_prefix": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"oidc_required_claim": schema.ListAttribute{
				CustomType: ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Computed:   true,
			},
			"oidc_signing_algs": schema.ListAttribute{
				CustomType: ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Computed:   true,
			},
			"oidc_ca_content": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
		},
	}
}

func (d *cloudProjectKubeOIDCDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudProjectKubeOIDCModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/openIdConnect",
		url.PathEscape(data.ServiceName.ValueString()),
		url.PathEscape(data.KubeId.ValueString()),
	)

	res := &CloudProjectKubeOIDCResponse{}

	log.Printf("[DEBUG] Will read OIDC from kube %s and project: %s", data.KubeId.ValueString(), data.ServiceName.ValueString())
	err := d.config.OVHClient.GetWithContext(ctx, endpoint, res)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get kube OIDC",
			fmt.Sprintf("error calling GET %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Read OIDC %+v", res)

	// Convert string slices to TfListNestedValue
	groupsClaimElements := make([]attr.Value, len(res.GroupsClaim))
	for i, v := range res.GroupsClaim {
		groupsClaimElements[i] = ovhtypes.NewTfStringValue(v)
	}
	data.OidcGroupsClaim = ovhtypes.TfListNestedValue[ovhtypes.TfStringValue]{
		ListValue: basetypes.NewListValueMust(ovhtypes.TfStringType{}, groupsClaimElements),
	}

	requiredClaimElements := make([]attr.Value, len(res.RequiredClaim))
	for i, v := range res.RequiredClaim {
		requiredClaimElements[i] = ovhtypes.NewTfStringValue(v)
	}
	data.OidcRequiredClaim = ovhtypes.TfListNestedValue[ovhtypes.TfStringValue]{
		ListValue: basetypes.NewListValueMust(ovhtypes.TfStringType{}, requiredClaimElements),
	}

	signingAlgsElements := make([]attr.Value, len(res.SigningAlgs))
	for i, v := range res.SigningAlgs {
		signingAlgsElements[i] = ovhtypes.NewTfStringValue(v)
	}
	data.OidcSigningAlgs = ovhtypes.TfListNestedValue[ovhtypes.TfStringValue]{
		ListValue: basetypes.NewListValueMust(ovhtypes.TfStringType{}, signingAlgsElements),
	}

	// Set the string data
	data.ClientId = ovhtypes.NewTfStringValue(res.ClientID)
	data.IssuerUrl = ovhtypes.NewTfStringValue(res.IssuerUrl)
	data.OidcUsernameClaim = ovhtypes.NewTfStringValue(res.UsernameClaim)
	data.OidcUsernamePrefix = ovhtypes.NewTfStringValue(res.UsernamePrefix)
	data.OidcGroupsPrefix = ovhtypes.NewTfStringValue(res.GroupsPrefix)
	data.OidcCaContent = ovhtypes.NewTfStringValue(res.CaContent)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
