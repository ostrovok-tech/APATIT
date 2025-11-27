FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION="v2.0.0"
RUN go build -v -ldflags="-s -w -X 'ping-admin-exporter/internal/version.Version=${VERSION}'" -o /app/ping-admin-exporter ./cmd/ping-admin-exporter

FROM alpine:3.20
ARG VERSION="latest"

RUN apk add --no-cache ca-certificates

LABEL org.opencontainers.image.title="Ping-Admin Exporter"
LABEL org.opencontainers.image.description="Prometheus Exporter for https://ping-admin.com/"
LABEL org.opencontainers.image.source="https://github.com/ostrovok-tech/ping-admin-exporter"
LABEL org.opencontainers.image.version="${VERSION}"

COPY --from=build /app/ping-admin-exporter /usr/local/bin/ping-admin-exporter
COPY locations.json /app/locations.json
WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser
EXPOSE 8080
ENV LISTEN_ADDRESS=:8080

ENTRYPOINT [ "ping-admin-exporter" ]