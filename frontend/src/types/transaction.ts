export interface Transaction {
  id: string;
  from: string;
  to: string;
  chain: string;
  amount: string;
  gas: string;
  gasPrice: string;
  status: 'pending' | 'confirmed' | 'failed';
  hash?: string;
  createdAt: string;
  receipt?: any; // Define melhor se houver estrutura
}

export interface CreateTransactionRequest {
  from: string;
  to: string;
  chain: string;
  amount: string;
  gas: string;
  gasPrice: string;
}