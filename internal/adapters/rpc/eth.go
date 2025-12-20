package rpc

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"encoding/hex"
	"encoding/json"
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
	url        string
}

// NewETHRPC constructs an ETHRPC. The httpClient parameter is optional; if nil,
// a default client with timeout is used.
func NewETHRPC(logger *zap.Logger, httpClient *http.Client) *ETHRPC {
	url := "https://ethereum-sepolia-rpc.publicnode.com"
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &ETHRPC{
		url:        url,
		httpClient: httpClient,
		logger:     logger,
	}
}

func (e *ETHRPC) SendRawTransaction(ctx context.Context, chain string, signedTx []byte) (string, error) {
	hexTx := "0x" + hex.EncodeToString(signedTx)
	return e.SendRawTransactionHex(ctx, "", hexTx)
}

func (e *ETHRPC) SendRawTransactionHex(ctx context.Context, chain string, signedTxHex string) (string, error) {
	var res string
	if err := e.rpcCall(ctx, "eth_sendRawTransaction", []interface{}{signedTxHex}, &res); err != nil {
		return "", err
	}
	return res, nil
}

func (e *ETHRPC) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	var res string
	if err := e.rpcCall(ctx, "eth_getBalance", []interface{}{address, "latest"}, &res); err != nil {
		return nil, err
	}
	return hexToBigInt(res)
}

func (e *ETHRPC) GetNonce(ctx context.Context, address string) (uint64, error) {
	var res string
	if err := e.rpcCall(ctx, "eth_getTransactionCount", []interface{}{address, "pending"}, &res); err != nil {
		return 0, err
	}
	return hexToUint64(res)
}

func (e *ETHRPC) GetBlockNumber(ctx context.Context) (uint64, error) {
	var res string
	if err := e.rpcCall(ctx, "eth_blockNumber", []interface{}{}, &res); err != nil {
		return 0, err
	}
	return hexToUint64(res)
}

func (e *ETHRPC) GetTransactionReceipt(ctx context.Context, txHash string) (*entity.Receipt, error) {
	var raw map[string]interface{}
	if err := e.rpcCall(ctx, "eth_getTransactionReceipt", []interface{}{txHash}, &raw); err != nil {
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
func (e *ETHRPC) rpcCall(ctx context.Context, method string, params interface{}, result interface{}) error {
	reqBody := map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", e.url, strings.NewReader(string(b)))
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

// GetLogs is not implemented for this adapter yet.
func (e *ETHRPC) GetLogs(ctx context.Context, f entity.LogFilter) ([]entity.Log, error) {
	return nil, fmt.Errorf("unsupported operation: GetLogs not implemented")
}

// EstimateFees is not implemented by this simple RPC adapter and returns an error.
func (e *ETHRPC) EstimateFees(ctx context.Context, chain string) (*big.Int, *big.Int, error) {
	return nil, nil, fmt.Errorf("unsupported operation: EstimateFees not implemented")
}
