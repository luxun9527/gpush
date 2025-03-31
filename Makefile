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
	dos2unix ./scripts/build.sh
	./scripts/build.sh
run:
	dos2unix ./scripts/run.sh
	dos2unix ./scripts/stop.sh
	./scripts/run.sh
dockerBuild:
	docker build -t='gpush' .
dockerRun:
	docker run -di --name=gpush -p 9992:9992 -p 10067:10067 -p 10068:10068 gpush:latest


stress:
	chmod +x ./bin/stress
	./bin/stress