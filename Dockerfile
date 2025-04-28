FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o reverse-soxy ./cmd/reverse-soxy

# Create a minimal image for running the application
FROM alpine:latest

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/reverse-soxy .

# Create a directory for configuration files
RUN mkdir -p /app/config && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
# SOCKS5 proxy port
EXPOSE 1080
# Tunnel listen port
EXPOSE 9000
# Relay listen port
EXPOSE 9000

# Set the entrypoint
ENTRYPOINT ["/app/reverse-soxy"]

# Default command (can be overridden)
CMD ["--proxy-listen-addr", "0.0.0.0:1080", "--tunnel-listen-port", "9000", "--secret", "changeme"]
