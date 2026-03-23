package ovh

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeNodePoolDataSource struct {
	config *Config
}

func NewCloudProjectKubeNodePoolDataSource() datasource.DataSource {
	return &cloudProjectKubeNodePoolDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeNodePoolDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeNodePoolDataSource{}
)

func (d *cloudProjectKubeNodePoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_nodepool"
}

func (d *cloudProjectKubeNodePoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudProjectKubeNodePoolDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CloudProjectKubeNodePoolDataSourceSchema(ctx)
}

func (d *cloudProjectKubeNodePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudProjectKubeNodePoolDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.KubeId.ValueString()
	nodepoolName := data.Name.ValueString()

	// Read API call logic — fetch all nodepools and filter by name
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	log.Printf("[DEBUG] Will read nodepools from cluster %s in project %s", kubeId, serviceName)

	var res []CloudProjectKubeNodePoolDataSourceModel
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to get kube nodepools",
			fmt.Sprintf("error calling GET %s: %s", endpoint, err),
		)
		return
	}

	// Find the nodepool matching the requested name
	var nodepoolTarget *CloudProjectKubeNodePoolDataSourceModel
	for i := range res {
		if res[i].Name.ValueString() == nodepoolName {
			nodepoolTarget = &res[i]
			break
		}
	}

	if nodepoolTarget == nil {
		resp.Diagnostics.AddError(
			"Nodepool not found",
			fmt.Sprintf("the nodepool named %s cannot be found for cluster %s in project %s", nodepoolName, kubeId, serviceName),
		)
		return
	}

	// Copy deserialized data from the matched nodepool
	data = *nodepoolTarget

	// flavor_name is a copy of flavor (same field from API)
	data.FlavorName = nodepoolTarget.Flavor

	// Flatten the nested autoscaling object into top-level attributes
	autoscaling := nodepoolTarget.Autoscaling
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

	log.Printf("[DEBUG] Read nodepool: %+v", nodepoolTarget)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
