# Build arguments for flexibility
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.22
ARG APP_VERSION=latest
ARG BUILD_DATE
ARG VCS_REF

# Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install necessary packages for building
RUN apk add --no-cache git ca-certificates tzdata upx

# Set working directory
WORKDIR /app

# Create non-root user for building
RUN addgroup -g 1001 -S buildgroup && \
    adduser -u 1001 -S builduser -G buildgroup

RUN chown builduser:buildgroup /app

# Copy go mod files first (better layer caching)
COPY --chown=builduser:buildgroup go.mod go.sum ./

# Download dependencies as non-root user
USER builduser
RUN go mod download && go mod verify

# Copy source code
COPY --chown=builduser:buildgroup . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${APP_VERSION} -X main.buildDate=${BUILD_DATE} -X main.commit=${VCS_REF}" \
    -o gokeki main.go

# Compress binary with UPX (optional, comment out if issues occur)
USER root
RUN upx --best --lzma gokeki

# Final stage - using distroless for better security
FROM gcr.io/distroless/static-debian12:nonroot

# Add OCI labels
LABEL org.opencontainers.image.title="Gokeki - 7TV Emote API"
LABEL org.opencontainers.image.description="A REST API for searching and managing 7TV emotes with Redis caching and Azure Storage"
LABEL org.opencontainers.image.version=${APP_VERSION}
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.source="https://github.com/yourusername/gokeki"
LABEL org.opencontainers.image.revision=${VCS_REF}
LABEL org.opencontainers.image.vendor="Your Organization"
LABEL org.opencontainers.image.licenses="MIT"
LABEL maintainer="your.email@example.com"

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/gokeki .

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Expose port
EXPOSE 8000

# Run the application
ENTRYPOINT ["./gokeki"]
