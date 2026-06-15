package monitor

import (
	"ChainConnector/internal/config"
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BlockProducerService struct {
	logger   *zap.Logger
	producer ports.BlockProducerPort
	enabled  bool
	chain    string
	wsURL    string
}

func NewBlockProducerService(cfg *config.Config, logger *zap.Logger, producer ports.BlockProducerPort) *BlockProducerService {
	if cfg == nil {
		cfg = &config.Config{}
	}
	chain := normalizeChainName(cfg.DefaultChain)
	wsURL := cfg.SepoliaWSURL
	if chain == "eth" || chain == "ethereum" {
		wsURL = cfg.EthWSURL
	}
	return &BlockProducerService{
		logger:   logger,
		producer: producer,
		enabled:  cfg.BlockProducerEnabled,
		chain:    chain,
		wsURL:    strings.TrimSpace(wsURL),
	}
}

func (s *BlockProducerService) Start(ctx context.Context) {
	if s == nil || ctx == nil || !s.enabled {
		if s != nil && s.logger != nil {
			s.logger.Info("block producer disabled")
		}
		return
	}
	if s.producer == nil || !s.producer.Enabled() {
		s.logger.Info("block producer queue disabled")
		return
	}
	if s.wsURL == "" {
		s.logger.Error("block producer missing websocket url")
		return
	}
	go s.run(ctx)
}

func (s *BlockProducerService) run(ctx context.Context) {
	backoff := 2 * time.Second
	for {
		if ctx.Err() != nil {
			s.logger.Info("block producer stopped")
			return
		}
		if err := s.subscribe(ctx); err != nil {
			if ctx.Err() != nil {
				return
			}
			s.logger.Error("block producer websocket subscription failed", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			if backoff < 30*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = 2 * time.Second
	}
}

func (s *BlockProducerService) subscribe(ctx context.Context) error {
	client, err := ethclient.DialContext(ctx, s.wsURL)
	if err != nil {
		return fmt.Errorf("dial websocket endpoint for %s: %w", s.chain, err)
	}
	defer client.Close()

	headers := make(chan *types.Header, 16)
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return fmt.Errorf("subscribe new heads: %w", err)
	}
	defer sub.Unsubscribe()

	s.logger.Info("block producer subscribed", zap.String("chain", s.chain))
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			if err == nil {
				return errors.New("websocket subscription closed")
			}
			return err
		case header := <-headers:
			if header == nil || header.Number == nil {
				continue
			}
			event := &entity.BlockEvent{
				Chain:       s.chain,
				BlockNumber: header.Number.Uint64(),
				BlockHash:   header.Hash().Hex(),
				ReceivedAt:  time.Now().UTC(),
			}
			if err := s.producer.EnqueueBlockEvent(ctx, event); err != nil {
				s.logger.Error("failed enqueue block event", zap.Error(err), zap.Uint64("block_number", event.BlockNumber))
			}
		}
	}
}

type BlockConsumerService struct {
	logger     *zap.Logger
	blockchain ports.BlockchainPort
	txRepo     ports.TxRepositoryPort
	walletRepo ports.WalletRepositoryPort
}

var _ ports.BlockEventProcessorPort = (*BlockConsumerService)(nil)

func NewBlockConsumerService(logger *zap.Logger, blockchain ports.BlockchainPort, txRepo ports.TxRepositoryPort, walletRepo ports.WalletRepositoryPort) *BlockConsumerService {
	return &BlockConsumerService{
		logger:     logger,
		blockchain: blockchain,
		txRepo:     txRepo,
		walletRepo: walletRepo,
	}
}

func (s *BlockConsumerService) ProcessBlockEvent(ctx context.Context, event *entity.BlockEvent) error {
	if event == nil {
		return errors.New("block event is required")
	}
	chain := normalizeChainName(event.Chain)
	block, err := s.blockchain.GetBlockByNumber(ctx, chain, event.BlockNumber)
	if err != nil {
		return fmt.Errorf("get block %d on %s: %w", event.BlockNumber, chain, err)
	}
	if block == nil {
		s.logger.Warn("block not found", zap.String("chain", chain), zap.Uint64("block_number", event.BlockNumber))
		return nil
	}
	wallets, err := s.walletRepo.ListAllWallets(ctx)
	if err != nil {
		return fmt.Errorf("list wallets for block monitor: %w", err)
	}
	walletIndex := buildWalletIndex(wallets, chain)
	if len(walletIndex) == 0 {
		s.logger.Info("block consumer skipped, no registered wallets", zap.String("chain", chain), zap.Uint64("block_number", event.BlockNumber))
		return nil
	}

	matches := 0
	for _, blockTx := range block.Transactions {
		if !matchesWallet(blockTx, walletIndex) {
			continue
		}
		matches++
		if err := s.saveMatchedTransaction(ctx, chain, blockTx); err != nil {
			s.logger.Error("failed save matched block transaction", zap.Error(err), zap.String("tx_hash", blockTx.Hash))
		}
	}
	s.logger.Info("block consumed", zap.String("chain", chain), zap.Uint64("block_number", block.Number), zap.Int("transactions", len(block.Transactions)), zap.Int("wallet_matches", matches))
	return nil
}

func (s *BlockConsumerService) saveMatchedTransaction(ctx context.Context, chain string, blockTx entity.BlockTransaction) error {
	if blockTx.Hash == "" {
		return nil
	}
	existing, err := s.txRepo.FindByHash(ctx, blockTx.Hash)
	if err != nil {
		return err
	}
	receipt, receiptErr := s.blockchain.GetTransactionReceipt(ctx, blockTx.Hash)
	if receiptErr != nil {
		s.logger.Warn("failed fetch receipt for matched transaction", zap.Error(receiptErr), zap.String("tx_hash", blockTx.Hash))
	}
	status := entity.TxStatusConfirmed
	if receipt != nil && receipt.Status == entity.ReceiptStatusFailed {
		status = entity.TxStatusFailed
	}
	if existing != nil {
		return s.txRepo.UpdateStatus(ctx, existing.ID, status, map[string]interface{}{
			"tx_hash": blockTx.Hash,
			"receipt": receipt,
		})
	}

	to := strings.TrimSpace(blockTx.To)
	var toPtr *string
	if to != "" {
		toPtr = &to
	}
	value := blockTx.Value
	if value == nil {
		value = big.NewInt(0)
	}
	gasPrice := blockTx.GasPrice
	if gasPrice == nil {
		gasPrice = big.NewInt(0)
	}
	now := time.Now().UTC()
	tx := &entity.Transaction{
		ID:        uuid.NewString(),
		From:      blockTx.From,
		To:        toPtr,
		Chain:     chain,
		Value:     new(big.Int).Set(value),
		Gas:       blockTx.Gas,
		GasPrice:  new(big.Int).Set(gasPrice),
		Nonce:     blockTx.Nonce,
		TxHash:    blockTx.Hash,
		Status:    status,
		Receipt:   receipt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.txRepo.Save(ctx, tx)
}

func buildWalletIndex(wallets []*entity.Wallet, chain string) map[string]struct{} {
	index := make(map[string]struct{})
	for _, wallet := range wallets {
		if wallet == nil || strings.TrimSpace(wallet.Address) == "" {
			continue
		}
		walletChain := normalizeChainName(wallet.Chain)
		if walletChain != "" && chain != "" && walletChain != chain {
			continue
		}
		index[strings.ToLower(strings.TrimSpace(wallet.Address))] = struct{}{}
	}
	return index
}

func matchesWallet(tx entity.BlockTransaction, wallets map[string]struct{}) bool {
	if _, ok := wallets[strings.ToLower(strings.TrimSpace(tx.From))]; ok {
		return true
	}
	if _, ok := wallets[strings.ToLower(strings.TrimSpace(tx.To))]; ok {
		return true
	}
	return false
}

func normalizeChainName(chain string) string {
	chain = strings.TrimSpace(strings.ToLower(chain))
	if chain == "" {
		return "sepolia"
	}
	return chain
}
