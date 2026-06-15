import React, { useState } from 'react'
import { CreateTransactionRequest } from '../types/transaction'
import { useCreateTransaction } from '../hooks/useCreateTransaction'

const emptyForm: CreateTransactionRequest = {
  from: '',
  to: '',
  chain: '',
  amount: '',
  gas: '',
  gas_price: '',
}

const TransactionForm: React.FC = () => {
  const [formData, setFormData] = useState<CreateTransactionRequest>(emptyForm)
  const { create, loading, error } = useCreateTransaction()

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [event.target.name]: event.target.value })
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    await create(formData)
    setFormData(emptyForm)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label>From:</label>
        <input type="text" name="from" value={formData.from} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <div>
        <label>To:</label>
        <input type="text" name="to" value={formData.to} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <div>
        <label>Chain:</label>
        <input type="text" name="chain" value={formData.chain} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <div>
        <label>Amount:</label>
        <input type="text" name="amount" value={formData.amount} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <div>
        <label>Gas:</label>
        <input type="text" name="gas" value={formData.gas} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <div>
        <label>Gas Price:</label>
        <input type="text" name="gas_price" value={formData.gas_price} onChange={handleChange} required className="border p-2 w-full" />
      </div>
      <button type="submit" disabled={loading} className="bg-blue-500 text-white p-2">
        {loading ? 'Creating...' : 'Create Transaction'}
      </button>
      {error && <p className="text-red-500">{error}</p>}
    </form>
  )
}

export default TransactionForm
