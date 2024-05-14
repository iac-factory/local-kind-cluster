# local-kind-cluster

Local Kubernetes Cluster(s) via Kind

## Installation

```bash
brew install kind
brew install tfk8s
```

## Usage

```bash
kind create cluster --config configuration.yaml

kubectl config set-context kind-kind
```

## Setup

```bash
# - https://istio.io/latest/docs/setup/getting-started/

# --> install istio
curl -L https://istio.io/downloadIstio | sh -

mv ./istio-1.21.2/bin/istioctl /usr/local/bin

# --> metrics server
kubectl apply -f ./metrics-server.yaml

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack --namespace prometheus --create-namespace --wait

helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

helm install istio-base istio/base -n istio-system --set defaultRevision=default --wait --create-namespace

helm install istiod istio/istiod -n istio-system --wait --create-namespace

kubectl get deployments -n istio-system --output wide --watch

helm install istio-ingressgateway istio/gateway -n istio-system --create-namespace
helm install istio-egressgateway istio/gateway -n istio-system --create-namespace

helm install istio-ingress istio/gateway -n development --create-namespace

kubectl apply -f ./addons/prometheus.yaml
kubectl apply -f ./addons/grafana.yaml
kubectl apply -f ./addons/jaeger.yaml
kubectl apply -f ./addons/loki.yaml
kubectl apply -f ./addons/kiali.yaml

kubectl apply -f istio-addons/extras/prometheus-operator.yaml
  
kubectl rollout status deployment/kiali -n istio-system

kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml

./registry.bash

cd applications/go-http-api
make build
kubectl apply -k ./kustomize/overlays/development

cd -

kubectl get svc istio-ingressgateway -n istio-system 

export INGRESS_HOST=127.0.0.1
export INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}')
export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].nodePort}')

export GATEWAY_URL=$INGRESS_HOST:$INGRESS_PORT

echo "$GATEWAY_URL"

echo "http://$GATEWAY_URL"

kubectl port-forward -n istio-system services/istio-ingressgateway 10000:80

for i in $(seq 1 100); do 
    curl -s -o /dev/null "http://localhost:10000";
done

istioctl dashboard kiali 
```

## Releases

- https://github.com/kubernetes-sigs/kind/releases
