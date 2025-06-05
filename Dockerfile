# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service .

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/wallet-service .
COPY config.env .

EXPOSE 8080

CMD ["./wallet-service"]
