import React from 'react'

interface OverviewCardsProps {
  totalTransactions: number
  activeBalances: number
  recentLogs: number
}

const OverviewCards: React.FC<OverviewCardsProps> = ({
  totalTransactions,
  activeBalances,
  recentLogs
}) => {
  return (
    <div className="overview-cards">
      <div className="card">
        <h3>Total Transactions</h3>
        <p>{totalTransactions}</p>
      </div>
      <div className="card">
        <h3>Active Balances</h3>
        <p>{activeBalances}</p>
      </div>
      <div className="card">
        <h3>Recent Logs</h3>
        <p>{recentLogs}</p>
      </div>
    </div>
  )
}

export default OverviewCards