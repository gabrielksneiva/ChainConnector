import React from 'react'
import { Wallet } from '../types/wallet'

interface WalletListProps {
  wallets: Wallet[]
  loading: boolean
}

const WalletList: React.FC<WalletListProps> = ({ wallets, loading }) => {
  if (loading) {
    return <p>Carregando carteiras...</p>
  }

  if (!wallets.length) {
    return <p className="empty-state">Nenhuma carteira encontrada. Use o formulario acima para criar sua primeira carteira.</p>
  }

  return (
    <div className="wallet-list section-block">
      <h3>Carteiras salvas</h3>
      <div className="wallet-grid">
        {wallets.map((wallet) => (
          <div key={wallet.id} className="wallet-card">
            <div className="wallet-card-header">
              <span>Chain</span>
              <strong>{wallet.chain}</strong>
            </div>
            <p className="wallet-address">{wallet.address}</p>
            <div className="wallet-meta">
              <span>ID: {wallet.id}</span>
              <span>{wallet.created_at ? new Date(wallet.created_at).toLocaleString() : 'Sem data'}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default WalletList
