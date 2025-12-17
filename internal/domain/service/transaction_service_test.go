package service

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"errors"
	"testing"
)

type mockRepo struct {
	saved   *entity.Transaction
	byID    map[string]*entity.Transaction
	byHash  map[string]*entity.Transaction
	updated map[string][]interface{}
}

func (m *mockRepo) Save(ctx context.Context, tx *entity.Transaction) error {
	if m.saved == nil {
		m.saved = tx
	}
	if m.byID == nil {
		m.byID = map[string]*entity.Transaction{}
	}
	m.byID[tx.ID] = tx
	return nil
}
func (m *mockRepo) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	if m.updated == nil {
		m.updated = map[string][]interface{}{}
	}
	m.updated[txID] = append(m.updated[txID], status)
	if tx, ok := m.byID[txID]; ok {
		tx.Status = status
	}
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	if m.byID == nil {
		return nil, errors.New("not found")
	}
	tx, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return tx, nil
}
func (m *mockRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	if m.byHash == nil {
		return nil, nil
	}
	return m.byHash[hash], nil
}
func (m *mockRepo) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	var out []*entity.Transaction
	for _, tx := range m.byID {
		if tx.Status == entity.TxStatusPending {
			out = append(out, tx)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func TestCreateTransaction_nil(t *testing.T) {
	svc := NewTransactionService(&mockRepo{})
	if err := svc.CreateTransaction(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil tx")
	}
}

func TestCreateTransaction_success(t *testing.T) {
	repo := &mockRepo{byID: map[string]*entity.Transaction{}}
	svc := NewTransactionService(repo)

	tx := &entity.Transaction{ID: "t1"}
	if err := svc.CreateTransaction(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	if tx.Status != entity.TxStatusPending {
		t.Fatalf("expected pending, got %v", tx.Status)
	}
	// if len(pub.published) == 0 {
	// 	t.Fatalf("expected event published")
	// }
}
