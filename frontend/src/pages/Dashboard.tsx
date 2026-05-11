import React, { useState } from 'react'
import Header from '../components/Header'
import Sidebar from '../components/Sidebar'
import OverviewCards from '../components/OverviewCards'
import TransactionForm from '../components/TransactionForm'
import TransactionList from '../components/TransactionList'
import TransactionDetail from '../components/TransactionDetail'
import BalanceQuery from '../components/BalanceQuery'
import BalanceUpdate from '../components/BalanceUpdate'
import BalanceDisplay from '../components/BalanceDisplay'
import BalanceHistory from '../components/BalanceHistory'
import InterestForm from '../components/InterestForm'
import InterestList from '../components/InterestList'
import LogViewer from '../components/LogViewer'
import LogFilters from '../components/LogFilters'
import { useTransactions } from '../hooks/useTransactions'
import { useTransactionDetail } from '../hooks/useTransactionDetail'
import { useBalance } from '../hooks/useBalance'
import { useBalanceHistory } from '../hooks/useBalanceHistory'
import { useBalanceUpdate } from '../hooks/useBalanceUpdate'
import { useInterests } from '../hooks/useInterests'
import { useLogs } from '../hooks/useLogs'

const Dashboard: React.FC = () => {
  const [activeSection, setActiveSection] = useState('transactions')
  const [selectedTransactionId, setSelectedTransactionId] = useState<string | null>(null)

  const { transactions, loading: txLoading, error: txError } = useTransactions()
  const { transaction: selectedTransaction } = useTransactionDetail(selectedTransactionId || '')

  const { balance, loading: balLoading } = useBalance()
  const { history: balanceHistory } = useBalanceHistory()
  const { updateBalance } = useBalanceUpdate()

  const { interests } = useInterests()
  const { logs } = useLogs()

  const handleSelectTransaction = (id: string) => {
    setSelectedTransactionId(id)
  }

  const handleCloseDetail = () => {
    setSelectedTransactionId(null)
  }

  const renderSection = () => {
    switch (activeSection) {
      case 'transactions':
        return (
          <div className="section-content">
            <h2>Transações</h2>
            <TransactionForm />
            <TransactionList transactions={transactions} onSelectTransaction={handleSelectTransaction} />
            {selectedTransaction && <TransactionDetail transaction={selectedTransaction} onClose={handleCloseDetail} />}
          </div>
        )
      case 'balances':
        return (
          <div className="section-content">
            <h2>Saldos</h2>
            <BalanceQuery />
            <BalanceUpdate />
            <BalanceDisplay balance={balance} loading={balLoading} />
            <BalanceHistory history={balanceHistory} />
          </div>
        )
      case 'monitoring':
        return (
          <div className="section-content">
            <h2>Monitoramento</h2>
            <InterestForm />
            <InterestList interests={interests} />
            <LogFilters />
            <LogViewer logs={logs} />
          </div>
        )
      default:
        return null
    }
  }

  return (
    <div className="dashboard">
      <Header title="ChainConnector Dashboard" />
      <div className="dashboard-body">
        <Sidebar activeSection={activeSection} onSectionChange={setActiveSection} />
        <main className="main-content">
          <OverviewCards
            totalTransactions={transactions.length}
            activeBalances={balance ? 1 : 0} // Placeholder
            recentLogs={logs.length}
          />
          {renderSection()}
        </main>
      </div>
    </div>
  )
}

export default Dashboard