package rpc

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"ChainConnector/internal/domain/entity"

	"go.uber.org/zap"
)

// Tests for internal parsing helpers and RPC methods using a local httptest.Server.
func TestHexParsers(t *testing.T) {
	// hexToBigInt
	bi, err := hexToBigInt("0x0a")
	if err != nil {
		t.Fatalf("hexToBigInt failed: %v", err)
	}
	if bi.Cmp(big.NewInt(10)) != 0 {
		t.Fatalf("expected 10, got %s", bi.String())
	}

	// hexToUint64
	u, err := hexToUint64("0x0f")
	if err != nil {
		t.Fatalf("hexToUint64 failed: %v", err)
	}
	if u != 15 {
		t.Fatalf("expected 15, got %d", u)
	}

	// hexToBytes
	b, err := hexToBytes("0x0102")
	if err != nil {
		t.Fatalf("hexToBytes failed: %v", err)
	}
	if len(b) != 2 || b[0] != 1 || b[1] != 2 {
		t.Fatalf("unexpected bytes: %v", b)
	}
}

func TestRPCMethodsAgainstTestServer(t *testing.T) {
	// Test server that returns appropriate json-rpc envelopes
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		method := req["method"].(string)
		switch method {
		case "eth_sendRawTransaction":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0xdeadbeef"})
		case "eth_getBalance":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0x0a"})
		case "eth_getTransactionCount":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0x01"})
		case "eth_blockNumber":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0x02"})
		case "eth_getTransactionReceipt":
			res := map[string]interface{}{
				"blockNumber":       "0x2",
				"blockHash":         "0xabc",
				"status":            "0x1",
				"contractAddress":   nil,
				"gasUsed":           "0x10",
				"cumulativeGasUsed": "0x20",
				"effectiveGasPrice": "0x05",
				"logs":              []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": res})
		default:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": nil})
		}
	}))
	defer srv.Close()

	eth := NewETHRPC(zap.NewNop(), nil)
	// override URL to point to our test server
	eth.url = srv.URL

	// SendRawTransactionHex
	hash, err := eth.SendRawTransactionHex(context.Background(), "", "0x01")
	if err != nil {
		t.Fatalf("SendRawTransactionHex error: %v", err)
	}
	if hash != "0xdeadbeef" {
		t.Fatalf("unexpected tx hash: %s", hash)
	}

	// GetBalance
	bal, err := eth.GetBalance(context.Background(), "0xaddr")
	if err != nil {
		t.Fatalf("GetBalance error: %v", err)
	}
	if bal.Cmp(big.NewInt(10)) != 0 {
		t.Fatalf("expected balance 10, got %s", bal.String())
	}

	// GetNonce
	nonce, err := eth.GetNonce(context.Background(), "0xaddr")
	if err != nil {
		t.Fatalf("GetNonce error: %v", err)
	}
	if nonce != 1 {
		t.Fatalf("expected nonce 1, got %d", nonce)
	}

	// GetBlockNumber
	bn, err := eth.GetBlockNumber(context.Background())
	if err != nil {
		t.Fatalf("GetBlockNumber error: %v", err)
	}
	if bn != 2 {
		t.Fatalf("expected blockNumber 2, got %d", bn)
	}

	// GetTransactionReceipt
	rec, err := eth.GetTransactionReceipt(context.Background(), "0xhash")
	if err != nil {
		t.Fatalf("GetTransactionReceipt error: %v", err)
	}
	if rec == nil || rec.Status != entity.ReceiptStatusSuccess {
		t.Fatalf("unexpected receipt: %+v", rec)
	}
}

func TestRPCErrorEnvelopeAndInvalidJSON(t *testing.T) {
	// Server that returns an error envelope
	srvErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "error": map[string]interface{}{"code": -32000, "message": "oops"}})
	}))
	defer srvErr.Close()

	eth := NewETHRPC(zap.NewNop(), nil)
	eth.url = srvErr.URL
	if _, err := eth.SendRawTransactionHex(context.Background(), "", "0x01"); err == nil {
		t.Fatalf("expected error from SendRawTransactionHex when server returns error envelope")
	}

	// Server that returns invalid JSON
	srvBad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srvBad.Close()
	eth.url = srvBad.URL
	if _, err := eth.GetBalance(context.Background(), "0xaddr"); err == nil {
		t.Fatalf("expected error from GetBalance when server returns invalid JSON")
	}
}

func TestReceiptParsingWithLogsAndNullReceipt(t *testing.T) {
	// Server that returns a receipt with logs including odd-length hex data
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["method"].(string) == "eth_getTransactionReceipt" {
			res := map[string]interface{}{
				"blockNumber": "0x2",
				"blockHash":   "0xabc",
				"status":      "0x1",
				"logs": []interface{}{
					map[string]interface{}{
						"address":         "0xcontract",
						"topics":          []interface{}{"0x01", "0x02"},
						"data":            "0xabc",
						"blockNumber":     "0x2",
						"transactionHash": "0xhash",
						"logIndex":        "0x1",
					},
				},
				"gasUsed":           "0x10",
				"cumulativeGasUsed": "0x20",
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": res})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": nil})
	}))
	defer srv.Close()

	eth := NewETHRPC(zap.NewNop(), nil)
	eth.url = srv.URL
	rec, err := eth.GetTransactionReceipt(context.Background(), "0xhash")
	if err != nil {
		t.Fatalf("GetTransactionReceipt error: %v", err)
	}
	if rec == nil || len(rec.Logs) != 1 {
		t.Fatalf("expected one log parsed, got %+v", rec)
	}
	lg := rec.Logs[0]
	if lg.Address != "0xcontract" {
		t.Fatalf("unexpected log address: %s", lg.Address)
	}
	// data "0xabc" should decode to 2 bytes after padding
	if len(lg.Data) == 0 {
		t.Fatalf("expected non-empty data in log")
	}

	// Now return null receipt
	srvNull := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": nil})
	}))
	defer srvNull.Close()
	eth.url = srvNull.URL
	rec2, err := eth.GetTransactionReceipt(context.Background(), "0xhash")
	if err != nil {
		t.Fatalf("GetTransactionReceipt error on null: %v", err)
	}
	if rec2 != nil {
		t.Fatalf("expected nil receipt when RPC returns null, got %+v", rec2)
	}
}

func TestSendRawTransactionWrapper(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["method"].(string) == "eth_sendRawTransaction" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": "0xabc"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "result": nil})
	}))
	defer srv.Close()

	eth := NewETHRPC(zap.NewNop(), nil)
	eth.url = srv.URL

	// call wrapper that accepts bytes
	h, err := eth.SendRawTransaction(context.Background(), "", []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("SendRawTransaction error: %v", err)
	}
	if h != "0xabc" {
		t.Fatalf("unexpected hash: %s", h)
	}
}

func TestGetLogsAndEstimateFeesUnsupported(t *testing.T) {
	eth := NewETHRPC(zap.NewNop(), nil)
	_, err := eth.GetLogs(context.Background(), entity.LogFilter{})
	if err == nil {
		t.Fatalf("expected error from GetLogs")
	}
	_, _, err2 := eth.EstimateFees(context.Background(), "")
	if err2 == nil {
		t.Fatalf("expected error from EstimateFees")
	}
}
