#!/bin/bash

# ChainConnector Development Setup with Automatic E2E Testing
# This script runs E2E tests first, then starts the main environment only if tests pass

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  ChainConnector - Development + Test Validation${NC}"
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo ""

# Cleanup function
cleanup() {
  if [ "$TEST_FAILED" = true ]; then
    echo -e "${RED}🧹 Cleaning up due to test failure...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose down --remove-orphans 2>/dev/null || true
  fi
}

trap cleanup EXIT

TEST_FAILED=false

# Step 1: Build images
echo -e "${YELLOW}📦 Building Docker images...${NC}"
cd "$PROJECT_ROOT"
docker-compose build 2>&1 | grep -E "Building|FINISHED" | head -5
echo ""

# Step 2: Start infrastructure
echo -e "${YELLOW}🐳 Starting services...${NC}"
docker-compose up -d 2>&1 | grep -E "Started|Running"
echo ""

# Step 3: Wait for services
echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
MAX_RETRIES=60
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if curl -s http://localhost:3001/health > /dev/null 2>&1 && \
     curl -s http://localhost:8080/ > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Services ready!${NC}"
    break
  fi
  RETRY_COUNT=$((RETRY_COUNT + 1))
  printf "  Waiting... ($RETRY_COUNT/$MAX_RETRIES)\r"
  sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo -e "${RED}❌ Services failed to start${NC}"
  TEST_FAILED=true
  exit 1
fi

echo ""

# Step 4: Run E2E tests
echo -e "${YELLOW}🧪 Running E2E tests...${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

cd "$PROJECT_ROOT/frontend"

if docker run --rm \
  --network host \
  -v "$PROJECT_ROOT/frontend/playwright-report:/app/playwright-report" \
  chainconnector-e2e 2>&1; then
  
  echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo ""
  echo -e "${GREEN}✅ All E2E tests passed!${NC}"
  echo ""
  echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
  echo -e "${GREEN}  🚀 Development environment is ready!${NC}"
  echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
  echo ""
  echo -e "${YELLOW}URLs:${NC}"
  echo "  Frontend:  http://localhost:8080"
  echo "  Backend:   http://localhost:3001"
  echo "  API:       http://localhost:8080/api"
  echo ""
  echo -e "${YELLOW}Commands:${NC}"
  echo "  View test report:  cd frontend && npx playwright show-report"
  echo "  Stop services:     docker-compose down"
  echo ""
  
else
  echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo ""
  echo -e "${RED}❌ E2E tests FAILED!${NC}"
  echo ""
  echo -e "${RED}════════════════════════════════════════════════════════════${NC}"
  echo -e "${RED}  Services are STOPPED - fix tests before trying again${NC}"
  echo -e "${RED}════════════════════════════════════════════════════════════${NC}"
  echo ""
  echo -e "${YELLOW}Debug:${NC}"
  echo "  View logs:         docker-compose logs -f"
  echo "  View test report:  cd frontend && npx playwright show-report"
  echo ""
  
  TEST_FAILED=true
  exit 1
fi
