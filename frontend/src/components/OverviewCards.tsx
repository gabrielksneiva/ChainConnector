import React from 'react'

interface OverviewCardsProps {
  totalTransactions: number
  activeBalances: number
  recentLogs: number
  activeWallets: number
}

const OverviewCards: React.FC<OverviewCardsProps> = ({
  totalTransactions,
  activeBalances,
  recentLogs,
  activeWallets,
}) => {
  return (
    <div className="overview-cards">
      <div className="card">
        <h3>Total de transacoes</h3>
        <p>{totalTransactions}</p>
        <span className="section-pill">Historico completo</span>
      </div>
      <div className="card">
        <h3>Carteiras</h3>
        <p>{activeWallets}</p>
        <span className="section-pill">Enderecos ativos</span>
      </div>
      <div className="card">
        <h3>Saldos</h3>
        <p>{activeBalances}</p>
        <span className="section-pill">Consultas locais</span>
      </div>
      <div className="card">
        <h3>Logs recentes</h3>
        <p>{recentLogs}</p>
        <span className="section-pill">Atualizacoes</span>
      </div>
    </div>
  )
}

export default OverviewCards
