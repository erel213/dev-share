# Build stage
FROM golang:1.24.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and runtime dependencies.
# Secrets are fetched on the host and mounted into the container via tmpfs
# (see scripts/fetch-secrets.sh + docker-compose.cloud-secrets.yml), so no
# cloud CLI is needed at runtime.
RUN apk --no-cache add ca-certificates curl sqlite unzip

# Install Terraform from HashiCorp releases
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; elif [ "$ARCH" = "aarch64" ]; then ARCH="arm64"; fi && \
    curl -fsSL "https://releases.hashicorp.com/terraform/1.12.0/terraform_1.12.0_linux_${ARCH}.zip" -o /tmp/terraform.zip && \
    unzip /tmp/terraform.zip -d /usr/local/bin/ && \
    rm /tmp/terraform.zip
RUN echo 'alias sqlite3="sqlite3 -cmd \".headers on\" -cmd \".mode table\""' >> /root/.profile

WORKDIR /root/
ENV HOME=/root

# Copy binaries from builder
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Copy migration files
COPY --from=builder /app/internal/infra/migrations/sqlite ./internal/infra/migrations/sqlite

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose port
EXPOSE 8080

# Run migrations then start server
CMD ["./entrypoint.sh"]
