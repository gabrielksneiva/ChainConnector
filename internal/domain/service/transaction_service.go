package service

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TransactionService struct {
	repo   ports.TxRepositoryPort
	logger *zap.Logger
}

func NewTransactionService(repo ports.TxRepositoryPort, logger *zap.Logger) *TransactionService {
	return &TransactionService{
		repo:   repo,
		logger: logger,
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
