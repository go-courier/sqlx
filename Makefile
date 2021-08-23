GO = go
PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = v$(shell cat .version)

fmt:
	goimports -l -w .

test: tidy
	$(GO) test -v -race ./...

cover: tidy
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

tidy:
	go mod tidy