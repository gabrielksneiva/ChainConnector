package http

import (
	"ChainConnector/internal/adapters/monitor"
	"ChainConnector/internal/adapters/wallet"
	"ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"context"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type transaction struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Chain       string `json:"chain"`
	Amount      string `json:"amount"`
	Gas         string `json:"gas,omitempty"`
	GasPrice    string `json:"gas_price,omitempty"`
	GasPriceAlt string `json:"gasPrice,omitempty"` // Accept camelCase for frontend compatibility
	WalletID    string `json:"wallet_id,omitempty"`
}

type walletRequest struct {
	Chain      string `json:"chain,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

type networkRequest struct {
	Name              string `json:"name"`
	ChainID           int64  `json:"chain_id"`
	ChainIDAlt        int64  `json:"chainId,omitempty"`
	RPCURL            string `json:"rpc_url,omitempty"`
	RPCURLAlt         string `json:"rpcUrl,omitempty"`
	CurrencySymbol    string `json:"currency_symbol,omitempty"`
	CurrencySymbolAlt string `json:"currencySymbol,omitempty"`
	ExplorerURL       string `json:"explorer_url,omitempty"`
	ExplorerURLAlt    string `json:"explorerUrl,omitempty"`
	Enabled           *bool  `json:"enabled,omitempty"`
}

type balanceUpdateRequest struct {
	Amount string `json:"amount"`
	Chain  string `json:"chain"`
}

type logQuery struct {
	FromBlock *uint64 `query:"fromBlock"`
	ToBlock   *uint64 `query:"toBlock"`
	Address   string  `query:"address"`
}

// FiberServer is an fx-friendly wrapper that contains the Fiber app and
// lifecycle/start logic. It is provided to the fx app via a constructor
// (NewFiberServer) and its Start method registers lifecycle hooks.
type fiberApp interface {
	Listen(string) error
	Shutdown() error
	Get(string, ...fiber.Handler) fiber.Router
	Post(string, ...fiber.Handler) fiber.Router
}

type FiberServer struct {
	app          fiberApp
	addr         string
	enabled      bool
	txSvc        *service.TransactionService
	networkSvc   *service.NetworkService
	repo         ports.TxRepositoryPort
	walletRepo   ports.WalletRepositoryPort
	blockchain   ports.BlockchainPort
	interest     *monitor.InterestStore
	filter       *monitor.BloomFilterCache
	logger       *zap.Logger
	bus          ports.EventBus
	defaultChain string
	queueEnabled bool
}

type interestRequest struct {
	Addresses []string   `json:"addresses"`
	Chain     string     `json:"chain,omitempty"`
	Topics    [][]string `json:"topics"`
	TxHashes  []string   `json:"tx_hashes"`
}

func CreateFiberServer() *fiber.App {
	app := fiber.New()
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	return app
}

// NewFiberServer constructs a FiberServer for fx. It accepts dependencies including config
// for reading the HTTP address from environment variables.
func NewFiberServer(
	logger *zap.Logger,
	cfg *config.Config,
	txSvc *service.TransactionService,
	networkSvc *service.NetworkService,
	repo ports.TxRepositoryPort,
	walletRepo ports.WalletRepositoryPort,
	blockchain ports.BlockchainPort,
	interest *monitor.InterestStore,
	filter *monitor.BloomFilterCache,
	bus ports.EventBus,
) *FiberServer {
	app := CreateFiberServer()
	srv := &FiberServer{
		app:          app,
		addr:         cfg.HTTPAddr,
		enabled:      cfg.HTTPEnabled,
		txSvc:        txSvc,
		networkSvc:   networkSvc,
		repo:         repo,
		walletRepo:   walletRepo,
		blockchain:   blockchain,
		interest:     interest,
		filter:       filter,
		logger:       logger,
		bus:          bus,
		defaultChain: cfg.DefaultChain,
		queueEnabled: cfg.SQSEnabled,
	}
	srv.router()
	return srv
}

// Start registers the lifecycle hooks to start and stop the Fiber server.
func (f *FiberServer) Start(lc fx.Lifecycle) {
	if lc == nil {
		return
	}
	if !f.enabled {
		f.logger.Info("fiber server disabled")
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := f.app.Listen(f.addr); err != nil {
					f.logger.Error("fiber start failed", zap.Error(err))
				}
			}()
			f.logger.Info("fiber server started", zap.String("addr", f.addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := f.app.Shutdown(); err != nil {
				f.logger.Error("fiber shutdown failed", zap.Error(err))
			}
			f.logger.Info("fiber server stopped")
			return nil
		},
	})
}

func (f *FiberServer) router() {
	f.app.Get("/health", f.handlerHeatlCheck)
	f.app.Post("/transaction", f.handlerTransaction)
	f.app.Get("/transaction/:id", f.handlerTransactionDetail)
	f.app.Get("/transactions", f.handlerTransactions)
	f.app.Post("/interest", f.handlerInterest)
	f.app.Get("/interests", f.handlerGetInterests)
	f.app.Get("/pending", f.handlerPendingTransactions)
	f.app.Post("/networks", f.handlerRegisterNetwork)
	f.app.Get("/networks", f.handlerListNetworks)
	f.app.Get("/networks/:id", f.handlerGetNetworkByID)
	f.app.Post("/wallets", f.handlerCreateWallet)
	f.app.Post("/wallets/import", f.handlerImportWallet)
	f.app.Get("/wallets", f.handlerListWallets)
	f.app.Get("/wallets/:id", f.handlerGetWalletByID)
	f.app.Get("/logs", f.handlerGetLogs)
	f.app.Get("/balance/:address", f.handlerGetBalance)
	f.app.Get("/balance/:address/history", f.handlerGetBalanceHistory)
	f.app.Post("/balance/:address", f.handlerUpdateBalance)
}

func (f *FiberServer) handlerHeatlCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (f *FiberServer) handlerTransaction(c *fiber.Ctx) error {
	var body transaction
	if err := c.BodyParser(&body); err != nil {
		f.logger.Error("invalid request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	if body.From == "" || body.To == "" || body.Chain == "" || body.Amount == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing required transaction fields")
	}

	// Support both gas_price and gasPrice field names
	gasPriceStr := body.GasPrice
	if gasPriceStr == "" {
		gasPriceStr = body.GasPriceAlt
	}
	if gasPriceStr == "" {
		_, maxFee, err := f.blockchain.EstimateFees(c.Context(), body.Chain)
		if err != nil {
			f.logger.Error("failed estimate gas price", zap.Error(err), zap.String("chain", body.Chain))
			return c.Status(fiber.StatusBadGateway).SendString("Unable to estimate gas price")
		}
		gasPriceStr = maxFee.String()
	}

	value := new(big.Int)
	_, ok := value.SetString(body.Amount, 10)
	if !ok || value.Sign() <= 0 {
		f.logger.Error("invalid amount value", zap.String("amount", body.Amount))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid amount value")
	}

	gasPriceInt := new(big.Int)
	_, ok = gasPriceInt.SetString(gasPriceStr, 10)
	if !ok || gasPriceInt.Sign() <= 0 {
		f.logger.Error("invalid gas price", zap.String("gas_price", gasPriceStr))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid gas price")
	}

	var gas uint64
	gasStr := strings.TrimSpace(body.Gas)
	if gasStr == "" {
		estimatedGas, err := f.blockchain.EstimateGas(c.Context(), body.Chain, body.From, body.To, value, nil)
		if err != nil {
			f.logger.Error("failed estimate gas", zap.Error(err), zap.String("chain", body.Chain))
			return c.Status(fiber.StatusBadGateway).SendString("Unable to estimate gas")
		}
		gas = estimatedGas
	} else {
		parsedGas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil || parsedGas == 0 {
			f.logger.Error("invalid gas value", zap.Error(err), zap.String("gas", gasStr))
			return c.Status(fiber.StatusBadRequest).SendString("Invalid gas value")
		}
		gas = parsedGas
	}

	tx := &entity.Transaction{
		From:     body.From,
		To:       &body.To,
		Chain:    body.Chain,
		Value:    value,
		Gas:      gas,
		GasPrice: gasPriceInt,
		Status:   entity.TxStatusPending,
	}

	if body.WalletID != "" {
		w, err := f.walletRepo.FindWalletByID(c.Context(), body.WalletID)
		if err != nil {
			f.logger.Error("failed to load wallet", zap.Error(err), zap.String("wallet_id", body.WalletID))
			return c.Status(fiber.StatusInternalServerError).SendString("Unable to load wallet")
		}
		if w == nil {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet not found")
		}
		if !strings.EqualFold(w.Address, body.From) {
			return c.Status(fiber.StatusBadRequest).SendString("Wallet address does not match from address")
		}
		signer, err := wallet.NewSignerFromPrivateKey(w.PrivateKey)
		if err != nil {
			f.logger.Error("failed create signer", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).SendString("Unable to create signer")
		}
		err = f.txSvc.SignAndSendTransaction(c.Context(), tx, signer, f.blockchain)
		if err != nil {
			f.logger.Error("failed sign and send transaction", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).SendString("Unable to send transaction")
		}
		return c.JSON(tx)
	}

	if err := f.txSvc.CreateTransaction(c.Context(), tx); err != nil {
		f.logger.Error("failed create transaction", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to create transaction")
	}

	if f.bus != nil {
		f.bus.Publish(c.Context(), "transaction.created", tx)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

func (f *FiberServer) handlerInterest(c *fiber.Ctx) error {
	var body interestRequest
	if err := c.BodyParser(&body); err != nil {
		f.logger.Error("invalid interest request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid interest request body")
	}

	chain := f.normalizeChain(body.Chain)

	for _, addr := range body.Addresses {
		if addr == "" {
			continue
		}
		f.interest.AddAddress(addr)
		if err := f.repo.AddInterestAddress(c.Context(), addr, chain); err != nil {
			f.logger.Error("failed to persist interest address", zap.Error(err), zap.String("address", addr), zap.String("chain", chain))
		}
	}
	for _, topicSet := range body.Topics {
		f.interest.AddTopics(topicSet)
	}
	for _, txHash := range body.TxHashes {
		f.interest.AddTxHash(txHash)
	}

	f.logger.Info("registered interest filter", zap.Int("addresses", len(body.Addresses)), zap.Int("topics", len(body.Topics)), zap.Int("tx_hashes", len(body.TxHashes)), zap.String("chain", chain))
	return c.SendStatus(fiber.StatusAccepted)
}

func (f *FiberServer) handlerPendingTransactions(c *fiber.Ctx) error {
	transactions, err := f.repo.ListPending(c.Context(), 50)
	if err != nil {
		f.logger.Error("failed to load pending transactions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load pending transactions")
	}
	return c.JSON(transactions)
}

func (f *FiberServer) handlerTransactions(c *fiber.Ctx) error {
	limit := 100
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid limit")
		}
		limit = parsed
	}
	transactions, err := f.repo.ListTransactions(c.Context(), limit)
	if err != nil {
		f.logger.Error("failed to load transactions", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load transactions")
	}
	return c.JSON(transactions)
}

func (f *FiberServer) handlerRegisterNetwork(c *fiber.Ctx) error {
	var body networkRequest
	if err := c.BodyParser(&body); err != nil {
		return f.errorResponse(c, fiber.StatusBadRequest, "Invalid network request", err)
	}

	chainID := body.ChainID
	if chainID == 0 {
		chainID = body.ChainIDAlt
	}
	rpcURL := body.RPCURL
	if rpcURL == "" {
		rpcURL = body.RPCURLAlt
	}
	currencySymbol := body.CurrencySymbol
	if currencySymbol == "" {
		currencySymbol = body.CurrencySymbolAlt
	}
	explorerURL := body.ExplorerURL
	if explorerURL == "" {
		explorerURL = body.ExplorerURLAlt
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	network := &entity.Network{
		Name:           body.Name,
		ChainID:        chainID,
		RPCURL:         rpcURL,
		CurrencySymbol: currencySymbol,
		ExplorerURL:    explorerURL,
		Enabled:        enabled,
	}

	if err := f.networkSvc.RegisterNetwork(c.Context(), network); err != nil {
		return f.errorResponse(c, networkRegistrationErrorStatus(err), "Unable to register network", err)
	}
	status := fiber.StatusCreated
	if f.queueEnabled {
		status = fiber.StatusAccepted
	}
	return c.Status(status).JSON(network)
}

func networkRegistrationErrorStatus(err error) int {
	if err == nil {
		return fiber.StatusInternalServerError
	}
	message := err.Error()
	if strings.Contains(message, "network name is required") ||
		strings.Contains(message, "chain_id must be greater than zero") ||
		strings.Contains(message, "must be a valid URL") ||
		strings.Contains(message, "must use http or https") {
		return fiber.StatusBadRequest
	}
	return fiber.StatusInternalServerError
}

func (f *FiberServer) handlerListNetworks(c *fiber.Ctx) error {
	networks, err := f.networkSvc.ListNetworks(c.Context())
	if err != nil {
		return f.errorResponse(c, fiber.StatusInternalServerError, "Unable to list networks", err)
	}
	return c.JSON(networks)
}

func (f *FiberServer) handlerGetNetworkByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if strings.TrimSpace(id) == "" {
		return f.errorResponse(c, fiber.StatusBadRequest, "Network ID is required", nil)
	}
	network, err := f.networkSvc.FindNetworkByID(c.Context(), id)
	if err != nil {
		return f.errorResponse(c, fiber.StatusInternalServerError, "Unable to load network", err)
	}
	if network == nil {
		return f.errorResponse(c, fiber.StatusNotFound, "Network not found", nil)
	}
	return c.JSON(network)
}

func (f *FiberServer) handlerTransactionDetail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Transaction id is required")
	}
	tx, err := f.repo.FindByID(c.Context(), id)
	if err != nil {
		f.logger.Error("failed load transaction detail", zap.Error(err), zap.String("id", id))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load transaction")
	}
	if tx == nil {
		return c.Status(fiber.StatusNotFound).SendString("Transaction not found")
	}
	return c.JSON(tx)
}

func (f *FiberServer) handlerGetInterests(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"addresses": f.interest.GetAddresses(),
		"topics":    f.interest.GetTopics(),
		"tx_hashes": f.interest.GetTxHashes(),
	})
}

func (f *FiberServer) handlerCreateWallet(c *fiber.Ctx) error {
	var body walletRequest
	if err := c.BodyParser(&body); err != nil {
		return f.errorResponse(c, fiber.StatusBadRequest, "Invalid wallet request", err)
	}

	chain := f.normalizeChain(body.Chain)

	var privateKey string
	if body.PrivateKey != "" {
		privateKey = strings.TrimSpace(body.PrivateKey)
		if _, err := hex.DecodeString(privateKey); err != nil {
			return f.errorResponse(c, fiber.StatusBadRequest, "Invalid private key format", err)
		}
	} else {
		key, err := crypto.GenerateKey()
		if err != nil {
			f.logger.Error("failed generate wallet key", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).SendString("Unable to create wallet")
		}
		privateKey = hex.EncodeToString(crypto.FromECDSA(key))
	}

	signer, err := wallet.NewSignerFromPrivateKey(privateKey)
	if err != nil {
		f.logger.Error("failed create signer", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to create wallet")
	}
	address, err := signer.Address(c.Context())
	if err != nil {
		f.logger.Error("failed derive wallet address", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to derive wallet address")
	}

	w := &entity.Wallet{
		ID:         uuid.NewString(),
		Address:    address,
		Chain:      chain,
		PrivateKey: privateKey,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := f.walletRepo.SaveWallet(c.Context(), w); err != nil {
		return f.errorResponse(c, fiber.StatusInternalServerError, "Unable to save wallet", err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": w.ID, "address": w.Address, "chain": w.Chain})
}

func (f *FiberServer) handlerGetWalletByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return f.errorResponse(c, fiber.StatusBadRequest, "Wallet ID is required", nil)
	}
	wallet, err := f.walletRepo.FindWalletByID(c.Context(), id)
	if err != nil {
		return f.errorResponse(c, fiber.StatusInternalServerError, "Unable to load wallet", err)
	}
	if wallet == nil {
		return f.errorResponse(c, fiber.StatusNotFound, "Wallet not found", nil)
	}
	return c.JSON(fiber.Map{"id": wallet.ID, "address": wallet.Address, "chain": wallet.Chain, "created_at": wallet.CreatedAt, "updated_at": wallet.UpdatedAt})
}

func (f *FiberServer) normalizeChain(chain string) string {
	chain = strings.TrimSpace(strings.ToLower(chain))
	if chain == "" {
		chain = f.defaultChain
	}
	return chain
}

func (f *FiberServer) errorResponse(c *fiber.Ctx, status int, message string, err error) error {
	if err != nil {
		f.logger.Error(message, zap.Error(err), zap.String("request_id", c.Get(fiber.HeaderXRequestID)))
	} else {
		f.logger.Warn(message, zap.String("request_id", c.Get(fiber.HeaderXRequestID)))
	}
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func (f *FiberServer) handlerImportWallet(c *fiber.Ctx) error {
	return f.handlerCreateWallet(c)
}

func (f *FiberServer) handlerListWallets(c *fiber.Ctx) error {
	chain := f.normalizeChain(c.Query("chain"))
	wallets, err := f.walletRepo.ListWallets(c.Context(), chain)
	if err != nil {
		f.logger.Error("failed list wallets", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to list wallets")
	}
	return c.JSON(wallets)
}

func (f *FiberServer) handlerGetLogs(c *fiber.Ctx) error {
	var query logQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid log query")
	}
	filter := entity.LogFilter{
		FromBlock: query.FromBlock,
		ToBlock:   query.ToBlock,
	}
	if query.Address != "" {
		filter.Addresses = []string{query.Address}
	}
	logs, err := f.blockchain.GetLogs(c.Context(), filter)
	if err != nil {
		f.logger.Error("failed fetch logs", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load logs")
	}
	return c.JSON(logs)
}

func (f *FiberServer) handlerGetBalance(c *fiber.Ctx) error {
	address := c.Params("address")
	chain := f.normalizeChain(c.Query("chain"))
	balance, err := f.repo.GetBalance(c.Context(), address, chain)
	if err != nil {
		f.logger.Error("failed fetch balance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load balance")
	}
	return c.JSON(fiber.Map{"address": address, "chain": chain, "amount": balance.String()})
}

func (f *FiberServer) handlerUpdateBalance(c *fiber.Ctx) error {
	address := c.Params("address")
	var body balanceUpdateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid balance request")
	}
	chain := f.normalizeChain(body.Chain)
	amount := new(big.Int)
	_, ok := amount.SetString(body.Amount, 10)
	if !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid amount")
	}
	if err := f.repo.UpdateBalance(c.Context(), address, chain, amount); err != nil {
		f.logger.Error("failed update balance", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to update balance")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (f *FiberServer) handlerGetBalanceHistory(c *fiber.Ctx) error {
	address := c.Params("address")
	if address == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Address is required")
	}

	// For now, historical balances are stored as interest or transaction activity.
	// This endpoint returns the latest balance from the repo and a lightweight history placeholder.
	balance, err := f.repo.GetBalance(c.Context(), address, f.normalizeChain(c.Query("chain")))
	if err != nil {
		f.logger.Error("failed fetch balance history", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).SendString("Unable to load balance history")
	}
	return c.JSON([]fiber.Map{{"address": address, "amount": balance.String()}})
}
