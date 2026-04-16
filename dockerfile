FROM golang:1.26-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /app/ ./cmd/web

FROM alpine:latest

WORKDIR /app
COPY --from=build-stage /app/web /app/web

CMD ["/app/web"]
