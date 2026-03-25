package ovh

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Resource Model ---

type cloudProjectKubeResourceModel struct {
	ID                          ovhtypes.TfStringValue                `tfsdk:"id"`
	ServiceName                 ovhtypes.TfStringValue                `tfsdk:"service_name"`
	Name                        ovhtypes.TfStringValue                `tfsdk:"name"`
	Version                     ovhtypes.TfStringValue                `tfsdk:"version"`
	Plan                        ovhtypes.TfStringValue                `tfsdk:"plan"`
	Region                      ovhtypes.TfStringValue                `tfsdk:"region"`
	KubeProxyMode               ovhtypes.TfStringValue                `tfsdk:"kube_proxy_mode"`
	PrivateNetworkId            ovhtypes.TfStringValue                `tfsdk:"private_network_id"`
	PrivateNetworkConfiguration *kubePrivateNetworkConfigurationModel `tfsdk:"private_network_configuration"`
	LoadBalancersSubnetId       ovhtypes.TfStringValue                `tfsdk:"load_balancers_subnet_id"`
	NodesSubnetId               ovhtypes.TfStringValue                `tfsdk:"nodes_subnet_id"`
	UpdatePolicy                ovhtypes.TfStringValue                `tfsdk:"update_policy"`
	CustomizationApiServer      *kubeCustomizationApiServerModel      `tfsdk:"customization_apiserver"`
	Customization               *kubeCustomizationDeprecatedModel     `tfsdk:"customization"`
	CustomizationKubeProxy      *kubeCustomizationKubeProxyModel      `tfsdk:"customization_kube_proxy"`

	// Computed
	ControlPlaneIsUpToDate ovhtypes.TfBoolValue                               `tfsdk:"control_plane_is_up_to_date"`
	IsUpToDate             ovhtypes.TfBoolValue                               `tfsdk:"is_up_to_date"`
	NextUpgradeVersions    ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"next_upgrade_versions"`
	NodesUrl               ovhtypes.TfStringValue                             `tfsdk:"nodes_url"`
	Status                 ovhtypes.TfStringValue                             `tfsdk:"status"`
	Url                    ovhtypes.TfStringValue                             `tfsdk:"url"`
	Kubeconfig             ovhtypes.TfStringValue                             `tfsdk:"kubeconfig"`
	KubeconfigAttributes   *kubeKubeconfigAttributesModel                     `tfsdk:"kubeconfig_attributes"`
}

// --- Nested Models ---

type kubeAdmissionPluginsModel struct {
	Enabled  ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"enabled"`
	Disabled ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"disabled"`
}

type kubeCustomizationApiServerModel struct {
	AdmissionPlugins *kubeAdmissionPluginsModel `tfsdk:"admissionplugins"`
}

// Deprecated model - wraps apiserver inside customization
type kubeCustomizationDeprecatedApiServerModel struct {
	AdmissionPlugins *kubeAdmissionPluginsModel `tfsdk:"admissionplugins"`
}

type kubeCustomizationDeprecatedModel struct {
	ApiServer *kubeCustomizationDeprecatedApiServerModel `tfsdk:"apiserver"`
}

type kubeProxyIPTablesModel struct {
	MinSyncPeriod ovhtypes.TfRFC3339DurationValue `tfsdk:"min_sync_period"`
	SyncPeriod    ovhtypes.TfRFC3339DurationValue `tfsdk:"sync_period"`
}

type kubeProxyIPVSModel struct {
	MinSyncPeriod ovhtypes.TfRFC3339DurationValue `tfsdk:"min_sync_period"`
	Scheduler     ovhtypes.TfStringValue          `tfsdk:"scheduler"`
	SyncPeriod    ovhtypes.TfRFC3339DurationValue `tfsdk:"sync_period"`
	TCPFinTimeout ovhtypes.TfRFC3339DurationValue `tfsdk:"tcp_fin_timeout"`
	TCPTimeout    ovhtypes.TfRFC3339DurationValue `tfsdk:"tcp_timeout"`
	UDPTimeout    ovhtypes.TfRFC3339DurationValue `tfsdk:"udp_timeout"`
}

type kubeCustomizationKubeProxyModel struct {
	IPTables *kubeProxyIPTablesModel `tfsdk:"iptables"`
	IPVS     *kubeProxyIPVSModel     `tfsdk:"ipvs"`
}

type kubePrivateNetworkConfigurationModel struct {
	DefaultVrackGateway            ovhtypes.TfStringValue `tfsdk:"default_vrack_gateway"`
	PrivateNetworkRoutingAsDefault ovhtypes.TfBoolValue   `tfsdk:"private_network_routing_as_default"`
}

type kubeKubeconfigAttributesModel struct {
	Host                 ovhtypes.TfStringValue `tfsdk:"host"`
	ClusterCACertificate ovhtypes.TfStringValue `tfsdk:"cluster_ca_certificate"`
	ClientCertificate    ovhtypes.TfStringValue `tfsdk:"client_certificate"`
	ClientKey            ovhtypes.TfStringValue `tfsdk:"client_key"`
}

// --- Conversion Helpers ---

// createOptsFromModel builds CloudProjectKubeCreateOpts from the framework resource model.
func createOptsFromModel(data *cloudProjectKubeResourceModel) *CloudProjectKubeCreateOpts {
	opts := &CloudProjectKubeCreateOpts{
		Region: data.Region.ValueString(),
	}

	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		v := data.Version.ValueString()
		opts.Version = &v
	}
	if !data.Plan.IsNull() && !data.Plan.IsUnknown() {
		v := data.Plan.ValueString()
		opts.Plan = &v
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		v := data.Name.ValueString()
		opts.Name = &v
	}
	if !data.UpdatePolicy.IsNull() && !data.UpdatePolicy.IsUnknown() {
		v := data.UpdatePolicy.ValueString()
		opts.UpdatePolicy = &v
	}
	if !data.LoadBalancersSubnetId.IsNull() && !data.LoadBalancersSubnetId.IsUnknown() {
		v := data.LoadBalancersSubnetId.ValueString()
		opts.LoadBalancersSubnetId = &v
	}
	if !data.NodesSubnetId.IsNull() && !data.NodesSubnetId.IsUnknown() {
		v := data.NodesSubnetId.ValueString()
		opts.NodesSubnetId = &v
	}
	if !data.PrivateNetworkId.IsNull() && !data.PrivateNetworkId.IsUnknown() {
		v := data.PrivateNetworkId.ValueString()
		opts.PrivateNetworkId = &v
	}
	if !data.KubeProxyMode.IsNull() && !data.KubeProxyMode.IsUnknown() {
		v := data.KubeProxyMode.ValueString()
		opts.KubeProxyMode = &v
	}

	// Private network configuration
	if data.PrivateNetworkConfiguration != nil {
		opts.PrivateNetworkConfiguration = &privateNetworkConfiguration{
			DefaultVrackGateway:            data.PrivateNetworkConfiguration.DefaultVrackGateway.ValueString(),
			PrivateNetworkRoutingAsDefault: data.PrivateNetworkConfiguration.PrivateNetworkRoutingAsDefault.ValueBool(),
		}
	}

	// Customization
	customization := &Customization{}

	// Kube proxy customization
	customization.KubeProxy = kubeProxyCustomizationFromModel(data.CustomizationKubeProxy)

	// API server customization - check deprecated vs new
	if data.Customization != nil && data.Customization.ApiServer != nil && data.Customization.ApiServer.AdmissionPlugins != nil {
		customization.APIServer = apiServerFromAdmissionPluginsModel(data.Customization.ApiServer.AdmissionPlugins)
	} else if data.CustomizationApiServer != nil && data.CustomizationApiServer.AdmissionPlugins != nil {
		customization.APIServer = apiServerFromAdmissionPluginsModel(data.CustomizationApiServer.AdmissionPlugins)
	}

	if customization.APIServer != nil || customization.KubeProxy != nil {
		opts.Customization = customization
	}

	return opts
}

// apiServerFromAdmissionPluginsModel converts the framework model to the API struct.
func apiServerFromAdmissionPluginsModel(m *kubeAdmissionPluginsModel) *APIServer {
	if m == nil {
		return nil
	}

	apiServer := &APIServer{
		AdmissionPlugins: &AdmissionPlugins{},
	}

	if !m.Enabled.IsNull() && !m.Enabled.IsUnknown() {
		enabled := stringSliceFromTfList(m.Enabled)
		apiServer.AdmissionPlugins.Enabled = &enabled
	}
	if !m.Disabled.IsNull() && !m.Disabled.IsUnknown() {
		disabled := stringSliceFromTfList(m.Disabled)
		apiServer.AdmissionPlugins.Disabled = &disabled
	}

	return apiServer
}

// kubeProxyCustomizationFromModel converts the framework model to the API struct.
func kubeProxyCustomizationFromModel(m *kubeCustomizationKubeProxyModel) *kubeProxyCustomization {
	if m == nil {
		return nil
	}

	kpc := &kubeProxyCustomization{}

	if m.IPTables != nil {
		kpc.IPTables = &kubeProxyCustomizationIPTables{}
		if !m.IPTables.MinSyncPeriod.IsNull() && !m.IPTables.MinSyncPeriod.IsUnknown() {
			v := m.IPTables.MinSyncPeriod.ValueString()
			kpc.IPTables.MinSyncPeriod = &v
		}
		if !m.IPTables.SyncPeriod.IsNull() && !m.IPTables.SyncPeriod.IsUnknown() {
			v := m.IPTables.SyncPeriod.ValueString()
			kpc.IPTables.SyncPeriod = &v
		}
	}

	if m.IPVS != nil {
		kpc.IPVS = &kubeProxyCustomizationIPVS{}
		if !m.IPVS.MinSyncPeriod.IsNull() && !m.IPVS.MinSyncPeriod.IsUnknown() {
			v := m.IPVS.MinSyncPeriod.ValueString()
			kpc.IPVS.MinSyncPeriod = &v
		}
		if !m.IPVS.Scheduler.IsNull() && !m.IPVS.Scheduler.IsUnknown() {
			v := m.IPVS.Scheduler.ValueString()
			kpc.IPVS.Scheduler = &v
		}
		if !m.IPVS.SyncPeriod.IsNull() && !m.IPVS.SyncPeriod.IsUnknown() {
			v := m.IPVS.SyncPeriod.ValueString()
			kpc.IPVS.SyncPeriod = &v
		}
		if !m.IPVS.TCPFinTimeout.IsNull() && !m.IPVS.TCPFinTimeout.IsUnknown() {
			v := m.IPVS.TCPFinTimeout.ValueString()
			kpc.IPVS.TCPFinTimeout = &v
		}
		if !m.IPVS.TCPTimeout.IsNull() && !m.IPVS.TCPTimeout.IsUnknown() {
			v := m.IPVS.TCPTimeout.ValueString()
			kpc.IPVS.TCPTimeout = &v
		}
		if !m.IPVS.UDPTimeout.IsNull() && !m.IPVS.UDPTimeout.IsUnknown() {
			v := m.IPVS.UDPTimeout.ValueString()
			kpc.IPVS.UDPTimeout = &v
		}
	}

	return kpc
}

// updateCustomizationOptsFromModel builds the customization update opts.
func updateCustomizationOptsFromModel(data *cloudProjectKubeResourceModel) *CloudProjectKubeUpdateCustomizationOpts {
	opts := &CloudProjectKubeUpdateCustomizationOpts{}

	// Kube proxy
	opts.KubeProxy = kubeProxyCustomizationFromModel(data.CustomizationKubeProxy)

	// API server - check deprecated vs new
	if data.Customization != nil && data.Customization.ApiServer != nil && data.Customization.ApiServer.AdmissionPlugins != nil {
		opts.APIServer = apiServerFromAdmissionPluginsModel(data.Customization.ApiServer.AdmissionPlugins)
	} else if data.CustomizationApiServer != nil && data.CustomizationApiServer.AdmissionPlugins != nil {
		opts.APIServer = apiServerFromAdmissionPluginsModel(data.CustomizationApiServer.AdmissionPlugins)
	}

	return opts
}

// modelFromResponse populates the framework resource model from the API response.
func modelFromResponse(ctx context.Context, res *CloudProjectKubeResponse, data *cloudProjectKubeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = ovhtypes.NewTfStringValue(res.Id)
	data.ControlPlaneIsUpToDate = ovhtypes.NewTfBoolValue(res.ControlPlaneIsUpToDate)
	data.IsUpToDate = ovhtypes.NewTfBoolValue(res.IsUpToDate)
	data.Name = ovhtypes.NewTfStringValue(res.Name)
	data.NodesSubnetId = ovhtypes.NewTfStringValue(res.NodesSubnetId)
	data.NodesUrl = ovhtypes.NewTfStringValue(res.NodesUrl)
	data.Region = ovhtypes.NewTfStringValue(res.Region)
	data.Status = ovhtypes.NewTfStringValue(res.Status)
	data.UpdatePolicy = ovhtypes.NewTfStringValue(res.UpdatePolicy)
	data.Url = ovhtypes.NewTfStringValue(res.Url)
	data.Plan = ovhtypes.NewTfStringValue(res.Plan)
	data.KubeProxyMode = ovhtypes.NewTfStringValue(res.KubeProxyMode)

	// For Optional-only fields where the API returns "" when unset,
	// preserve the prior state value (null) to avoid inconsistent plan errors.
	if res.PrivateNetworkId != "" {
		data.PrivateNetworkId = ovhtypes.NewTfStringValue(res.PrivateNetworkId)
	}
	if res.LoadBalancersSubnetId != "" {
		data.LoadBalancersSubnetId = ovhtypes.NewTfStringValue(res.LoadBalancersSubnetId)
	}

	// Version: strip patch version (return only major.minor)
	versionStr := res.Version
	versionPatch, err := version.NewVersion(versionStr)
	if err != nil {
		// Fallback: strip everything after last dot
		if idx := strings.LastIndex(versionStr, "."); idx >= 0 {
			versionStr = versionStr[:idx]
		}
	} else {
		s := versionPatch.String()
		if idx := strings.LastIndex(s, "."); idx >= 0 {
			versionStr = s[:idx]
		}
	}
	data.Version = ovhtypes.NewTfStringValue(versionStr)

	// Next upgrade versions
	data.NextUpgradeVersions = tfListFromStringSlice(ctx, res.NextUpgradeVersions)

	// API server customization
	if res.Customization.APIServer != nil && res.Customization.APIServer.AdmissionPlugins != nil {
		plugins := res.Customization.APIServer.AdmissionPlugins
		admissionModel := &kubeAdmissionPluginsModel{
			Enabled:  tfStringListFromOptionalSlice(ctx, plugins.Enabled),
			Disabled: tfStringListFromOptionalSlice(ctx, plugins.Disabled),
		}

		// Set on the non-deprecated attribute by default
		if data.Customization != nil && data.Customization.ApiServer != nil {
			// User is using deprecated syntax, populate it
			data.Customization.ApiServer.AdmissionPlugins = admissionModel
		} else {
			if data.CustomizationApiServer == nil {
				data.CustomizationApiServer = &kubeCustomizationApiServerModel{}
			}
			data.CustomizationApiServer.AdmissionPlugins = admissionModel
		}
	}

	// Kube proxy customization
	if res.Customization.KubeProxy != nil && data.CustomizationKubeProxy != nil {
		if res.Customization.KubeProxy.IPTables != nil && data.CustomizationKubeProxy.IPTables != nil {
			data.CustomizationKubeProxy.IPTables.MinSyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPTables.MinSyncPeriod)
			data.CustomizationKubeProxy.IPTables.SyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPTables.SyncPeriod)
		}

		if res.Customization.KubeProxy.IPVS != nil && data.CustomizationKubeProxy.IPVS != nil {
			data.CustomizationKubeProxy.IPVS.MinSyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.MinSyncPeriod)
			data.CustomizationKubeProxy.IPVS.Scheduler = optionalStringToTfValue(res.Customization.KubeProxy.IPVS.Scheduler)
			data.CustomizationKubeProxy.IPVS.SyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.SyncPeriod)
			data.CustomizationKubeProxy.IPVS.TCPFinTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.TCPFinTimeout)
			data.CustomizationKubeProxy.IPVS.TCPTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.TCPTimeout)
			data.CustomizationKubeProxy.IPVS.UDPTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.UDPTimeout)
		}
	}

	return diags
}

// setKubeconfigOnModel fetches the kubeconfig and populates the model.
func setKubeconfigOnModel(config *Config, serviceName, kubeId string, data *cloudProjectKubeResourceModel) error {
	kubeConfig, err := getKubeconfig(config, serviceName, kubeId)
	if err != nil {
		return err
	}

	if len(kubeConfig.Clusters) == 0 || len(kubeConfig.Users) == 0 {
		return fmt.Errorf("kubeconfig is invalid")
	}

	data.Kubeconfig = ovhtypes.NewTfStringValue(*kubeConfig.Raw)
	data.KubeconfigAttributes = &kubeKubeconfigAttributesModel{
		Host:                 ovhtypes.NewTfStringValue(kubeConfig.Clusters[0].Cluster.Server),
		ClusterCACertificate: ovhtypes.NewTfStringValue(kubeConfig.Clusters[0].Cluster.CertificateAuthorityData),
		ClientCertificate:    ovhtypes.NewTfStringValue(kubeConfig.Users[0].User.ClientCertificateData),
		ClientKey:            ovhtypes.NewTfStringValue(kubeConfig.Users[0].User.ClientKeyData),
	}

	return nil
}

// --- Helper functions ---

// optionalStringToTfValue converts a *string to a TfStringValue.
func optionalStringToTfValue(s *string) ovhtypes.TfStringValue {
	if s == nil {
		return ovhtypes.TfStringValue{StringValue: basetypes.NewStringValue("")}
	}
	return ovhtypes.NewTfStringValue(*s)
}

// optionalDurationToTfValue converts a *string to a TfRFC3339DurationValue.
func optionalDurationToTfValue(s *string) ovhtypes.TfRFC3339DurationValue {
	if s == nil {
		return ovhtypes.TfRFC3339DurationValue{StringValue: basetypes.NewStringValue("")}
	}
	return ovhtypes.NewTfRFC3339DurationValue(*s)
}

// tfStringListFromOptionalSlice converts a *[]string to TfListNestedValue.
func tfStringListFromOptionalSlice(ctx context.Context, slice *[]string) ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] {
	if slice == nil {
		emptySlice := make([]string, 0)
		return tfListFromStringSlice(ctx, emptySlice)
	}
	return tfListFromStringSlice(ctx, *slice)
}

// isUsingDeprecatedCustomizationSyntax checks if the user's config uses the deprecated customization block.
func isUsingDeprecatedCustomizationSyntax(data *cloudProjectKubeResourceModel) bool {
	return data.Customization != nil && data.Customization.ApiServer != nil
}
