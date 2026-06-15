//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
)

const (
	defaultAPIBaseURL       = "http://localhost:3001"
	defaultRPCURL           = "http://localhost:8545"
	defaultDatabaseURL      = "postgres://user:password@localhost:5432/chainconnector?sslmode=disable"
	defaultChainName        = "sepolia"
	defaultChainID          = int64(11155111)
	defaultAnvilPrivateKey  = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	integrationEnableEnvKey = "CHAINCONNECTOR_INTEGRATION"
)

func TestBlockMonitorStackCapturesRegisteredWalletTransaction(t *testing.T) {
	if os.Getenv(integrationEnableEnvKey) != "1" {
		t.Skipf("set %s=1 and run docker compose up -d --build before running this integration test", integrationEnableEnvKey)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	apiBaseURL := getenv("CHAINCONNECTOR_API_BASE_URL", defaultAPIBaseURL)
	rpcURL := getenv("CHAINCONNECTOR_RPC_URL", defaultRPCURL)
	databaseURL := getenv("CHAINCONNECTOR_DATABASE_URL", defaultDatabaseURL)
	chainName := getenv("CHAINCONNECTOR_TEST_CHAIN", defaultChainName)
	chainID := big.NewInt(getenvInt64("CHAINCONNECTOR_TEST_CHAIN_ID", defaultChainID))

	waitForBackend(ctx, t, apiBaseURL)

	recipientKey, recipientAddress := generateRecipientWallet(t)
	importWallet(ctx, t, apiBaseURL, chainName, recipientKey, recipientAddress)

	txHash := sendAnvilTransaction(ctx, t, rpcURL, chainID, recipientAddress)
	waitForCapturedTransaction(ctx, t, databaseURL, txHash.Hex())
}

func waitForBackend(ctx context.Context, t *testing.T, apiBaseURL string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBaseURL+"/health", nil)
		if err != nil {
			t.Fatalf("build health request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			return
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled waiting for backend: %v", ctx.Err())
		case <-time.After(500 * time.Millisecond):
		}
	}
	t.Fatalf("backend health endpoint did not become ready at %s", apiBaseURL)
}

func generateRecipientWallet(t *testing.T) (string, common.Address) {
	t.Helper()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("generate recipient key: %v", err)
	}
	return hex.EncodeToString(crypto.FromECDSA(key)), crypto.PubkeyToAddress(key.PublicKey)
}

func importWallet(ctx context.Context, t *testing.T, apiBaseURL string, chain string, privateKey string, expectedAddress common.Address) {
	t.Helper()
	payload := map[string]string{
		"chain":       chain,
		"private_key": privateKey,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal wallet payload: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/wallets/import", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build wallet request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("import wallet request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("import wallet returned status %d", resp.StatusCode)
	}

	var response struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("decode wallet response: %v", err)
	}
	if common.HexToAddress(response.Address) != expectedAddress {
		t.Fatalf("imported wallet address mismatch: got %s want %s", response.Address, expectedAddress.Hex())
	}
}

func sendAnvilTransaction(ctx context.Context, t *testing.T, rpcURL string, chainID *big.Int, recipient common.Address) common.Hash {
	t.Helper()
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		t.Fatalf("dial anvil rpc: %v", err)
	}
	defer client.Close()

	senderKey := mustPrivateKey(t, getenv("CHAINCONNECTOR_ANVIL_PRIVATE_KEY", defaultAnvilPrivateKey))
	senderAddress := crypto.PubkeyToAddress(senderKey.PublicKey)

	nonce, err := client.PendingNonceAt(ctx, senderAddress)
	if err != nil {
		t.Fatalf("load sender nonce: %v", err)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		t.Fatalf("suggest gas price: %v", err)
	}

	tx := types.NewTransaction(nonce, recipient, big.NewInt(12345), 21_000, gasPrice, nil)
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), senderKey)
	if err != nil {
		t.Fatalf("sign transaction: %v", err)
	}
	if err := client.SendTransaction(ctx, signed); err != nil {
		t.Fatalf("send transaction: %v", err)
	}
	waitForReceipt(ctx, t, client, signed.Hash())
	return signed.Hash()
}

func waitForReceipt(ctx context.Context, t *testing.T, client *ethclient.Client, txHash common.Hash) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	for time.Now().Before(deadline) {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil && receipt != nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled waiting for receipt %s: %v", txHash.Hex(), ctx.Err())
		case <-time.After(750 * time.Millisecond):
		}
	}
	t.Fatalf("receipt not found for transaction %s", txHash.Hex())
}

func waitForCapturedTransaction(ctx context.Context, t *testing.T, databaseURL string, txHash string) {
	t.Helper()
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		err := db.QueryRowContext(ctx, `SELECT status FROM transactions WHERE tx_hash = $1`, txHash).Scan(&status)
		if err == nil {
			if status != "confirmed" {
				t.Fatalf("captured transaction %s with unexpected status %s", txHash, status)
			}
			return
		}
		if err != sql.ErrNoRows {
			t.Fatalf("query captured transaction: %v", err)
		}
		select {
		case <-ctx.Done():
			t.Fatalf("context cancelled waiting for captured tx %s: %v", txHash, ctx.Err())
		case <-time.After(1 * time.Second):
		}
	}
	t.Fatalf("transaction %s was not captured by block consumer", txHash)
}

func mustPrivateKey(t *testing.T, value string) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.HexToECDSA(value)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	return key
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", key, err))
	}
	return parsed
}
