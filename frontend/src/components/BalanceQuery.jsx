import React, { useState } from 'react';
import { useBalance } from '../hooks/useBalance';

const BalanceQuery: React.FC = () => {
  const [address, setAddress] = useState('');
  const [chain, setChain] = useState('');
  const { balance, loading, error, fetchBalance } = useBalance();

  const isValidAddress = (addr: string) => /^0x[a-fA-F0-9]{40}$/.test(addr);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isValidAddress(address)) {
      alert('Invalid Ethereum address');
      return;
    }
    fetchBalance(address, chain);
  };

  return (
    <div className="balance-query">
      <h2>Query Balance</h2>
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
        <div>
          <label>Chain:</label>
          <input
            type="text"
            value={chain}
            onChange={(e) => setChain(e.target.value)}
            required
            className="border p-2 w-full"
          />
        </div>
        <button type="submit" disabled={loading} className="bg-blue-500 text-white p-2">
          {loading ? 'Loading...' : 'Query Balance'}
        </button>
      </form>
      {error && <p className="text-red-500">{error}</p>}
      {balance && (
        <div>
          <h3>Balance</h3>
          <p>Address: {balance.address}</p>
          <p>Chain: {balance.chain}</p>
          <p>Amount: {balance.amount}</p>
        </div>
      )}
    </div>
  );
};

export default BalanceQuery;