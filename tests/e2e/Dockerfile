FROM golang:1.24-alpine

WORKDIR /app

# Install build dependencies if needed (e.g. gcc for CGO)
RUN apk add --no-cache gcc musl-dev

# Copy module files first for caching
COPY go.work go.work.sum ./
COPY core/go.mod core/go.sum ./core/
COPY cli/go.mod cli/go.sum ./cli/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Default command (overridden by docker-compose)
CMD ["go", "test", "-v", "./tests/e2e/..."]
