import { FormEvent, ReactNode, useEffect, useMemo, useState } from 'react'
import {
  CreateTransactionRequest,
  CreateWalletRequest,
  Transaction,
  Wallet,
  createTransaction,
  createWallet,
  getTransactionByID,
  importWallet,
  listTransactions,
  listWallets,
  pingHealth,
} from '../services/api'

type Section = 'wallets' | 'transactions'
type OperationMode = 'guided' | 'advanced'

const sections: Array<{ id: Section; label: string }> = [
  { id: 'wallets', label: 'Carteiras' },
  { id: 'transactions', label: 'Transacoes' },
]

const defaultTransaction: CreateTransactionRequest = {
  from: '',
  to: '',
  chain: 'sepolia',
  amount: '',
  gas: '',
  gas_price: '',
  wallet_id: '',
}

function Dashboard() {
  const [activeSection, setActiveSection] = useState<Section>('wallets')
  const [health, setHealth] = useState('checking')
  const [notice, setNotice] = useState('')
  const [error, setError] = useState('')
  const [busyCount, setBusyCount] = useState(0)

  const [wallets, setWallets] = useState<Wallet[]>([])
  const [walletForm, setWalletForm] = useState<CreateWalletRequest>({ chain: 'sepolia', private_key: '' })
  const [showPrivateKey, setShowPrivateKey] = useState(false)
  const [selectedWallet, setSelectedWallet] = useState<Wallet | null>(null)

  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [transactionForm, setTransactionForm] = useState<CreateTransactionRequest>(defaultTransaction)
  const [transactionLookupID, setTransactionLookupID] = useState('')
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null)
  const [transactionMode, setTransactionMode] = useState<OperationMode>('guided')

  const walletAddresses = useMemo(
    () => new Set(wallets.map((wallet) => wallet.address.toLowerCase())),
    [wallets],
  )
  const pendingCount = transactions.filter((tx) => statusLabel(tx.status) === 'pending').length

  const operationCards = [
    {
      label: 'API',
      value: health,
      tone: health === 'online' ? 'success' : health === 'offline' ? 'danger' : 'warning',
      detail: health === 'online' ? 'Backend respondendo' : 'Verifique o servico local',
    },
    {
      label: 'Carteiras monitoradas',
      value: String(wallets.length),
      tone: wallets.length > 0 ? 'success' : 'neutral',
      detail: 'Toda carteira cadastrada entra no monitoramento',
    },
    {
      label: 'Historico',
      value: String(transactions.length),
      tone: transactions.length > 0 ? 'success' : 'neutral',
      detail: 'Depositos, saques e transacoes internas',
    },
    {
      label: 'Pendentes',
      value: String(pendingCount),
      tone: pendingCount > 0 ? 'warning' : 'neutral',
      detail: 'Aguardando confirmacao ou processamento',
    },
  ]

  const isBusy = busyCount > 0

  useEffect(() => {
    void bootstrap()
    // Bootstrap only once; later refreshes are explicit user actions.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function runAction<T>(
    action: () => Promise<T>,
    successMessage: string,
    options: { silent?: boolean } = {},
  ): Promise<T | null> {
    setError('')
    if (!options.silent) setNotice('')
    setBusyCount((current) => current + 1)
    try {
      const result = await action()
      if (!options.silent) setNotice(successMessage)
      return result
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro inesperado')
      return null
    } finally {
      setBusyCount((current) => Math.max(0, current - 1))
    }
  }

  async function bootstrap() {
    const status = await runAction(async () => pingHealth(), 'Backend conectado', { silent: true })
    setHealth(status ? 'online' : 'offline')
    await Promise.all([
      refreshWallets({ silent: true }),
      refreshTransactions({ silent: true }),
    ])
  }

  async function refreshWallets(options: { silent?: boolean } = {}) {
    const result = await runAction(() => listWallets(), 'Carteiras atualizadas', options)
    if (result) setWallets(result)
  }

  async function refreshTransactions(options: { silent?: boolean } = {}) {
    const result = await runAction(() => listTransactions(200), 'Historico atualizado', options)
    if (result) setTransactions(result)
  }

  async function submitWallet(event: FormEvent, mode: 'create' | 'import') {
    event.preventDefault()
    const payload = cleanWalletPayload(walletForm)
    if (!payload.chain) {
      setError('Informe a chain da carteira')
      return
    }
    if (mode === 'import' && !payload.private_key) {
      setError('Informe a private key local para importar a carteira')
      return
    }

    const result = await runAction(
      () => (mode === 'import' ? importWallet(payload) : createWallet(payload)),
      mode === 'import' ? 'Carteira importada' : 'Carteira criada',
    )
    if (result) {
      setSelectedWallet(result)
      setWalletForm({ chain: result.chain, private_key: '' })
      await refreshWallets()
    }
  }

  async function submitTransaction(event: FormEvent) {
    event.preventDefault()
    const result = await runAction(
      () => createTransaction(cleanTransactionPayload(transactionForm)),
      'Transacao enviada',
    )
    if (result && typeof result === 'object') {
      setSelectedTransaction(result)
    }
    await refreshTransactions()
  }

  async function lookupTransaction(event: FormEvent) {
    event.preventDefault()
    const result = await runAction(() => getTransactionByID(transactionLookupID), 'Transacao carregada')
    if (result) setSelectedTransaction(result)
  }

  function selectWalletForSource(wallet: Wallet) {
    setSelectedWallet(wallet)
    setTransactionForm((current) => ({
      ...current,
      from: wallet.address,
      chain: wallet.chain,
      wallet_id: wallet.id,
    }))
    setActiveSection('transactions')
    setNotice('Carteira aplicada como origem do saque')
    setError('')
  }

  function selectWalletForDestination(wallet: Wallet) {
    setSelectedWallet(wallet)
    setTransactionForm((current) => ({
      ...current,
      to: wallet.address,
      chain: wallet.chain,
    }))
    setActiveSection('transactions')
    setNotice('Carteira aplicada como destino monitorado')
    setError('')
  }

  async function copyValue(value: string, label: string) {
    if (!value || value === '-') return
    try {
      await navigator.clipboard.writeText(value)
      setNotice(`${label} copiado`)
      setError('')
    } catch {
      setError('Nao foi possivel copiar para a area de transferencia')
    }
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">ChainConnector</p>
          <h1>Painel operacional</h1>
        </div>
        <div className={`health health-${health}`} aria-label={`Status do backend: ${health}`}>
          {health}
        </div>
      </header>

      <div className="layout">
        <aside className="sidebar">
          <nav>
            {sections.map((section) => (
              <button
                key={section.id}
                className={activeSection === section.id ? 'active' : ''}
                onClick={() => setActiveSection(section.id)}
                aria-current={activeSection === section.id ? 'page' : undefined}
                type="button"
              >
                {section.label}
              </button>
            ))}
          </nav>
          <button className="secondary" type="button" onClick={() => void bootstrap()} disabled={isBusy}>
            Atualizar tudo
          </button>
        </aside>

        <main className="content">
          <section className="metrics-grid operational-metrics" aria-label="Status operacional">
            {operationCards.map((item) => (
              <article className={`metric metric-${item.tone}`} key={item.label}>
                <span>{item.label}</span>
                <strong>{item.value}</strong>
                <small>{item.detail}</small>
              </article>
            ))}
          </section>

          {(notice || error) && (
            <div className={`alert ${error ? 'danger' : 'success'}`}>{error || notice}</div>
          )}

          {activeSection === 'wallets' && (
            <section className="panel">
              <PanelHeader title="Carteiras monitoradas" actionLabel="Recarregar" onAction={refreshWallets} disabled={isBusy} />
              <section className="wallet-monitoring-banner" aria-label="Monitoramento automatico de carteiras">
                <div>
                  <span className="section-pill">Monitoramento automatico</span>
                  <h3>Se a carteira esta registrada, ela esta monitorada.</h3>
                  <p>
                    Use Receber teste para preencher o destino de uma transacao e validar o fluxo de deposito.
                    Use Enviar desta para assinar/enviar a partir da carteira local.
                  </p>
                </div>
                <strong>{wallets.length}</strong>
              </section>

              <form className="form-grid" onSubmit={(event) => submitWallet(event, 'create')}>
                <Field label="Chain">
                  <input
                    value={walletForm.chain}
                    onChange={(event) => setWalletForm({ ...walletForm, chain: event.target.value })}
                    required
                  />
                </Field>
                <Field label="Private key local">
                  <div className="input-with-action">
                    <input
                      value={walletForm.private_key || ''}
                      onChange={(event) => setWalletForm({ ...walletForm, private_key: event.target.value })}
                      placeholder="Opcional para criar; obrigatoria para importar"
                      type={showPrivateKey ? 'text' : 'password'}
                      autoComplete="off"
                    />
                    <button
                      className="secondary compact-button"
                      type="button"
                      aria-label={showPrivateKey ? 'Ocultar private key' : 'Mostrar private key'}
                      onClick={() => setShowPrivateKey((current) => !current)}
                    >
                      {showPrivateKey ? 'Ocultar' : 'Mostrar'}
                    </button>
                  </div>
                </Field>
                <div className="form-actions horizontal">
                  <button className="primary" type="submit" disabled={isBusy}>
                    Criar carteira
                  </button>
                  <button
                    className="secondary"
                    type="button"
                    disabled={isBusy}
                    onClick={(event) => void submitWallet(event, 'import')}
                  >
                    Importar carteira
                  </button>
                </div>
              </form>

              <DataTable
                caption="Carteiras registradas"
                empty="Nenhuma carteira registrada. Crie ou importe uma carteira local para monitorar depósitos e saques."
                onCopy={copyValue}
                headers={['Monitoramento', 'Endereco', 'Chain', 'Wallet ID', 'Criada em', 'Acoes']}
                rows={wallets.map((wallet) => [
                  statusBadge('monitorada'),
                  compactCell(wallet.address, 'Endereco'),
                  wallet.chain,
                  compactCell(wallet.id, 'Wallet ID'),
                  formatDate(wallet.created_at),
                  {
                    actions: [
                      {
                        label: 'Receber teste',
                        title: 'Usar esta carteira como destino monitorado',
                        action: () => selectWalletForDestination(wallet),
                      },
                      {
                        label: 'Enviar desta',
                        title: 'Usar esta carteira como origem',
                        action: () => selectWalletForSource(wallet),
                      },
                    ],
                  },
                ])}
              />
              <JsonPanel title="Carteira selecionada" data={selectedWallet} />
            </section>
          )}

          {activeSection === 'transactions' && (
            <section className="panel">
              <PanelHeader title="Transacoes" actionLabel="Recarregar historico" onAction={refreshTransactions} disabled={isBusy} />
              <div className="segmented-control" aria-label="Modo da transacao">
                <button
                  className={transactionMode === 'guided' ? 'active' : ''}
                  type="button"
                  onClick={() => setTransactionMode('guided')}
                >
                  Guiado
                </button>
                <button
                  className={transactionMode === 'advanced' ? 'active' : ''}
                  type="button"
                  onClick={() => setTransactionMode('advanced')}
                >
                  Avancado
                </button>
              </div>

              <form className="flow-grid" onSubmit={submitTransaction}>
                <StepBlock step="1" title="Origem" hint="Para saque assinado, use uma carteira registrada como origem.">
                  <div className="form-grid compact-grid">
                    <Field label="Origem">
                      <input
                        value={transactionForm.from}
                        onChange={(event) => setTransactionForm({ ...transactionForm, from: event.target.value })}
                        required
                      />
                    </Field>
                    <Field label="Wallet ID opcional">
                      <input
                        value={transactionForm.wallet_id || ''}
                        onChange={(event) => setTransactionForm({ ...transactionForm, wallet_id: event.target.value })}
                      />
                    </Field>
                    <Field label="Chain">
                      <input
                        value={transactionForm.chain}
                        onChange={(event) => setTransactionForm({ ...transactionForm, chain: event.target.value })}
                        required
                      />
                    </Field>
                  </div>
                </StepBlock>

                <StepBlock step="2" title="Destino e valor" hint="Para validar deposito, use uma carteira registrada como destino.">
                  <div className="form-grid compact-grid">
                    <Field label="Destino">
                      <input
                        value={transactionForm.to}
                        onChange={(event) => setTransactionForm({ ...transactionForm, to: event.target.value })}
                        required
                      />
                    </Field>
                    <Field label="Valor na chain">
                      <input
                        value={transactionForm.amount}
                        onChange={(event) => setTransactionForm({ ...transactionForm, amount: event.target.value })}
                        required
                      />
                    </Field>
                  </div>
                </StepBlock>

                <StepBlock step="3" title="Execucao" hint="Revise taxas e envie para o backend processar.">
                  <div className="execution-summary">
                    <div>
                      <span>Gas</span>
                      <strong>Calculado pelo backend</strong>
                    </div>
                    <div>
                      <span>Fee</span>
                      <strong>Estimativa via RPC da chain</strong>
                    </div>
                  </div>
                  <div className="form-grid compact-grid">
                    <div className="form-actions">
                      <button className="primary" type="submit" disabled={isBusy}>
                        Criar transacao
                      </button>
                    </div>
                  </div>
                </StepBlock>
              </form>

              {transactionMode === 'advanced' && (
                <JsonPanel title="Payload da transacao" data={cleanTransactionPayload(transactionForm)} />
              )}

              <LookupForm
                label="Buscar transacao por ID"
                value={transactionLookupID}
                onChange={setTransactionLookupID}
                onSubmit={lookupTransaction}
                disabled={isBusy}
              />

              <DataTable
                caption="Historico de transacoes"
                empty="Nenhuma transacao registrada ainda. Crie uma transacao ou envie fundos para uma carteira monitorada."
                onCopy={copyValue}
                headers={['Tipo', 'Status', 'From', 'To', 'Valor', 'Hash', 'Criada em']}
                rows={transactions.map((tx) => [
                  typeBadge(transactionDirection(tx, walletAddresses)),
                  statusBadge(statusLabel(tx.status)),
                  compactCell(tx.from, 'Origem'),
                  compactCell(tx.to || '', 'Destino'),
                  tx.value !== undefined ? String(tx.value) : '-',
                  tx.tx_hash ? compactCell(tx.tx_hash, 'Transaction hash') : '-',
                  formatDate(tx.created_at),
                ])}
              />
              <JsonPanel title="Transacao selecionada" data={selectedTransaction} />
            </section>
          )}
        </main>
      </div>
    </div>
  )
}

function PanelHeader(props: {
  title: string
  actionLabel: string
  onAction: () => void | Promise<void>
  disabled?: boolean
}) {
  return (
    <div className="panel-header">
      <h2>{props.title}</h2>
      <button className="secondary" type="button" disabled={props.disabled} onClick={() => void props.onAction()}>
        {props.actionLabel}
      </button>
    </div>
  )
}

function Field(props: { label: string; children: ReactNode }) {
  return (
    <label className="field">
      <span>{props.label}</span>
      {props.children}
    </label>
  )
}

function StepBlock(props: { step: string; title: string; hint?: string; children: ReactNode }) {
  return (
    <div className="step-block">
      <div className="step-heading">
        <span className="step-kicker">{props.step}</span>
        <div>
          <h3>{props.title}</h3>
          {props.hint && <p>{props.hint}</p>}
        </div>
      </div>
      {props.children}
    </div>
  )
}

function LookupForm(props: {
  label: string
  value: string
  onChange: (value: string) => void
  onSubmit: (event: FormEvent) => void
  disabled?: boolean
}) {
  return (
    <form className="lookup" onSubmit={props.onSubmit}>
      <label className="field">
        <span>{props.label}</span>
        <input value={props.value} onChange={(event) => props.onChange(event.target.value)} disabled={props.disabled} />
      </label>
      <button className="secondary" type="submit" disabled={props.disabled}>
        Buscar
      </button>
    </form>
  )
}

type TableCell =
  | string
  | number
  | { label: string; title?: string; action?: () => void; copyValue?: string; copyLabel?: string; status?: boolean; kind?: boolean }
  | { actions: Array<{ label: string; title?: string; action: () => void }> }

function DataTable(props: {
  caption: string
  headers: string[]
  rows: TableCell[][]
  empty: string
  onCopy?: (value: string, label: string) => void
}) {
  if (props.rows.length === 0) {
    return <div className="empty">{props.empty}</div>
  }
  return (
    <div className="table-wrap">
      <table>
        <caption>{props.caption}</caption>
        <thead>
          <tr>
            {props.headers.map((header) => (
              <th key={header} scope="col">
                {header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {props.rows.map((row, rowIndex) => (
            <tr key={rowKey(row, rowIndex)}>
              {row.map((cell, cellIndex) => (
                <td key={`${cellIndex}-${cellLabel(cell)}`} data-label={props.headers[cellIndex]}>
                  {renderTableCell(cell, props.onCopy)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function renderTableCell(cell: TableCell, onCopy?: (value: string, label: string) => void) {
  if (typeof cell === 'string' || typeof cell === 'number') return String(cell || '-')
  if ('actions' in cell) {
    return (
      <div className="action-cluster">
        {cell.actions.map((action) => (
          <button
            className="secondary table-action"
            type="button"
            key={action.label}
            title={action.title}
            onClick={action.action}
          >
            {action.label}
          </button>
        ))}
      </div>
    )
  }
  if (cell.action) {
    return (
      <button className="secondary table-action" type="button" onClick={cell.action}>
        {cell.label}
      </button>
    )
  }
  if (cell.status || cell.kind) {
    return (
      <span className={`status-badge ${cell.kind ? 'type-badge' : ''} status-${normalizeStatusClass(cell.label)}`}>
        {cell.label || '-'}
      </span>
    )
  }
  return (
    <span className="copy-cell">
      <span className="mono" title={cell.title} aria-label={cell.title}>
        {cell.label || '-'}
      </span>
      {cell.copyValue && onCopy && (
        <button
          className="secondary copy-button"
          type="button"
          title={`Copiar ${cell.copyLabel || 'valor'}`}
          aria-label={`Copiar ${cell.copyLabel || 'valor'}`}
          onClick={() => onCopy(cell.copyValue || '', cell.copyLabel || 'Valor')}
        >
          Copiar
        </button>
      )}
    </span>
  )
}

function rowKey(row: TableCell[], rowIndex: number): string {
  return `${row.map(cellLabel).join('-')}-${rowIndex}`
}

function cellLabel(cell: TableCell): string {
  if (typeof cell === 'string' || typeof cell === 'number') return String(cell)
  if ('actions' in cell) return cell.actions.map((action) => action.label).join('-')
  return cell.title || cell.label
}

function JsonPanel(props: { title: string; data: unknown }) {
  return (
    <details className="json-panel" open={Boolean(props.data)}>
      <summary>{props.title}</summary>
      <pre>{props.data ? JSON.stringify(props.data, null, 2) : 'null'}</pre>
    </details>
  )
}

function cleanWalletPayload(payload: CreateWalletRequest): CreateWalletRequest {
  return {
    chain: payload.chain.trim(),
    private_key: payload.private_key?.trim() || undefined,
  }
}

function cleanTransactionPayload(payload: CreateTransactionRequest): CreateTransactionRequest {
  return {
    ...payload,
    from: payload.from.trim(),
    to: payload.to.trim(),
    chain: payload.chain.trim(),
    amount: payload.amount.trim(),
    gas: payload.gas.trim(),
    gas_price: payload.gas_price.trim(),
    wallet_id: payload.wallet_id?.trim() || undefined,
  }
}

function compact(value: string): string {
  if (!value) return '-'
  return value.length > 18 ? `${value.slice(0, 10)}...${value.slice(-6)}` : value
}

function compactCell(value: string, copyLabel = 'valor'): TableCell {
  return { label: compact(value), title: value || '-', copyValue: value || '', copyLabel }
}

function formatDate(value?: string): string {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

function statusLabel(value?: string | number): string {
  const statusMap: Record<number, string> = {
    0: 'unknown',
    1: 'pending',
    2: 'signed',
    3: 'sent',
    4: 'confirmed',
    5: 'failed',
    6: 'cancelled',
  }
  if (typeof value === 'number') return statusMap[value] || String(value)
  return value || '-'
}

function statusBadge(value: string): TableCell {
  return { label: value || '-', title: value || '-', status: true }
}

function typeBadge(value: string): TableCell {
  return { label: value, title: value, kind: true }
}

function normalizeStatusClass(value: string): string {
  return value.toLowerCase().replace(/[^a-z0-9]+/g, '-')
}

function transactionDirection(tx: Transaction, walletAddresses: Set<string>): string {
  const fromKnown = walletAddresses.has(tx.from.toLowerCase())
  const toKnown = tx.to ? walletAddresses.has(tx.to.toLowerCase()) : false
  if (fromKnown && toKnown) return 'interno'
  if (toKnown) return 'deposito'
  if (fromKnown) return 'saque'
  return 'externa'
}

export default Dashboard
