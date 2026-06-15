import { useState, useEffect } from 'react'
import { Wallet, CreateWalletRequest } from '../types/wallet'
import { createWallet as apiCreateWallet, listWallets } from '../services/api'

export const useWallets = (pollInterval: number = 10000) => {
  const [wallets, setWallets] = useState<Wallet[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchWallets = async () => {
    try {
      const data = await listWallets()
      setWallets(Array.isArray(data) ? data : [])
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const createWallet = async (data: CreateWalletRequest): Promise<boolean> => {
    setLoading(true)
    setError(null)
    try {
      await apiCreateWallet(data)
      await fetchWallets()
      return true
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error')
      return false
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchWallets()
    const interval = setInterval(fetchWallets, pollInterval)
    return () => clearInterval(interval)
  }, [pollInterval])

  return { wallets, loading, error, refetch: fetchWallets, createWallet }
}
