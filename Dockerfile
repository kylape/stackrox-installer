# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for dependency downloads
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the installer binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o installer ./installer

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 installer && \
    adduser -D -u 1001 -G installer installer

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/installer /usr/local/bin/installer

# Copy default configuration
COPY installer.yaml /app/installer.yaml

# Create config directory for mounted configs
RUN mkdir -p /app/config

# Set ownership
RUN chown -R installer:installer /app

# Switch to non-root user
USER installer

# Set the entrypoint to the installer binary
ENTRYPOINT ["/usr/local/bin/installer"]

# Default command shows help
CMD ["--help"]