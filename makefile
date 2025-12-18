cover:
	go test ./... -v -count=1 -covermode=atomic -coverprofile=coverage.out && go tool cover -func=coverage.out && rm coverage.out
.PHONY: cover

run:
	go run cmd/chainconnector/main.go
.PHONY: run

lint:
	golangci-lint run --timeout 5m && go fmt ./... && go vet ./...
.PHONY: lint