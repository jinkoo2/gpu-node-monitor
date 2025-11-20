# Use the official Golang image to build the application (Builder Stage)
FROM golang:1.22-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the go module and source files
COPY go.mod go.sum ./
COPY main.go .

# Download Go modules
RUN go mod download

# Build the application
# We use CGO_ENABLED=0 to create a statically linked binary for the final stage
RUN go build -ldflags "-s -w" -o /alertmanager-adapter .

# Use a minimal Alpine image for the final, small runtime image
FROM alpine:latest

# Expose the port the adapter listens on
EXPOSE 8080

# Copy the built binary from the builder stage
COPY --from=builder /alertmanager-adapter /usr/local/bin/alertmanager-adapter

# Set the entry point to run the application
CMD ["alertmanager-adapter"]