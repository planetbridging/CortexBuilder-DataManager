# Start from the latest golang base image
FROM golang:1.22.3

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Run go mod tidy to clean up the go.mod file
RUN go mod tidy

RUN go clean -modcache


# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Expose ports 4123 and 12345 to the outside world
EXPOSE 4123
EXPOSE 12345

# Command to run the Go application
CMD ["go", "run", "."]
