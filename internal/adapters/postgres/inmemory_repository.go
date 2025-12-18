package postgres

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"sync"
)

type InMemoryTxRepository struct {
	mu     sync.RWMutex
	byID   map[string]*entity.Transaction
	byHash map[string]*entity.Transaction
}

func NewInMemoryTxRepository() ports.TxRepositoryPort {
	return &InMemoryTxRepository{
		byID:   make(map[string]*entity.Transaction),
		byHash: make(map[string]*entity.Transaction),
	}
}

func (r *InMemoryTxRepository) Save(ctx context.Context, tx *entity.Transaction) error {
	if tx == nil || tx.ID == "" {
		return errors.New("invalid transaction")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[tx.ID] = tx
	if tx.TxHash != "" {
		r.byHash[tx.TxHash] = tx
	}
	return nil
}

func (r *InMemoryTxRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.byID[id]
	if !ok {
		return nil, nil
	}
	return tx, nil
}

func (r *InMemoryTxRepository) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.byHash[hash]
	if !ok {
		return nil, nil
	}
	return tx, nil
}

func (r *InMemoryTxRepository) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.byID[txID]
	if !ok {
		return errors.New("transaction not found")
	}
	tx.Status = status
	if updates != nil {
		if h, ok := updates["tx_hash"].(string); ok && h != "" {
			tx.TxHash = h
			r.byHash[h] = tx
		}
	}
	return nil
}

func (r *InMemoryTxRepository) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]*entity.Transaction, 0, 10)
	for _, tx := range r.byID {
		if tx.Status == entity.TxStatusPending {
			res = append(res, tx)
			if limit > 0 && len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}
