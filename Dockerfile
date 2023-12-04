# Use a lightweight version of Go for building the application
FROM golang:1.18-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files for dependency management
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch using Alpine for a smaller, secure base image
FROM alpine:latest  

# Add CA certificates to make HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory in the container
WORKDIR /root/

# Copy the compiled application from the builder stage
COPY --from=builder /app/main .

# Copy static assets, templates, and data from the build stage
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/data ./data

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
