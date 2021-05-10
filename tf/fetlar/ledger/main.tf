variable "cluster_workspace" {
  type = string
  default = "cluster"
}
output "cluster_workspace" { value = var.cluster_workspace }

variable "cluster_name" {
    type = string
    default = "kluster"
}
output "cluster_name" { value = var.cluster_name }

variable "max_prefunded_nodes" {
  type = number
  default = 30
}
output "max_quorum_nodes" { value = var.max_prefunded_nodes }

output "gcp_project_id" { value = local.gcp_project_id }
output "gcp_project_region" { value = local.gcp_project_region }
output "gcp_project_zone" { value = local.gcp_project_zone }
output "cluster_bucket" { value = "${local.gcp_project_id}-${random_uuid.cluster_bucket.result}" }

# exists to fource re apply when only output values have changed
resource "null_resource" "nn" {}

locals {
  # All remote state references are via variables with short cuts in the
  # locals.
  gcp_project_id = data.terraform_remote_state.cluster.outputs.gcp_project_id
  gcp_project_region = data.terraform_remote_state.cluster.outputs.gcp_project_region
  gcp_project_zone = data.terraform_remote_state.cluster.outputs.gcp_project_zone
  # this is the workload identity base for the cluster. All workload identities
  # are constructed from this - thats how they work.
  gcp_project_sa_fqn = "serviceAccount:${data.terraform_remote_state.cluster.outputs.gcp_project_id}.svc.id.goog"
  static_ingress = data.terraform_remote_state.cluster.outputs.static_ingress
}

provider "random" {}

provider "null" {}

provider "google" {
  version     = "3.4.0"
  #project = data.terraform_remote_state.cluster.outputs.gcp_project_id
  #project = "ledger-2"
  project = local.gcp_project_id
}

data "terraform_remote_state" "cluster" {
  backend = "remote"
  config = {
    organization = "robustroundrobin"
    workspaces = {
      # name = "ledger-2"
      name = var.cluster_workspace
    }
  }
}

provider "kubernetes" {
  load_config_file = "false"
  host = "https://${data.google_container_cluster.fetlar.endpoint}"
  token = data.google_client_config.provider.access_token
  cluster_ca_certificate = base64decode(
    data.google_container_cluster.fetlar.master_auth[0].cluster_ca_certificate,
  )
}

data "google_client_config" "provider" {}
data "google_container_cluster" "fetlar" {
  name = var.cluster_name
  project = data.terraform_remote_state.cluster.outputs.gcp_project_id
  location = data.terraform_remote_state.cluster.outputs.gcp_project_zone
}

resource "random_uuid" "cluster_bucket" { }

resource "google_storage_bucket" "cluster" {
  # requires that the cpe default service account is added as a delegated
  # owner at https://www.google.com/webmasters/verification
  # name = "${local.gcp_project_id}-cluster.${var.gcp_buckets_tld}"
  name = "${local.gcp_project_id}-${random_uuid.cluster_bucket.result}"
  project = local.gcp_project_id
  # location is the 'region' here!
  location = local.gcp_project_region
  storage_class = "STANDARD"
}
