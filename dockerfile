FROM golang:1.24-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

RUN go build -o /app/ ./cmd/web

FROM alpine:latest

WORKDIR /app
COPY --from=build-stage /app/web /app/web
EXPOSE 2345

# CMD ["/app/web", "-debug"]
CMD ["/app/web"]
