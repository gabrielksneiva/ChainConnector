import { useState } from 'react'
import TransactionForm from '../components/TransactionForm'
import TransactionList from '../components/TransactionList'
import TransactionDetail from '../components/TransactionDetail'
import { useTransactions } from '../hooks/useTransactions'
import { useTransactionDetail } from '../hooks/useTransactionDetail'

function Dashboard() {
  const [selectedTransactionId, setSelectedTransactionId] = useState(null)
  const { transactions, loading, error } = useTransactions()
  const { transaction: selectedTransaction } = useTransactionDetail(selectedTransactionId)

  const handleSelectTransaction = (id) => {
    setSelectedTransactionId(id)
  }

  const handleCloseDetail = () => {
    setSelectedTransactionId(null)
  }

  return (
    <div className="dashboard">
      <h1>ChainConnector Dashboard</h1>
      
      <div className="overview">
        <div className="card">
          <h3>Total Transactions</h3>
          <p>{transactions.length}</p>
        </div>
      </div>

      <div className="sections">
        <section>
          <h2>Criar Transação</h2>
          <TransactionForm />
        </section>

        <section>
          <h2>Transações Pendentes</h2>
          {loading ? <p>Loading...</p> : error ? <p>Error: {error}</p> : <TransactionList transactions={transactions} onSelectTransaction={handleSelectTransaction} />}
        </section>
      </div>

      {selectedTransaction && <TransactionDetail transaction={selectedTransaction} onClose={handleCloseDetail} />}
    </div>
  )
}

export default Dashboard