FROM golang:1.18-alpine3.15 as build

WORKDIR $GOPATH/src/memphis-broker
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w" -a -o  .

FROM alpine:3.15
ENV GOPATH="/go/src"
WORKDIR /run

RUN apk update && apk add --no-cache make protobuf-dev
RUN apk add --update ca-certificates && mkdir -p /nats/bin && mkdir /nats/conf

COPY --from=build $GOPATH/memphis-broker/memphis-broker /bin/nats-server
COPY --from=build $GOPATH/memphis-broker/conf/* conf/
COPY --from=build $GOPATH/memphis-broker/version.conf .

EXPOSE 5555 6666 8222

ENTRYPOINT ["/bin/nats-server"]
