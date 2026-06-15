const API_BASE = '/api'
const REQUEST_TIMEOUT_MS = 15000

export interface ApiErrorBody {
  error?: string | { message?: string; code?: string }
  message?: string
}

export interface Network {
  id: string
  name: string
  chain_id: number
  rpc_url?: string
  currency_symbol?: string
  explorer_url?: string
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface RegisterNetworkRequest {
  name: string
  chain_id: number
  rpc_url?: string
  currency_symbol?: string
  explorer_url?: string
  enabled: boolean
}

export interface Wallet {
  id: string
  address: string
  chain: string
  created_at?: string
  updated_at?: string
}

export interface CreateWalletRequest {
  chain: string
  private_key?: string
}

export interface Transaction {
  id: string
  from: string
  to?: string
  chain?: string
  value?: string
  gas?: number
  gas_price?: string
  tx_hash?: string
  status?: string | number
  created_at?: string
  updated_at?: string
  receipt?: unknown
  error_message?: string
}

export interface CreateTransactionRequest {
  from: string
  to: string
  chain: string
  amount: string
  gas: string
  gas_price: string
  wallet_id?: string
}

export interface Balance {
  address: string
  chain: string
  amount: string
}

export interface BalanceHistoryItem {
  address?: string
  chain?: string
  amount: string
  timestamp?: string
}

export interface InterestRequest {
  addresses: string[]
  chain?: string
  topics: string[][]
  tx_hashes: string[]
}

export interface InterestSnapshot {
  addresses: string[]
  topics: string[][]
  tx_hashes: string[]
}

export interface LogEntry {
  address: string
  topics: string[]
  data: string | number[]
  block_number: number
  tx_hash: string
  log_index: number
}

export interface LogQuery {
  fromBlock?: number
  toBlock?: number
  address?: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const controller = new AbortController()
  const timeoutID = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...init,
      signal: init?.signal ?? controller.signal,
      headers: {
        ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
        ...init?.headers,
      },
    })
  } catch (err) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      throw new Error('A requisição demorou demais para responder')
    }
    throw err
  } finally {
    window.clearTimeout(timeoutID)
  }

  const contentType = response.headers.get('content-type') || ''
  const bodyText = await response.text()
  const body = parseResponseBody(bodyText, contentType)

  if (!response.ok) {
    const message = extractErrorMessage(body) || bodyText || `Request failed with status ${response.status}`
    throw new Error(message)
  }

  return body as T
}

function parseResponseBody(bodyText: string, contentType: string): unknown {
  if (!bodyText || !contentType.includes('application/json')) return bodyText

  try {
    return JSON.parse(bodyText)
  } catch {
    return bodyText
  }
}

function extractErrorMessage(body: unknown): string {
  if (typeof body === 'string') return body
  if (!body || typeof body !== 'object') return ''

  const errorBody = body as ApiErrorBody
  if (typeof errorBody.error === 'string') return errorBody.error
  if (errorBody.error && typeof errorBody.error === 'object' && errorBody.error.message) {
    return errorBody.error.message
  }
  return errorBody.message || ''
}

export const pingHealth = () => request<string>('/health')

export const registerNetwork = (data: RegisterNetworkRequest) =>
  request<Network>('/networks', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const listNetworks = () => request<Network[]>('/networks')

export const getNetworkByID = (id: string) => request<Network>(`/networks/${encodeURIComponent(id)}`)

export const createWallet = (data: CreateWalletRequest) =>
  request<Wallet>('/wallets', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const importWallet = (data: CreateWalletRequest) =>
  request<Wallet>('/wallets/import', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const listWallets = (chain?: string) => {
  const query = new URLSearchParams()
  if (chain) {
    query.set('chain', chain)
  }
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<Wallet[]>(`/wallets${suffix}`)
}

export const getWalletByID = (id: string) => request<Wallet>(`/wallets/${encodeURIComponent(id)}`)

export const createTransaction = (data: CreateTransactionRequest) =>
  request<Transaction | string>('/transaction', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const listPendingTransactions = () => request<Transaction[]>('/pending')

export const listTransactions = (limit = 100) => request<Transaction[]>(`/transactions?limit=${limit}`)

export const getTransactionByID = (id: string) => request<Transaction>(`/transaction/${encodeURIComponent(id)}`)

export const getBalance = (address: string, chain?: string) => {
  const query = new URLSearchParams()
  if (chain) {
    query.set('chain', chain)
  }
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<Balance>(`/balance/${encodeURIComponent(address)}${suffix}`)
}

export const updateBalance = (address: string, amount: string, chain: string) =>
  request<string>(`/balance/${encodeURIComponent(address)}`, {
    method: 'POST',
    body: JSON.stringify({ amount, chain }),
  })

export const getBalanceHistory = (address: string, chain?: string) => {
  const query = new URLSearchParams()
  if (chain) {
    query.set('chain', chain)
  }
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<BalanceHistoryItem[]>(`/balance/${encodeURIComponent(address)}/history${suffix}`)
}

export const registerInterest = (data: InterestRequest) =>
  request<string>('/interest', {
    method: 'POST',
    body: JSON.stringify(data),
  })

export const getInterests = () => request<InterestSnapshot>('/interests')

export const getLogs = (params: LogQuery) => {
  const query = new URLSearchParams()
  if (params.fromBlock !== undefined) query.set('fromBlock', String(params.fromBlock))
  if (params.toBlock !== undefined) query.set('toBlock', String(params.toBlock))
  if (params.address) query.set('address', params.address)
  const suffix = query.toString() ? `?${query.toString()}` : ''
  return request<LogEntry[]>(`/logs${suffix}`)
}
