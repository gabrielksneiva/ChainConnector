package app

import (
	"ChainConnector/internal/adapters/http"
	"ChainConnector/internal/domain/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Modules = fx.Options(
	fx.Provide(
		newZapLogger,
		service.NewTransactionService,
		http.NewFiberServer,
	),
	fx.Invoke(func(lc fx.Lifecycle, h *http.FiberServer) {
		h.Start(lc)
	}),
)

func newZapLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}
