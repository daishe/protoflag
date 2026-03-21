
BIN_DIR = $(shell pwd)/bin

.PHONY: all
all: lint test

.PHONY: clean
clean: tools-clean
	rm -rf dependencies

.PHONY: tidy
tidy:
	go mod tidy
	cd examples && go mod tidy
	cd tools && go mod tidy

.PHONY: update
update:
	go get -t -u all
	go mod tidy
	cd examples && go get -t -u all
	cd examples && go mod tidy
	cd tools && go list -f '{{if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' -m all | xargs go get --tags tools -t -u
	cd tools && go mod tidy

dependencies: go.mod go.sum
	go mod download
	touch dependencies

examples/dependencies: examples/go.mod examples/go.sum
	cd examples && go mod download
	touch examples/dependencies

.PHONY: test
test: dependencies
	go test -count 1 -timeout 5m ./...
	cd examples && go test -count 1 -timeout 5m ./...

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint config verify
	GOWORK="$(shell pwd)/.golangci.go.work" bin/golangci-lint run ./... ./examples/...

.PHONY: tools
tools: bin/buf bin/golangci-lint bin/protoc-gen-buf-breaking bin/protoc-gen-buf-lint bin/protoc-gen-go

.PHONY: tools-clean
tools-clean:
	rm -rf bin
	rm -rf tools/dependencies

tools/dependencies: tools/go.mod tools/go.sum tools/tools.go
	cd tools && go mod download
	touch tools/dependencies

bin/buf: tools/dependencies
	cd tools && go build -o ../bin/buf github.com/bufbuild/buf/cmd/buf

bin/golangci-lint: tools/dependencies
	cd tools && go build -o ../bin/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint

bin/protoc-gen-buf-breaking: tools/dependencies
	cd tools && go build -o ../bin/protoc-gen-buf-breaking github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking

bin/protoc-gen-buf-lint: tools/dependencies
	cd tools && go build -o ../bin/protoc-gen-buf-lint github.com/bufbuild/buf/cmd/protoc-gen-buf-lint

bin/protoc-gen-go: tools/dependencies
	cd tools && go build -o ../bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: proto
proto: bin/buf bin/protoc-gen-buf-breaking bin/protoc-gen-buf-lint bin/protoc-gen-go
	PATH="$(BIN_DIR):$$PATH" buf lint
	find examples -type f -name '*.pb.go' -exec rm {} \;
	(cd examples && PATH="$(BIN_DIR):$$PATH" buf generate --template buf.gen.yaml)
	find internal -type f -name '*.pb.go' -exec rm {} \;
	(cd internal && PATH="$(BIN_DIR):$$PATH" buf generate --template buf.gen.yaml)
