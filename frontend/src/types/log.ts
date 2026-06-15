export interface LogEntry {
  address: string;
  topics: string[];
  data: string | number[];
  block_number: number;
  tx_hash: string;
  log_index: number;
}
