import { Transaction, CreateTransactionRequest } from '../types/transaction';
import { Balance, BalanceUpdateRequest, BalanceHistoryItem } from '../types/balance';
import { Interest } from '../types/interest';
import { LogEntry } from '../types/log';

const API_BASE = '/api'; // Ajustar conforme necessário

export const createTransaction = async (data: CreateTransactionRequest): Promise<Transaction> => {
  const response = await fetch(`${API_BASE}/transaction`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to create transaction');
  }
  return response.json();
};

export const getPendingTransactions = async (): Promise<Transaction[]> => {
  const response = await fetch(`${API_BASE}/pending`);
  if (!response.ok) {
    throw new Error('Failed to fetch pending transactions');
  }
  return response.json();
};

export const getTransactionDetail = async (id: string): Promise<Transaction> => {
  const response = await fetch(`${API_BASE}/transaction/${id}`);
  if (!response.ok) {
    throw new Error('Failed to fetch transaction detail');
  }
  return response.json();
};

export const getBalance = async (address: string, chain: string): Promise<Balance> => {
  const response = await fetch(`${API_BASE}/balance/${address}?chain=${chain}`);
  if (!response.ok) {
    throw new Error('Failed to fetch balance');
  }
  return response.json();
};

export const updateBalance = async (address: string, data: BalanceUpdateRequest): Promise<void> => {
  const response = await fetch(`${API_BASE}/balance/${address}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to update balance');
  }
};

export const getBalanceHistory = async (address: string): Promise<BalanceHistoryItem[]> => {
  const response = await fetch(`${API_BASE}/balance/${address}/history`);
  if (!response.ok) {
    throw new Error('Failed to fetch balance history');
  }
  return response.json();
};

export const registerInterest = async (data: Interest): Promise<void> => {
  const response = await fetch(`${API_BASE}/interest`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error('Failed to register interest');
  }
};

export const getInterests = async (): Promise<Interest[]> => {
  const response = await fetch(`${API_BASE}/interests`);
  if (!response.ok) {
    throw new Error('Failed to fetch interests');
  }
  return response.json();
};

export const getLogs = async (params: { fromBlock?: number; toBlock?: number; address?: string }): Promise<LogEntry[]> => {
  const query = new URLSearchParams();
  if (params.fromBlock !== undefined) query.append('fromBlock', params.fromBlock.toString());
  if (params.toBlock !== undefined) query.append('toBlock', params.toBlock.toString());
  if (params.address) query.append('address', params.address);
  const response = await fetch(`${API_BASE}/logs?${query.toString()}`);
  if (!response.ok) {
    throw new Error('Failed to fetch logs');
  }
  return response.json();
};