package ovh

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/ovh/go-ovh/ovh"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/helpers"
	"github.com/ovh/terraform-provider-ovh/v2/ovh/ovhwrap"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
	ovhvalidators "github.com/ovh/terraform-provider-ovh/v2/ovh/validators"
)

// --- Resource struct ---

type cloudProjectKubeResource struct {
	config *Config
}

func NewCloudProjectKubeResource() resource.Resource {
	return &cloudProjectKubeResource{}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &cloudProjectKubeResource{}
	_ resource.ResourceWithConfigure   = &cloudProjectKubeResource{}
	_ resource.ResourceWithImportState = &cloudProjectKubeResource{}
)

// --- Metadata / Configure / Schema ---

func (r *cloudProjectKubeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_project_kube"
}

func (r *cloudProjectKubeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudProjectKubeResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"name": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "Cluster name",
			},
			"version": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "Kubernetes version",
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
				CustomType:  ovhtypes.TfStringType{},
				Required:    true,
				Description: "Cluster region",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kube_proxy_mode": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "Kube proxy mode",
				Validators: []validator.String{
					stringvalidator.OneOf("iptables", "ipvs"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_network_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "Private network ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"load_balancers_subnet_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Description: "Load balancers subnet ID",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("private_network_id")),
				},
			},
			"nodes_subnet_id": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "Nodes subnet ID",
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("private_network_id")),
				},
			},
			"update_policy": schema.StringAttribute{
				CustomType:  ovhtypes.TfStringType{},
				Optional:    true,
				Computed:    true,
				Description: "Cluster update policy",
			},

			// Computed attributes
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
				Validators:  []validator.Object{},
				Blocks: map[string]schema.Block{
					"admissionplugins": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Optional:    true,
								Computed:    true,
								Description: "Enabled admission plugins",
							},
							"disabled": schema.ListAttribute{
								CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
								Optional:    true,
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
										Optional:    true,
										Computed:    true,
										Description: "Enabled admission plugins",
									},
									"disabled": schema.ListAttribute{
										CustomType:  ovhtypes.NewTfListNestedType[ovhtypes.TfStringValue](ctx),
										Optional:    true,
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
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Minimum period that iptables rules are refreshed, in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
							"sync_period": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Period that iptables rules are refreshed, in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
						},
					},
					"ipvs": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"min_sync_period": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Minimum period that IPVS rules are refreshed, in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
							"scheduler": schema.StringAttribute{
								CustomType:  ovhtypes.TfStringType{},
								Optional:    true,
								Description: "IPVS scheduler",
								Validators: []validator.String{
									stringvalidator.OneOf("rr", "lc", "dh", "sh", "sed", "nq"),
								},
							},
							"sync_period": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Period that IPVS rules are refreshed, in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
							"tcp_fin_timeout": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Timeout value used for IPVS TCP sessions after receiving a FIN in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
							"tcp_timeout": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Timeout value used for idle IPVS TCP sessions in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
							"udp_timeout": schema.StringAttribute{
								CustomType:  ovhtypes.TfRFC3339DurationType{},
								Optional:    true,
								Description: "Timeout value used for IPVS UDP packets in RFC3339 duration format",
								Validators: []validator.String{
									ovhvalidators.RFC3339Duration(),
								},
							},
						},
					},
				},
			},
			"private_network_configuration": schema.SingleNestedBlock{
				Description: "Private network configuration",
				Validators: []validator.Object{
					ovhvalidators.RequireAttributesWhenBlockPresent(
						"default_vrack_gateway",
						"private_network_routing_as_default",
					),
				},
				Attributes: map[string]schema.Attribute{
					"default_vrack_gateway": schema.StringAttribute{
						CustomType:  ovhtypes.TfStringType{},
						Optional:    true,
						Description: "If defined, all egress traffic will be routed towards this IP address, which should belong to the private network. Empty string means disabled.",
					},
					"private_network_routing_as_default": schema.BoolAttribute{
						CustomType:  ovhtypes.TfBoolType{},
						Optional:    true,
						Description: "Defines whether routing should default to using the nodes' private interface, instead of their public interface. Default is false.",
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

// --- CRUD Operations ---

func (r *cloudProjectKubeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data cloudProjectKubeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ServiceName.IsNull() {
		data.ServiceName = ovhtypes.NewTfStringValue(os.Getenv("OVH_CLOUD_PROJECT_SERVICE"))
	}

	serviceName := data.ServiceName.ValueString()
	params := createOptsFromModel(&data)
	res := &CloudProjectKubeResponse{}

	log.Printf("[DEBUG] Will create kube: %s", params)
	endpoint := fmt.Sprintf("/cloud/project/%s/kube", url.PathEscape(serviceName))
	if err := r.config.OVHClient.PostWithContext(ctx, endpoint, params, res); err != nil {
		resp.Diagnostics.AddError(
			"Failed to create kube cluster",
			fmt.Sprintf("error calling POST %s with params %s: %s", endpoint, params, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for kube %s to be available", res.Id)
	kubeEndpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", url.PathEscape(serviceName), url.PathEscape(res.Id))
	if err := helpers.WaitAvailable(r.config.OVHClient, kubeEndpoint, 15*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Failed waiting for kube to be available",
			err.Error(),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for kube %s to be READY", res.Id)
	if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, res.Id, 15*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for kube to be READY",
			fmt.Sprintf("timeout while waiting kube %s to be READY: %s", res.Id, err),
		)
		return
	}
	log.Printf("[DEBUG] kube %s is READY", res.Id)

	data.ServiceName = ovhtypes.NewTfStringValue(serviceName)

	// Read back the resource
	if err := r.readKube(ctx, serviceName, res.Id, &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube after create",
			err.Error(),
		)
		return
	}

	// Fetch kubeconfig
	if err := setKubeconfigOnModel(r.config, serviceName, data.ID.ValueString(), &data); err != nil {
		resp.Diagnostics.AddError(
			"Failed to fetch kubeconfig",
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data cloudProjectKubeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.ID.ValueString()

	log.Printf("[DEBUG] Will read kube %s from project: %s", kubeId, serviceName)

	if err := r.readKube(ctx, serviceName, kubeId, &data); err != nil {
		helpers.CheckDeletedWithContext(ctx, resp, err, fmt.Sprintf("/cloud/project/%s/kube/%s", serviceName, kubeId))
		return
	}

	// Fetch kubeconfig if not yet set
	if data.Kubeconfig.IsNull() || data.Kubeconfig.ValueString() == "" || data.KubeconfigAttributes == nil {
		if err := setKubeconfigOnModel(r.config, serviceName, kubeId, &data); err != nil {
			resp.Diagnostics.AddError(
				"Failed to fetch kubeconfig",
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudProjectKubeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudProjectKubeResourceModel
	var state cloudProjectKubeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := plan.ServiceName.ValueString()
	kubeId := state.ID.ValueString()

	// --- Customization changes ---
	if customizationChanged(&plan, &state) {
		opts := updateCustomizationOptsFromModel(&plan)

		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/customization",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
		)
		if err := r.config.OVHClient.PutWithContext(ctx, endpoint, opts, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update kube customization",
				fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
			)
			return
		}

		log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
		if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, time.Hour); err != nil {
			resp.Diagnostics.AddError(
				"Timeout waiting for kube to be READY",
				fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
			)
			return
		}
		log.Printf("[DEBUG] kube %s is READY", kubeId)
	}

	// --- Version change ---
	if !plan.Version.Equal(state.Version) && !plan.Version.IsNull() && !plan.Version.IsUnknown() {
		oldValue := state.Version.ValueString()
		newValue := plan.Version.ValueString()

		log.Printf("[DEBUG] cluster version change from %s to %s", oldValue, newValue)

		oldVersion, err := version.NewVersion(oldValue)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid version",
				fmt.Sprintf("version %s does not match a semver", oldValue),
			)
			return
		}
		newVersion, err := version.NewVersion(newValue)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid version",
				fmt.Sprintf("version %s does not match a semver", newValue),
			)
			return
		}

		oldSegments := oldVersion.Segments()
		newSegments := newVersion.Segments()

		if oldSegments[0] != 1 || newSegments[0] != 1 {
			resp.Diagnostics.AddError(
				"Unsupported version",
				"the only supported major version is 1",
			)
			return
		}
		if len(oldSegments) < 2 || len(newSegments) < 2 {
			resp.Diagnostics.AddError(
				"Invalid version format",
				"the version should only specify the major and minor versions (e.g. \"1.20\")",
			)
			return
		}
		if newVersion.LessThan(oldVersion) {
			resp.Diagnostics.AddError(
				"Version downgrade not allowed",
				fmt.Sprintf("cannot downgrade cluster from %s to %s", oldValue, newValue),
			)
			return
		}
		if oldSegments[1]+1 != newSegments[1] {
			resp.Diagnostics.AddError(
				"Version upgrade not allowed",
				fmt.Sprintf("cannot upgrade cluster from %s to %s, only next minor version is authorized", oldValue, newValue),
			)
			return
		}

		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/update",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
		)
		if err := r.config.OVHClient.PostWithContext(ctx, endpoint, CloudProjectKubeUpdateOpts{
			Strategy: "NEXT_MINOR",
		}, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update kube version",
				fmt.Sprintf("error calling POST %s: %s", endpoint, err),
			)
			return
		}

		log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
		if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, time.Hour); err != nil {
			resp.Diagnostics.AddError(
				"Timeout waiting for kube to be READY",
				fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
			)
			return
		}
		log.Printf("[DEBUG] kube %s is READY", kubeId)
	}

	// --- Plan change ---
	if !plan.Plan.Equal(state.Plan) && !plan.Plan.IsNull() && !plan.Plan.IsUnknown() {
		oldPlan := state.Plan.ValueString()
		newPlan := plan.Plan.ValueString()
		if oldPlan == "standard" {
			resp.Diagnostics.AddError(
				"Plan change not allowed",
				fmt.Sprintf("you cannot migrate from %s to %s", oldPlan, newPlan),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Plan change not available",
			fmt.Sprintf("migrate from %s to %s is not available yet", oldPlan, newPlan),
		)
		return
	}

	// --- Update policy change ---
	if !plan.UpdatePolicy.Equal(state.UpdatePolicy) && !plan.UpdatePolicy.IsNull() && !plan.UpdatePolicy.IsUnknown() {
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/updatePolicy",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
		)
		if err := r.config.OVHClient.PutWithContext(ctx, endpoint, CloudProjectKubeUpdatePolicyOpts{
			UpdatePolicy: plan.UpdatePolicy.ValueString(),
		}, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update kube update policy",
				fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
			)
			return
		}
	}

	// --- Load balancers subnet ID change ---
	if !plan.LoadBalancersSubnetId.Equal(state.LoadBalancersSubnetId) && !plan.LoadBalancersSubnetId.IsNull() && !plan.LoadBalancersSubnetId.IsUnknown() {
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/updateLoadBalancersSubnetId",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
		)
		if err := r.config.OVHClient.PutWithContext(ctx, endpoint, CloudProjectKubeUpdateLoadBalancersSubnetIdOpts{
			LoadBalancersSubnetId: plan.LoadBalancersSubnetId.ValueString(),
		}, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update load balancers subnet ID",
				fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
			)
			return
		}

		if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, time.Hour); err != nil {
			resp.Diagnostics.AddError(
				"Timeout waiting for kube to be READY",
				fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
			)
			return
		}
	}

	// --- Name change ---
	if !plan.Name.Equal(state.Name) && !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		name := plan.Name.ValueString()
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s",
			url.PathEscape(serviceName),
			url.PathEscape(kubeId),
		)
		if err := r.config.OVHClient.PutWithContext(ctx, endpoint, CloudProjectKubePutOpts{
			Name: &name,
		}, nil); err != nil {
			resp.Diagnostics.AddError(
				"Failed to update kube name",
				fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
			)
			return
		}
	}

	// --- Private network configuration change ---
	if privateNetworkConfigChanged(&plan, &state) {
		if plan.PrivateNetworkConfiguration != nil {
			endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s/privateNetworkConfiguration",
				url.PathEscape(serviceName),
				url.PathEscape(kubeId),
			)
			if err := r.config.OVHClient.PutWithContext(ctx, endpoint, CloudProjectKubeUpdatePNCOpts{
				DefaultVrackGateway:            plan.PrivateNetworkConfiguration.DefaultVrackGateway.ValueString(),
				PrivateNetworkRoutingAsDefault: plan.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault.ValueBool(),
			}, nil); err != nil {
				resp.Diagnostics.AddError(
					"Failed to update private network configuration",
					fmt.Sprintf("error calling PUT %s: %s", endpoint, err),
				)
				return
			}

			log.Printf("[DEBUG] Waiting for kube %s to be READY", kubeId)
			if err := waitForCloudProjectKubeReadyCtx(ctx, r.config.OVHClient, serviceName, kubeId, time.Hour); err != nil {
				resp.Diagnostics.AddError(
					"Timeout waiting for kube to be READY",
					fmt.Sprintf("timeout while waiting kube %s to be READY: %s", kubeId, err),
				)
				return
			}
			log.Printf("[DEBUG] kube %s is READY", kubeId)
		}
	}

	// Read back the resource
	if err := r.readKube(ctx, serviceName, kubeId, &plan); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read kube after update",
			err.Error(),
		)
		return
	}

	// Preserve kubeconfig from state (it doesn't change on update)
	plan.Kubeconfig = state.Kubeconfig
	plan.KubeconfigAttributes = state.KubeconfigAttributes

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudProjectKubeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data cloudProjectKubeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceName := data.ServiceName.ValueString()
	kubeId := data.ID.ValueString()

	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	log.Printf("[DEBUG] Will delete kube %s from project: %s", kubeId, serviceName)
	if err := r.config.OVHClient.DeleteWithContext(ctx, endpoint, nil); err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete kube",
			fmt.Sprintf("error calling DELETE %s: %s", endpoint, err),
		)
		return
	}

	log.Printf("[DEBUG] Waiting for kube %s to be DELETED", kubeId)
	if err := waitForCloudProjectKubeDeletedCtx(ctx, r.config.OVHClient, serviceName, kubeId, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError(
			"Timeout waiting for kube to be DELETED",
			fmt.Sprintf("timeout while waiting kube %s to be DELETED: %s", kubeId, err),
		)
		return
	}
	log.Printf("[DEBUG] kube %s is DELETED", kubeId)
}

func (r *cloudProjectKubeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), kubeId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_name"), serviceName)...)
}

// --- Helpers ---

// readKube reads the kube cluster from the API and populates the model.
func (r *cloudProjectKubeResource) readKube(ctx context.Context, serviceName, kubeId string, data *cloudProjectKubeResourceModel) error {
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s",
		url.PathEscape(serviceName),
		url.PathEscape(kubeId),
	)

	res := &CloudProjectKubeResponse{}

	if err := r.config.OVHClient.GetWithContext(ctx, endpoint, res); err != nil {
		return fmt.Errorf("error calling GET %s: %w", endpoint, err)
	}

	diags := modelFromResponse(ctx, res, data)
	if diags.HasError() {
		return fmt.Errorf("error mapping response to model: %v", diags.Errors())
	}

	log.Printf("[DEBUG] Read kube %+v", res)
	return nil
}

// customizationChanged checks if any customization attributes changed between plan and state.
func customizationChanged(plan, state *cloudProjectKubeResourceModel) bool {
	// Check customization_apiserver
	if (plan.CustomizationApiServer == nil) != (state.CustomizationApiServer == nil) {
		return true
	}
	if plan.CustomizationApiServer != nil && state.CustomizationApiServer != nil {
		if (plan.CustomizationApiServer.AdmissionPlugins == nil) != (state.CustomizationApiServer.AdmissionPlugins == nil) {
			return true
		}
		if plan.CustomizationApiServer.AdmissionPlugins != nil && state.CustomizationApiServer.AdmissionPlugins != nil {
			if !plan.CustomizationApiServer.AdmissionPlugins.Enabled.Equal(state.CustomizationApiServer.AdmissionPlugins.Enabled) {
				return true
			}
			if !plan.CustomizationApiServer.AdmissionPlugins.Disabled.Equal(state.CustomizationApiServer.AdmissionPlugins.Disabled) {
				return true
			}
		}
	}

	// Check deprecated customization
	if (plan.Customization == nil) != (state.Customization == nil) {
		return true
	}
	if plan.Customization != nil && state.Customization != nil {
		if (plan.Customization.ApiServer == nil) != (state.Customization.ApiServer == nil) {
			return true
		}
		if plan.Customization.ApiServer != nil && state.Customization.ApiServer != nil {
			if (plan.Customization.ApiServer.AdmissionPlugins == nil) != (state.Customization.ApiServer.AdmissionPlugins == nil) {
				return true
			}
			if plan.Customization.ApiServer.AdmissionPlugins != nil && state.Customization.ApiServer.AdmissionPlugins != nil {
				if !plan.Customization.ApiServer.AdmissionPlugins.Enabled.Equal(state.Customization.ApiServer.AdmissionPlugins.Enabled) {
					return true
				}
				if !plan.Customization.ApiServer.AdmissionPlugins.Disabled.Equal(state.Customization.ApiServer.AdmissionPlugins.Disabled) {
					return true
				}
			}
		}
	}

	// Check customization_kube_proxy
	if (plan.CustomizationKubeProxy == nil) != (state.CustomizationKubeProxy == nil) {
		return true
	}
	if plan.CustomizationKubeProxy != nil && state.CustomizationKubeProxy != nil {
		// Check iptables
		if (plan.CustomizationKubeProxy.IPTables == nil) != (state.CustomizationKubeProxy.IPTables == nil) {
			return true
		}
		if plan.CustomizationKubeProxy.IPTables != nil && state.CustomizationKubeProxy.IPTables != nil {
			if !plan.CustomizationKubeProxy.IPTables.MinSyncPeriod.Equal(state.CustomizationKubeProxy.IPTables.MinSyncPeriod) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPTables.SyncPeriod.Equal(state.CustomizationKubeProxy.IPTables.SyncPeriod) {
				return true
			}
		}
		// Check ipvs
		if (plan.CustomizationKubeProxy.IPVS == nil) != (state.CustomizationKubeProxy.IPVS == nil) {
			return true
		}
		if plan.CustomizationKubeProxy.IPVS != nil && state.CustomizationKubeProxy.IPVS != nil {
			if !plan.CustomizationKubeProxy.IPVS.MinSyncPeriod.Equal(state.CustomizationKubeProxy.IPVS.MinSyncPeriod) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPVS.Scheduler.Equal(state.CustomizationKubeProxy.IPVS.Scheduler) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPVS.SyncPeriod.Equal(state.CustomizationKubeProxy.IPVS.SyncPeriod) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPVS.TCPFinTimeout.Equal(state.CustomizationKubeProxy.IPVS.TCPFinTimeout) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPVS.TCPTimeout.Equal(state.CustomizationKubeProxy.IPVS.TCPTimeout) {
				return true
			}
			if !plan.CustomizationKubeProxy.IPVS.UDPTimeout.Equal(state.CustomizationKubeProxy.IPVS.UDPTimeout) {
				return true
			}
		}
	}

	return false
}

// privateNetworkConfigChanged checks if private network config changed.
func privateNetworkConfigChanged(plan, state *cloudProjectKubeResourceModel) bool {
	if (plan.PrivateNetworkConfiguration == nil) != (state.PrivateNetworkConfiguration == nil) {
		return true
	}
	if plan.PrivateNetworkConfiguration != nil && state.PrivateNetworkConfiguration != nil {
		if !plan.PrivateNetworkConfiguration.DefaultVrackGateway.Equal(state.PrivateNetworkConfiguration.DefaultVrackGateway) {
			return true
		}
		if !plan.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault.Equal(state.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault) {
			return true
		}
	}
	return false
}

// cloudProjectKubeExists checks if a kube cluster exists.
func cloudProjectKubeExists(serviceName, id string, client *ovhwrap.Client) error {
	res := &CloudProjectKubeResponse{}
	endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", serviceName, id)
	return client.Get(endpoint, res)
}

// waitForCloudProjectKubeDeletedCtx waits for a kube cluster to be fully deleted.
func waitForCloudProjectKubeDeletedCtx(ctx context.Context, client *ovhwrap.Client, serviceName, kubeId string, timeout time.Duration) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		res := &CloudProjectKubeResponse{}
		endpoint := fmt.Sprintf("/cloud/project/%s/kube/%s", serviceName, kubeId)
		err := client.GetWithContext(ctx, endpoint, res)
		if err != nil {
			if errOvh, ok := err.(*ovh.APIError); ok && errOvh.Code == 404 {
				return nil
			}
			return retry.NonRetryableError(fmt.Errorf("error reading kube %s: %w", kubeId, err))
		}
		if res.Status == "DELETED" {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("kube %s is in state %s, waiting for DELETED", kubeId, res.Status))
	})
}
