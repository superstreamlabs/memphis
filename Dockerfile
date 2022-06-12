FROM golang:1.18-alpine3.15 as build

WORKDIR $GOPATH/src/memphis-broker
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w" -a -o /memphis-broker

FROM alpine:3.15
# COPY --from=build . .
# COPY --from=build $GOPATH/src/memphis-broker/config/* .
# COPY --from=build $GOPATH/src/memphis-broker/version.conf .
COPY --from=build /memphis-broker /memphis-broker

EXPOSE 5555

CMD ["/memphis-broker"]
