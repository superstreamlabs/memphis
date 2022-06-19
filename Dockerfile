FROM golang:1.18-alpine3.15 as build

WORKDIR $GOPATH/src/memphis-broker
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w" -a -o /memphis-broker

FROM alpine:3.15
ENV GOPATH="/go/src"
WORKDIR /run
COPY --from=build $GOPATH/memphis-broker/config/* config/
COPY --from=build $GOPATH/memphis-broker/version.conf .
COPY --from=build /memphis-broker memphis-broker

EXPOSE 5555

CMD ["/run/memphis-broker"]
