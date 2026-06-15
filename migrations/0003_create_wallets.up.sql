-- Migration: create wallets table for on-chain transaction signing

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS wallets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  address text NOT NULL UNIQUE,
  chain text NOT NULL DEFAULT 'sepolia',
  private_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets (address);
CREATE INDEX IF NOT EXISTS idx_wallets_chain ON wallets (chain);

DROP TRIGGER IF EXISTS set_updated_at_wallets ON wallets;
CREATE TRIGGER set_updated_at_wallets
BEFORE UPDATE ON wallets
FOR EACH ROW
EXECUTE FUNCTION ecs_set_updated_at();
