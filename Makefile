.PHONY: build build-version lint lint-fix vet clean check

build:
	go build -o lazyjira ./cmd/lazyjira

build-version:
	go build -ldflags "-s -w -X main.version=$$(git rev-parse --short HEAD)" -o lazyjira ./cmd/lazyjira

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

vet:
	go vet ./...

clean:
	rm -f lazyjira

check: lint vet build
