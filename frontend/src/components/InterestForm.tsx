import React, { useState } from 'react';
import { Interest } from '../types/interest';
import { useInterests } from '../hooks/useInterests';

const InterestForm: React.FC = () => {
  const [addresses, setAddresses] = useState('');
  const [topics, setTopics] = useState('');
  const [txHashes, setTxHashes] = useState('');
  const { addInterest, error } = useInterests();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const interest: Interest = {
      addresses: addresses.split(',').map(a => a.trim()).filter(a => a),
      topics: topics.split(',').map(t => t.trim()).filter(t => t),
      tx_hashes: txHashes.split(',').map(h => h.trim()).filter(h => h),
    };
    await addInterest(interest);
    setAddresses('');
    setTopics('');
    setTxHashes('');
  };

  return (
    <form onSubmit={handleSubmit} className="interest-form">
      <h3>Registrar Interesse</h3>
      <div>
        <label>Endereços (separados por vírgula):</label>
        <input
          type="text"
          value={addresses}
          onChange={(e) => setAddresses(e.target.value)}
          placeholder="0x123..., 0x456..."
        />
      </div>
      <div>
        <label>Tópicos (separados por vírgula):</label>
        <input
          type="text"
          value={topics}
          onChange={(e) => setTopics(e.target.value)}
          placeholder="0xabc..., 0xdef..."
        />
      </div>
      <div>
        <label>Hashes de Transação (separados por vírgula):</label>
        <input
          type="text"
          value={txHashes}
          onChange={(e) => setTxHashes(e.target.value)}
          placeholder="0x789..., 0x012..."
        />
      </div>
      <button type="submit">Registrar</button>
      {error && <p className="error">{error}</p>}
    </form>
  );
};

export default InterestForm;