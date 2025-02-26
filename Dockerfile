# Use Go 1.22 as the builder image
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Install air for live reloading
RUN go install github.com/cosmtrek/air@v1.42.0

# Copy go.mod and go.sum to download dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy application source code
COPY backend .

# Build the application binary
RUN go build -o /app/main .

# Use a lightweight Go runtime for the final container
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Install dependencies (InSpec, PostgreSQL client)
RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    postgresql-client \
    && curl https://omnitruck.chef.io/install.sh | bash -s -- -P inspec \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy built Go binary and air from the builder stage
COPY --from=builder /go/bin/air /usr/local/bin/air
COPY --from=builder /app/main /app/main
COPY --from=builder /app/docs /app/docs

# Expose the application port
EXPOSE 8080

# Use air for live reloading
CMD ["air"]
