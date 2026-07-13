# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

FROM alpine:3.20
# ca-certificates is required for TLS to MongoDB Atlas (mongodb+srv:// URIs).
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/api /usr/local/bin/api

# Railway injects PORT at runtime; internal/config reads it directly.
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
