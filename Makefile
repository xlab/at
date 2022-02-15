FILES_TO_FMT      ?= $(shell find . -path ./vendor -prune -o -name '*.go' -print)

## format: Formats Go code
.PHONY: format
format:
	@echo ">> formatting code"
	@gofmt -s -w $(FILES_TO_FMT)

## test-integration: Run all Go integration tests
.PHONY: test-integration
test-integration:
	echo 'mode: atomic' > coverage.out
	go list ./... | xargs -I{} sh -c 'go test -race -tags=integration -covermode=atomic -coverprofile=coverage.tmp -coverpkg $(go list ./... | tr "\n" ",") {} && tail -n +2 coverage.tmp >> coverage.out || exit 255'
	rm coverage.tmp
