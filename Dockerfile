FROM ghcr.io/battleofthebots/botb-base-image AS builder
RUN apt-get install -y golang

WORKDIR /opt
COPY go.mod .
RUN go mod tidy
COPY server.go .
RUN go build server.go

FROM ghcr.io/battleofthebots/botb-base-image

USER user

COPY --from=builder /opt/server  /bin
CMD /bin/server
