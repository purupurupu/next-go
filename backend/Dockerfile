# Development Dockerfile for Go backend
FROM golang:1.22-bookworm

# Install air for hot reloading
RUN go install github.com/air-verse/air@latest

# Install PostgreSQL client (for debugging)
RUN apt-get update && apt-get install -y postgresql-client && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod first for better caching
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Run with air for hot reloading
CMD ["air", "-c", ".air.toml"]
