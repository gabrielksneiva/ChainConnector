package monitor

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"math/big"
	"testing"

	"go.uber.org/zap"
)

type fakeBlockChain struct {
	block   *entity.Block
	receipt *entity.Receipt
}

func (f *fakeBlockChain) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (f *fakeBlockChain) GetNonce(ctx context.Context, address string) (uint64, error) {
	return 0, nil
}

func (f *fakeBlockChain) GetTransactionReceipt(ctx context.Context, txHash string) (*entity.Receipt, error) {
	return f.receipt, nil
}

func (f *fakeBlockChain) GetLogs(ctx context.Context, filter entity.LogFilter) ([]entity.Log, error) {
	return nil, nil
}

func (f *fakeBlockChain) GetBlockNumber(ctx context.Context) (uint64, error) {
	return 0, nil
}

func (f *fakeBlockChain) GetBlockByNumber(ctx context.Context, chain string, number uint64) (*entity.Block, error) {
	return f.block, nil
}

func (f *fakeBlockChain) EstimateFees(ctx context.Context, chain string) (*big.Int, *big.Int, error) {
	return big.NewInt(1), big.NewInt(1), nil
}

func (f *fakeBlockChain) EstimateGas(ctx context.Context, chain string, from string, to string, value *big.Int, data []byte) (uint64, error) {
	return 21000, nil
}

func (f *fakeBlockChain) SendRawTransaction(ctx context.Context, chain string, signedTx []byte) (string, error) {
	return "0xhash", nil
}

func (f *fakeBlockChain) SendRawTransactionHex(ctx context.Context, chain string, signedTxHex string) (string, error) {
	return "0xhash", nil
}

type fakeBlockRepo struct {
	wallets []*entity.Wallet
	saved   []*entity.Transaction
	byHash  map[string]*entity.Transaction
}

func (f *fakeBlockRepo) Save(ctx context.Context, tx *entity.Transaction) error {
	f.saved = append(f.saved, tx)
	if f.byHash == nil {
		f.byHash = make(map[string]*entity.Transaction)
	}
	f.byHash[tx.TxHash] = tx
	return nil
}

func (f *fakeBlockRepo) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	return nil, nil
}

func (f *fakeBlockRepo) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	if f.byHash == nil {
		return nil, nil
	}
	return f.byHash[hash], nil
}

func (f *fakeBlockRepo) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	return nil
}

func (f *fakeBlockRepo) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	return nil, nil
}

func (f *fakeBlockRepo) ListTransactions(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	return f.saved, nil
}

func (f *fakeBlockRepo) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (f *fakeBlockRepo) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	return nil
}

func (f *fakeBlockRepo) AddInterestAddress(ctx context.Context, address string, chain string) error {
	return nil
}

func (f *fakeBlockRepo) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	return nil, nil
}

func (f *fakeBlockRepo) SaveWallet(ctx context.Context, wallet *entity.Wallet) error {
	f.wallets = append(f.wallets, wallet)
	return nil
}

func (f *fakeBlockRepo) FindWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	return nil, nil
}

func (f *fakeBlockRepo) FindWalletByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	return nil, nil
}

func (f *fakeBlockRepo) ListWallets(ctx context.Context, chain string) ([]*entity.Wallet, error) {
	return f.wallets, nil
}

func (f *fakeBlockRepo) ListAllWallets(ctx context.Context) ([]*entity.Wallet, error) {
	return f.wallets, nil
}

func TestBlockConsumerSavesTransactionsForRegisteredWallets(t *testing.T) {
	walletAddress := "0x1111111111111111111111111111111111111111"
	blockchain := &fakeBlockChain{
		block: &entity.Block{
			Chain:  "sepolia",
			Number: 10,
			Transactions: []entity.BlockTransaction{
				{Hash: "0xmatch", From: "0x2222222222222222222222222222222222222222", To: walletAddress, Value: big.NewInt(7), Gas: 21000, GasPrice: big.NewInt(1)},
				{Hash: "0xskip", From: "0x3333333333333333333333333333333333333333", To: "0x4444444444444444444444444444444444444444"},
			},
		},
		receipt: &entity.Receipt{TxHash: "0xmatch", Status: entity.ReceiptStatusSuccess},
	}
	repo := &fakeBlockRepo{
		wallets: []*entity.Wallet{{ID: "wallet-1", Address: walletAddress, Chain: "sepolia"}},
	}
	service := NewBlockConsumerService(zap.NewNop(), blockchain, repo, repo)

	err := service.ProcessBlockEvent(context.Background(), &entity.BlockEvent{Chain: "sepolia", BlockNumber: 10})
	if err != nil {
		t.Fatalf("ProcessBlockEvent returned error: %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 matched transaction, got %d", len(repo.saved))
	}
	if repo.saved[0].TxHash != "0xmatch" {
		t.Fatalf("expected 0xmatch, got %s", repo.saved[0].TxHash)
	}
	if repo.saved[0].Status != entity.TxStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", repo.saved[0].Status)
	}
}

func TestBlockConsumerIgnoresWalletsFromOtherChains(t *testing.T) {
	walletAddress := "0x1111111111111111111111111111111111111111"
	blockchain := &fakeBlockChain{
		block: &entity.Block{
			Chain:  "sepolia",
			Number: 10,
			Transactions: []entity.BlockTransaction{
				{Hash: "0xmatch", From: walletAddress},
			},
		},
	}
	repo := &fakeBlockRepo{
		wallets: []*entity.Wallet{{ID: "wallet-1", Address: walletAddress, Chain: "polygon"}},
	}
	service := NewBlockConsumerService(zap.NewNop(), blockchain, repo, repo)

	err := service.ProcessBlockEvent(context.Background(), &entity.BlockEvent{Chain: "sepolia", BlockNumber: 10})
	if err != nil {
		t.Fatalf("ProcessBlockEvent returned error: %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("expected no saved transactions, got %d", len(repo.saved))
	}
}
