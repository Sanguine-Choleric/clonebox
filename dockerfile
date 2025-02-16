FROM golang:1.24-alpine

WORKDIR /container

COPY . .

RUN go mod download

RUN go build -o /app/ ./cmd/web

EXPOSE 8080

CMD ["/app/web"]