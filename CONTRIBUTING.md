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
    kind create cluster --config "configuration.yaml"
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```
1. Bootstrap.
    ```bash
    flux bootstrap github --repository "https://github.com/iac-factory/cluster-management" \
        --owner "iac-factory" \
        --private "false" \
        --personal "false" \
        --path "clusters/local"
    ```
1. Sync local cluster repository's `vendors`.
    ```bash
    git submodule update --remote --recursive
    ```
1. Add `kustomization.yaml` to new cluster directory.
    ```bash
    cat << EOF > ./vendors/cluster-management/clusters/local/kustomization.yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources: []
    EOF
    ```
1. Update.
    ```bash
    git submodule foreach "git add . && git commit --message \"Git Submodule Update(s)\" && git push -u origin HEAD:main" 
    ```
1. Establish Secret(s).
    ```bash
    mkdir -p ./kustomize/secrets/.secrets
   
    printf "%s" "${GITHUB_USER}" > ./kustomize/secrets/.secrets/username
    printf "%s" "${GITHUB_TOKEN}" > ./kustomize/secrets/.secrets/password

    function access-key-id() {
        printf "%s" "$(aws secretsmanager get-secret-value --secret-id "local/external-secrets/provider/aws/credentials" --query SecretString | jq -r | jq -r ".\"aws-access-key-id\"")"
    }

    function secret-access-key() {
        printf "%s" "$(aws secretsmanager get-secret-value --secret-id "local/external-secrets/provider/aws/credentials" --query SecretString | jq -r | jq -r ".\"aws-secret-access-key\"")"
    }

    printf "%s" "$(access-key-id)" > ./kustomize/secrets/.secrets/aws-access-key-id
    printf "%s" "$(secret-access-key)" > ./kustomize/secrets/.secrets/aws-secret-access-key
    
    # printf "[default]\naws_access_key_id=%s\naws_secret_access_key=%s" "$(aws configure get aws_access_key_id)" "$(aws configure get aws_secret_access_key)" > ./kustomize/secrets/.secrets/profile
   
    kubectl apply --kustomize ./kustomize/secrets --wait
    ```
1. Start the local registry.
    ```bash
    bash registry.bash
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
    curl "http://localhost:8080/v1/test-service-1/alpha"
    curl "http://localhost:8080/v1/test-service-1/bravo"
    curl "http://localhost:8080/v1/test-service-2"
    curl "http://localhost:8080/v1/test-service-3"
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
