#!/bin/bash

# ChainConnector Development Setup Script

set -e

echo "🚀 ChainConnector Development Environment Setup"
echo "================================================"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose."
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ .env file created. Please review and update if necessary."
else
    echo "✅ .env file already exists"
fi

# Check if migrations directory exists and contains files
if [ -d "migrations" ] && [ "$(ls -A migrations)" ]; then
    echo "✅ Migrations directory found"
else
    echo "⚠️  No migrations found. Ensure migrations are in the ./migrations directory"
fi

# Build Docker images
echo ""
echo "🐳 Building Docker images..."
docker-compose build

# Start services
echo ""
echo "🚀 Starting services (Anvil, PostgreSQL, ChainConnector)..."
docker-compose up -d

# Wait for services to be healthy
echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check if services are running
ANVIL_READY=false
POSTGRES_READY=false
CHAINCONNECTOR_READY=false

for i in {1..30}; do
    echo -n "."
    
    if [ "$ANVIL_READY" = false ]; then
        if curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' > /dev/null 2>&1; then
            ANVIL_READY=true
            echo ""
            echo "✅ Anvil is ready at http://localhost:8545"
        fi
    fi
    
    if [ "$POSTGRES_READY" = false ]; then
        if docker-compose exec -T postgres pg_isready -U user -d chainconnector > /dev/null 2>&1; then
            POSTGRES_READY=true
            echo "✅ PostgreSQL is ready at localhost:5432"
        fi
    fi
    
    if [ "$CHAINCONNECTOR_READY" = false ]; then
        if curl -s http://localhost:3000/health > /dev/null 2>&1; then
            CHAINCONNECTOR_READY=true
            echo "✅ ChainConnector is ready at http://localhost:3000"
        fi
    fi
    
    if [ "$ANVIL_READY" = true ] && [ "$POSTGRES_READY" = true ] && [ "$CHAINCONNECTOR_READY" = true ]; then
        break
    fi
    
    sleep 1
done

echo ""
echo "================================================"
echo "✨ Development environment is ready!"
echo "================================================"
echo ""
echo "📚 Useful commands:"
echo "  docker-compose logs -f chainconnector  - View ChainConnector logs"
echo "  docker-compose logs -f anvil           - View Anvil (RPC) logs"
echo "  docker-compose down                    - Stop all services"
echo "  docker-compose ps                      - Show service status"
echo ""
echo "🔗 API Endpoints:"
echo "  Health:           http://localhost:3000/health"
echo "  Create Tx:        POST http://localhost:3000/transaction"
echo "  Register Interest: POST http://localhost:3000/interest"
echo "  List Pending:     GET http://localhost:3000/pending"
echo ""
echo "🪙 Anvil Test Accounts (10 accounts with 1000 ETH each):"
docker-compose exec -T anvil curl -s http://localhost:8545 -X POST -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' | jq '.result[]' 2>/dev/null || echo "Use: cast rpc eth_accounts --rpc-url http://localhost:8545"
echo ""
