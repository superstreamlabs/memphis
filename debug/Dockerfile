FROM golang:1.18-alpine3.15 as build

WORKDIR $GOPATH/src/memphis-broker
COPY . .

RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -o  .

FROM alpine:3.15
ENV GOPATH="/go/src"
WORKDIR /run

RUN apk add --update ca-certificates && mkdir -p /nats/bin && mkdir /nats/conf

COPY --from=build $GOPATH/memphis-broker/memphis-broker /bin/nats-server
COPY --from=build $GOPATH/memphis-broker/conf/* conf/
COPY --from=build $GOPATH/memphis-broker/version.conf .
COPY --from=build /go/bin/dlv /bin/dlv

EXPOSE 5555 6666 8222 4000

ENTRYPOINT ["/bin/dlv", "--listen=:4000", "--headless=true", "--log=true", "--accept-multiclient", "--api-version=2", "exec", "/bin/nats-server" ]
