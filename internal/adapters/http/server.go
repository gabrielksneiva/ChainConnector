package http

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"context"
	"encoding/json"
	"math/big"
	"strconv"

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
	app    fiberApp
	addr   string
	txSvc  *service.TransactionService
	logger *zap.Logger
	bus    ports.EventBus
}

func CreateFiberServer() *fiber.App {
	app := fiber.New()
	// lightweight health route for tests and quick checks
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	return app
}

// NewFiberServer constructs a FiberServer for fx. It accepts a zap.Logger
// and sets a default address. Modify to read config when available.
func NewFiberServer(logger *zap.Logger, txSvc *service.TransactionService, bus ports.EventBus) *FiberServer {
	app := CreateFiberServer()
	srv := &FiberServer{app: app, addr: ":3000", logger: logger, txSvc: txSvc, bus: bus}
	// register routes so router() is used
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
}

// HANDLERS
func (f *FiberServer) handlerHeatlCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (f *FiberServer) handlerTransaction(c *fiber.Ctx) error {
	var body transaction

	err := json.Unmarshal(c.Body(), &body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	value := new(big.Int)
	value.SetString(body.Amount, 10)

	gasPrice := new(big.Int)
	gasPrice.SetString(body.GasPrice, 10)

	gas, err := strconv.ParseUint(body.Gas, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid gas value")
	}

	tx := &entity.Transaction{
		To:       &body.To,
		Chain:    body.Chain,
		Value:    value,
		Gas:      gas,
		GasPrice: gasPrice,
	}

	f.bus.Publish(context.Background(), "transaction.created", tx)

	return c.SendStatus(fiber.StatusAccepted)

}
