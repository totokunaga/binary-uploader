FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

# Install necessary certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Create the storage directory
RUN mkdir -p /app/storage

# Expose the port (this will be overridden by docker-compose)
EXPOSE 8080

# Run the server
CMD ["./server"] 