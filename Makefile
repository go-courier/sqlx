GO = go
PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = v$(shell cat .version)

fmt:
	goimports -l -w .
	gofmt -l -w .

test:
	$(GO) test -v -race ./...

cover:
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...