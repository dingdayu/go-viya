.PHONY: fmt vet test race tidy check release-check release-snapshot

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test -race -cover ./...

race:
	go test -race ./...

tidy:
	go mod tidy

check: fmt tidy vet test

release-check:
	goreleaser check
	go run golang.org/x/exp/cmd/gorelease@latest -base=latest

release-snapshot:
	goreleaser release --snapshot --clean
