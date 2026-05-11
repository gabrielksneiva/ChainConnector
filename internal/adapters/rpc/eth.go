package rpc

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type ETHRPC struct {
	httpClient *http.Client
	logger     *zap.Logger
	urls       map[string]string
	defaultURL string
}

// NewETHRPC constructs an ETHRPC. The httpClient parameter is optional; if nil,
// a default client with timeout is used. The default RPC endpoint is Sepolia.
func NewETHRPC(logger *zap.Logger, httpClient *http.Client) *ETHRPC {
	return NewETHRPCWithURLs(logger, httpClient, map[string]string{"sepolia": "https://ethereum-sepolia-rpc.publicnode.com"})
}

// NewETHRPCWithURL constructs an ETHRPC with an explicit URL for Sepolia.
func NewETHRPCWithURL(logger *zap.Logger, httpClient *http.Client, url string) *ETHRPC {
	return NewETHRPCWithURLs(logger, httpClient, map[string]string{"sepolia": url})
}

// NewETHRPCWithURLs constructs an ETHRPC with explicit endpoint URLs per chain.
func NewETHRPCWithURLs(logger *zap.Logger, httpClient *http.Client, urls map[string]string) *ETHRPC {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	if urls == nil {
		urls = map[string]string{"sepolia": "https://ethereum-sepolia-rpc.publicnode.com"}
	}
	defaultURL := urls["sepolia"]
	if defaultURL == "" {
		for _, u := range urls {
			if u != "" {
				defaultURL = u
				break
			}
		}
	}
	if defaultURL == "" {
		defaultURL = "https://ethereum-sepolia-rpc.publicnode.com"
	}

	normalized := make(map[string]string, len(urls))
	for chain, u := range urls {
		normalized[strings.ToLower(chain)] = u
	}
	return &ETHRPC{
		urls:       normalized,
		defaultURL: defaultURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (e *ETHRPC) SendRawTransaction(ctx context.Context, chain string, signedTx []byte) (string, error) {
	hexTx := "0x" + hex.EncodeToString(signedTx)
	return e.SendRawTransactionHex(ctx, chain, hexTx)
}

func (e *ETHRPC) SendRawTransactionHex(ctx context.Context, chain string, signedTxHex string) (string, error) {
	var res string
	if err := e.rpcCall(ctx, chain, "eth_sendRawTransaction", []interface{}{signedTxHex}, &res); err != nil {
		return "", err
	}
	return res, nil
}

func (e *ETHRPC) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	var res string
	if err := e.rpcCall(ctx, "", "eth_getBalance", []interface{}{address, "latest"}, &res); err != nil {
		return nil, err
	}
	return hexToBigInt(res)
}

func (e *ETHRPC) GetNonce(ctx context.Context, address string) (uint64, error) {
	var res string
	if err := e.rpcCall(ctx, "", "eth_getTransactionCount", []interface{}{address, "pending"}, &res); err != nil {
		return 0, err
	}
	return hexToUint64(res)
}

func (e *ETHRPC) GetBlockNumber(ctx context.Context) (uint64, error) {
	var res string
	if err := e.rpcCall(ctx, "", "eth_blockNumber", []interface{}{}, &res); err != nil {
		return 0, err
	}
	return hexToUint64(res)
}

func (e *ETHRPC) GetTransactionReceipt(ctx context.Context, txHash string) (*entity.Receipt, error) {
	var raw map[string]interface{}
	if err := e.rpcCall(ctx, "", "eth_getTransactionReceipt", []interface{}{txHash}, &raw); err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	r := &entity.Receipt{TxHash: txHash}
	if v, ok := raw["blockNumber"].(string); ok {
		bn, _ := hexToUint64(v)
		r.BlockNumber = bn
	}
	if v, ok := raw["blockHash"].(string); ok {
		r.BlockHash = v
	}
	if v, ok := raw["status"].(string); ok {
		st, _ := hexToUint64(v)
		if st == 1 {
			r.Status = entity.ReceiptStatusSuccess
		} else {
			r.Status = entity.ReceiptStatusFailed
		}
	}
	if v, ok := raw["contractAddress"].(string); ok {
		r.ContractAddress = v
	}
	if v, ok := raw["gasUsed"].(string); ok {
		gu, _ := hexToUint64(v)
		r.GasUsed = gu
	}
	if v, ok := raw["cumulativeGasUsed"].(string); ok {
		cgu, _ := hexToUint64(v)
		r.CumulativeGasUsed = cgu
	}
	if v, ok := raw["effectiveGasPrice"].(string); ok {
		egp, _ := hexToBigInt(v)
		r.EffectiveGasPrice = egp
	}
	// logs
	if logsRaw, ok := raw["logs"].([]interface{}); ok {
		logs := make([]entity.Log, 0, len(logsRaw))
		for _, lr := range logsRaw {
			m, ok := lr.(map[string]interface{})
			if !ok {
				continue
			}
			var lg entity.Log
			if addr, ok := m["address"].(string); ok {
				lg.Address = addr
			}
			if topics, ok := m["topics"].([]interface{}); ok {
				for _, t := range topics {
					if ts, ok := t.(string); ok {
						lg.Topics = append(lg.Topics, ts)
					}
				}
			}
			if data, ok := m["data"].(string); ok {
				b, _ := hexToBytes(data)
				lg.Data = b
			}
			if bn, ok := m["blockNumber"].(string); ok {
				n, _ := hexToUint64(bn)
				lg.BlockNumber = n
			}
			if txh, ok := m["transactionHash"].(string); ok {
				lg.TxHash = txh
			}
			if li, ok := m["logIndex"].(string); ok {
				ix, _ := hexToUint64(li)
				lg.LogIndex = uint32(ix)
			}
			logs = append(logs, lg)
		}
		r.Logs = logs
	}
	return r, nil
}

// --- low-level JSON-RPC call ---
func (e *ETHRPC) rpcCall(ctx context.Context, chain string, method string, params interface{}, result interface{}) error {
	url, err := e.getURL(chain)
	if err != nil {
		return err
	}
	reqBody := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if e.logger != nil {
				e.logger.Warn("failed closing response body", zap.Error(cerr))
			}
		}
	}()
	body, _ := io.ReadAll(resp.Body)

	var envelope struct {
		Jsonrpc string          `json:"jsonrpc"`
		ID      interface{}     `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("invalid rpc response: %w; body=%s", err, string(body))
	}
	if envelope.Error != nil {
		return fmt.Errorf("rpc error: %d %s", envelope.Error.Code, envelope.Error.Message)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("failed decode result: %w; raw=%s", err, string(envelope.Result))
	}
	return nil
}

func (e *ETHRPC) getURL(chain string) (string, error) {
	if strings.TrimSpace(chain) == "" {
		return e.defaultURL, nil
	}
	key := strings.ToLower(strings.TrimSpace(chain))
	if url, ok := e.urls[key]; ok && url != "" {
		return url, nil
	}
	if key == "ethereum" {
		if url, ok := e.urls["eth"]; ok && url != "" {
			return url, nil
		}
	}
	if e.defaultURL != "" {
		return e.defaultURL, nil
	}
	return "", errors.New("unsupported chain endpoint")
}

// --- helper parsers ---
func hexToBigInt(hexs string) (*big.Int, error) {
	if hexs == "" || hexs == "0x" {
		return big.NewInt(0), nil
	}
	s := strings.TrimPrefix(hexs, "0x")
	// If hex string has odd length, pad with a leading 0 so DecodeString accepts it.
	if len(s)%2 == 1 {
		s = "0" + s
	}
	i := new(big.Int)
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	i.SetBytes(b)
	return i, nil
}

func hexToUint64(hexs string) (uint64, error) {
	if hexs == "" || hexs == "0x" {
		return 0, nil
	}
	s := strings.TrimPrefix(hexs, "0x")
	if len(s)%2 == 1 {
		s = "0" + s
	}
	v := new(big.Int)
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	v.SetBytes(b)
	return v.Uint64(), nil
}

func hexToBytes(hexs string) ([]byte, error) {
	s := strings.TrimPrefix(hexs, "0x")
	if s == "" {
		return []byte{}, nil
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return hex.DecodeString(s)
}

// GetLogs executes an eth_getLogs JSON-RPC request using the provided filter.
func (e *ETHRPC) GetLogs(ctx context.Context, f entity.LogFilter) ([]entity.Log, error) {
	params := map[string]interface{}{}
	if f.FromBlock != nil {
		params["fromBlock"] = fmt.Sprintf("0x%x", *f.FromBlock)
	}
	if f.ToBlock != nil {
		params["toBlock"] = fmt.Sprintf("0x%x", *f.ToBlock)
	}
	if len(f.Addresses) > 0 {
		params["address"] = f.Addresses
	}
	if len(f.Topics) > 0 {
		params["topics"] = f.Topics
	}

	var rawLogs []map[string]interface{}
	if err := e.rpcCall(ctx, "", "eth_getLogs", []interface{}{params}, &rawLogs); err != nil {
		return nil, err
	}

	logs := make([]entity.Log, 0, len(rawLogs))
	for _, raw := range rawLogs {
		var logEntry entity.Log
		if addr, ok := raw["address"].(string); ok {
			logEntry.Address = addr
		}
		if topics, ok := raw["topics"].([]interface{}); ok {
			for _, t := range topics {
				if ts, ok := t.(string); ok {
					logEntry.Topics = append(logEntry.Topics, ts)
				}
			}
		}
		if data, ok := raw["data"].(string); ok {
			b, _ := hexToBytes(data)
			logEntry.Data = b
		}
		if bn, ok := raw["blockNumber"].(string); ok {
			n, _ := hexToUint64(bn)
			logEntry.BlockNumber = n
		}
		if txh, ok := raw["transactionHash"].(string); ok {
			logEntry.TxHash = txh
		}
		if li, ok := raw["logIndex"].(string); ok {
			ix, _ := hexToUint64(li)
			logEntry.LogIndex = uint32(ix)
		}
		logs = append(logs, logEntry)
	}
	return logs, nil
}

// EstimateFees attempts a basic fee estimate using eth_gasPrice.
func (e *ETHRPC) EstimateFees(ctx context.Context, chain string) (*big.Int, *big.Int, error) {
	var gasPrice string
	if err := e.rpcCall(ctx, chain, "eth_gasPrice", []interface{}{}, &gasPrice); err != nil {
		return nil, nil, err
	}
	price, err := hexToBigInt(gasPrice)
	if err != nil {
		return nil, nil, err
	}
	return price, price, nil
}
