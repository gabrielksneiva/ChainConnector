package postgres

import (
	"context"
	"testing"

	"ChainConnector/internal/domain/entity"
)

func TestInMemoryRepository_BasicLifecycle(t *testing.T) {
	repo := NewInMemoryTxRepository()
	ctx := context.Background()

	tx := &entity.Transaction{ID: "t1", TxHash: "h1", Status: entity.TxStatusPending}
	if err := repo.Save(ctx, tx); err != nil {
		t.Fatalf("save error: %v", err)
	}

	if got, _ := repo.FindByID(ctx, "t1"); got == nil {
		t.Fatalf("expected to find saved tx by id")
	}
	if got, _ := repo.FindByHash(ctx, "h1"); got == nil {
		t.Fatalf("expected to find saved tx by hash")
	}

	// update status and hash
	if err := repo.UpdateStatus(ctx, "t1", entity.TxStatusSent, map[string]interface{}{"tx_hash": "h2"}); err != nil {
		t.Fatalf("update status error: %v", err)
	}
	got, _ := repo.FindByID(ctx, "t1")
	if got.TxHash != "h2" || got.Status != entity.TxStatusSent {
		t.Fatalf("unexpected state after update: %+v", got)
	}

	// list pending should not include sent tx
	list, _ := repo.ListPending(ctx, 10)
	for _, v := range list {
		if v.ID == "t1" {
			t.Fatalf("sent tx still in pending list")
		}
	}
}
