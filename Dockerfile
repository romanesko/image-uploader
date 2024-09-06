# Stage 1: Build the Go application
FROM golang:1.20 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download the Go Modules. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Stage 2: Run the Go application
FROM debian:bullseye-slim

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Create an uploads directory
RUN mkdir -p /app/uploads

# Expose port 8080 to the outside world
EXPOSE 8086

# Command to run the executable
CMD ["./main"]
