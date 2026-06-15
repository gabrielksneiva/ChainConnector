import { useState } from 'react';
import { BalanceHistoryItem } from '../types/balance';
import { getBalanceHistory } from '../services/api';

export const useBalanceHistory = () => {
  const [history, setHistory] = useState<BalanceHistoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHistory = async (address: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await getBalanceHistory(address);
      setHistory(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return { history, loading, error, fetchHistory };
};