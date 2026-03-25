package ovh

import (
	"fmt"
)

// Constants for kube cluster attribute keys (used in tests).
const (
	kubeClusterLoadBalancersSubnetIdKey       = "load_balancers_subnet_id"
	kubeClusterNodesSubnetIdKey               = "nodes_subnet_id"
	kubeClusterNameKey                        = "name"
	kubeClusterPrivateNetworkIDKey            = "private_network_id"
	kubeClusterPrivateNetworkConfigurationKey = "private_network_configuration"
	kubeClusterUpdatePolicyKey                = "update_policy"
	kubeClusterVersionKey                     = "version"
	kubeClusterPlanKey                        = "plan"
	kubeClusterProxyModeKey                   = "kube_proxy_mode"
	kubeClusterCustomization                  = "customization" // Deprecated
	kubeClusterCustomizationApiServerKey      = "customization_apiserver"
	kubeClusterCustomizationKubeProxyKey      = "customization_kube_proxy"
)

type CloudProjectKubeUpdatePolicyOpts struct {
	UpdatePolicy string `json:"updatePolicy"`
}

// CloudProjectKubePutOpts update cluster options
type CloudProjectKubePutOpts struct {
	Name *string `json:"name,omitempty"`
}

type privateNetworkConfiguration struct {
	DefaultVrackGateway            string `json:"defaultVrackGateway"`
	PrivateNetworkRoutingAsDefault bool   `json:"privateNetworkRoutingAsDefault"`
}

type CloudProjectKubeUpdateLoadBalancersSubnetIdOpts struct {
	LoadBalancersSubnetId string `json:"loadBalancersSubnetId"`
}

type CloudProjectKubeCreateOpts struct {
	Name                        *string                      `json:"name,omitempty"`
	PrivateNetworkId            *string                      `json:"privateNetworkId,omitempty"`
	PrivateNetworkConfiguration *privateNetworkConfiguration `json:"privateNetworkConfiguration,omitempty"`
	Region                      string                       `json:"region"`
	Version                     *string                      `json:"version,omitempty"`
	Plan                        *string                      `json:"plan,omitempty"`
	UpdatePolicy                *string                      `json:"updatePolicy,omitempty"`
	Customization               *Customization               `json:"customization,omitempty"`
	KubeProxyMode               *string                      `json:"kubeProxyMode,omitempty"`
	LoadBalancersSubnetId       *string                      `json:"loadBalancersSubnetId,omitempty"`
	NodesSubnetId               *string                      `json:"nodesSubnetId,omitempty"`
}

type Customization struct {
	APIServer *APIServer              `json:"apiServer,omitempty"`
	KubeProxy *kubeProxyCustomization `json:"kubeProxy,omitempty"`
}

type APIServer struct {
	AdmissionPlugins *AdmissionPlugins `json:"admissionPlugins,omitempty"`
}

type kubeProxyCustomization struct {
	IPTables *kubeProxyCustomizationIPTables `json:"iptables,omitempty"`
	IPVS     *kubeProxyCustomizationIPVS     `json:"ipvs,omitempty"`
}

type kubeProxyCustomizationIPTables struct {
	MinSyncPeriod *string `json:"minSyncPeriod,omitempty"`
	SyncPeriod    *string `json:"syncPeriod,omitempty"`
}

type kubeProxyCustomizationIPVS struct {
	MinSyncPeriod *string `json:"minSyncPeriod,omitempty"`
	Scheduler     *string `json:"scheduler,omitempty"`
	SyncPeriod    *string `json:"syncPeriod,omitempty"`
	TCPFinTimeout *string `json:"tcpFinTimeout,omitempty"`
	TCPTimeout    *string `json:"tcpTimeout,omitempty"`
	UDPTimeout    *string `json:"udpTimeout,omitempty"`
}

type AdmissionPlugins struct {
	Enabled  *[]string `json:"enabled,omitempty"`
	Disabled *[]string `json:"disabled,omitempty"`
}

func (opts *CloudProjectKubeCreateOpts) String() string {
	var str string
	if opts.Name != nil {
		str = *opts.Name
	}

	str += fmt.Sprintf(" (%s)", opts.Region)

	if opts.Version != nil {
		str += fmt.Sprintf(": %s", *opts.Version)
	}

	return str
}

type CloudProjectKubeResponse struct {
	ControlPlaneIsUpToDate bool          `json:"controlPlaneIsUpToDate"`
	Id                     string        `json:"id"`
	IsUpToDate             bool          `json:"isUpToDate"`
	LoadBalancersSubnetId  string        `json:"loadBalancersSubnetId"`
	Name                   string        `json:"name"`
	NextUpgradeVersions    []string      `json:"nextUpgradeVersions"`
	NodesSubnetId          string        `json:"nodesSubnetId"`
	NodesUrl               string        `json:"nodesUrl"`
	PrivateNetworkId       string        `json:"privateNetworkId"`
	Region                 string        `json:"region"`
	Status                 string        `json:"status"`
	UpdatePolicy           string        `json:"updatePolicy"`
	Url                    string        `json:"url"`
	Version                string        `json:"version"`
	Plan                   string        `json:"plan"`
	Customization          Customization `json:"customization"`
	KubeProxyMode          string        `json:"kubeProxyMode"`
}

func (v *CloudProjectKubeResponse) String() string {
	return fmt.Sprintf("%s(%s): %s", v.Name, v.Id, v.Status)
}

type CloudProjectKubeKubeConfigResponse struct {
	Content string `json:"content"`
}

type CloudProjectKubeUpdateOpts struct {
	Strategy string `json:"strategy"`
}

type CloudProjectKubeResetOpts struct {
	PrivateNetworkId *string `json:"privateNetworkId,omitempty"`
}

type CloudProjectKubeUpdatePNCOpts struct {
	DefaultVrackGateway            string `json:"defaultVrackGateway"`
	PrivateNetworkRoutingAsDefault bool   `json:"privateNetworkRoutingAsDefault"`
}

type CloudProjectKubeUpdateCustomizationOpts struct {
	APIServer *APIServer              `json:"apiServer,omitempty"`
	KubeProxy *kubeProxyCustomization `json:"kubeProxy,omitempty"`
}

type CloudProjectKubeNodeResponse struct {
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	DeployedAt string `json:"deployedAt"`
	Flavor     string `json:"flavor"`
	Id         string `json:"id"`
	InstanceId string `json:"instanceId"`
	IsUpToDate bool   `json:"isUpToDate"`
	Name       string `json:"name"`
	NodePoolId string `json:"nodePoolId"`
	ProjectId  string `json:"projectId"`
	Status     string `json:"status"`
	Version    string `json:"version"`
}

func (v CloudProjectKubeNodeResponse) ToMap() map[string]interface{} {
	obj := make(map[string]interface{})
	obj["created_at"] = v.CreatedAt
	obj["deployed_at"] = v.DeployedAt
	obj["flavor"] = v.Flavor
	obj["id"] = v.Id
	obj["instance_id"] = v.InstanceId
	obj["is_up_to_date"] = v.IsUpToDate
	obj["name"] = v.Name
	obj["node_pool_id"] = v.NodePoolId
	obj["project_id"] = v.ProjectId
	obj["status"] = v.Status
	obj["updated_at"] = v.UpdatedAt
	obj["version"] = v.Version
	return obj
}
