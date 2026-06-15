import React, { useState } from 'react'
import { CreateWalletRequest } from '../types/wallet'

interface WalletFormProps {
  onCreate: (data: CreateWalletRequest) => Promise<boolean>
  loading: boolean
  error: string | null
}

const WalletForm: React.FC<WalletFormProps> = ({ onCreate, loading, error }) => {
  const [formData, setFormData] = useState<CreateWalletRequest>({ chain: 'sepolia' })
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const handleChange = (event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [event.target.name]: event.target.value })
    setSuccessMessage(null)
  }

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    const success = await onCreate(formData)
    if (success) {
      setSuccessMessage('Carteira criada com sucesso.')
      setFormData({ chain: formData.chain })
    }
  }

  return (
    <section className="wallet-form section-block">
      <div className="section-header">
        <div>
          <h3>Criar ou importar carteira</h3>
          <p>Gere uma nova carteira Sepolia ou importe uma existente com sua chave privada.</p>
        </div>
      </div>
      <form className="form-grid" onSubmit={handleSubmit}>
        <div>
          <label>Chain</label>
          <select name="chain" value={formData.chain} onChange={handleChange} className="input-field">
            <option value="sepolia">Sepolia</option>
            <option value="mainnet">Mainnet</option>
          </select>
        </div>
        <div>
          <label>Private Key (opcional)</label>
          <input
            name="private_key"
            value={formData.private_key ?? ''}
            onChange={handleChange}
            placeholder="Cole a chave privada para importar"
            className="input-field"
          />
        </div>
        <div className="form-actions">
          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Processando...' : 'Criar / Importar carteira'}
          </button>
          <p className="hint">Deixe vazio para gerar uma nova carteira automaticamente.</p>
        </div>
      </form>
      {successMessage && <p className="success-message">{successMessage}</p>}
      {error && <p className="error-message">{error}</p>}
    </section>
  )
}

export default WalletForm
