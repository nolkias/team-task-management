FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY app/go.mod app/go.sum* ./
RUN go mod download

COPY app/ .
RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/api ./api
COPY app/migrations ./migrations
COPY app/config.yaml ./config.yaml

EXPOSE 8080

CMD ["./api"]
