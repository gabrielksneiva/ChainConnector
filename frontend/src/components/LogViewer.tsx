import React, { useState } from 'react';
import { useLogs } from '../hooks/useLogs';
import LogFilters from './LogFilters';

const LogViewer: React.FC = () => {
  const [filters, setFilters] = useState<{ fromBlock?: number; toBlock?: number; address?: string }>({});
  const { logs, loading, error } = useLogs(filters);

  return (
    <div className="log-viewer">
      <h3>Logs de Blockchain</h3>
      <LogFilters onFilterChange={setFilters} />
      {loading && <p>Loading logs...</p>}
      {error && <p>Error: {error}</p>}
      <div className="logs-list">
        {logs.map((log, index) => (
          <div key={index} className="log-entry">
            <p><strong>Bloco:</strong> {log.blockNumber}</p>
            <p><strong>Endereço:</strong> {log.address}</p>
            <p><strong>Tópicos:</strong> {log.topics.join(', ')}</p>
            <p><strong>Dados:</strong> {log.data}</p>
            <p><strong>Hash da Transação:</strong> {log.transactionHash}</p>
            <p><strong>Índice do Log:</strong> {log.logIndex}</p>
            <p><strong>Removido:</strong> {log.removed ? 'Sim' : 'Não'}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

export default LogViewer;