import { useState } from 'react';
import { BalanceUpdateRequest } from '../types/balance';
import { updateBalance } from '../services/api';

export const useBalanceUpdate = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const update = async (address: string, data: BalanceUpdateRequest) => {
    setLoading(true);
    setError(null);
    try {
      await updateBalance(address, data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return { update, loading, error };
};