CWD := ${CURDIR}

run:
	@go run cli/main.go

gen-proto:
ifndef PROTOC_GEN_GO
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
endif
	@protoc --proto_path=${CWD}/lib/proto ${CWD}/lib/proto/*.proto --go_out=.

get-vendor:
	@go mod tidy
	@go mod vendor

compose:
	@docker-compose up