# Dockerfile inside /services/auth-service

FROM golang:1.21

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o main ./cmd/main.go

CMD ["./main"]
