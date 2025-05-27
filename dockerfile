FROM golang:1.24 AS build-stage

WORKDIR /container

COPY . .

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go mod download

# RUN go build -o /app/ ./cmd/web
RUN go build -gcflags="all=-N -l" -o /app/ ./cmd/web

FROM debian:bookworm

WORKDIR /
COPY --from=build-stage /app/web /app/web
COPY --from=build-stage /go/bin/dlv /

EXPOSE 2345

# CMD ["/app/web"]
CMD ["/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/app/web"]