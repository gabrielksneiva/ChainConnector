import React from 'react';
import { Balance } from '../types/balance';

interface BalanceDisplayProps {
  balance: Balance | null;
}

const BalanceDisplay: React.FC<BalanceDisplayProps> = ({ balance }) => {
  if (!balance) {
    return <p>No balance data available.</p>;
  }

  return (
    <div className="balance-display">
      <h3>Current Balance</h3>
      <p>Address: {balance.address}</p>
      <p>Chain: {balance.chain}</p>
      <p>Amount: {balance.amount}</p>
    </div>
  );
};

export default BalanceDisplay;