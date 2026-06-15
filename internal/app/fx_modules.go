package app

import (
	"ChainConnector/internal/adapters/eventbus"
	"ChainConnector/internal/adapters/http"
	"ChainConnector/internal/adapters/monitor"
	"ChainConnector/internal/adapters/postgres"
	"ChainConnector/internal/adapters/rpc"
	"ChainConnector/internal/adapters/sqsqueue"
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
		service.NewNetworkService,
		func() ports.EventBus { return eventbus.NewInMemoryBus(4, 1024) },
		sqsqueue.NewNetworkQueue,
		sqsqueue.NewNetworkConsumer,
		sqsqueue.NewBlockQueue,
		sqsqueue.NewBlockQueueConsumer,
		providerPostgresTxRepository,
		providerTxRepositoryPort,
		providerWalletRepositoryPort,
		providerNetworkRepositoryPort,
		providerNetworkProducerPort,
		providerBlockProducerPort,
		providerBlockEventProcessorPort,
		monitor.NewInterestStore,
		monitor.NewBloomFilterCache,
		monitor.NewBlockProducerService,
		monitor.NewBlockConsumerService,
		http.NewFiberServer,
		providerETHRPC,
		monitor.NewMonitorService,
	),
	fx.Invoke(func(lc fx.Lifecycle, cfg *config.Config, h *http.FiberServer) {
		if cfg.HTTPEnabled {
			h.Start(lc)
		}
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
	fx.Invoke(func(lc fx.Lifecycle, consumer *sqsqueue.NetworkConsumer) {
		var cancel context.CancelFunc
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				runCtx, stop := context.WithCancel(context.Background())
				cancel = stop
				consumer.Start(runCtx)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if cancel != nil {
					cancel()
				}
				return nil
			},
		})
	}),
	fx.Invoke(func(lc fx.Lifecycle, producer *monitor.BlockProducerService, consumer *sqsqueue.BlockQueueConsumer) {
		var producerCancel context.CancelFunc
		var consumerCancel context.CancelFunc
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				producerCtx, stopProducer := context.WithCancel(context.Background())
				consumerCtx, stopConsumer := context.WithCancel(context.Background())
				producerCancel = stopProducer
				consumerCancel = stopConsumer
				producer.Start(producerCtx)
				consumer.Start(consumerCtx)
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if producerCancel != nil {
					producerCancel()
				}
				if consumerCancel != nil {
					consumerCancel()
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

func providerPostgresTxRepository(logger *zap.Logger, cfg *config.Config) (*postgres.PostgresTxRepository, error) {
	return postgres.NewPostgresTxRepository(cfg.DatabaseURL, cfg.MigrationsDir, logger)
}

func providerTxRepositoryPort(repo *postgres.PostgresTxRepository) ports.TxRepositoryPort {
	return repo
}

func providerWalletRepositoryPort(repo *postgres.PostgresTxRepository) ports.WalletRepositoryPort {
	return repo
}

func providerNetworkRepositoryPort(repo *postgres.PostgresTxRepository) ports.NetworkRepositoryPort {
	return repo
}

func providerNetworkProducerPort(queue *sqsqueue.NetworkQueue) ports.NetworkProducerPort {
	return queue
}

func providerBlockProducerPort(queue *sqsqueue.BlockQueue) ports.BlockProducerPort {
	return queue
}

func providerBlockEventProcessorPort(processor *monitor.BlockConsumerService) ports.BlockEventProcessorPort {
	return processor
}
