import React from 'react';
import { Transaction } from '../types/transaction';
import TransactionStatus from './TransactionStatus';

interface TransactionListProps {
  transactions: Transaction[];
  onSelectTransaction: (id: string) => void;
}

const TransactionList: React.FC<TransactionListProps> = ({ transactions, onSelectTransaction }) => {
  return (
    <table className="w-full border-collapse border border-gray-300">
      <thead>
        <tr>
          <th className="border border-gray-300 p-2">Status</th>
          <th className="border border-gray-300 p-2">Hash</th>
          <th className="border border-gray-300 p-2">Amount</th>
          <th className="border border-gray-300 p-2">Date</th>
          <th className="border border-gray-300 p-2">Actions</th>
        </tr>
      </thead>
      <tbody>
        {transactions.map((tx) => (
          <tr key={tx.id}>
            <td className="border border-gray-300 p-2">
              <TransactionStatus status={tx.status} />
            </td>
            <td className="border border-gray-300 p-2">{tx.hash || 'N/A'}</td>
            <td className="border border-gray-300 p-2">{tx.amount}</td>
            <td className="border border-gray-300 p-2">{new Date(tx.createdAt).toLocaleString()}</td>
            <td className="border border-gray-300 p-2">
              <button onClick={() => onSelectTransaction(tx.id)} className="text-blue-500">
                View Details
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
};

export default TransactionList;