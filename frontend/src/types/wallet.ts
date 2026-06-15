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
