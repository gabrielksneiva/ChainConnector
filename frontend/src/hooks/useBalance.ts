import { useState } from 'react';
import { Balance } from '../types/balance';
import { getBalance } from '../services/api';

export const useBalance = () => {
  const [balance, setBalance] = useState<Balance | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchBalance = async (address: string, chain: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await getBalance(address, chain);
      setBalance(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return { balance, loading, error, fetchBalance };
};