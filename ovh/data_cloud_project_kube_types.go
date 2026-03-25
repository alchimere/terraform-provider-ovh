package ovh

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	ovhtypes "github.com/ovh/terraform-provider-ovh/v2/ovh/types"
)

// --- Data Source Model ---

type cloudProjectKubeDataSourceModel struct {
	ServiceName            ovhtypes.TfStringValue                             `tfsdk:"service_name"`
	KubeId                 ovhtypes.TfStringValue                             `tfsdk:"kube_id"`
	Name                   ovhtypes.TfStringValue                             `tfsdk:"name"`
	Version                ovhtypes.TfStringValue                             `tfsdk:"version"`
	Plan                   ovhtypes.TfStringValue                             `tfsdk:"plan"`
	Region                 ovhtypes.TfStringValue                             `tfsdk:"region"`
	KubeProxyMode          ovhtypes.TfStringValue                             `tfsdk:"kube_proxy_mode"`
	PrivateNetworkId       ovhtypes.TfStringValue                             `tfsdk:"private_network_id"`
	LoadBalancersSubnetId  ovhtypes.TfStringValue                             `tfsdk:"load_balancers_subnet_id"`
	NodesSubnetId          ovhtypes.TfStringValue                             `tfsdk:"nodes_subnet_id"`
	UpdatePolicy           ovhtypes.TfStringValue                             `tfsdk:"update_policy"`
	CustomizationApiServer *kubeCustomizationApiServerModel                   `tfsdk:"customization_apiserver"`
	Customization          *kubeCustomizationDeprecatedModel                  `tfsdk:"customization"`
	CustomizationKubeProxy *kubeCustomizationKubeProxyModel                   `tfsdk:"customization_kube_proxy"`
	ControlPlaneIsUpToDate ovhtypes.TfBoolValue                               `tfsdk:"control_plane_is_up_to_date"`
	IsUpToDate             ovhtypes.TfBoolValue                               `tfsdk:"is_up_to_date"`
	NextUpgradeVersions    ovhtypes.TfListNestedValue[ovhtypes.TfStringValue] `tfsdk:"next_upgrade_versions"`
	NodesUrl               ovhtypes.TfStringValue                             `tfsdk:"nodes_url"`
	Status                 ovhtypes.TfStringValue                             `tfsdk:"status"`
	Url                    ovhtypes.TfStringValue                             `tfsdk:"url"`
	Kubeconfig             ovhtypes.TfStringValue                             `tfsdk:"kubeconfig"`
	KubeconfigAttributes   []kubeKubeconfigAttributesModel                    `tfsdk:"kubeconfig_attributes"`
}

// dataSourceModelFromResponse populates the data source model from the API response.
func dataSourceModelFromResponse(ctx context.Context, res *CloudProjectKubeResponse, data *cloudProjectKubeDataSourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.Name = ovhtypes.NewTfStringValue(res.Name)
	data.ControlPlaneIsUpToDate = ovhtypes.NewTfBoolValue(res.ControlPlaneIsUpToDate)
	data.IsUpToDate = ovhtypes.NewTfBoolValue(res.IsUpToDate)
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

	// Version: strip patch version
	versionStr := res.Version
	versionPatch, err := version.NewVersion(versionStr)
	if err != nil {
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

		if data.Customization != nil && data.Customization.ApiServer != nil {
			data.Customization.ApiServer.AdmissionPlugins = admissionModel
		} else {
			if data.CustomizationApiServer == nil {
				data.CustomizationApiServer = &kubeCustomizationApiServerModel{}
			}
			data.CustomizationApiServer.AdmissionPlugins = admissionModel
		}
	}

	// Kube proxy customization
	if res.Customization.KubeProxy != nil {
		if data.CustomizationKubeProxy == nil {
			data.CustomizationKubeProxy = &kubeCustomizationKubeProxyModel{}
		}

		if res.Customization.KubeProxy.IPTables != nil {
			if data.CustomizationKubeProxy.IPTables == nil {
				data.CustomizationKubeProxy.IPTables = &kubeProxyIPTablesModel{}
			}
			data.CustomizationKubeProxy.IPTables.MinSyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPTables.MinSyncPeriod)
			data.CustomizationKubeProxy.IPTables.SyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPTables.SyncPeriod)
		}

		if res.Customization.KubeProxy.IPVS != nil {
			if data.CustomizationKubeProxy.IPVS == nil {
				data.CustomizationKubeProxy.IPVS = &kubeProxyIPVSModel{}
			}
			data.CustomizationKubeProxy.IPVS.MinSyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.MinSyncPeriod)
			data.CustomizationKubeProxy.IPVS.Scheduler = optionalStringToTfValue(res.Customization.KubeProxy.IPVS.Scheduler)
			data.CustomizationKubeProxy.IPVS.SyncPeriod = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.SyncPeriod)
			data.CustomizationKubeProxy.IPVS.TCPFinTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.TCPFinTimeout)
			data.CustomizationKubeProxy.IPVS.TCPTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.TCPTimeout)
			data.CustomizationKubeProxy.IPVS.UDPTimeout = optionalDurationToTfValue(res.Customization.KubeProxy.IPVS.UDPTimeout)
		}

		// If both iptables and ipvs are nil, remove the block
		if data.CustomizationKubeProxy.IPTables == nil && data.CustomizationKubeProxy.IPVS == nil {
			data.CustomizationKubeProxy = nil
		}
	}

	return diags
}

// setKubeconfigOnDataSourceModel fetches the kubeconfig and populates the data source model.
func setKubeconfigOnDataSourceModel(config *Config, serviceName, kubeId string, data *cloudProjectKubeDataSourceModel) error {
	kubeConfig, err := getKubeconfig(config, serviceName, kubeId)
	if err != nil {
		return err
	}

	if len(kubeConfig.Clusters) == 0 || len(kubeConfig.Users) == 0 {
		return fmt.Errorf("kubeconfig is invalid")
	}

	data.Kubeconfig = ovhtypes.NewTfStringValue(*kubeConfig.Raw)
	data.KubeconfigAttributes = []kubeKubeconfigAttributesModel{{
		Host:                 ovhtypes.NewTfStringValue(kubeConfig.Clusters[0].Cluster.Server),
		ClusterCACertificate: ovhtypes.NewTfStringValue(kubeConfig.Clusters[0].Cluster.CertificateAuthorityData),
		ClientCertificate:    ovhtypes.NewTfStringValue(kubeConfig.Users[0].User.ClientCertificateData),
		ClientKey:            ovhtypes.NewTfStringValue(kubeConfig.Users[0].User.ClientKeyData),
	}}

	return nil
}
