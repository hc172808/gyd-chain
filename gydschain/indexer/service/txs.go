package service

import (
	"database/sql"

	"github.com/gydschain/gydschain/internal/chain"
	"github.com/gydschain/gydschain/internal/tx"
)

// TransactionIndexer indexes transaction data
type TransactionIndexer struct {
	db *sql.DB
}

// NewTransactionIndexer creates a new transaction indexer
func NewTransactionIndexer(db *sql.DB) *TransactionIndexer {
	return &TransactionIndexer{db: db}
}

// IndexTransaction indexes a transaction
func (ti *TransactionIndexer) IndexTransaction(dbTx *sql.Tx, block *chain.Block, txn *tx.Transaction, txIndex int) error {
	_, err := dbTx.Exec(`
		INSERT INTO transactions (hash, block_number, block_hash, tx_index, from_address,
		                         to_address, value, asset, fee, nonce, data, signature,
		                         tx_type, status, gas_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (hash) DO NOTHING
	`,
		txn.Hash(),
		block.Number,
		block.Hash(),
		txIndex,
		txn.From,
		txn.To,
		txn.Value.String(),
		txn.Asset,
		txn.Fee.String(),
		txn.Nonce,
		txn.Data,
		txn.Signature,
		txn.Type.String(),
		1, // Status - would come from receipt
		0, // Gas used - would come from receipt
	)
	return err
}

// GetTransaction retrieves a transaction by hash
func (ti *TransactionIndexer) GetTransaction(hash string) (*IndexedTransaction, error) {
	txn := &IndexedTransaction{}
	
	err := ti.db.QueryRow(`
		SELECT hash, block_number, block_hash, tx_index, from_address, to_address,
		       value, asset, fee, nonce, data, signature, tx_type, status, gas_used, created_at
		FROM transactions WHERE hash = $1
	`, hash).Scan(
		&txn.Hash, &txn.BlockNumber, &txn.BlockHash, &txn.TxIndex,
		&txn.From, &txn.To, &txn.Value, &txn.Asset, &txn.Fee, &txn.Nonce,
		&txn.Data, &txn.Signature, &txn.Type, &txn.Status, &txn.GasUsed, &txn.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return txn, err
}

// GetTransactionsByBlock retrieves transactions for a block
func (ti *TransactionIndexer) GetTransactionsByBlock(blockNumber uint64) ([]*IndexedTransaction, error) {
	rows, err := ti.db.Query(`
		SELECT hash, block_number, block_hash, tx_index, from_address, to_address,
		       value, asset, fee, nonce, tx_type, status, gas_used, created_at
		FROM transactions
		WHERE block_number = $1
		ORDER BY tx_index ASC
	`, blockNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return ti.scanTransactions(rows)
}

// GetRecentTransactions retrieves recent transactions
func (ti *TransactionIndexer) GetRecentTransactions(limit int) ([]*IndexedTransaction, error) {
	rows, err := ti.db.Query(`
		SELECT hash, block_number, block_hash, tx_index, from_address, to_address,
		       value, asset, fee, nonce, tx_type, status, gas_used, created_at
		FROM transactions
		ORDER BY block_number DESC, tx_index DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return ti.scanTransactions(rows)
}

// GetTransactionsByType retrieves transactions by type
func (ti *TransactionIndexer) GetTransactionsByType(txType string, limit, offset int) ([]*IndexedTransaction, error) {
	rows, err := ti.db.Query(`
		SELECT hash, block_number, block_hash, tx_index, from_address, to_address,
		       value, asset, fee, nonce, tx_type, status, gas_used, created_at
		FROM transactions
		WHERE tx_type = $1
		ORDER BY block_number DESC, tx_index DESC
		LIMIT $2 OFFSET $3
	`, txType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return ti.scanTransactions(rows)
}

// GetTransactionCount returns total transaction count
func (ti *TransactionIndexer) GetTransactionCount() (uint64, error) {
	var count uint64
	err := ti.db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	return count, err
}

// GetTransactionCountByAddress returns transaction count for an address
func (ti *TransactionIndexer) GetTransactionCountByAddress(address string) (uint64, error) {
	var count uint64
	err := ti.db.QueryRow(`
		SELECT COUNT(*) FROM transactions 
		WHERE from_address = $1 OR to_address = $1
	`, address).Scan(&count)
	return count, err
}

// GetDailyTransactionStats returns daily transaction statistics
func (ti *TransactionIndexer) GetDailyTransactionStats(days int) ([]*DailyStats, error) {
	rows, err := ti.db.Query(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as tx_count,
			SUM(CAST(value AS NUMERIC)) as total_value,
			SUM(CAST(fee AS NUMERIC)) as total_fees
		FROM transactions
		WHERE created_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []*DailyStats
	for rows.Next() {
		s := &DailyStats{}
		if err := rows.Scan(&s.Date, &s.TxCount, &s.TotalValue, &s.TotalFees); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	
	return stats, nil
}

// scanTransactions scans transaction rows
func (ti *TransactionIndexer) scanTransactions(rows *sql.Rows) ([]*IndexedTransaction, error) {
	var txs []*IndexedTransaction
	for rows.Next() {
		txn := &IndexedTransaction{}
		if err := rows.Scan(
			&txn.Hash, &txn.BlockNumber, &txn.BlockHash, &txn.TxIndex,
			&txn.From, &txn.To, &txn.Value, &txn.Asset, &txn.Fee, &txn.Nonce,
			&txn.Type, &txn.Status, &txn.GasUsed, &txn.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, txn)
	}
	return txs, nil
}

// IndexedTransaction represents an indexed transaction
type IndexedTransaction struct {
	Hash        string  `json:"hash"`
	BlockNumber uint64  `json:"block_number"`
	BlockHash   string  `json:"block_hash"`
	TxIndex     int     `json:"tx_index"`
	From        string  `json:"from"`
	To          *string `json:"to,omitempty"`
	Value       string  `json:"value"`
	Asset       string  `json:"asset"`
	Fee         string  `json:"fee"`
	Nonce       uint64  `json:"nonce"`
	Data        []byte  `json:"data,omitempty"`
	Signature   string  `json:"signature"`
	Type        string  `json:"type"`
	Status      int     `json:"status"`
	GasUsed     uint64  `json:"gas_used"`
	CreatedAt   string  `json:"created_at"`
}

// DailyStats represents daily transaction statistics
type DailyStats struct {
	Date       string `json:"date"`
	TxCount    uint64 `json:"tx_count"`
	TotalValue string `json:"total_value"`
	TotalFees  string `json:"total_fees"`
}
