MODULE       := github.com/repeter513/shop-proto
PROTO_ROOT   := proto
GEN_ROOT     := gen/go
PROTO_FILES  := $(shell find $(PROTO_ROOT) -name '*.proto')
PROTOC       := protoc
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc
GOOGLEAPIS   := $(shell go list -m -f '{{.Dir}}' google.golang.org/protobuf)/types

.PHONY: all proto deps clean

all: proto

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: deps
	@command -v protoc >/dev/null || (echo "install protoc: brew install protobuf" && exit 1)
	@test -x "$(PROTOC_GEN_GO)" || (echo "run: make deps" && exit 1)
	@test -x "$(PROTOC_GEN_GO_GRPC)" || (echo "run: make deps" && exit 1)
	@mkdir -p $(GEN_ROOT)
	$(PROTOC) \
		-I $(PROTO_ROOT) \
		-I $(GOOGLEAPIS) \
		--go_out=. --go_opt=module=$(MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

clean:
	rm -rf $(GEN_ROOT)