import React from 'react';
import { Transaction } from '../types/transaction';
import TransactionStatus from './TransactionStatus';

interface TransactionDetailProps {
  transaction: Transaction | null;
  onClose: () => void;
}

const TransactionDetail: React.FC<TransactionDetailProps> = ({ transaction, onClose }) => {
  if (!transaction) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex justify-center items-center">
      <div className="bg-white p-6 rounded shadow-lg max-w-md w-full">
        <h2 className="text-xl mb-4">Transaction Details</h2>
        <p><strong>ID:</strong> {transaction.id}</p>
        <p><strong>From:</strong> {transaction.from}</p>
        <p><strong>To:</strong> {transaction.to}</p>
        <p><strong>Chain:</strong> {transaction.chain}</p>
        <p><strong>Amount:</strong> {transaction.amount}</p>
        <p><strong>Gas:</strong> {transaction.gas}</p>
        <p><strong>Gas Price:</strong> {transaction.gasPrice}</p>
        <p><strong>Status:</strong> <TransactionStatus status={transaction.status} /></p>
        <p><strong>Hash:</strong> {transaction.hash || 'N/A'}</p>
        <p><strong>Created At:</strong> {new Date(transaction.createdAt).toLocaleString()}</p>
        {transaction.receipt && (
          <div>
            <strong>Receipt:</strong>
            <pre className="bg-gray-100 p-2 mt-2">{JSON.stringify(transaction.receipt, null, 2)}</pre>
          </div>
        )}
        <button onClick={onClose} className="mt-4 bg-red-500 text-white p-2">Close</button>
      </div>
    </div>
  );
};

export default TransactionDetail;