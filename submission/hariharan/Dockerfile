# --- Build stage ---
FROM golang:1.26 AS builder

WORKDIR /app

# Copy go files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.19

WORKDIR /app

# Add CA certs (important for DB connections in real setups)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/server /server

# Expose port
EXPOSE 8080

# Run app
CMD ["/server"]
