terraform {
  required_version = "~> 0.13"
  required_providers {
    google-beta = ">= 3.8"
  }
  backend "remote" {
    organization = "robustroundrobin"
    workspaces {
      name = "fetlar"
    }
  }
}
