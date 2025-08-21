# --- Build Stage ---
FROM golang:1.22-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for creating a static binary that can run in a minimal container.
# -o /app/goly-app specifies the output path for the binary.
RUN CGO_ENABLED=0 go build -o /app/goly-app goly/main.go

# --- Final Stage ---
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/goly-app .

# Expose port 3000
EXPOSE 3000

# Set the command to run the application
CMD ["./goly-app"]
