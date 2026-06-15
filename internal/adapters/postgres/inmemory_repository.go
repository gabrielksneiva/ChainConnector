package postgres

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"log"
	"math/big"
	"sort"
	"strconv"
	"sync"
)

type InMemoryTxRepository struct {
	mu        sync.RWMutex
	byID      map[string]*entity.Transaction
	byHash    map[string]*entity.Transaction
	balances  map[string]*big.Int // key: "chain:address"
	interests map[string][]string // key: chain, value: []addresses
	wallets   map[string]*entity.Wallet
	networks  map[string]*entity.Network
}

func NewInMemoryTxRepository() ports.TxRepositoryPort {
	return &InMemoryTxRepository{
		byID:      make(map[string]*entity.Transaction),
		byHash:    make(map[string]*entity.Transaction),
		balances:  make(map[string]*big.Int),
		interests: make(map[string][]string),
		wallets:   make(map[string]*entity.Wallet),
		networks:  make(map[string]*entity.Network),
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
	log.Printf("repository: saved tx id=%s hash=%s status=%s", tx.ID, tx.TxHash, tx.Status)
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
	log.Printf("repository: updated status tx id=%s status=%s tx_hash=%s", txID, status, tx.TxHash)
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

func (r *InMemoryTxRepository) ListTransactions(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]*entity.Transaction, 0, len(r.byID))
	for _, tx := range r.byID {
		res = append(res, tx)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.After(res[j].CreatedAt)
	})
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

// AddInterestAddress adds an address to monitor for a specific chain
func (r *InMemoryTxRepository) AddInterestAddress(ctx context.Context, address string, chain string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.interests[chain] == nil {
		r.interests[chain] = []string{}
	}
	// Check if address already exists
	for _, addr := range r.interests[chain] {
		if addr == address {
			return nil
		}
	}
	r.interests[chain] = append(r.interests[chain], address)
	log.Printf("repository: added interest address=%s chain=%s", address, chain)
	return nil
}

// GetInterestAddresses returns all addresses being monitored for a chain
func (r *InMemoryTxRepository) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	addresses := make([]string, len(r.interests[chain]))
	copy(addresses, r.interests[chain])
	return addresses, nil
}

// SaveWallet stores or updates a wallet in memory
func (r *InMemoryTxRepository) SaveWallet(ctx context.Context, wallet *entity.Wallet) error {
	if wallet == nil || wallet.Address == "" {
		return errors.New("invalid wallet")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wallets[wallet.ID] = wallet
	return nil
}

func (r *InMemoryTxRepository) FindWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wallet, ok := r.wallets[id]
	if !ok {
		return nil, nil
	}
	return wallet, nil
}

func (r *InMemoryTxRepository) FindWalletByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, wallet := range r.wallets {
		if wallet.Address == address {
			return wallet, nil
		}
	}
	return nil, nil
}

func (r *InMemoryTxRepository) ListWallets(ctx context.Context, chain string) ([]*entity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var wallets []*entity.Wallet
	for _, wallet := range r.wallets {
		if wallet.Chain == chain {
			wallets = append(wallets, wallet)
		}
	}
	return wallets, nil
}

func (r *InMemoryTxRepository) ListAllWallets(ctx context.Context) ([]*entity.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	wallets := make([]*entity.Wallet, 0, len(r.wallets))
	for _, wallet := range r.wallets {
		copy := *wallet
		wallets = append(wallets, &copy)
	}
	return wallets, nil
}

// GetBalance returns the current balance for an address on a chain
func (r *InMemoryTxRepository) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	key := chain + ":" + address
	balance, ok := r.balances[key]
	if !ok {
		return big.NewInt(0), nil
	}
	return new(big.Int).Set(balance), nil
}

// UpdateBalance updates the balance for an address on a chain
func (r *InMemoryTxRepository) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	if amount == nil {
		return errors.New("amount cannot be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	key := chain + ":" + address
	r.balances[key] = new(big.Int).Set(amount)
	log.Printf("repository: updated balance address=%s chain=%s amount=%s", address, chain, amount.String())
	return nil
}

func (r *InMemoryTxRepository) SaveNetwork(ctx context.Context, network *entity.Network) error {
	if network == nil || network.ID == "" || network.Name == "" || network.ChainID <= 0 {
		return errors.New("invalid network")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	chainKey := strconv.FormatInt(network.ChainID, 10)
	for key, existing := range r.networks {
		if existing.ID == network.ID || existing.ChainID == network.ChainID {
			delete(r.networks, key)
		}
	}
	copy := *network
	r.networks[network.ID] = &copy
	r.networks[chainKey] = &copy
	log.Printf("repository: saved network id=%s name=%s chain_id=%d", network.ID, network.Name, network.ChainID)
	return nil
}

func (r *InMemoryTxRepository) FindNetworkByID(ctx context.Context, id string) (*entity.Network, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	network, ok := r.networks[id]
	if !ok {
		return nil, nil
	}
	copy := *network
	return &copy, nil
}

func (r *InMemoryTxRepository) ListNetworks(ctx context.Context) ([]*entity.Network, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	networks := make([]*entity.Network, 0)
	seen := make(map[string]struct{})
	for _, network := range r.networks {
		if _, ok := seen[network.ID]; ok {
			continue
		}
		copy := *network
		networks = append(networks, &copy)
		seen[network.ID] = struct{}{}
	}
	return networks, nil
}
