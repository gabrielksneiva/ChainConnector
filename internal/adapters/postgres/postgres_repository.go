package postgres

import (
	"ChainConnector/internal/domain/entity"
	"ChainConnector/internal/domain/ports"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgresTxRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

var _ ports.TxRepositoryPort = (*PostgresTxRepository)(nil)
var _ ports.WalletRepositoryPort = (*PostgresTxRepository)(nil)
var _ ports.NetworkRepositoryPort = (*PostgresTxRepository)(nil)

func NewPostgresTxRepository(dsn string, migrationsDir string, logger *zap.Logger) (*PostgresTxRepository, error) {
	if dsn == "" {
		return nil, errors.New("missing DATABASE_URL")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(15 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := RunMigrations(context.Background(), db, migrationsDir, logger); err != nil {
		return nil, err
	}

	return &PostgresTxRepository{db: db, logger: logger}, nil
}

func (r *PostgresTxRepository) Save(ctx context.Context, tx *entity.Transaction) error {
	if tx == nil || tx.ID == "" {
		return errors.New("invalid transaction")
	}

	payload, err := json.Marshal(tx.Data)
	if err != nil {
		return err
	}
	receiptJSON, err := json.Marshal(tx.Receipt)
	if err != nil {
		return err
	}

	query := `INSERT INTO transactions (
        id, tx_hash, chain_id, from_address, to_address, value, nonce, gas_limit,
        gas_price, raw_tx, payload, receipt, status, created_at, updated_at
      ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
      ON CONFLICT (id) DO UPDATE SET
        tx_hash = EXCLUDED.tx_hash,
        chain_id = EXCLUDED.chain_id,
        from_address = EXCLUDED.from_address,
        to_address = EXCLUDED.to_address,
        value = EXCLUDED.value,
        nonce = EXCLUDED.nonce,
        gas_limit = EXCLUDED.gas_limit,
        gas_price = EXCLUDED.gas_price,
        raw_tx = EXCLUDED.raw_tx,
        payload = EXCLUDED.payload,
        receipt = EXCLUDED.receipt,
        status = EXCLUDED.status,
        updated_at = EXCLUDED.updated_at`

	_, err = r.db.ExecContext(ctx, query,
		tx.ID,
		tx.TxHash,
		tx.Chain,
		tx.From,
		tx.To,
		bigIntToNullString(tx.Value),
		tx.Nonce,
		tx.Gas,
		bigIntToNullString(tx.GasPrice),
		tx.RawTxHex,
		payload,
		receiptJSON,
		tx.Status.String(),
		tx.CreatedAt,
		tx.UpdatedAt,
	)
	return err
}

func (r *PostgresTxRepository) FindByID(ctx context.Context, id string) (*entity.Transaction, error) {
	query := `SELECT id, tx_hash, chain_id, from_address, to_address, value, nonce,
        gas_limit, gas_price, raw_tx, receipt, status, created_at, updated_at
      FROM transactions WHERE id = $1`
	return r.scanTransaction(ctx, query, id)
}

func (r *PostgresTxRepository) FindByHash(ctx context.Context, hash string) (*entity.Transaction, error) {
	if hash == "" {
		return nil, nil
	}
	query := `SELECT id, tx_hash, chain_id, from_address, to_address, value, nonce,
        gas_limit, gas_price, raw_tx, receipt, status, created_at, updated_at
      FROM transactions WHERE tx_hash = $1`
	return r.scanTransaction(ctx, query, hash)
}

func (r *PostgresTxRepository) UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error {
	if txID == "" {
		return errors.New("transaction id required")
	}

	setClauses := []string{"status = $1", "updated_at = $2"}
	args := []interface{}{status.String(), time.Now().UTC()}
	idx := 3
	if h, ok := updates["tx_hash"].(string); ok && h != "" {
		setClauses = append(setClauses, "tx_hash = $"+strconv.Itoa(idx))
		args = append(args, h)
		idx++
	}
	if receipt, ok := updates["receipt"].(*entity.Receipt); ok && receipt != nil {
		receiptJSON, err := json.Marshal(receipt)
		if err != nil {
			return err
		}
		setClauses = append(setClauses, "receipt = $"+strconv.Itoa(idx))
		args = append(args, receiptJSON)
		idx++
	}
	if errMsg, ok := updates["error_message"].(string); ok && errMsg != "" {
		setClauses = append(setClauses, "error_message = $"+strconv.Itoa(idx))
		args = append(args, errMsg)
		idx++
	}
	args = append(args, txID)
	query := fmt.Sprintf("UPDATE transactions SET %s WHERE id = $%d", strings.Join(setClauses, ", "), idx)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresTxRepository) ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, tx_hash, chain_id, from_address, to_address, value, nonce,
        gas_limit, gas_price, raw_tx, receipt, status, created_at, updated_at
      FROM transactions WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.Transaction, 0)
	for rows.Next() {
		tx, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, tx)
	}
	return result, rows.Err()
}

func (r *PostgresTxRepository) ListTransactions(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT id, tx_hash, chain_id, from_address, to_address, value, nonce,
        gas_limit, gas_price, raw_tx, receipt, status, created_at, updated_at
      FROM transactions ORDER BY created_at DESC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*entity.Transaction, 0)
	for rows.Next() {
		tx, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, tx)
	}
	return result, rows.Err()
}

func (r *PostgresTxRepository) scanTransaction(ctx context.Context, query string, args ...interface{}) (*entity.Transaction, error) {
	row := r.db.QueryRowContext(ctx, query, args...)
	return r.scanRow(row)
}

func (r *PostgresTxRepository) scanRow(scanner interface {
	Scan(dest ...interface{}) error
}) (*entity.Transaction, error) {
	var (
		tx           entity.Transaction
		value        sql.NullString
		gasPrice     sql.NullString
		receiptJSON  sql.NullString
		toAddress    sql.NullString
		statusString string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := scanner.Scan(
		&tx.ID,
		&tx.TxHash,
		&tx.Chain,
		&tx.From,
		&toAddress,
		&value,
		&tx.Nonce,
		&tx.Gas,
		&gasPrice,
		&tx.RawTxHex,
		&receiptJSON,
		&statusString,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if toAddress.Valid {
		tx.To = &toAddress.String
	}
	if value.Valid {
		tx.Value = new(big.Int)
		_, ok := tx.Value.SetString(value.String, 10)
		if !ok {
			return nil, fmt.Errorf("invalid value stored for transaction %s", tx.ID)
		}
	}
	if gasPrice.Valid {
		tx.GasPrice = new(big.Int)
		_, ok := tx.GasPrice.SetString(gasPrice.String, 10)
		if !ok {
			return nil, fmt.Errorf("invalid gas_price stored for transaction %s", tx.ID)
		}
	}
	tx.Status = entity.ParseTxStatus(statusString)
	if receiptJSON.Valid {
		var receipt entity.Receipt
		if err := json.Unmarshal([]byte(receiptJSON.String), &receipt); err == nil {
			tx.Receipt = &receipt
		}
	}
	tx.CreatedAt = createdAt
	tx.UpdatedAt = updatedAt
	return &tx, nil
}

func bigIntToNullString(v *big.Int) interface{} {
	if v == nil {
		return nil
	}
	return v.String()
}

// AddInterestAddress adds an address to monitor for a specific chain
func (r *PostgresTxRepository) AddInterestAddress(ctx context.Context, address string, chain string) error {
	query := `INSERT INTO interest_addresses (address, chain, created_at) VALUES ($1, $2, $3)
	          ON CONFLICT (address, chain) DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, address, chain, time.Now().UTC())
	return err
}

// GetInterestAddresses returns all addresses being monitored for a chain
func (r *PostgresTxRepository) GetInterestAddresses(ctx context.Context, chain string) ([]string, error) {
	query := `SELECT address FROM interest_addresses WHERE chain = $1`
	rows, err := r.db.QueryContext(ctx, query, chain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []string
	for rows.Next() {
		var addr string
		if err := rows.Scan(&addr); err != nil {
			return nil, err
		}
		addresses = append(addresses, addr)
	}
	return addresses, rows.Err()
}

func (r *PostgresTxRepository) SaveWallet(ctx context.Context, wallet *entity.Wallet) error {
	if wallet == nil || wallet.Address == "" {
		return errors.New("invalid wallet")
	}
	query := `INSERT INTO wallets (id, address, chain, private_key, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          ON CONFLICT (address) DO UPDATE SET
	            chain = EXCLUDED.chain,
	            private_key = EXCLUDED.private_key,
	            updated_at = EXCLUDED.updated_at`
	_, err := r.db.ExecContext(ctx, query,
		wallet.ID,
		wallet.Address,
		wallet.Chain,
		wallet.PrivateKey,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)
	return err
}

func (r *PostgresTxRepository) FindWalletByID(ctx context.Context, id string) (*entity.Wallet, error) {
	query := `SELECT id, address, chain, private_key, created_at, updated_at FROM wallets WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var wallet entity.Wallet
	if err := row.Scan(&wallet.ID, &wallet.Address, &wallet.Chain, &wallet.PrivateKey, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *PostgresTxRepository) FindWalletByAddress(ctx context.Context, address string) (*entity.Wallet, error) {
	query := `SELECT id, address, chain, private_key, created_at, updated_at FROM wallets WHERE address = $1`
	row := r.db.QueryRowContext(ctx, query, address)
	var wallet entity.Wallet
	if err := row.Scan(&wallet.ID, &wallet.Address, &wallet.Chain, &wallet.PrivateKey, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *PostgresTxRepository) ListWallets(ctx context.Context, chain string) ([]*entity.Wallet, error) {
	query := `SELECT id, address, chain, created_at, updated_at FROM wallets WHERE chain = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, chain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wallets := make([]*entity.Wallet, 0)
	for rows.Next() {
		var wallet entity.Wallet
		if err := rows.Scan(&wallet.ID, &wallet.Address, &wallet.Chain, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}
	return wallets, rows.Err()
}

func (r *PostgresTxRepository) ListAllWallets(ctx context.Context) ([]*entity.Wallet, error) {
	query := `SELECT id, address, chain, created_at, updated_at FROM wallets ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wallets := make([]*entity.Wallet, 0)
	for rows.Next() {
		var wallet entity.Wallet
		if err := rows.Scan(&wallet.ID, &wallet.Address, &wallet.Chain, &wallet.CreatedAt, &wallet.UpdatedAt); err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}
	return wallets, rows.Err()
}

// GetBalance returns the current balance for an address on a chain
func (r *PostgresTxRepository) GetBalance(ctx context.Context, address string, chain string) (*big.Int, error) {
	query := `SELECT balance FROM user_balances WHERE address = $1 AND chain = $2`
	row := r.db.QueryRowContext(ctx, query, address, chain)

	var balanceStr string
	err := row.Scan(&balanceStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return big.NewInt(0), nil
		}
		return nil, err
	}

	balance := new(big.Int)
	_, ok := balance.SetString(balanceStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid balance stored for %s@%s", address, chain)
	}
	return balance, nil
}

// UpdateBalance updates the balance for an address on a chain
func (r *PostgresTxRepository) UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error {
	if amount == nil {
		return errors.New("amount cannot be nil")
	}

	query := `INSERT INTO user_balances (address, chain, balance, updated_at) VALUES ($1, $2, $3, $4)
	          ON CONFLICT (address, chain) DO UPDATE SET
	            balance = EXCLUDED.balance,
	            updated_at = EXCLUDED.updated_at`
	_, err := r.db.ExecContext(ctx, query, address, chain, amount.String(), time.Now().UTC())
	return err
}

func (r *PostgresTxRepository) SaveNetwork(ctx context.Context, network *entity.Network) error {
	if network == nil || network.ID == "" || network.Name == "" || network.ChainID <= 0 {
		return errors.New("invalid network")
	}

	query := `INSERT INTO networks (
	    id, name, chain_id, rpc_url, currency_symbol, explorer_url, enabled, created_at, updated_at
	  ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	  ON CONFLICT (chain_id) DO UPDATE SET
	    id = EXCLUDED.id,
	    name = EXCLUDED.name,
	    rpc_url = EXCLUDED.rpc_url,
	    currency_symbol = EXCLUDED.currency_symbol,
	    explorer_url = EXCLUDED.explorer_url,
	    enabled = EXCLUDED.enabled,
	    updated_at = EXCLUDED.updated_at`
	_, err := r.db.ExecContext(ctx, query,
		network.ID,
		network.Name,
		network.ChainID,
		nullString(network.RPCURL),
		nullString(network.CurrencySymbol),
		nullString(network.ExplorerURL),
		network.Enabled,
		network.CreatedAt,
		network.UpdatedAt,
	)
	return err
}

func (r *PostgresTxRepository) FindNetworkByID(ctx context.Context, id string) (*entity.Network, error) {
	query := `SELECT id, name, chain_id, rpc_url, currency_symbol, explorer_url, enabled, created_at, updated_at
	  FROM networks WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanNetwork(row)
}

func (r *PostgresTxRepository) ListNetworks(ctx context.Context) ([]*entity.Network, error) {
	query := `SELECT id, name, chain_id, rpc_url, currency_symbol, explorer_url, enabled, created_at, updated_at
	  FROM networks ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	networks := make([]*entity.Network, 0)
	for rows.Next() {
		network, err := scanNetwork(rows)
		if err != nil {
			return nil, err
		}
		networks = append(networks, network)
	}
	return networks, rows.Err()
}

func scanNetwork(scanner interface {
	Scan(dest ...interface{}) error
}) (*entity.Network, error) {
	var (
		network        entity.Network
		rpcURL         sql.NullString
		currencySymbol sql.NullString
		explorerURL    sql.NullString
	)
	err := scanner.Scan(
		&network.ID,
		&network.Name,
		&network.ChainID,
		&rpcURL,
		&currencySymbol,
		&explorerURL,
		&network.Enabled,
		&network.CreatedAt,
		&network.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if rpcURL.Valid {
		network.RPCURL = rpcURL.String
	}
	if currencySymbol.Valid {
		network.CurrencySymbol = currencySymbol.String
	}
	if explorerURL.Valid {
		network.ExplorerURL = explorerURL.String
	}
	return &network, nil
}

func nullString(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
