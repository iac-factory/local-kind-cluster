provider "kubernetes" {
  host = "https://127.0.0.1:65300" # kubectl cluster-info
  config_path = "~/.kube/config"
}

provider "helm" {
  kubernetes {
    host = "https://127.0.0.1:65300" # kubectl cluster-info
    config_path = "~/.kube/config"
  }
}
