import { useState, useEffect } from 'react';
import { LogEntry } from '../types/log';
import { getLogs } from '../services/api';

interface LogFilters {
  fromBlock?: number;
  toBlock?: number;
  address?: string;
}

export const useLogs = (filters: LogFilters, pollInterval: number = 5000) => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchLogs = async () => {
    try {
      const data = await getLogs(filters);
      setLogs(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
    const interval = setInterval(fetchLogs, pollInterval);
    return () => clearInterval(interval);
  }, [filters, pollInterval]);

  return { logs, loading, error, refetch: fetchLogs };
};