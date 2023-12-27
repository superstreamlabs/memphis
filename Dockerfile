# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.18 as build

WORKDIR $GOPATH/src/memphis
COPY . .

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w" -a -o  .

FROM alpine:3.18
ENV GOPATH="/go/src"
WORKDIR /run

RUN apk update && apk add --no-cache make protobuf-dev
RUN apk add --update ca-certificates && mkdir -p /nats/bin && mkdir /nats/conf

COPY --from=build $GOPATH/memphis/memphis /bin/nats-server
# COPY --from=build $GOPATH/memphis/conf/* conf/
COPY --from=build $GOPATH/memphis/version.conf .

ENTRYPOINT ["/bin/nats-server"]
