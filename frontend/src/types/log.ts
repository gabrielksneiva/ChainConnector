export interface LogEntry {
  blockNumber: number;
  address: string;
  topics: string[];
  data: string;
  transactionHash: string;
  logIndex: number;
  removed: boolean;
}