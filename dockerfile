# Use the official Golang Alpine image as the base image
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-simulator .

# Declare the final step
# Still use the official Golang Docker image
FROM alpine:3.22

# Copy the build result from the previous step
# into the /usr/local/bin directory
COPY --from=builder /app/api-simulator /usr/local/bin/api-simulator

# Copy the api-data.json into the /data directory
COPY --from=builder /app/data/api-data.json /data/api-data.json

# Declare a volume for persistent data storage
VOLUME ["/data"]

# Expose port 8800 for the application
EXPOSE 8800

# Command to run the executable
ENTRYPOINT ["api-simulator"]