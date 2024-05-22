# Contributing Guide

## Usage

1. Spin up the cluster, locally.

    ```bash
    kind delete cluster
    kind create cluster --config "configuration.yaml"
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```

1. Start the local registry.

    ```bash
    bash registry.bash
    ```

1. Establish Secret(s).

    ```bash
    mkdir -p ./kustomize/.secrets
   
    printf "%s" "${GITHUB_USER}" > ./kustomize/.secrets/username
    printf "%s" "${GITHUB_TOKEN}" > ./kustomize/.secrets/password

    printf "%s" "$(aws configure get aws_access_key_id)" > ./kustomize/.secrets/aws-access-key-id
    printf "%s" "$(aws configure get aws_secret_access_key)" > ./kustomize/.secrets/aws-secret-access-key
    
    printf "[default]\naws_access_key_id=%s\naws_secret_access_key=%s" "$(aws configure get aws_access_key_id)" "$(aws configure get aws_secret_access_key)" > ./kustomize/.secrets/profile 
    
    kubectl apply --kustomize ./kustomize
    ```

1. Install.

    ```bash
    kubectl apply --kustomize ./kustomize
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
