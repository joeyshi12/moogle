FROM golang:1.24-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy Go modules and files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the application
COPY . .

# Build the application
RUN go build -o page-rank ./cmd/page-rank/main.go

# Run tests
RUN go test -v ./...

# Use a smaller base image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy thhe built binary from the builder stage
COPY --from=builder /app/page-rank .

# EXPOSE ports?
# EXPOSE 8080

# Run the application
CMD ["./page-rank"]
