package service

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TransactionService struct {
	repo       ports.TxRepositoryPort
	blockchain ports.BlockchainPort
	logger     *zap.Logger
}

func NewTransactionService(repo ports.TxRepositoryPort, blockchain ports.BlockchainPort, logger *zap.Logger) *TransactionService {
	return &TransactionService{
		repo:       repo,
		blockchain: blockchain,
		logger:     logger,
	}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx *entity.Transaction) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	if tx.From == "" {
		return errors.New("transaction from address is required")
	}
	if tx.To == nil || *tx.To == "" {
		return errors.New("transaction to address is required")
	}
	if tx.Chain == "" {
		return errors.New("transaction chain is required")
	}
	if tx.Value == nil || tx.Value.Sign() <= 0 {
		return errors.New("transaction amount must be greater than zero")
	}
	if tx.Gas == 0 {
		return errors.New("transaction gas amount must be greater than zero")
	}
	if tx.GasPrice == nil || tx.GasPrice.Sign() <= 0 {
		return errors.New("transaction gas price must be greater than zero")
	}

	if tx.ID == "" {
		tx.ID = uuid.NewString()
	}

	now := time.Now().UTC()
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = now
	}
	tx.UpdatedAt = now

	if tx.Status == entity.TxStatusUnknown {
		tx.Status = entity.TxStatusPending
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		return err
	}

	s.logger.Sugar().Infof("Transaction created with ID %s and hash %s", tx.ID, tx.TxHash)
	return nil
}

func (s *TransactionService) SignAndSendTransaction(ctx context.Context, tx *entity.Transaction, signer ports.WalletSignerPort, blockchain ports.BlockchainPort) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	if signer == nil {
		return errors.New("signer is required")
	}
	if blockchain == nil {
		return errors.New("blockchain port is required")
	}
	if tx.ChainID == nil {
		tx.ChainID = chainIDFromName(tx.Chain)
		if tx.ChainID == nil {
			return errors.New("transaction chain id is required for signing")
		}
	}

	signedTx, err := signer.SignTransaction(ctx, tx)
	if err != nil {
		return err
	}
	if len(signedTx) == 0 {
		return errors.New("signed transaction is empty")
	}

	tx.RawTxHex = "0x" + hex.EncodeToString(signedTx)
	txHash, err := blockchain.SendRawTransaction(ctx, tx.Chain, signedTx)
	if err != nil {
		return err
	}
	tx.TxHash = txHash
	tx.Status = entity.TxStatusPending
	if err := s.repo.Save(ctx, tx); err != nil {
		return err
	}

	s.logger.Sugar().Infof("Transaction signed and sent id=%s hash=%s", tx.ID, txHash)
	return nil
}

func chainIDFromName(name string) *big.Int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mainnet", "eth", "ethereum":
		return big.NewInt(1)
	case "sepolia":
		return big.NewInt(11155111)
	case "goerli":
		return big.NewInt(5)
	case "polygon":
		return big.NewInt(137)
	case "mumbai":
		return big.NewInt(80001)
	default:
		return nil
	}
}
