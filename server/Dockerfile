FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application for production
# Build the entire package in the cmd directory
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-w -s' -o server ./cmd

# --- Final Production Stage ---
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/server .

# Create the storage directory (if needed, adjust path as necessary)
RUN mkdir -p /app/storage

EXPOSE 8080

# Command to run the executable
CMD ["./server"] 