FROM golang:1.24.0 as build
WORKDIR /exporter
COPY go.mod go.sum /exporter/
RUN go mod download
COPY . .
RUN export CGO_ENABLED=0 && make build

FROM alpine:3.21.3
RUN apk update && apk add --no-cache ca-certificates
RUN mkdir /exporter
COPY --from=build /exporter/bin/exporter /exporter
WORKDIR /exporter
USER nobody
ENTRYPOINT  [ "/exporter/exporter" ]
