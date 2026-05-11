package http

import (
	"ChainConnector/internal/adapters/monitor"
	"ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"context"
	"math/big"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type transaction struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Chain    string `json:"chain"`
	Amount   string `json:"amount"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gas_price"`
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
	app      fiberApp
	addr     string
	txSvc    *service.TransactionService
	repo     ports.TxRepositoryPort
	interest *monitor.InterestStore
	filter   *monitor.BloomFilterCache
	logger   *zap.Logger
	bus      ports.EventBus
}

type interestRequest struct {
	Addresses []string   `json:"addresses"`
	Chain     string     `json:"chain,omitempty"`
	Topics    [][]string `json:"topics"`
	TxHashes  []string   `json:"tx_hashes"`
}

func CreateFiberServer() *fiber.App {
	app := fiber.New()
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
	repo ports.TxRepositoryPort,
	interest *monitor.InterestStore,
	filter *monitor.BloomFilterCache,
	bus ports.EventBus,
) *FiberServer {
	app := CreateFiberServer()
	srv := &FiberServer{
		app:      app,
		addr:     cfg.HTTPAddr,
		txSvc:    txSvc,
		repo:     repo,
		interest: interest,
		filter:   filter,
		logger:   logger,
		bus:      bus,
	}
	srv.router()
	return srv
}

// Start registers the lifecycle hooks to start and stop the Fiber server.
func (f *FiberServer) Start(lc fx.Lifecycle) {
	if lc == nil {
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
	f.app.Post("/interest", f.handlerInterest)
	f.app.Get("/pending", f.handlerPendingTransactions)
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

	if body.From == "" || body.To == "" || body.Chain == "" || body.Amount == "" || body.Gas == "" || body.GasPrice == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing required transaction fields")
	}

	value := new(big.Int)
	_, ok := value.SetString(body.Amount, 10)
	if !ok || value.Sign() <= 0 {
		f.logger.Error("invalid amount value", zap.String("amount", body.Amount))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid amount value")
	}

	gasPrice := new(big.Int)
	_, ok = gasPrice.SetString(body.GasPrice, 10)
	if !ok || gasPrice.Sign() <= 0 {
		f.logger.Error("invalid gas price", zap.String("gas_price", body.GasPrice))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid gas price")
	}

	gas, err := strconv.ParseUint(body.Gas, 10, 64)
	if err != nil || gas == 0 {
		f.logger.Error("invalid gas value", zap.Error(err), zap.String("gas", body.Gas))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid gas value")
	}

	tx := &entity.Transaction{
		From:     body.From,
		To:       &body.To,
		Chain:    body.Chain,
		Value:    value,
		Gas:      gas,
		GasPrice: gasPrice,
	}

	topic := "transaction.created"
	f.bus.Publish(c.Context(), topic, tx)
	f.logger.Info("published transaction to bus", zap.String("topic", topic), zap.String("from", body.From), zap.String("to", body.To), zap.String("chain", body.Chain), zap.String("amount", body.Amount))

	return c.SendStatus(fiber.StatusAccepted)
}

func (f *FiberServer) handlerInterest(c *fiber.Ctx) error {
	var body interestRequest
	if err := c.BodyParser(&body); err != nil {
		f.logger.Error("invalid interest request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid interest request body")
	}

	chain := strings.TrimSpace(strings.ToLower(body.Chain))
	if chain == "" {
		chain = "sepolia"
	}

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
