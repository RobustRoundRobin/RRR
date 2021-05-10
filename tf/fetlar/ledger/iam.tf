# -----------------------------------------------------------------------------
# kubeip service account and role
# -----------------------------------------------------------------------------
resource "google_project_iam_custom_role" "kubeip" {
  role_id = "kubeip"
  title = "kubeip role"
  project = local.gcp_project_id

  permissions = [
    "compute.addresses.list",
    "compute.instances.addAccessConfig", "compute.instances.deleteAccessConfig",
    "compute.instances.get",
    "compute.instances.list",
    "compute.projects.get",
    "container.clusters.get",
    "container.clusters.list",
    "resourcemanager.projects.get",
    "compute.networks.useExternalIp",
    "compute.subnetworks.useExternalIp",
    "compute.addresses.use",
  ]
}

resource "google_project_iam_member" "iam_member_kubeip" {
  project = local.gcp_project_id
  role = "projects/${local.gcp_project_id}/roles/kubeip"
  member = "serviceAccount:${google_service_account.workloads["kubeip"].email}"
}

# -----------------------------------------------------------------------------
# cluster storage bucket iam bindings
# -----------------------------------------------------------------------------

# Set the appropriate iam member and roles for quourm service account. This is additive
#
# Consider using - https://github.com/terraform-google-modules/terraform-google-iam/tree/master/modules/storage_buckets_iam
# * https://cloud.google.com/iam/docs/overview
# * https://www.terraform.io/docs/providers/google/r/storage_bucket_iam.html
# Notes
# * policy - setting the *whole* policy clobbers any existing policy already on the
#   resource - hence google_storage_bucket_iam_policy is a bit of a shotgun/foot.
# * binding - sets *all* members (discarding previous) for a particular role.
#   After application, any previous members are dropped. no other (previous) members 
# * member - not authorative (additive)
resource "google_storage_bucket_iam_member" "cluster_bucket_quorum_members" {
  # TODO: configure this in module "cluster" like we do for the node pools
  for_each = {
    genesis_objectadmin = ["roles/storage.objectAdmin", "queth_genesis"]
    genesis_objectview = ["roles/storage.objectViewer", "queth_genesis"]
    membership_objectadmin = ["roles/storage.objectAdmin", "queth_membership"]
    membership_objectview = ["roles/storage.objectViewer", "queth_membership"]
    node_objecadmin = ["roles/storage.objectAdmin", "queth_node"]
    node_objectview = ["roles/storage.objectViewer", "queth_node"]
    client_objectview = ["roles/storage.objectViewer", "queth_node"]
  }
  bucket = google_storage_bucket.cluster.name
  role = each.value[0]
  member = "serviceAccount:${google_service_account.workloads[each.value[1]].email}"
}
# iam for secrets in secrets.tf
