# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod first for better caching
COPY go.mod ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o seedpoet ./cmd/bip39poem

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy binary and any static assets if needed
COPY --from=builder /app/seedpoet .
# Note: The templates are embedded in the Go binary using //go:embed in your source

EXPOSE 8080

CMD ["./seedpoet", "serve"]
