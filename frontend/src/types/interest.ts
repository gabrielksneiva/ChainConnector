export interface Interest {
  id?: string;
  addresses: string[];
  topics: string[];
  tx_hashes: string[];
}

export interface LogEntry {
  blockNumber: number;
  address: string;
  topics: string[];
  data: string;
  transactionHash: string;
  logIndex: number;
  removed: boolean;
}