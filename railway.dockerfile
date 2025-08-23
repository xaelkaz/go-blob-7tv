# Railway-optimized Dockerfile
# Uses Railway's environment variables for build info

# Build arguments automatically provided by Railway
ARG RAILWAY_GIT_COMMIT_SHA
ARG RAILWAY_GIT_COMMIT_DATE  
ARG RAILWAY_GIT_BRANCH

# Build stage
FROM golang:1.25-alpine3.19 AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata upx

WORKDIR /app

# Create build user
RUN addgroup -g 1001 -S buildgroup && \
    adduser -u 1001 -S builduser -G buildgroup

# Copy and download dependencies (better caching)
COPY --chown=builduser:buildgroup go.mod go.sum ./
USER builduser
RUN go mod download && go mod verify

# Copy source
COPY --chown=builduser:buildgroup . .

# Build with Railway metadata
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.version=${RAILWAY_GIT_COMMIT_SHA} -X main.buildDate=${RAILWAY_GIT_COMMIT_DATE} -X main.branch=${RAILWAY_GIT_BRANCH}" \
    -o gokeki main.go

# Compress binary
USER root
RUN upx --best --lzma gokeki

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

# Railway-specific labels
LABEL railway.app="gokeki"
LABEL railway.service="7tv-emote-api"

WORKDIR /app
COPY --from=builder /app/gokeki .
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

EXPOSE 8000
ENTRYPOINT ["./gokeki"]
