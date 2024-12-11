# Step 1: Build the Go binary
FROM golang:1.23 as builder

# Set the working directory for the build process
WORKDIR /app

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN GOOS=linux GOARCH=amd64 go build -o /app/main .

# Step 2: Create the final container image
FROM debian:bullseye-slim

# Install dependencies needed to run the Go binary (e.g., libc)
RUN apt-get update && apt-get install -y ca-certificates

# Set the working directory
WORKDIR /root

# Copy the compiled binary from the builder image
COPY --from=builder /app/main /root/main

# Make sure the binary is executable
RUN chmod +x /root/main

# Set the entrypoint for the container
CMD ["/root/main"]
