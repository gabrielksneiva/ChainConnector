-- Migration: create networks table for registered EVM networks

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS networks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL UNIQUE,
  chain_id bigint NOT NULL UNIQUE,
  rpc_url text,
  currency_symbol text,
  explorer_url text,
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_networks_chain_id ON networks (chain_id);
CREATE INDEX IF NOT EXISTS idx_networks_enabled ON networks (enabled);

DROP TRIGGER IF EXISTS set_updated_at_networks ON networks;
CREATE TRIGGER set_updated_at_networks
BEFORE UPDATE ON networks
FOR EACH ROW
EXECUTE FUNCTION ecs_set_updated_at();
