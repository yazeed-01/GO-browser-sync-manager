# Use an official Go runtime as a parent image (builder stage)
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go Modules manifests to the container
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the entire Go project to the container
COPY . .

# Build the Go app with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o main .

# Start a new image from scratch (lighter image)
FROM alpine:latest

# Install necessary dependencies
RUN apk --no-cache add ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the compiled binary from the builder image
COPY --from=builder /app/main .

# Copy the templates folder into the container
COPY --from=builder /app/templates /root/templates

# Copy the links.json file into the container
COPY --from=builder /app/links.json /root/links.json

# Copy the email.json file into the container
COPY --from=builder /app/emails.json /root/emails.json

# Copy the extensions.json file into the container
COPY --from=builder /app/extensions.json /root/extensions.json

# Set GIN_MODE to release to optimize performance
ENV GIN_MODE=release

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
