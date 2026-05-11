package monitor

import (
	"ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type MonitorService struct {
	logger      *zap.Logger
	blockchain  ports.BlockchainPort
	repo        ports.TxRepositoryPort
	filter      *BloomFilterCache
	interest    *InterestStore
	interval    time.Duration
	enabled     bool
	chain       string
	lastScanned uint64
}

func NewMonitorService(cfg *config.Config, logger *zap.Logger, blockchain ports.BlockchainPort, repo ports.TxRepositoryPort, filter *BloomFilterCache, interest *InterestStore) *MonitorService {
	if cfg == nil {
		cfg = &config.Config{}
	}
	chainName := strings.TrimSpace(strings.ToLower(cfg.DefaultChain))
	if chainName == "" {
		chainName = "sepolia"
	}
	interval := cfg.MonitorInterval
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &MonitorService{
		logger:     logger,
		blockchain: blockchain,
		repo:       repo,
		filter:     filter,
		interest:   interest,
		interval:   interval,
		enabled:    cfg.MonitorEnabled,
		chain:      chainName,
	}
}

func (m *MonitorService) Start(ctx context.Context) {
	if ctx == nil || !m.enabled {
		m.logger.Info("monitor service disabled or missing context")
		return
	}
	go m.run(ctx)
}

func (m *MonitorService) run(ctx context.Context) {
	m.logger.Info("monitor service started", zap.String("chain", m.chain), zap.Duration("interval", m.interval))
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		if err := m.scan(ctx); err != nil {
			m.logger.Error("monitor scan failed", zap.Error(err))
		}
		select {
		case <-ctx.Done():
			m.logger.Info("monitor service stopped")
			return
		case <-ticker.C:
		}
	}
}

func (m *MonitorService) scan(ctx context.Context) error {
	latest, err := m.blockchain.GetBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed get latest block: %w", err)
	}

	from := m.lastScanned + 1
	if m.lastScanned == 0 {
		if latest > 0 {
			from = latest - 1
		} else {
			from = 0
		}
	}
	if latest < from {
		return nil
	}
	if latest == from {
		m.lastScanned = latest
		return nil
	}

	if err := m.syncInterestAddresses(ctx); err != nil {
		m.logger.Warn("failed to synchronize interest addresses", zap.Error(err))
	}

	filter := m.interest.ToLogFilter(&from, &latest)
	if len(filter.Addresses) == 0 && len(filter.Topics) == 0 && len(m.interest.GetTxHashes()) == 0 {
		m.logger.Info("monitor scan skipped, no interest addresses, topics or tx hashes configured")
		m.lastScanned = latest
		return nil
	}

	logs, err := m.blockchain.GetLogs(ctx, filter)
	if err != nil {
		return fmt.Errorf("get logs failed: %w", err)
	}

	m.logger.Info("monitor scan found logs", zap.Int("count", len(logs)), zap.Uint64("from", from), zap.Uint64("to", latest))

	for _, l := range logs {
		if l.TxHash == "" {
			continue
		}
		if m.filter.TestString(l.TxHash) {
			continue
		}
		m.filter.AddString(l.TxHash)

		tx, err := m.repo.FindByHash(ctx, l.TxHash)
		if err != nil {
			m.logger.Error("repo find by hash failed", zap.Error(err), zap.String("tx_hash", l.TxHash))
			continue
		}
		if tx == nil {
			tx = &entity.Transaction{
				Chain:     m.chain,
				TxHash:    l.TxHash,
				Status:    entity.TxStatusPending,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if receipt, err := m.blockchain.GetTransactionReceipt(ctx, l.TxHash); err == nil && receipt != nil {
				tx.Receipt = receipt
				if receipt.Status == entity.ReceiptStatusFailed {
					tx.Status = entity.TxStatusFailed
				} else {
					tx.Status = entity.TxStatusConfirmed
				}
			}
			if err := m.repo.Save(ctx, tx); err != nil {
				m.logger.Error("repo save tx failed", zap.Error(err), zap.String("tx_hash", l.TxHash))
			}
		}
	}

	if err := m.processInterestTxHashes(ctx); err != nil {
		m.logger.Warn("failed to process interest tx hashes", zap.Error(err))
	}

	return m.updatePending(ctx)
}

func (m *MonitorService) syncInterestAddresses(ctx context.Context) error {
	addresses, err := m.repo.GetInterestAddresses(ctx, m.chain)
	if err != nil {
		return err
	}
	for _, addr := range addresses {
		m.interest.AddAddress(addr)
	}
	return nil
}

func (m *MonitorService) processInterestTxHashes(ctx context.Context) error {
	for _, txHash := range m.interest.GetTxHashes() {
		if txHash == "" || m.filter.TestString(txHash) {
			continue
		}

		receipt, err := m.blockchain.GetTransactionReceipt(ctx, txHash)
		if err != nil {
			m.logger.Warn("failed fetch receipt for interest tx hash", zap.Error(err), zap.String("tx_hash", txHash))
			continue
		}
		if receipt == nil {
			continue
		}

		status := entity.TxStatusConfirmed
		if receipt.Status == entity.ReceiptStatusFailed {
			status = entity.TxStatusFailed
		}

		tx, err := m.repo.FindByHash(ctx, txHash)
		if err != nil {
			m.logger.Error("repo find by hash failed", zap.Error(err), zap.String("tx_hash", txHash))
			continue
		}
		if tx == nil {
			tx = &entity.Transaction{
				Chain:     m.chain,
				TxHash:    txHash,
				Status:    status,
				Receipt:   receipt,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := m.repo.Save(ctx, tx); err != nil {
				m.logger.Error("repo save tx failed", zap.Error(err), zap.String("tx_hash", txHash))
			}
			continue
		}
		if tx.Status != status {
			if err := m.repo.UpdateStatus(ctx, tx.ID, status, map[string]interface{}{"tx_hash": txHash}); err != nil {
				m.logger.Error("failed update tx status", zap.Error(err), zap.String("tx_hash", txHash), zap.String("tx_id", tx.ID))
			}
		}
	}
	return nil
}

func (m *MonitorService) updatePending(ctx context.Context) error {
	pending, err := m.repo.ListPending(ctx, 100)
	if err != nil {
		return err
	}

	for _, tx := range pending {
		if tx.TxHash == "" {
			continue
		}
		receipt, err := m.blockchain.GetTransactionReceipt(ctx, tx.TxHash)
		if err != nil {
			m.logger.Warn("failed fetch receipt for pending tx", zap.Error(err), zap.String("tx_hash", tx.TxHash))
			continue
		}
		if receipt == nil {
			continue
		}
		status := entity.TxStatusConfirmed
		if receipt.Status == entity.ReceiptStatusFailed {
			status = entity.TxStatusFailed
		}
		if err := m.repo.UpdateStatus(ctx, tx.ID, status, map[string]interface{}{"tx_hash": tx.TxHash}); err != nil {
			m.logger.Error("failed update pending tx status", zap.Error(err), zap.String("tx_id", tx.ID))
		}
	}
	return nil
}
