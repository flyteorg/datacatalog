export GO111MODULE=off
export REPOSITORY=datacatalog
include boilerplate/lyft/docker_build/Makefile
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o datacatalog ./cmd/main.go && mv ./datacatalog ./bin

.PHONY: linux_compile
linux_compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts/datacatalog ./cmd/

.PHONY: generate_idl
generate_idl:
        which grpc_tools.protoc || (pip install grpcio-tools)
        python -m grpc_tools.protoc -I ../flyteidl/protos/ -I ./protos/idl/ --python_out=./protos/gen/pb_python/ --grpc_python_out=./protos/gen/pb_python/  ./protos/idl/datacatalog/service.proto	
	protoc -I ./vendor/github.com/lyft/flyteidl/protos/ -I ./protos/idl/datacatalog/. --go_out=plugins=grpc:protos/gen ./protos/idl/datacatalog/service.proto

.PHONY: generate
generate:
	which pflags || (go get github.com/lyft/flytestdlib/cli/pflags)
	which mockery || (go get github.com/enghabu/mockery/cmd/mockery)
	@go generate ./...
