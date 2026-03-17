package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeIPRestrictionsDataSource struct {
	config *Config
}

type cloudProjectKubeIPRestrictionsModel struct {
	ServiceName ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId      ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	IPs         ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"ips"`
}

func NewCloudProjectKubeIPRestrictionsDataSource() datasource.DataSource {
	return &cloudProjectKubeIPRestrictionsDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeIPRestrictionsDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeIPRestrictionsDataSource{}
)

func (d *cloudProjectKubeIPRestrictionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_iprestrictions"
}

func (d *cloudProjectKubeIPRestrictionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudProjectKubeIPRestrictionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"ips": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Computed:    true,
				Description: "List of IP restrictions for the cluster",
			},
		},
	}
}

func (d *cloudProjectKubeIPRestrictionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudProjectKubeIPRestrictionsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	// Read API call logic
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/ipRestrictions",
		url.PathEscape(data.ServiceName.ValueString()),
		url.PathEscape(data.KubeId.ValueString()),
	)

	log.Printf("[DEBUG] Will read iprestrictions from cluster %s in project %s",
		data.KubeId.ValueString(), data.ServiceName.ValueString())

	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &data.IPs); err != nil {
		resp.Diagnostics.AddError(
			"Failed to get kube IP restrictions",
			fmt.Sprintf("error calling GET %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Read iprestrictions: %+v", data.IPs)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
