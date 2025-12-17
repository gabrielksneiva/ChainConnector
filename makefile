cover:
	go test ./... -v -covermode=atomic -coverprofile=coverage.out && go tool cover -func=coverage.out && rm coverage.out

.PHONY: cover