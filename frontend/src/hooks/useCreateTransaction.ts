import { useState } from 'react';
import { Transaction, CreateTransactionRequest } from '../types/transaction';
import { createTransaction as apiCreateTransaction } from '../services/api';

export const useCreateTransaction = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const create = async (data: CreateTransactionRequest): Promise<Transaction | null> => {
    setLoading(true);
    setError(null);
    try {
      const transaction = await apiCreateTransaction(data);
      return transaction;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
      return null;
    } finally {
      setLoading(false);
    }
  };

  return { create, loading, error };
};