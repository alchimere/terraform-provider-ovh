package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

type cloudProjectKubeDataSource struct {
	config *Config
}

func NewCloudProjectKubeDataSource() datasource.DataSource {
	return &cloudProjectKubeDataSource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &cloudProjectKubeDataSource{}
	_ datasource.DataSourceWithConfigure = &cloudProjectKubeDataSource{}
)

func (d *cloudProjectKubeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube"
}

func (d *cloudProjectKubeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudProjectKubeDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"name": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"version": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"plan": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "Cluster plan",
				Validators: []validator.String{
					stringvalidator.OneOf("standard", "free"),
				},
			},
			"region": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"kube_proxy_mode": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.OneOf("iptables", "ipvs"),
				},
			},
			"private_network_id": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"load_balancers_subnet_id": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"nodes_subnet_id": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"update_policy": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"control_plane_is_up_to_date": schema.BoolAttribute{
				CustomType: ovhtypes.TfBoolType{},
				Computed:   true,
			},
			"is_up_to_date": schema.BoolAttribute{
				CustomType: ovhtypes.TfBoolType{},
				Computed:   true,
			},
			"next_upgrade_versions": schema.ListAttribute{
				CustomType: ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
				Computed:   true,
			},
			"nodes_url": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"status": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"url": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
			},
			"kubeconfig": schema.StringAttribute{
				CustomType: ovhtypes.TfStringType{},
				Computed:   true,
				Sensitive:  true,
			},
		},
		Blocks: map[string]schema.Block{
			"customization_apiserver": schema.SingleNestedBlock{
				Description: "Kubernetes API server customization",
				Blocks: map[string]schema.Block{
					"admissionplugins": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Computed:    true,
								Description: "Enabled admission plugins",
							},
							"disabled": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Computed:    true,
								Description: "Disabled admission plugins",
							},
						},
					},
				},
			},
			"customization": schema.SingleNestedBlock{
				Description:        "Kubernetes cluster customization (deprecated)",
				DeprecationMessage: "Use customization_apiserver instead",
				Blocks: map[string]schema.Block{
					"apiserver": schema.SingleNestedBlock{
						DeprecationMessage: "Use customization_apiserver instead",
						Blocks: map[string]schema.Block{
							"admissionplugins": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.ListAttribute{
										CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
										Computed:    true,
										Description: "Enabled admission plugins",
									},
									"disabled": schema.ListAttribute{
										CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
										Computed:    true,
										Description: "Disabled admission plugins",
									},
								},
							},
						},
					},
				},
			},
			"customization_kube_proxy": schema.SingleNestedBlock{
				Description: "Kubernetes kube-proxy customization",
				Blocks: map[string]schema.Block{
					"iptables": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min_sync_period": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"sync_period": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
						},
					},
					"ipvs": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min_sync_period": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"scheduler": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"sync_period": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"tcp_fin_timeout": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"tcp_timeout": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
							"udp_timeout": schema.StringAttribute{
								CustomType: ovhtypes.TfStringType{},
								Computed:   true,
							},
						},
					},
				},
			},
			"kubeconfig_attributes": schema.SingleNestedBlock{
				Description: "The kubeconfig configuration file of the Kubernetes cluster",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						CustomType: ovhtypes.TfStringType{},
						Computed:   true,
					},
					"cluster_ca_certificate": schema.StringAttribute{
						CustomType: ovhtypes.TfStringType{},
						Computed:   true,
						Sensitive:  true,
					},
					"client_certificate": schema.StringAttribute{
						CustomType: ovhtypes.TfStringType{},
						Computed:   true,
						Sensitive:  true,
					},
					"client_key": schema.StringAttribute{
						CustomType: ovhtypes.TfStringType{},
						Computed:   true,
						Sensitive:  true,
					},
				},
			},
		},
	}
}

func (d *cloudProjectKubeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data cloudProjectKubeDataSourceModel

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

	log.Printf("[DEBUG] Will read public cloud kube %s for project: %s", kubeId, serviceName)

	res := &CloudProjectKubeResponse{}
	endpoint := fmt.Sprintf(
		"/cloud/project/%s/kube/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)
	if err := d.config.OVHClient.GetWithContext(ctx, endpoint, res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube cluster",
			fmt.Sprintf("error calling GET %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Read kube %+v", res)

	// Populate the model from the API response
	diags := dataSourceModelFromResponse(ctx, res, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch kubeconfig
	if err := setKubeconfigOnDataSourceModel(d.config, serviceName, kubeId, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kubeconfig",
			fmt.Sprintf("error reading kubeconfig for kube %s: %s", kubeId, err),
		)
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
