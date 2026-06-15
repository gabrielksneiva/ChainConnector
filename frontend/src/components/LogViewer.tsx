import React, { useState } from 'react'
import { useLogs } from '../hooks/useLogs'
import LogFilters from './LogFilters'

const safeJoin = (items?: string[]) => (Array.isArray(items) && items.length > 0 ? items.join(', ') : 'N/A')
const formatData = (data: string | number[]) => (Array.isArray(data) ? data.join(', ') : data)

const LogViewer: React.FC = () => {
  const [filters, setFilters] = useState<{ fromBlock?: number; toBlock?: number; address?: string }>({})
  const { logs, loading, error } = useLogs(filters)

  return (
    <div className="log-viewer">
      <h3>Logs de Blockchain</h3>
      <LogFilters onFilterChange={setFilters} />
      {loading && <p>Loading logs...</p>}
      {error && <p>Error: {error}</p>}
      <div className="logs-list">
        {logs.map((log, index) => (
          <div key={index} className="log-entry">
            <p><strong>Bloco:</strong> {log.block_number}</p>
            <p><strong>Endereco:</strong> {log.address}</p>
            <p><strong>Topicos:</strong> {safeJoin(log.topics)}</p>
            <p><strong>Dados:</strong> {formatData(log.data)}</p>
            <p><strong>Hash da Transacao:</strong> {log.tx_hash}</p>
            <p><strong>Indice do Log:</strong> {log.log_index}</p>
          </div>
        ))}
      </div>
    </div>
  )
}

export default LogViewer
