# ═══════════════════════════════════════════════════════════════════════════════
# Regional Agent Dockerfile
# Go 1.21 with Raft support
# ═══════════════════════════════════════════════════════════════════════════════

FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o regional-agent .

# ═══════════════════════════════════════════════════════════════════════════════

FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/regional-agent .

# Create data directory
RUN mkdir -p /data

# Expose ports
EXPOSE 12000 12001 50052 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the agent
CMD ["./regional-agent"]
