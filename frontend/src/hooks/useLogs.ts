import { useCallback, useEffect, useState } from 'react';
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

  const fetchLogs = useCallback(async () => {
    try {
      const data = await getLogs(filters);
      setLogs(Array.isArray(data) ? data : []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchLogs();
    const interval = setInterval(fetchLogs, pollInterval);
    return () => clearInterval(interval);
  }, [fetchLogs, pollInterval]);

  return { logs, loading, error, refetch: fetchLogs };
};
