# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies (make and git)
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy Makefile
COPY Makefile ./

# Download dependencies using make
RUN make deps

# Copy source code
COPY . .

# Generate Swagger docs and build the application using make
RUN make swagger
RUN CGO_ENABLED=0 GOOS=linux make build

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and wget for healthcheck
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy the binary from builder (Makefile outputs to bin/server)
COPY --from=builder /app/bin/server .

# Expose port
EXPOSE 8080

# Run the server
CMD ["./server"]

