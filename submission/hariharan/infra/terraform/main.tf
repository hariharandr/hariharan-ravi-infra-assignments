terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.27"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.13"
    }
  }
}

provider "kubernetes" {
  config_path    = "~/.kube/config"
  config_context = "kind-config-service"
}

provider "helm" {
  kubernetes {
    config_path    = "~/.kube/config"
    config_context = "kind-config-service"
  }
}

resource "kubernetes_namespace" "config_service" {
  metadata {
    name = "config-service"
    labels = {
      managed-by = "terraform"
    }
  }
}

resource "kubernetes_secret" "db_credentials" {
  metadata {
    name      = "db-credentials"
    namespace = kubernetes_namespace.config_service.metadata[0].name
  }

  data = {
    DATABASE_URL = "postgres://configuser:configsecret@postgres-postgresql.config-service.svc.cluster.local:5432/configdb"
    DB_PASSWORD  = "configsecret"
    DB_USER      = "configuser"
  }

  type = "Opaque"
}

resource "helm_release" "postgres" {
  name       = "postgres"
  repository = "oci://registry-1.docker.io/bitnamicharts"
  chart      = "postgresql"
  namespace  = kubernetes_namespace.config_service.metadata[0].name
  version    = "18.5.23"

  set {
    name  = "auth.username"
    value = "configuser"
  }

  set {
    name  = "auth.password"
    value = "configsecret"
  }

  set {
    name  = "auth.database"
    value = "configdb"
  }

  set {
    name  = "primary.persistence.enabled"
    value = "false"
  }
}

output "namespace" {
  value = kubernetes_namespace.config_service.metadata[0].name
}

output "db_secret_name" {
  value = kubernetes_secret.db_credentials.metadata[0].name
}
