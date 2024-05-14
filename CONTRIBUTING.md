# Contribution Guide

## Usage

1. Spin up the cluster, locally.

    ```bash
    kind delete cluster
    kind create cluster --config "configuration.yaml"
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```

## General

**Helm Chart Versions**

```bash
helm search repo kyverno -l
```
