all :: deploy

prepare:
	go mod vendor

build: prepare
	docker build --file Dockerfile --tag localhost:5050/authentication:latest .
	docker push localhost:5050/authentication:latest

apply: build
	kubectl apply --kustomize ./kustomize/overlays/development --wait

deploy: apply
	kubectl --namespace development rollout restart deployments/authentication
	kubectl --namespace development rollout status deployments/authentication
