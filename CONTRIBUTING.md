# Contribution Guide

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
    kubectl port-forward -n "development" services/api-ingress-gateway-istio 10000:80
    kubectl port-forward -n development services/postgres 5432:5432
    ```

7. Simulate load.

    ```bash
    for i in $(seq 1 100); do 
        curl -s -o /dev/null "http://localhost:10000/v1";
    done
    ```
