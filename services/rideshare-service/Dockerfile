# Build stage
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o rideshare-api ./cmd/server

# Final stage
FROM alpine:3.18

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/rideshare-api .

# Expose port
EXPOSE 8080

# Set entry point
ENTRYPOINT ["/app/rideshare-api"]