PKGS := $(shell go list ./... | grep -v /vendor)

.PHONY: lint
lint:
	golangci-lint run -v ./...

.PHONY: test
test:
	go test -v $(PKGS)

.PHONY: staticcheck
staticcheck:
	staticcheck $(PKGS)
