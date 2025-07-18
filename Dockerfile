# Build stage
FROM ubuntu:24.04 AS builder

# Install dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN curl -fsSL https://go.dev/dl/go1.24.3.linux-amd64.tar.gz | tar -xzC /usr/local
ENV PATH="/usr/local/go/bin:${PATH}"

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o proxy .

# Runtime stage
FROM ubuntu:24.04

# Install only runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy only the binary from builder stage
COPY --from=builder /app/proxy .

# Expose port 9090
EXPOSE 9090

# Run the proxy serve command
CMD ["./proxy", "serve", "--address", ":9090"]
