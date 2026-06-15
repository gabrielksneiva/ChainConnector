package http

import (
	"ChainConnector/internal/adapters/monitor"
	"ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var capturedHook fx.Hook

type fakeApp struct{}

func (f *fakeApp) Listen(_ string) error                          { return nil }
func (f *fakeApp) Shutdown() error                                { return nil }
func (f *fakeApp) Get(_ string, _ ...fiber.Handler) fiber.Router  { return nil }
func (f *fakeApp) Post(_ string, _ ...fiber.Handler) fiber.Router { return nil }

type fakeBlockchain struct{}

func (f *fakeBlockchain) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (f *fakeBlockchain) GetNonce(ctx context.Context, address string) (uint64, error) {
	return 0, nil
}
func (f *fakeBlockchain) GetTransactionReceipt(ctx context.Context, txHash string) (*entity.Receipt, error) {
	return nil, nil
}
func (f *fakeBlockchain) GetLogs(ctx context.Context, filter entity.LogFilter) ([]entity.Log, error) {
	return nil, nil
}
func (f *fakeBlockchain) GetBlockNumber(ctx context.Context) (uint64, error) {
	return 0, nil
}
func (f *fakeBlockchain) GetBlockByNumber(ctx context.Context, chain string, number uint64) (*entity.Block, error) {
	return nil, nil
}
func (f *fakeBlockchain) EstimateFees(ctx context.Context, chain string) (*big.Int, *big.Int, error) {
	return big.NewInt(1), big.NewInt(1), nil
}
func (f *fakeBlockchain) EstimateGas(ctx context.Context, chain string, from string, to string, value *big.Int, data []byte) (uint64, error) {
	return 21000, nil
}
func (f *fakeBlockchain) SendRawTransaction(ctx context.Context, chain string, signedTx []byte) (string, error) {
	return "0xhash", nil
}
func (f *fakeBlockchain) SendRawTransactionHex(ctx context.Context, chain string, signedTxHex string) (string, error) {
	return "0xhash", nil
}

type fakeLc struct{}

func (f *fakeLc) Append(h fx.Hook) { capturedHook = h }

// Ensure CreateFiberServer registers healthcheck and root routes.
func TestCreateFiberServer_Healthcheck(t *testing.T) {
	app := CreateFiberServer()
	// /health should return 200 OK with body "OK"
	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error from app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFiberServer_StartRegistersHooks(t *testing.T) {
	// Should not panic when registering hooks
	// Start hook registration relies on fx lifecycle internals and
	// is exercised indirectly via integration. Skip direct invocation
	// here to avoid lifecycle initialization complexity in unit tests.
	s := makeTestServer(nil)

	// Use zero-value lifecycle; Start should handle nil Append without panicking.
	var lc fx.Lifecycle
	s.Start(lc)
}

func TestNewFiberServer_ConstructsWithLogger(t *testing.T) {
	s := makeTestServer(nil)
	if s == nil || s.app == nil {
		t.Fatalf("expected non-nil FiberServer and app")
	}
	if s.logger == nil {
		t.Fatalf("expected logger to be set on FiberServer")
	}
}

func TestFiberServer_HookExecution(t *testing.T) {
	s := makeTestServer(nil)
	// inject fake app to avoid real network Listen
	s.app = &fakeApp{}

	// clear captured hook and start
	capturedHook = fx.Hook{}
	s.Start(&fakeLc{})

	// Execute OnStart and OnStop directly; they should not panic.
	if capturedHook.OnStart != nil {
		if err := capturedHook.OnStart(context.Background()); err != nil {
			t.Fatalf("OnStart returned error: %v", err)
		}
	}
	if capturedHook.OnStop != nil {
		if err := capturedHook.OnStop(context.Background()); err != nil {
			t.Fatalf("OnStop returned error: %v", err)
		}
	}
}

type fakeBus struct {
	lastTopic   string
	lastPayload interface{}
}

func (f *fakeBus) Publish(ctx context.Context, topic string, payload interface{}) {
	f.lastTopic = topic
	f.lastPayload = payload
}
func (f *fakeBus) Subscribe(topic string, handler ports.EventHandler) func() { return func() {} }
func (f *fakeBus) Close() error                                              { return nil }

type fakeRepo struct{}

func (f *fakeRepo) Save(ctx context.Context, tx *entity.Transaction) error {
	return nil
}
func (f *fakeRepo) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	return nil, nil
}
func (f *fakeRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	return nil, nil
}
func (f *fakeRepo) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	return nil
}
func (f *fakeRepo) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	return nil, nil
}
func (f *fakeRepo) ListTransactions(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	return nil, nil
}

func (f *fakeRepo) AddInterestAddress(ctx context.Context, address string, chain string) error {
	return nil
}

func (f *fakeRepo) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	return []string{}, nil
}

func (f *fakeRepo) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (f *fakeRepo) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	return nil
}

func (f *fakeRepo) SaveNetwork(ctx context.Context, network *entity.Network) error {
	return nil
}

func (f *fakeRepo) FindNetworkByID(ctx context.Context, id string) (*entity.Network, error) {
	return &entity.Network{ID: id, Name: "sepolia", ChainID: 11155111, Enabled: true}, nil
}

func (f *fakeRepo) ListNetworks(ctx context.Context) ([]*entity.Network, error) {
	return []*entity.Network{{ID: "network-1", Name: "sepolia", ChainID: 11155111, Enabled: true}}, nil
}

func (f *fakeRepo) SaveWallet(ctx context.Context, wallet *entity.Wallet) error {
	return nil
}

func (f *fakeRepo) FindWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	return nil, nil
}

func (f *fakeRepo) FindWalletByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	return nil, nil
}

func (f *fakeRepo) ListWallets(ctx context.Context, chain string) ([]*entity.Wallet, error) {
	return nil, nil
}

func (f *fakeRepo) ListAllWallets(ctx context.Context) ([]*entity.Wallet, error) {
	return nil, nil
}

func makeTestServer(bus ports.EventBus) *FiberServer {
	logger := zap.NewNop()
	cfg := &config.Config{
		HTTPAddr:     ":3000",
		HTTPEnabled:  true,
		DefaultChain: "sepolia",
	}
	repo := &fakeRepo{}
	blockchain := &fakeBlockchain{}
	txSvc := service.NewTransactionService(repo, blockchain, zap.NewNop())
	networkSvc := service.NewNetworkService(repo, nil, zap.NewNop())
	interest := monitor.NewInterestStore()
	filter := monitor.NewBloomFilterCache()
	return NewFiberServer(logger, cfg, txSvc, networkSvc, repo, nil, blockchain, interest, filter, bus)
}

func TestHandlerTransactionPublishes(t *testing.T) {
	bus := &fakeBus{}
	s := makeTestServer(bus)
	app := s.app.(*fiber.App)

	body := map[string]string{
		"from":      "0xfrom",
		"to":        "0xto",
		"chain":     "sepolia",
		"amount":    "10",
		"gas":       "21000",
		"gas_price": "1",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	if bus.lastTopic != "transaction.created" {
		t.Fatalf("expected topic transaction.created, got %s", bus.lastTopic)
	}
	if bus.lastPayload == nil {
		t.Fatalf("expected payload, got nil")
	}
}

func TestHandlerTransactionInvalidJSON(t *testing.T) {
	bus := &fakeBus{}
	s := makeTestServer(bus)
	app := s.app.(*fiber.App)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", resp.StatusCode)
	}
}

func TestHandlerTransactionInvalidGas(t *testing.T) {
	bus := &fakeBus{}
	s := makeTestServer(bus)
	app := s.app.(*fiber.App)

	body := map[string]string{
		"from":      "0xfrom",
		"to":        "0xto",
		"chain":     "sepolia",
		"amount":    "10",
		"gas":       "notanumber",
		"gas_price": "1",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid gas, got %d", resp.StatusCode)
	}
}

func TestHandlerHeatlCheckMethod(t *testing.T) {
	s := makeTestServer(nil)
	app := s.app.(*fiber.App)

	// register a route that uses the method receiver so we invoke handlerHeatlCheck
	app.Get("/direct-health", s.handlerHeatlCheck)
	req, _ := http.NewRequest("GET", "/direct-health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHandlerRegisterNetwork(t *testing.T) {
	s := makeTestServer(nil)
	app := s.app.(*fiber.App)

	body := map[string]interface{}{
		"name":            "Sepolia",
		"chain_id":        11155111,
		"rpc_url":         "http://localhost:8545",
		"currency_symbol": "ETH",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/networks", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}
