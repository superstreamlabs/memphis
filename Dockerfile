# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.19-alpine3.17 as build

WORKDIR $GOPATH/src/memphis-broker
COPY . .

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w" -a -o  .

FROM alpine:3.17
ENV GOPATH="/go/src"
WORKDIR /run

RUN apk update && apk add --no-cache make protobuf-dev
RUN apk add --update ca-certificates && mkdir -p /nats/bin && mkdir /nats/conf

COPY --from=build $GOPATH/memphis-broker/memphis-broker /bin/nats-server
COPY --from=build $GOPATH/memphis-broker/conf/* conf/
COPY --from=build $GOPATH/memphis-broker/version.conf .

ENTRYPOINT ["/bin/nats-server"]
