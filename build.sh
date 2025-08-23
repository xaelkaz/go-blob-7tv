#!/bin/bash

# Build script for Gokeki Docker image
# Usage: ./scripts/build.sh [tag] [--push]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
TAG=${1:-latest}
PUSH=${2}
IMAGE_NAME="gokeki"

# Build arguments
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
APP_VERSION=${TAG}

echo -e "${GREEN}Building Gokeki Docker image...${NC}"
echo "Tag: $TAG"
echo "Build Date: $BUILD_DATE"
echo "VCS Ref: $VCS_REF"
echo ""

# Build the image
docker build \
  --build-arg APP_VERSION="$APP_VERSION" \
  --build-arg BUILD_DATE="$BUILD_DATE" \
  --build-arg VCS_REF="$VCS_REF" \
  --tag "$IMAGE_NAME:$TAG" \
  --tag "$IMAGE_NAME:latest" \
  .

echo -e "${GREEN}✅ Build completed successfully!${NC}"

# Show image size
IMAGE_SIZE=$(docker images "$IMAGE_NAME:$TAG" --format "table {{.Size}}" | tail -n 1)
echo -e "${YELLOW}Image size: $IMAGE_SIZE${NC}"

# Push if requested
if [ "$PUSH" = "--push" ]; then
  echo -e "${GREEN}Pushing image to registry...${NC}"
  docker push "$IMAGE_NAME:$TAG"
  if [ "$TAG" != "latest" ]; then
    docker push "$IMAGE_NAME:latest"
  fi
  echo -e "${GREEN}✅ Push completed!${NC}"
fi

echo ""
echo -e "${GREEN}Available commands:${NC}"
echo "  Run locally: docker run -p 8000:8000 --env-file .env $IMAGE_NAME:$TAG"
echo "  With compose: docker-compose up"
echo "  Check image:  docker inspect $IMAGE_NAME:$TAG"
