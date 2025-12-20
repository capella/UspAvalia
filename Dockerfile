# Build stage
FROM golang:1.24.6-alpine AS builder

# Install build dependencies for SQLite
RUN apk add --no-cache git ca-certificates gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -o uspavalia .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests (OAuth), wget for healthcheck/cron, and SQLite runtime
RUN apk --no-cache add ca-certificates wget sqlite-libs

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/uspavalia .

# Copy configuration file
COPY --from=builder /app/.uspavalia.yaml .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/matrusp ./matrusp

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Create non-root user
RUN adduser -D -s /bin/sh uspavalia

# Change ownership of app directory
RUN chown -R uspavalia:uspavalia /app

# Switch to non-root user
USER uspavalia

# Expose port
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 -O /dev/null http://localhost:8080/ || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]

# Run the application
CMD ["./uspavalia", "serve"]
