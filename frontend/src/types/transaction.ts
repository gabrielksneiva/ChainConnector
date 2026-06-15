export interface Transaction {
  id: string;
  from: string;
  to?: string;
  chain?: string;
  value?: string;
  gas?: number;
  gas_price?: string;
  status?: string | number;
  tx_hash?: string;
  created_at?: string;
  updated_at?: string;
  receipt?: unknown; // Define melhor se houver estrutura
  error_message?: string;
}

export interface CreateTransactionRequest {
  from: string;
  to: string;
  chain: string;
  amount: string;
  gas: string;
  gas_price: string;
  wallet_id?: string;
}
