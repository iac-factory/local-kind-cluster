# local-kind-cluster

Local Kubernetes Cluster(s) via Kind

## Installation

```bash
brew install kind
```

## Usage

```bash
kind create cluster --name local --config configuration.yaml

kubectl config set-context kind-local
```

## Releases

- https://github.com/kubernetes-sigs/kind/releases