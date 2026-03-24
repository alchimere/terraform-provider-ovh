package ovh

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Schema ---

func CloudProjectKubeNodepoolNodesDataSourceSchema(ctx context.Context) schema.Schema {
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
			"nodes": schema.ListAttribute{
				CustomType:  ovhtypes.NewTfListNestedType[CloudProjectKubeNodeValue](ctx),
				Computed:    true,
				Description: "Nodes composing the node pool",
			},
		},
	}
}

// --- Model ---

type CloudProjectKubeNodepoolNodesDataSourceModel struct {
	ServiceName ovhtypes.TfStringValue                                `tfsdk:"service_name"`
	KubeId      ovhtypes.TfStringValue                                `tfsdk:"kube_id"`
	Name        ovhtypes.TfStringValue                                `tfsdk:"name"`
	Nodes       ovhtypes.TfListNestedValue[CloudProjectKubeNodeValue] `tfsdk:"nodes"`
}
