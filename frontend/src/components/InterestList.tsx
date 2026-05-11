import React from 'react';
import { useInterests } from '../hooks/useInterests';

const InterestList: React.FC = () => {
  const { interests, loading, error } = useInterests();

  if (loading) return <p>Loading interests...</p>;
  if (error) return <p>Error: {error}</p>;

  return (
    <div className="interest-list">
      <h3>Interesses Registrados</h3>
      {interests.length === 0 ? (
        <p>Nenhum interesse registrado.</p>
      ) : (
        <ul>
          {interests.map((interest, index) => (
            <li key={index}>
              <strong>Endereços:</strong> {interest.addresses.join(', ')}<br />
              <strong>Tópicos:</strong> {interest.topics.join(', ')}<br />
              <strong>Hashes:</strong> {interest.tx_hashes.join(', ')}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default InterestList;