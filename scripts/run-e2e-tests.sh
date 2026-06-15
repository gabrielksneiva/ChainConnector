#!/bin/bash

# Script para rodar testes E2E do ChainConnector
# Este script:
# 1. Inicia Docker Compose
# 2. Aguarda a API estar pronta
# 3. Executa os testes Playwright
# 4. Para os containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "📦 ChainConnector E2E Test Runner"
echo "=================================="

# Function to cleanup on exit
cleanup() {
  echo ""
  echo "🧹 Limpando recursos..."
  cd "$PROJECT_ROOT"
  docker-compose down 2>/dev/null || true
}

trap cleanup EXIT

# Start Docker Compose
echo "🐳 Iniciando Docker Compose..."
cd "$PROJECT_ROOT"
docker-compose up -d

# Wait for backend to be ready
echo "⏳ Aguardando API estar pronta..."
MAX_RETRIES=60
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if curl -s http://localhost:3001/health > /dev/null 2>&1; then
    echo "✅ API pronta!"
    break
  fi
  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "  Tentativa $RETRY_COUNT/$MAX_RETRIES..."
  sleep 1
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
  echo "❌ API não ficou pronta em tempo"
  exit 1
fi

# Run Playwright tests
echo ""
echo "🧪 Executando testes E2E..."
cd "$PROJECT_ROOT/frontend"

if [ "$1" == "ui" ]; then
  npm run test:e2e:ui
elif [ "$1" == "debug" ]; then
  npm run test:e2e:debug
elif [ "$1" == "headed" ]; then
  npm run test:e2e:headed
else
  npm run test:e2e
fi

TEST_RESULT=$?

if [ $TEST_RESULT -eq 0 ]; then
  echo ""
  echo "✅ Todos os testes passaram!"
else
  echo ""
  echo "❌ Alguns testes falharam"
fi

exit $TEST_RESULT
