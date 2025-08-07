# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the provider
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o terraform-provider-uptime .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S terraform && \
    adduser -u 1000 -S terraform -G terraform

# Copy binary from builder
COPY --from=builder /app/terraform-provider-uptime /usr/local/bin/terraform-provider-uptime

# Set ownership
RUN chown terraform:terraform /usr/local/bin/terraform-provider-uptime

# Switch to non-root user
USER terraform

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/terraform-provider-uptime"]