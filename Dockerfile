# Build stage
FROM --platform=linux/amd64 golang:1.24 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go mod and sum files to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code from the current directory to the container's working directory
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM --platform=linux/amd64 alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the built Go binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8080 for the application
EXPOSE 8080

# Command to run the application
CMD ["./main"]
