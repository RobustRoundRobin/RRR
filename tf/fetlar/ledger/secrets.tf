# NOTICE: These need to be imported via terraform import, eg:
# terraform import \
#   module.cluster.google_secret_manager_secret.qnode[\"qnode-$i-$kind\"] \
#   projects/fetlar/secrets/qnode-$i-$kind
resource "google_secret_manager_secret" "qnode" {
  project = local.gcp_project_id
  provider = google-beta

  for_each = toset([
    for pair in setproduct(
      range(0, var.max_prefunded_nodes),
        ["enode", "key", "wallet-address", "wallet-key", "wallet-password"]
      ): "qnode-${tostring(pair[0])}-${pair[1]}"])

  secret_id = each.key
  replication {
    automatic = true
  }
}

# access to node keys is granted only to the "node" sa
resource "google_secret_manager_secret_iam_member" "qnode-key" {
  project = local.gcp_project_id
  count = var.max_prefunded_nodes
  provider = google-beta
  secret_id = google_secret_manager_secret.qnode["qnode-${count.index}-key"].secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.workloads["queth_node"].email}"
}

# the enode isn't secret, its just there for convenience.
resource "google_secret_manager_secret_iam_member" "qnode-enode" {
  project = local.gcp_project_id
  count = var.max_prefunded_nodes
  provider = google-beta
  secret_id = google_secret_manager_secret.qnode["qnode-${count.index}-enode"].secret_id
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.workloads["queth_node"].email}"
}

resource "google_secret_manager_secret_iam_member" "qnode-wallet" {
  provider = google-beta
  project = local.gcp_project_id

  for_each = toset([
    for pair in setproduct(
      range(0, var.max_prefunded_nodes), ["address", "key", "password"]
      ): "qnode-${tostring(pair[0])}-wallet-${pair[1]}"])

  secret_id = each.key
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.workloads["queth_client"].email}"
}

# For convenience genesis pre-funds a wallet for up to max_prefunded_nodes. the
# wallet addresses in the genesis config. its easiest if it reads them direct
# from secretmanager. max_prefunded_nodes is a b
resource "google_secret_manager_secret_iam_member" "qnode-wallet-addresses" {
  provider = google-beta
  project = local.gcp_project_id

  for_each = toset([
    for pair in setproduct(
      range(0, var.max_prefunded_nodes), ["address"]
      ): "qnode-${tostring(pair[0])}-wallet-${pair[1]}"])

  secret_id = each.key
  role = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.workloads["queth_node"].email}"
}
