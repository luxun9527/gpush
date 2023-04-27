.PHONY: proto
proto:
	protoc -Iproto/googleapis -Iproto \
        --grpc-gateway_out ./proto \
        --grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt generate_unbound_methods=true \
        --go_out=./proto\
        --go-grpc_out=./proto  \
        proto/Proxy.proto
build:
	./scripts/run.sh
