
variable "gcp_project_id" {
  type = string
}

output "gcp_project_id" { value = var.gcp_project_id }

variable "gcp_project_name" {
    type = string
}

output "gcp_project_name" { value = var.gcp_project_name }

variable "gcp_project_region" {
  type = string
  default = "europe-west2"
}
output "gcp_project_region" { value = var.gcp_project_region }

variable "gcp_project_zone" {
    type = string
    default = "europe-west2-a"
}
output "gcp_project_zone" { value = var.gcp_project_zone }
output "static_ingress" {
  value = module.cluster.static_ingress
}

variable "cluster_name" {
    type = string
    default = "kluster"
}
output "cluster_name" { value = var.cluster_name }

# This exists to enable output variables to be added without changing other
# resources
# * [issue](https://github.com/hashicorp/terraform/issues/22607) remains open.
# * [Workaround](https://support.hashicorp.com/hc/en-us/articles/360000958148-Terraform-outputs-fail-to-update-in-Terraform-Enterprise)
resource "null_resource" "n" {}

module "cluster" {

  project                             = var.gcp_project_id
  gcp_project_name                    = var.gcp_project_name
  source                              = "./gke"
  region                              = var.gcp_project_region
  location                            = var.gcp_project_zone
  zone                                = var.gcp_project_zone
  cluster_name                        = var.cluster_name
  cluster_range_name                  = "gke-pods"
  services_range_name                 = "gke-services"
  daily_maintenance_window_start_time = "03:00"
  subnet_cidr_range                   = "10.0.0.0/16" # 10.0.0.0 -> 10.0.255.255
  master_ipv4_cidr_block              = "10.1.0.0/28" # 10.1.0.0 -> 10.1.0.15
  cluster_range_cidr                  = "10.2.0.0/16" # 10.2.0.0 -> 10.2.255.255
  services_range_cidr                 = "10.3.0.0/16" # 10.3.0.0 -> 10.3.255.255
  nat_ip_allocate_option              = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat  = "LIST_OF_SUBNETWORKS"
  source_ip_ranges_to_nat             = ["ALL_IP_RANGES"]
  nat_log_filter                      = "ERRORS_ONLY"
  logging_service                     = "none" # $$$
  monitoring_service                  = "none" # $$$

  # We seperately define the node pools, as per recomendation here
  # https://www.terraform.io/docs/providers/google/guides/using_gke_with_terraform.html
  node_pools = {
    ingress-pool = {
      machine_type       = "g1-small" # $$$
      initial_node_count = 1
      min_node_count     = 1
      max_node_count     = 1
      # Can make this !preemptible for reliability but for development its not worth it
      preemptible        = true
      auto_repair        = true
      auto_upgrade       = false
      disk_size_gb       = 10
      disk_type          = "pd-standard"
      image_type         = "COS"
      service_account    = "kluster-serviceaccount@${var.gcp_project_id}.iam.gserviceaccount.com"
    }
    work-pool = {
      machine_type       = "n1-standard-2" # $$$
      initial_node_count = 3
      min_node_count     = 1
      max_node_count     = 5
      preemptible        = true
      auto_repair        = true
      auto_upgrade       = true
      disk_size_gb       = 64
      disk_type          = "pd-standard"
      image_type         = "COS"
      service_account    = "kluster-serviceaccount@${var.gcp_project_id}.iam.gserviceaccount.com"

    }
  }

  node_pools_taints = {
    ingress-pool = [
      {
        key    = "ingress-pool"
        value  = true
        effect = "NO_EXECUTE"
      }
    ]
    work-pool = []
  }

  node_pools_tags = {
    ingress-pool = [
      "ingress-pool"
    ]
    work-pool = [
      "work-pool"
    ]
  }

  node_pools_oauth_scopes = {
    custom-node-pool = [
      "https://www.googleapis.com/auth/cloud-platform",
      "https://www.googleapis.com/auth/service.management",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }
}

