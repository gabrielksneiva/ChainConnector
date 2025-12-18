cover:
	go test ./... -v -count=1 -covermode=atomic -coverprofile=coverage.out && go tool cover -func=coverage.out && rm coverage.out

.PHONY: cover