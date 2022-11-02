
.PHONY: run
run: fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...
