package app

import (
	"ChainConnector/internal/adapters/eventbus"
	"ChainConnector/internal/adapters/http"
	"ChainConnector/internal/adapters/monitor"
	"ChainConnector/internal/adapters/postgres"
	"ChainConnector/internal/adapters/rpc"
	"ChainConnector/internal/config"
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
		config.Load,
		newZapLogger,
		service.NewTransactionService,
		func() ports.EventBus { return eventbus.NewInMemoryBus(4, 1024) },
		providerPostgresTxRepository,
		monitor.NewInterestStore,
		monitor.NewBloomFilterCache,
		http.NewFiberServer,
		providerETHRPC,
		monitor.NewMonitorService,
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
	fx.Invoke(func(lc fx.Lifecycle, ms *monitor.MonitorService) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				ms.Start(ctx)
				return nil
			},
		})
	}),
)

func newZapLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func providerETHRPC(logger *zap.Logger, cfg *config.Config) ports.BlockchainPort {
	eth := rpc.NewETHRPCWithURLs(logger, nil, map[string]string{
		"eth":     cfg.EthRPCURL,
		"sepolia": cfg.SepoliaRPCURL,
	})
	return eth
}

func providerPostgresTxRepository(logger *zap.Logger, cfg *config.Config) (ports.TxRepositoryPort, error) {
	return postgres.NewPostgresTxRepository(cfg.DatabaseURL, logger)
}
