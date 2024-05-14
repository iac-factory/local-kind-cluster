resource "helm_release" "external-secrets" {
  chart      = "external-secrets"
  name       = "external-secrets"
  repository = "https://charts.external-secrets.io"
  namespace  = "external-secrets"

  version = "0.9.17"

  create_namespace = true
  skip_crds = false

  atomic  = true
  wait    = true
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}

resource "helm_release" "prometheus" {
  name       = "kube-prometheus-stack"
  chart      = "kube-prometheus-stack"
  repository = "https://prometheus-community.github.io/helm-charts"
  namespace = "prometheus"

  version = "58.5.1"

  create_namespace = true
  skip_crds = false

  atomic  = true
  wait    = true
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}

resource "helm_release" "kyverno" {
  name       = "kyverno"
  chart      = "kyverno"
  repository = "https://kyverno.github.io/kyverno"
  namespace = "kyverno"

  version = "3.2.2"

  create_namespace = true
  skip_crds = false

  atomic  = true
  wait    = true
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}

resource "helm_release" "istio-base" {
  name       = "istio-base"
  chart      = "base"
  repository = "https://istio-release.storage.googleapis.com/charts"
  namespace = "istio-system"

  create_namespace = true
  skip_crds = false

  set {
    name  = "defaultRevision"
    value = "default"
  }

  atomic  = true
  wait    = true
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}

resource "helm_release" "istiod" {
  name       = "istiod"
  chart      = "istiod"
  repository = "https://istio-release.storage.googleapis.com/charts"
  namespace = "istio-system"

  create_namespace = true
  skip_crds = false

  atomic  = true
  wait    = true
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}

resource "helm_release" "istio-gateway" {
  name       = "istio-ingressgateway"
  chart      = "gateway"
  repository = "https://istio-release.storage.googleapis.com/charts"
  namespace = "istio-system"

  create_namespace = true
  skip_crds = false

  atomic  = false
  wait    = false
  timeout = 900

  cleanup_on_fail = true

  dependency_update = true
  force_update = true
}
