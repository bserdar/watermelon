PREFIX?=$(shell pwd)

.DEFAULT: all
all: watermelon test

# Package list
PKGS=$(shell go list  ./cmd/... ./server/...|  grep -v test )


server/pb/inventory.pb.go: proto/inventory.proto
	protoc --proto_path=proto  --go_out=plugins=grpc:../../.. $< 
server/pb/module.pb.go: proto/module.proto
	protoc --proto_path=proto  --go_out=plugins=grpc:../../.. $< 
server/pb/host.pb.go: proto/host.proto
	protoc --proto_path=proto  --go_out=plugins=grpc:../../.. $< 
server/pb/remote.pb.go: proto/remote.proto
	protoc --proto_path=proto  --go_out=plugins=grpc:../../.. $< 
server/pb/empty.pb.go: proto/empty.proto
	protoc --proto_path=proto  --go_out=plugins=grpc:../../.. $< 

proto: server/pb/inventory.pb.go server/pb/module.pb.go server/pb/remote.pb.go server/pb/host.pb.go server/pb/empty.pb.go

.PHONY: vet
vet:
	@echo "+ $@"
	@go vet  $(PKGS)

.PHONY: fmt
fmt:
	@echo "+ $@"
	@test -z "$$(gofmt -s -l server 2>&1 | grep -v pb\.go | tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)
	@test -z "$$(gofmt -s -l cmd 2>&1 |  tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)
	@test -z "$$(gofmt -s -l client 2>&1 |  tee /dev/stderr)" || \
		(echo >&2 "+ please format Go code with 'gofmt -s'" && false)

watermelon: proto fmt vet
	@echo "+ $@"
	@go build -o watermelon main.go

.PHONY: test
test:
	@echo "+ $@"
	@go test  $(PKGS)

.PHONY: clean
clean:
	@echo "+ $@"
	rm watermelon

