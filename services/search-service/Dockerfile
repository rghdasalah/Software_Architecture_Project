# Start from an official Go base image
FROM golang:1.22

# Install Git (required for go get) and CA certificates
#RUN apk add --no-cache git ca-certificates
RUN apt-get update && apt-get install -y git ca-certificates



# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the app
COPY . .

# Build the Go binary
RUN go build -o search-service cmd/main.go

# Expose the HTTP port
EXPOSE 8080

# Command to run the binary
CMD ["./search-service"]
