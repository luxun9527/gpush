FROM  debian:stretch-slim


RUN mkdir "/ws"
WORKDIR "/ws"

COPY ./bin/proxy /ws/proxy
COPY ./config /ws/config
RUN chmod +x /ws/proxy
ENTRYPOINT ["./proxy","--config=/ws/config/socket/config.toml"]