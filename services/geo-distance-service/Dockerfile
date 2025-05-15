# Start from a minimal Go base image
FROM golang:1.22

# Install certificates + git (needed for module fetching)
#RUN apk add --no-cache git ca-certificates

# Install certificates + git (needed for module fetching)
RUN apt-get update && apt-get install -y git ca-certificates


# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN go build -o geo-distance-service cmd/main.go

# Expose the gRPC port
EXPOSE 50051

# Run the service
CMD ["./geo-distance-service"]
