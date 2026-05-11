import React, { useState, useEffect } from 'react';
import { useBalanceHistory } from '../hooks/useBalanceHistory';

const BalanceHistory: React.FC = () => {
  const [address, setAddress] = useState('');
  const { history, loading, error, fetchHistory } = useBalanceHistory();

  const isValidAddress = (addr: string) => /^0x[a-fA-F0-9]{40}$/.test(addr);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isValidAddress(address)) {
      alert('Invalid Ethereum address');
      return;
    }
    fetchHistory(address);
  };

  return (
    <div className="balance-history">
      <h2>Balance History</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label>Address:</label>
          <input
            type="text"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            required
            className="border p-2 w-full"
          />
        </div>
        <button type="submit" disabled={loading} className="bg-purple-500 text-white p-2">
          {loading ? 'Loading...' : 'Fetch History'}
        </button>
      </form>
      {error && <p className="text-red-500">{error}</p>}
      {history.length > 0 && (
        <ul className="space-y-2">
          {history.map((item, index) => (
            <li key={index} className="border p-2">
              <p>Amount: {item.amount}</p>
              <p>Chain: {item.chain}</p>
              <p>Timestamp: {item.timestamp}</p>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default BalanceHistory;