FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION="v2.0.0"
RUN go build -v -ldflags="-s -w -X 'apatit/internal/version.Version=${VERSION}'" -o /app/apatit ./cmd/apatit

FROM alpine:3.20
ARG VERSION="latest"

RUN apk add --no-cache ca-certificates

LABEL org.opencontainers.image.title="APATIT (Advanced Ping-Admin Tasks Indicators Transducer)"
LABEL org.opencontainers.image.description="Transducer for Tasks Indicators from https://ping-admin.com/"
LABEL org.opencontainers.image.source="https://github.com/ostrovok-tech/apatit"
LABEL org.opencontainers.image.version="${VERSION}"

COPY --from=build /app/apatit /usr/local/bin/apatit
COPY locations.json /app/locations.json
WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser
EXPOSE 8080
ENV LISTEN_ADDRESS=:8080

ENTRYPOINT [ "apatit" ]