.PHONY: proto
proto:
	protoc  -Iproto --go_out=. --go-grpc_out=.  Proxy.proto

run:
	nohup ./proxy --config=config/proxy/config.toml &
	nohup ./ws --config=config/socket/config.toml &
build:
	cd scipts