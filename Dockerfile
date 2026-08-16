# Build stage: compile the Go binary
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

# Copy module files first and download dependencies for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/task-manager-api .

# Runtime stage: minimal image that only runs the binary
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the build stage
COPY --from=builder /app/task-manager-api .

EXPOSE 8080

# Run the API binary as the container command
CMD ["./task-manager-api"]
