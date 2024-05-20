# Contributing Guide

## Usage

1. Spin up the cluster, locally.

    ```bash
    kind delete cluster
    kind create cluster --config "configuration.yaml"
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```

2. Execute `terraform apply`.

    ```bash
    terraform -chdir=./ci/terraform apply 
    ```

3. Add additional manifests.

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/addons/extras/prometheus-operator.yaml
    ```

4. Start the local registry.

    ```bash
    bash registry.bash
    ```

5. Deploy application(s).

    ```bash
    cd applications/authentication
    make
    cd - 
    ```

6. Begin port-forwarding.

    ```bash
    # kubectl port-forward -n istio-system services/istio-ingressgateway 10000:80
    kubectl port-forward -n "development" services/postgres 5432:5432
    kubectl port-forward -n "development" services/api-ingress-gateway-istio 10000:80
    ```

7. Simulate load.

    ```bash
    for i in $(seq 1 100); do 
        curl -s -o /dev/null "http://localhost:10000";
    done
    ```

8. View dashboard.

    ```bash
    istioctl dashboard kiali 
    ```

## External Secrets Setup

**Disclaimer** - _Local Development Purposes Only_. Do not use the following method outside local development.

### External Secrets Keyverno Permissions

***Note*** - Keyverno permissions are *additive*; `ClusterRole` -- when applied -- mutate permissions, rather than
overwrite.

```bash
kubectl apply --filename "./policies/kyverno-external-secrets-permissions.yaml"
```

### Provider Setup

```bash
kubectl create namespace "cloud-provider-system"
kubectl create secret generic --namespace "cloud-provider-system" "aws-secrets-manager-bootstrap" \
    --from-literal="aws-access-key-id=$(aws configure get aws_access_key_id)" \
    --from-literal="aws-secret-access-key=$(aws configure get aws_secret_access_key)"

kubectl apply --filename "./policies/aws-cluster-secret-store.yaml"
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
