FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

FROM alpine:latest

WORKDIR /app

# Install necessary certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/migrate .

# Create migrations directory
RUN mkdir -p ./db/migrations

# Command to run migrations
ENTRYPOINT ["./migrate", "-command", "up"] 