import React, { useState } from 'react';

interface LogFiltersProps {
  onFilterChange: (filters: { fromBlock?: number; toBlock?: number; address?: string }) => void;
}

const LogFilters: React.FC<LogFiltersProps> = ({ onFilterChange }) => {
  const [fromBlock, setFromBlock] = useState('');
  const [toBlock, setToBlock] = useState('');
  const [address, setAddress] = useState('');

  const handleApplyFilters = () => {
    const filters = {
      fromBlock: fromBlock ? parseInt(fromBlock) : undefined,
      toBlock: toBlock ? parseInt(toBlock) : undefined,
      address: address || undefined,
    };
    onFilterChange(filters);
  };

  return (
    <div className="log-filters">
      <h3>Filtros de Logs</h3>
      <div>
        <label>Bloco Inicial:</label>
        <input
          type="number"
          value={fromBlock}
          onChange={(e) => setFromBlock(e.target.value)}
          placeholder="Ex: 1000000"
        />
      </div>
      <div>
        <label>Bloco Final:</label>
        <input
          type="number"
          value={toBlock}
          onChange={(e) => setToBlock(e.target.value)}
          placeholder="Ex: 2000000"
        />
      </div>
      <div>
        <label>Endereço:</label>
        <input
          type="text"
          value={address}
          onChange={(e) => setAddress(e.target.value)}
          placeholder="0x123..."
        />
      </div>
      <button onClick={handleApplyFilters}>Aplicar Filtros</button>
    </div>
  );
};

export default LogFilters;