package ovh

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCloudProjectKubeNodePoolDataSource_basic(t *testing.T) {
	name := acctest.RandomWithPrefix(test_prefix)
	region := os.Getenv("OVH_CLOUD_PROJECT_KUBE_REGION_TEST")

	config := fmt.Sprintf(
		testAccCloudProjectKubeNodePoolDataSourceConfig,
		os.Getenv("OVH_CLOUD_PROJECT_SERVICE_TEST"),
		name,
		region,
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckKubernetes(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: true, // expect resource to be read after creation
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "max_nodes", "2"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_unneeded_time_seconds", "222"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_unready_time_seconds", "2222"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_utilization_threshold", "0.2"),
				),
			},
		},
	})
}

var testAccCloudProjectKubeNodePoolDataSourceConfig = `
resource "ovh_cloud_project_kube" "cluster" {
	service_name  = "%s"
	name          = "%s"
	region        = "%s"
}

resource "ovh_cloud_project_kube_nodepool" "pool" {
	service_name  = ovh_cloud_project_kube.cluster.service_name
	kube_id       = ovh_cloud_project_kube.cluster.id
	name          = ovh_cloud_project_kube.cluster.name
	flavor_name   = "b3-8"
	desired_nodes = 1
	min_nodes     = 0
	max_nodes     = 2
	autoscaling_scale_down_unneeded_time_seconds = 222
	autoscaling_scale_down_unready_time_seconds = 2222
	autoscaling_scale_down_utilization_threshold = 0.2

	depends_on = [
		ovh_cloud_project_kube.cluster
	]
}

data "ovh_cloud_project_kube_nodepool" "poolDataSource" {
  service_name  = ovh_cloud_project_kube.cluster.service_name
  kube_id       = ovh_cloud_project_kube.cluster.id
  name          = ovh_cloud_project_kube_nodepool.pool.name

  depends_on = [
    ovh_cloud_project_kube_nodepool.pool
  ]
}
`

func TestAccCloudProjectKubeNodePoolDataSource_full(t *testing.T) {
	name := acctest.RandomWithPrefix(test_prefix)
	region := os.Getenv("OVH_CLOUD_PROJECT_KUBE_REGION_TEST")

	config := fmt.Sprintf(
		testAccCloudProjectKubeNodePoolDataSourceFullConfig,
		os.Getenv("OVH_CLOUD_PROJECT_SERVICE_TEST"),
		name,
		region,
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckKubernetes(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:             config,
				ExpectNonEmptyPlan: true, // expect resource to be read after creation
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "flavor_name", "b3-8"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "desired_nodes", "1"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "min_nodes", "1"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "max_nodes", "2"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "anti_affinity", "true"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscale", "true"),

					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_unneeded_time_seconds", "222"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_unready_time_seconds", "2222"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "autoscaling_scale_down_utilization_threshold", "0.2"),

					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.metadata.annotations.annotation-is-set", "true"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.metadata.labels.label-is-set", "true"),
					// Finalizers not included in this test because they prevent nodes to be deleted.
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.spec.unschedulable", "true"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.spec.taints.0.key", "taint-here"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.spec.taints.0.effect", "PreferNoSchedule"),
					resource.TestCheckResourceAttr("data.ovh_cloud_project_kube_nodepool.poolDataSource", "template.spec.taints.0.value", "true"),
				),
			},
		},
	})
}

var testAccCloudProjectKubeNodePoolDataSourceFullConfig = `
resource "ovh_cloud_project_kube" "cluster" {
	service_name  = "%s"
	name          = "%s"
	region        = "%s"
}

resource "ovh_cloud_project_kube_nodepool" "pool" {
	service_name  = ovh_cloud_project_kube.cluster.service_name
	kube_id       = ovh_cloud_project_kube.cluster.id
	name          = ovh_cloud_project_kube.cluster.name
	flavor_name   = "b3-8"
	desired_nodes = 1
	min_nodes     = 1
	max_nodes     = 2
	anti_affinity = true
	autoscale     = true

	autoscaling_scale_down_unneeded_time_seconds = 222
	autoscaling_scale_down_unready_time_seconds = 2222
	autoscaling_scale_down_utilization_threshold = 0.2

	template {
		metadata {
			annotations {
				annotation-is-set = "true"
			}
			finalizers = [ ]
			labels {
				label-is-set = "true"
			}
		}
		spec {
			unschedulable = true
			taints = [
				{
					key    = "taint-here"
					effect = "PreferNoSchedule"
					value  = "true"
				}
			]
		}
	}

	depends_on = [
		ovh_cloud_project_kube.cluster
	]
}

data "ovh_cloud_project_kube_nodepool" "poolDataSource" {
  service_name  = ovh_cloud_project_kube.cluster.service_name
  kube_id       = ovh_cloud_project_kube.cluster.id
  name          = ovh_cloud_project_kube_nodepool.pool.name

  depends_on = [
    ovh_cloud_project_kube_nodepool.pool
  ]
}
`
