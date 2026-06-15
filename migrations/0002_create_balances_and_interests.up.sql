-- Migration: create user_balances and interest_addresses tables
-- Dependências: pgcrypto extension

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- User balances table for offchain balance tracking
CREATE TABLE IF NOT EXISTS user_balances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  address text NOT NULL,
  chain text NOT NULL DEFAULT 'sepolia',
  balance numeric NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(address, chain)
);

-- Interest addresses table for monitoring
CREATE TABLE IF NOT EXISTS interest_addresses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  address text NOT NULL,
  chain text NOT NULL DEFAULT 'sepolia',
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(address, chain)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_user_balances_address ON user_balances (address);
CREATE INDEX IF NOT EXISTS idx_user_balances_chain ON user_balances (chain);
CREATE INDEX IF NOT EXISTS idx_interest_addresses_address ON interest_addresses (address);
CREATE INDEX IF NOT EXISTS idx_interest_addresses_chain ON interest_addresses (chain);

-- Triggers for updated_at
DROP TRIGGER IF EXISTS set_updated_at_balances ON user_balances;
CREATE TRIGGER set_updated_at_balances
BEFORE UPDATE ON user_balances
FOR EACH ROW
EXECUTE FUNCTION ecs_set_updated_at();

DROP TRIGGER IF EXISTS set_updated_at_interest ON interest_addresses;
CREATE TRIGGER set_updated_at_interest
BEFORE UPDATE ON interest_addresses
FOR EACH ROW
EXECUTE FUNCTION ecs_set_updated_at();