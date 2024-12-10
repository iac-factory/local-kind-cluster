# Contributing Guide

## Usage

1. Setup Load Balancer.

    ```bash
    go install sigs.k8s.io/cloud-provider-kind@latest
    sudo install "$(go env --json | jq -r ".GOPATH")/bin/cloud-provider-kind" /usr/local/bin
    sudo cloud-provider-kind
    ```

1. *Local* - Create cluster.

    ```bash
    kind create cluster --config "configuration.yaml" --verbosity 9
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```

1. Install `flux`.

    ```bash
        # flux bootstrap github --repository "https://github.com/iac-factory/cluster-management" \
        #     --owner "iac-factory" \
        #     --private "false" \
        #     --personal "false" \
        #     --path "clusters/local"
   
    kubectl apply -f https://github.com/fluxcd/flux2/releases/latest/download/install.yaml
    ```

1. Start the local registry.

    ```bash
    bash registry.bash
    ```

1. Install `postgres-operator`.

    ```bash
    kubectl apply --server-side -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.24/releases/cnpg-1.24.1.yaml
    ```
   
1. Port forward `postgres`.

    ```bash
    kubectl port-forward --namespace development services/postgres-cluster-rw 5432:5432
    ```

1. Apply the manifests.

    ```bash
    kubectl apply --kustomize ./cluster
    ```

## Istio

### Dashboard

```bash
# kubectl --namespace istio-system create token kiali | pbcopy

istioctl dashboard kiali
istioctl dashboard jaeger
```

### Port-Forwarding

```bash
kubectl apply --kustomize ./applications

kubectl port-forward --namespace development services/api-gateway-istio 8080:80

for i in $(seq 1 100); do 
    curl "http://localhost:8080/v1/test-service-1"
    curl "http://localhost:8080/v1/test-service-2"
    curl "http://localhost:8080/v1/test-service-2/alpha"
done

```

## Kyverno Debugging

```bash
kubectl get --raw /api/v1/namespaces | jq

# --> true || false
kubectl get --raw /api/v1/namespaces | kyverno jp query "items[*].metadata.name | contains(@, 'flux-system')"

# --> yes || no
kubectl auth can-i create ExternalSecret --as system:serviceaccount:kyverno:kyverno-background-controller

kubectl get clusterrole kyverno:background-controller -o yaml
```

## Git Submodule Update(s)

```bash
git submodule update --recursive --remote
```
