import { useState, useEffect } from 'react';
import { Transaction } from '../types/transaction';
import { listPendingTransactions } from '../services/api';

export const useTransactions = (pollInterval: number = 5000) => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchTransactions = async () => {
    try {
      const data = await listPendingTransactions();
      setTransactions(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTransactions();
    const interval = setInterval(fetchTransactions, pollInterval);
    return () => clearInterval(interval);
  }, [pollInterval]);

  return { transactions, loading, error, refetch: fetchTransactions };
};
