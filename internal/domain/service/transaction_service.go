package service

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"

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

	tx.Status = entity.TxStatusPending

	if err := s.repo.Save(ctx, tx); err != nil {
		return err
	}

	// TODO: Add logger when fx provider is set up
	s.logger.Sugar().Infof("Transaction created with ID %s and hash %s\n", tx.ID, tx.TxHash)

	return nil
}
