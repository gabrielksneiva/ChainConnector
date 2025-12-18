package http

import (
	"ChainConnector/internal/domain/service"
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// FiberServer is an fx-friendly wrapper that contains the Fiber app and
// lifecycle/start logic. It is provided to the fx app via a constructor
// (NewFiberServer) and its Start method registers lifecycle hooks.
type fiberApp interface {
	Listen(string) error
	Shutdown() error
}

type FiberServer struct {
	app    fiberApp
	addr   string
	txSvc  *service.TransactionService
	logger *zap.Logger
}

// Start registers the lifecycle hooks to start and stop the Fiber server.
func (s *FiberServer) Start(lc fx.Lifecycle) {
	if lc == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := s.app.Listen(s.addr); err != nil {
					s.logger.Error("fiber start failed", zap.Error(err))
				}
			}()
			s.logger.Info("fiber server started", zap.String("addr", s.addr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := s.app.Shutdown(); err != nil {
				s.logger.Error("fiber shutdown failed", zap.Error(err))
			}
			s.logger.Info("fiber server stopped")
			return nil
		},
	})
}

// NewFiberServer constructs a FiberServer for fx. It accepts a zap.Logger
// and sets a default address. Modify to read config when available.
func NewFiberServer(logger *zap.Logger, txSvc *service.TransactionService) *FiberServer {
	app := CreateFiberServer()
	return &FiberServer{app: app, addr: ":3000", logger: logger, txSvc: txSvc}
}

func CreateFiberServer() *fiber.App {
	app := fiber.New()

	// ROUTES
	app.Get("/health", handlerHeatlCheck)

	return app
}

// HANDLERS
func handlerHeatlCheck(c *fiber.Ctx) error {
	return c.SendString("OK")
}
