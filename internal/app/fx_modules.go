package app

import (
	"ChainConnector/internal/adapters/eventbus"
	"ChainConnector/internal/adapters/http"
	"ChainConnector/internal/adapters/postgres"
	"ChainConnector/internal/adapters/rpc"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"context"
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Modules = fx.Options(
	fx.Provide(
		newZapLogger,
		service.NewTransactionService,
		func() ports.EventBus { return eventbus.NewInMemoryBus(4, 1024) },
		postgres.NewInMemoryTxRepository,
		http.NewFiberServer,
		providerETHRPC,
	),
	fx.Invoke(func(lc fx.Lifecycle, h *http.FiberServer) {
		h.Start(lc)
	}),
	fx.Invoke(func(lc fx.Lifecycle, bus ports.EventBus, svc *service.TransactionService, logger *zap.Logger) {
		var unsub func()
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				unsub = bus.Subscribe("transactions.create", func(ctx context.Context, payload interface{}) error {
					tx, ok := payload.(*entity.Transaction)
					if !ok {
						return errors.New("invalid payload for transactions.create")
					}
					return svc.CreateTransaction(ctx, tx)
				})
				logger.Info("subscribed to transactions.create")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if unsub != nil {
					unsub()
				}
				if err := bus.Close(); err != nil {
					logger.Error("error closing bus", zap.Error(err))
				}
				return nil
			},
		})
	}),
)

func newZapLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func providerETHRPC(logger *zap.Logger) *rpc.ETHRPC {
	eth := rpc.NewETHRPC(logger, nil)
	return eth
}
