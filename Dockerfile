# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o server ./cmd/server

# Final stage - minimal image
FROM alpine:3.19

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Expose port
EXPOSE 3000

# Run the server
CMD ["./server"]
