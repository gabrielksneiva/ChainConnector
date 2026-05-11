package service

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"errors"
	"math/big"
	"testing"

	"go.uber.org/zap"
)

type mockRepo struct {
	saved   *entity.Transaction
	byID    map[string]*entity.Transaction
	byHash  map[string]*entity.Transaction
	updated map[string][]interface{}
}

// repoErr is a test repo implementation that returns an error on Save.
type repoErr struct{}

func (r *repoErr) Save(ctx context.Context, tx *entity.Transaction) error {
	return errors.New("save failed")
}
func (r *repoErr) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	return nil
}
func (r *repoErr) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	return nil, nil
}
func (r *repoErr) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	return nil, nil
}
func (r *repoErr) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	return nil, nil
}

func (r *repoErr) AddInterestAddress(ctx context.Context, address string, chain string) error {
	return nil
}

func (r *repoErr) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	return []string{}, nil
}

func (r *repoErr) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (r *repoErr) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	return nil
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

func (m *mockRepo) AddInterestAddress(ctx context.Context, address string, chain string) error {
	return nil
}

func (m *mockRepo) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	return []string{}, nil
}

func (m *mockRepo) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (m *mockRepo) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	return nil
}

func TestCreateTransaction_nil(t *testing.T) {
	svc := NewTransactionService(&mockRepo{}, zap.NewNop())
	if err := svc.CreateTransaction(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil tx")
	}
}

func TestCreateTransaction_success(t *testing.T) {
	repo := &mockRepo{byID: map[string]*entity.Transaction{}}
	svc := NewTransactionService(repo, zap.NewNop())

	amount := big.NewInt(1000)
	gasPrice := big.NewInt(100)
	to := "0xabc123"
	tx := &entity.Transaction{
		ID:       "t1",
		From:     "0xfrom",
		To:       &to,
		Chain:    "sepolia",
		Value:    amount,
		Gas:      21000,
		GasPrice: gasPrice,
	}
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

func TestCreateTransaction_SaveError(t *testing.T) {
	svc := NewTransactionService(&repoErr{}, zap.NewNop())
	tx := &entity.Transaction{ID: "t2"}
	if err := svc.CreateTransaction(context.Background(), tx); err == nil {
		t.Fatalf("expected save error propagated")
	}
}
