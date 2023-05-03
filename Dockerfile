FROM golang:1.18.1-alpine AS builder

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct
WORKDIR /build
COPY . /build
RUN mkdir ./bin\
    && cd "/build/cmd/socket"\
    && go build -buildvcs=false  -o "/build/bin/socket"\
    && cd "/build/cmd/proxy"\
    && go build -buildvcs=false  -o "/build/bin/proxy"

FROM debian:stretch-slim

WORKDIR "/gpush"
COPY --from=builder /build/bin /gpush/bin
COPY --from=builder /build/config /gpush/config
COPY --from=builder /build/scripts /gpush/scripts
EXPOSE 9992
EXPOSE 10067
EXPOSE 10068
ENTRYPOINT ["./scripts/docker_run.sh"]