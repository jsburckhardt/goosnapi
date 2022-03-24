lint:
	golangci-lint run --timeout=20m

vet:
	go vet

fmt:
	gofmt -l -w -s .

.PHONY: build
build:
	go build

test-ci:
	go-acc ./... -- -v -coverprofile=cover.out 2>&1 | go-junit-report > report.xml
	gocov convert cover.out | gocov-xml > coverage.xml

test:
	go-acc -o cover.out ./...
	go tool cover -html=cover.out -o cover.html

prepare-checkin: lint
	go mod tidy