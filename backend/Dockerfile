# Use Go 1.25.3 Alpine image
FROM golang:1.25.3-alpine

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the app as "server" (same as before)
RUN go build -o server .

# Expose port
EXPOSE 8080

# Start the app
CMD ["./server"]
