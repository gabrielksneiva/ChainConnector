# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o chainconnector \
    ./cmd/chainconnector

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies only
RUN apk add --no-cache ca-certificates postgresql-client curl && \
    addgroup -S app && \
    adduser -S app -G app && \
    rm -rf /var/cache/apk/*

WORKDIR /app

# Copy the built application from builder
COPY --from=builder /app/chainconnector .

# Copy migrations if they exist
COPY migrations ./migrations

RUN chown -R app:app /app
USER app

# Expose the HTTP server port
EXPOSE 3000

# Set environment defaults
ENV HTTP_ADDR=0.0.0.0:3000 \
    LOG_LEVEL=info \
    MONITOR_ENABLED=true

# Health check
HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --retries=3 \
    CMD curl -f http://localhost:3000/health --silent --show-error || exit 1

# Run the application
CMD ["./chainconnector"]
