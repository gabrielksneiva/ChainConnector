import { useState, useEffect } from 'react';
import { Interest } from '../types/interest';
import { registerInterest, getInterests } from '../services/api';

export const useInterests = (pollInterval: number = 10000) => {
  const [interests, setInterests] = useState<Interest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchInterests = async () => {
    try {
      const data = await getInterests();
      setInterests(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const addInterest = async (interest: Interest) => {
    try {
      await registerInterest(interest);
      await fetchInterests(); // Refetch after adding
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    }
  };

  useEffect(() => {
    fetchInterests();
    const interval = setInterval(fetchInterests, pollInterval);
    return () => clearInterval(interval);
  }, [pollInterval]);

  return { interests, loading, error, refetch: fetchInterests, addInterest };
};