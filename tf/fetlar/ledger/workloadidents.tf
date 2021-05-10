resource "google_service_account" "workloads" {
  for_each = {
    kubeip = ["kube-system", "kubeip-sa"]
    queth_genesis = ["queth", "quorum-genesis-sa"]
    queth_node = ["queth", "quorum-node-sa"]
    queth_membership = ["queth", "quorum-membership-sa"]
    queth_client = ["queth", "quorum-client-sa"]
  }
  account_id   = each.value[1]
  display_name = substr("Workload Identity ${local.gcp_project_sa_fqn}[${each.value[0]}/${each.value[1]}]", 0, 100)
  project      = local.gcp_project_id
}

resource "google_service_account_iam_member" "workloads" {
  for_each = {
    kubeip = ["kube-system", "kubeip-sa"]
    queth_genesis = ["queth", "quorum-genesis-sa"]
    queth_node = ["queth", "quorum-node-sa"]
    queth_membership = ["queth", "quorum-membership-sa"]
    queth_client = ["caas", "quorum-client-sa"]
  }

  service_account_id = google_service_account.workloads[each.key].name
  role               = "roles/iam.workloadIdentityUser"
  member             = "${local.gcp_project_sa_fqn}[${each.value[0]}/${each.value[1]}]"
}
