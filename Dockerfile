# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /artf-agent \
    ./cmd/agent

# Runtime stage
FROM scratch

# Import certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /artf-agent /artf-agent

# Create non-root user (required by ARTF spec)
# Note: scratch image doesn't support adduser, so we set USER to numeric ID
USER 65534:65534

# Expose ports
EXPOSE 50051 8080

# Health check
HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/artf-agent", "-health-check"] || exit 1

# Set entrypoint
ENTRYPOINT ["/artf-agent"]

# Default arguments
CMD ["-grpc-port=50051", "-health-port=8080"]
