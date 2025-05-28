FROM golang:1.24-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum .
# RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go mod download

COPY . .

RUN go build -o /app/ ./cmd/web
# RUN go build -gcflags="all=-N -l" -o /app/ ./cmd/web

CMD ["/app/web"]
FROM alpine:latest

WORKDIR /app
COPY --from=build-stage /app/web /app/web
# COPY --from=build-stage /go/bin/dlv /
# COPY --from=build-stage /usr/local/go/bin/go /

EXPOSE 2345

CMD ["/app/web"]
# CMD ["/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/app/web"]