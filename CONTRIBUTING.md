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
