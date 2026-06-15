import React from 'react'

interface TransactionStatusProps {
  status?: string | number
}

const TransactionStatus: React.FC<TransactionStatusProps> = ({ status }) => {
  const normalizedStatus = String(status ?? 'unknown').toLowerCase()

  const getStatusColor = () => {
    switch (normalizedStatus) {
      case 'pending':
        return 'bg-yellow-500'
      case 'confirmed':
        return 'bg-green-500'
      case 'failed':
        return 'bg-red-500'
      default:
        return 'bg-gray-500'
    }
  }

  return (
    <span className={`px-2 py-1 text-white text-sm rounded ${getStatusColor()}`}>
      {normalizedStatus.charAt(0).toUpperCase() + normalizedStatus.slice(1)}
    </span>
  )
}

export default TransactionStatus
