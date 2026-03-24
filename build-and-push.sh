#!/bin/bash
set -e

REGISTRY="registry.germainleignel.com/personal"
IMAGE_NAME="rss-digest"
TAG="${1:-latest}"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${TAG}"

echo "🔨 Building Docker image..."
docker build -t "${FULL_IMAGE}" .

echo "📤 Pushing to registry..."
docker push "${FULL_IMAGE}"

echo ""
echo "✅ Image pushed: ${FULL_IMAGE}"
echo ""
echo "To deploy to k8s:"
echo "  kubectl apply -k k8s/"
echo ""
echo "To update the deployment:"
echo "  kubectl rollout restart deployment/rss-digest"
