FROM golang:1.24-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /app/ ./cmd/web

FROM alpine:latest

WORKDIR /app
COPY --from=build-stage /app/web /app/web

CMD ["/app/web"]
