import React, { useState } from 'react'
import { Interest } from '../types/interest'
import { useInterests } from '../hooks/useInterests'

const splitCSV = (value: string) => value.split(',').map((item) => item.trim()).filter(Boolean)

const InterestForm: React.FC = () => {
  const [addresses, setAddresses] = useState('')
  const [topics, setTopics] = useState('')
  const [txHashes, setTxHashes] = useState('')
  const { addInterest, error } = useInterests()

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    const interest: Interest = {
      addresses: splitCSV(addresses),
      topics: splitCSV(topics).map((topic) => [topic]),
      tx_hashes: splitCSV(txHashes),
    }
    await addInterest(interest)
    setAddresses('')
    setTopics('')
    setTxHashes('')
  }

  return (
    <form onSubmit={handleSubmit} className="interest-form">
      <h3>Registrar Interesse</h3>
      <div>
        <label>Enderecos (separados por virgula):</label>
        <input type="text" value={addresses} onChange={(event) => setAddresses(event.target.value)} placeholder="0x123..., 0x456..." />
      </div>
      <div>
        <label>Topicos (separados por virgula):</label>
        <input type="text" value={topics} onChange={(event) => setTopics(event.target.value)} placeholder="0xabc..., 0xdef..." />
      </div>
      <div>
        <label>Hashes de Transacao (separados por virgula):</label>
        <input type="text" value={txHashes} onChange={(event) => setTxHashes(event.target.value)} placeholder="0x789..., 0x012..." />
      </div>
      <button type="submit">Registrar</button>
      {error && <p className="error">{error}</p>}
    </form>
  )
}

export default InterestForm
