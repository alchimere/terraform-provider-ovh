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

type cloudProjectKubeNodepoolNodesDataSource struct {
	config *Config
}

func NewCloudProjectKubeNodepoolNodesDataSource() datasource.DataSource {
	return &cloudProjectKubeNodepoolNodesDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeNodepoolNodesDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeNodepoolNodesDataSource{}
)

func (d *cloudProjectKubeNodepoolNodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube_nodepool_nodes"
}

func (d *cloudProjectKubeNodepoolNodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudProjectKubeNodepoolNodesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = CloudProjectKubeNodepoolNodesDataSourceSchema(ctx)
}

func (d *cloudProjectKubeNodepoolNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudProjectKubeNodepoolNodesDataSourceModel

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
	name := data.Name.ValueString()

	endpointNodepool := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool", serviceName, kubeId)
	var nodePoolRes []CloudProjectKubeNodePoolResponse
	log.Printf("[DEBUG] Will read nodepools from cluster %s in project %s", kubeId, serviceName)
	if err := d.config.OVHClient.GetWithContext(ctx, endpointNodepool, &nodePoolRes); err != nil {
		resp.Diagnostics.AddError(
			"Failed to get nodepools",
			fmt.Sprintf("error calling GET %s: %s", endpointNodepool, err),
		)
		return
	}

	var nodepoolTarget *CloudProjectKubeNodePoolResponse

	for _, nodepool := range nodePoolRes {
		if nodepool.Name == name {
			nodepoolTarget = &nodepool
			break
		}
	}

	if nodepoolTarget == nil {
		resp.Diagnostics.AddError(
			"Nodepool not found",
			fmt.Sprintf("The nodepool named %s cannot be found for cluster %s in project %s", name, kubeId, serviceName),
		)
		return
	}

	endpointNodepoolNodes := fmt.Sprintf("/cloud/project/%s/kube/%s/nodepool/%s/nodes",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
		url.PathEscape(nodepoolTarget.Id))
	var resNodepoolNodes []CloudProjectKubeNodeResponse

	log.Printf("[DEBUG] Will read nodes from node pool %s and cluster %s in project %s", name, kubeId, serviceName)
	if err := d.config.OVHClient.GetWithContext(ctx, endpointNodepoolNodes, &resNodepoolNodes); err != nil {
		resp.Diagnostics.AddError(
			"Failed to get nodepool nodes",
			fmt.Sprintf("error calling GET %s: %s", endpointNodepoolNodes, err),
		)
		return
	}

	nodes := make([]CloudProjectKubeNodeValue, len(resNodepoolNodes))
	ids := make([]string, len(resNodepoolNodes))

	for i, node := range resNodepoolNodes {
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

	log.Printf("[DEBUG] Read nodepool nodes: %+v", resNodepoolNodes)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
