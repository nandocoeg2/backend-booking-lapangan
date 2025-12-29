# Start from Golang base image
FROM golang:1.23-alpine

# Install git for fetching dependencies
RUN apk update && apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/api

# Expose port 8080
EXPOSE 8080

# Run the executable
CMD ["./main"]
