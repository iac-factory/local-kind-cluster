all :: deploy

service = $(shell basename $(CURDIR))

version = $(shell [ -f VERSION ] && head VERSION || echo "0.0.0")

major      		= $(shell echo $(version) | sed "s/^\([0-9]*\).*/\1/")
minor      		= $(shell echo $(version) | sed "s/[0-9]*\.\([0-9]*\).*/\1/")
patch      		= $(shell echo $(version) | sed "s/[0-9]*\.[0-9]*\.\([0-9]*\).*/\1/")

major-upgrade 	= $(shell expr $(major) + 1).$(minor).$(patch)
minor-upgrade 	= $(major).$(shell expr $(minor) + 1).$(patch)
patch-upgrade 	= $(major).$(minor).$(shell expr $(patch) + 1)

bump:
	@echo $(patch-upgrade) > VERSION

prepare: bump
	@echo "$(service) - $(version)"
	@go mod vendor

build: prepare
	@docker build --file Dockerfile --tag "localhost:5050/$(service):$(version)" --build-arg="SERVICE=$(service)" .
	@docker push "localhost:5050/$(service):$(version)"

update: build
	@go run ./private --service "$(service)" --version "$(version)"

apply: update
	@kubectl apply --kustomize ./kustomize --wait

deploy: apply
	@kubectl --namespace development rollout restart deployments/$(service)
	@kubectl --namespace development rollout status deployments/$(service)
