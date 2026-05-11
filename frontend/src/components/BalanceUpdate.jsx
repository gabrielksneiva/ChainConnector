import React, { useState } from 'react';
import { useBalanceUpdate } from '../hooks/useBalanceUpdate';

const BalanceUpdate: React.FC = () => {
  const [address, setAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [chain, setChain] = useState('');
  const { update, loading, error } = useBalanceUpdate();

  const isValidAddress = (addr: string) => /^0x[a-fA-F0-9]{40}$/.test(addr);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!isValidAddress(address)) {
      alert('Invalid Ethereum address');
      return;
    }
    await update(address, { amount, chain });
    // Reset form
    setAddress('');
    setAmount('');
    setChain('');
  };

  return (
    <div className="balance-update">
      <h2>Update Balance</h2>
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
          <label>Amount:</label>
          <input
            type="text"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
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
        <button type="submit" disabled={loading} className="bg-green-500 text-white p-2">
          {loading ? 'Updating...' : 'Update Balance'}
        </button>
      </form>
      {error && <p className="text-red-500">{error}</p>}
    </div>
  );
};

export default BalanceUpdate;