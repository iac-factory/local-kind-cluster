build:
	docker build --file Dockerfile --tag localhost:5050/go-http-api:latest .
	docker push localhost:5050/go-http-api:latest

deploy: build
	kubectl --namespace development rollout restart deployments/go-http-api
