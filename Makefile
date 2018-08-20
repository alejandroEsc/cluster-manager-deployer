
.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

go-lint: ## quick run of linting
	golint -set_exit_status $(shell git ls-files "**/*.go" "*.go" | grep -v -e "vendor" | xargs echo)

compile: ## Compile the binaries into the ./bin directory
	mkdir -p ./bin

clean: ## Clean up built files
	rm -rf ./bin

goimport: ## Run goimports
	goimports -w $(shell git ls-files "**/*.go" "*.go" | grep -v -e "vendor" | xargs echo)

