GO = go
PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = v$(shell cat .version)

fmt: install.goimports
	goimports -l -w .

test: tidy
	$(GO) test -v -race ./...

cover: tidy
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

tidy: fmt
	go mod tidy

install.goimports:
	go install golang.org/x/tools/cmd/goimports

install.go-mod-outdated:
	go install github.com/psampaz/go-mod-outdated@v0.8.0