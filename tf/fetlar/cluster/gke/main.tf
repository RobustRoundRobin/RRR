provider "google" {
  version     = "3.4.0"
  #project = var.project
  #region = var.region
  #zone = var.zone
}

output "static_ingress" {
  value = google_compute_address.static-ingress.address
}

data "google_client_config" "provider" {}

resource "google_project_service" "cloudresourcemanager" {
  project = var.project
  service = "cloudresourcemanager.googleapis.com"

  disable_dependent_services = true
}

resource "google_project_service" "iam" {
  project = var.project
  service    = "iam.googleapis.com"
  depends_on = [google_project_service.cloudresourcemanager]

  disable_dependent_services = true
}

resource "google_project_service" "container" {
  project = var.project
  service    = "container.googleapis.com"
  depends_on = [google_project_service.iam]

  disable_dependent_services = true
}

resource "google_container_cluster" "k8s" {
  provider           = google-beta
  name               = var.cluster_name
  project            = var.project
  depends_on         = [google_project_service.container]
  location           = var.location
  logging_service    = var.logging_service
  monitoring_service = var.monitoring_service

  # need to create a default node pool
  # delete this immediately
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.gke-network.self_link
  subnetwork = google_compute_subnetwork.gke-subnet.self_link

  workload_identity_config {
    identity_namespace = "${var.project}.svc.id.goog"
  }
  private_cluster_config {
    master_ipv4_cidr_block  = var.master_ipv4_cidr_block
    enable_private_nodes    = var.enable_private_nodes
    enable_private_endpoint = var.enable_private_endpoint
  }

  ip_allocation_policy {
    cluster_secondary_range_name  = var.cluster_range_name
    services_secondary_range_name = var.services_range_name
  }
}
