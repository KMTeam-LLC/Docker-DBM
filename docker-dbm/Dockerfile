# =============================================================================
# Docker-DBM - Multi-stage Dockerfile
# Powered by KMTeam LLC
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Build the Go binary
# -----------------------------------------------------------------------------
FROM golang:1.22-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static linking
# -ldflags="-s -w" to strip debug info and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o docker-dbm \
    .

# -----------------------------------------------------------------------------
# Stage 2: Create minimal runtime image
# -----------------------------------------------------------------------------
FROM alpine:latest

# Install CA certificates for potential TLS connections
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/docker-dbm .

# Use non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["/app/docker-dbm"]
