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
	./scripts/build.sh
run:
	./scripts/run.sh
dockerBuild:
	docker build -t='gpush' .
dockerRun:
	docker run -di --name=gpush -p 9992:9992 -p 10067:10067 -p 10068:10068 gpush:latest