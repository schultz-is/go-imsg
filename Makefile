.PHONY: test
test:
	go test -v -coverprofile coverage.out ./...

.PHONY: coverage
coverage:
	go tool cover -html coverage.out

.PHONY: vet
vet:
	go vet -v ./...

.PHONY: clean
clean:
	rm coverage.out
