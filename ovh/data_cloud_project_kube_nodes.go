package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeNodesDataSource struct {
	config *Config
}

func NewCloudProjectKubeNodesDataSource() datasource.DataSource {
	return &cloudProjectKubeNodesDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeNodesDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeNodesDataSource{}
)

func (d *cloudProjectKubeNodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_nodes"
}

func (d *cloudProjectKubeNodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudProjectKubeNodesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CloudProjectKubeNodesDataSourceSchema(ctx)
}

func (d *cloudProjectKubeNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudProjectKubeNodesDataSourceModel

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

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/node",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId))
	var res []CloudProjectKubeNodeResponse

	log.Printf("[DEBUG] Will read nodes from cluster %s in project %s", kubeId, serviceName)
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, &res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to get kube nodes",
			fmt.Sprintf("error calling GET %s: %s", endpoint, err),
		)
		return
	}

	nodes := make([]CloudProjectKubeNodeValue, len(res))
	ids := make([]string, len(res))

	for i, node := range res {
		nodes[i] = CloudProjectKubeNodeValue{
			CreatedAt:  ovhtypes.NewTfStringValue(node.CreatedAt),
			UpdatedAt:  ovhtypes.NewTfStringValue(node.UpdatedAt),
			DeployedAt: ovhtypes.NewTfStringValue(node.DeployedAt),
			Flavor:     ovhtypes.NewTfStringValue(node.Flavor),
			Id:         ovhtypes.NewTfStringValue(node.Id),
			InstanceId: ovhtypes.NewTfStringValue(node.InstanceId),
			IsUpToDate: ovhtypes.NewTfBoolValue(node.IsUpToDate),
			Name:       ovhtypes.NewTfStringValue(node.Name),
			NodePoolId: ovhtypes.NewTfStringValue(node.NodePoolId),
			ProjectId:  ovhtypes.NewTfStringValue(node.ProjectId),
			Status:     ovhtypes.NewTfStringValue(node.Status),
			Version:    ovhtypes.NewTfStringValue(node.Version),
			state:      attr.ValueStateKnown,
		}
		ids = append(ids, node.Id)
	}

	// sort.Strings sorts in place, returns nothing
	sort.Strings(ids)

	// Convert nodes to attr.Value slice
	nodeValues := make([]attr.Value, len(nodes))
	for i, node := range nodes {
		nodeValues[i] = node
	}

	data.Nodes = ovhtypes.TfListNestedValue[CloudProjectKubeNodeValue]{
		ListValue: basetypes.NewListValueMust(CloudProjectKubeNodeValue{}.Type(ctx), nodeValues),
	}

	log.Printf("[DEBUG] Read nodes: %+v", res)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
