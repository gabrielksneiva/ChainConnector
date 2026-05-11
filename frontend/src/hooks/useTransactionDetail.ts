import { useState, useEffect } from 'react';
import { Transaction } from '../types/transaction';
import { getTransactionDetail } from '../services/api';

export const useTransactionDetail = (id: string | null, pollInterval: number = 5000) => {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchDetail = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const data = await getTransactionDetail(id);
      setTransaction(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (id) {
      fetchDetail();
      const interval = setInterval(fetchDetail, pollInterval);
      return () => clearInterval(interval);
    }
  }, [id, pollInterval]);

  return { transaction, loading, error, refetch: fetchDetail };
};