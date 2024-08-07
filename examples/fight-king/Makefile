IMG ?= imroc/fight-king:latest

CONTAINER_TOOL ?= docker

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push ${IMG}

.PHONY: docker-build-push
docker-build-push: docker-build docker-push

.PHONY: update
update: docker-build-push
	kubectl -n fight-king delete pod --all
