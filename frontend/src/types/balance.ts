export interface Balance {
  address: string;
  chain: string;
  amount: string;
}

export interface BalanceUpdateRequest {
  amount: string;
  chain: string;
}

export interface BalanceHistoryItem {
  amount: string;
  chain: string;
  timestamp: string;
}