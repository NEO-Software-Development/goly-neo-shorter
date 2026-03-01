terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.0.0"
    }
  }
}

provider "kubernetes" {
  // The provider configuration is left empty. Terraform will use the
  // default kubeconfig file to connect to the Kubernetes cluster.
  // Make sure your kubeconfig is pointing to the correct cluster.
}

resource "kubernetes_secret" "db_secret" {
  metadata {
    name = "goly-db-secret"
  }
  data = {
    DB_USER     = var.db_user
    DB_PASSWORD = var.db_password
  }
}

resource "kubernetes_config_map" "app_config" {
  metadata {
    name = "goly-app-config"
  }
  data = {
    DB_HOST = var.db_host
    DB_NAME = var.db_name
  }
}

resource "kubernetes_deployment" "goly_deployment" {
  metadata {
    name = "goly-deployment"
  }
  spec {
    replicas = var.replicas
    selector {
      match_labels = {
        app = "goly"
      }
    }
    template {
      metadata {
        labels = {
          app = "goly"
        }
      }
      spec {
        container {
          image = var.app_image
          name  = "goly"
          ports {
            container_port = 3000
          }
          env_from {
            secret_ref {
              name = kubernetes_secret.db_secret.metadata[0].name
            }
            config_map_ref {
              name = kubernetes_config_map.app_config.metadata[0].name
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "goly_service" {
  metadata {
    name = "goly-service"
  }
  spec {
    selector = {
      app = "goly"
    }
    ports {
      port        = 80
      target_port = 3000
    }
    type = "LoadBalancer"
  }
}
