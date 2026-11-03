# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
# The website (web/dist) is embedded in the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/seedpoet ./cmd/seedpoet

# Final stage: the binary only, as an unprivileged user
FROM alpine:latest
RUN adduser -D -H -u 10001 seedpoet
COPY --from=builder /out/seedpoet /usr/local/bin/seedpoet
USER seedpoet
EXPOSE 8080
CMD ["seedpoet", "serve", "0.0.0.0:8080"]
