# Build stage
FROM golang:1.23 AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /artf-agent \
    ./cmd/agent

# Runtime stage
FROM ubuntu:24.04

# Install CA certificates for HTTPS
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary
COPY --from=builder /artf-agent /artf-agent

# Use non-root user (nobody already exists in Ubuntu with UID 65534)
USER nobody

# Expose ports
EXPOSE 50051 8080

# Health check
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/artf-agent", "-health-check"] || exit 1

# Set entrypoint
ENTRYPOINT ["/artf-agent"]

# Default arguments
CMD ["--grpc-port=50051", "--health-port=8080"]
