package service

import (
	"database/sql"
	"fmt"

	"github.com/gydschain/gydschain/internal/tx"
)

// AccountIndexer indexes account data
type AccountIndexer struct {
	db *sql.DB
}

// NewAccountIndexer creates a new account indexer
func NewAccountIndexer(db *sql.DB) *AccountIndexer {
	return &AccountIndexer{db: db}
}

// UpdateFromTransaction updates account data from a transaction
func (ai *AccountIndexer) UpdateFromTransaction(dbTx *sql.Tx, txn *tx.Transaction, blockNumber uint64) error {
	// Update sender account
	if err := ai.updateAccount(dbTx, txn.From, blockNumber); err != nil {
		return fmt.Errorf("update sender: %w", err)
	}
	
	// Update recipient account
	if txn.To != "" {
		if err := ai.updateAccount(dbTx, txn.To, blockNumber); err != nil {
			return fmt.Errorf("update recipient: %w", err)
		}
	}
	
	// Update balances
	if err := ai.updateBalance(dbTx, txn.From, txn.Asset, txn.Value.String(), false); err != nil {
		return fmt.Errorf("update sender balance: %w", err)
	}
	
	if txn.To != "" {
		if err := ai.updateBalance(dbTx, txn.To, txn.Asset, txn.Value.String(), true); err != nil {
			return fmt.Errorf("update recipient balance: %w", err)
		}
	}
	
	return nil
}

// updateAccount updates or creates an account
func (ai *AccountIndexer) updateAccount(dbTx *sql.Tx, address string, blockNumber uint64) error {
	_, err := dbTx.Exec(`
		INSERT INTO accounts (address, nonce, tx_count, first_seen_block, last_seen_block)
		VALUES ($1, 0, 1, $2, $2)
		ON CONFLICT (address) DO UPDATE SET
			tx_count = accounts.tx_count + 1,
			last_seen_block = $2,
			updated_at = NOW()
	`, address, blockNumber)
	return err
}

// updateBalance updates account balance
func (ai *AccountIndexer) updateBalance(dbTx *sql.Tx, address, asset, amount string, isCredit bool) error {
	// This is a simplified version - in production you'd need proper big integer handling
	var operator string
	if isCredit {
		operator = "+"
	} else {
		operator = "-"
	}
	
	_, err := dbTx.Exec(fmt.Sprintf(`
		INSERT INTO account_balances (address, asset, balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (address, asset) DO UPDATE SET
			balance = (CAST(account_balances.balance AS NUMERIC) %s CAST($3 AS NUMERIC))::TEXT,
			updated_at = NOW()
	`, operator), address, asset, amount)
	return err
}

// GetAccount retrieves an account by address
func (ai *AccountIndexer) GetAccount(address string) (*Account, error) {
	account := &Account{Address: address}
	
	err := ai.db.QueryRow(`
		SELECT nonce, tx_count, first_seen_block, last_seen_block
		FROM accounts WHERE address = $1
	`, address).Scan(&account.Nonce, &account.TxCount, &account.FirstSeenBlock, &account.LastSeenBlock)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	// Get balances
	rows, err := ai.db.Query(`
		SELECT asset, balance FROM account_balances WHERE address = $1
	`, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	account.Balances = make(map[string]string)
	for rows.Next() {
		var asset, balance string
		if err := rows.Scan(&asset, &balance); err != nil {
			return nil, err
		}
		account.Balances[asset] = balance
	}
	
	return account, nil
}

// GetAccountBalance retrieves balance for a specific asset
func (ai *AccountIndexer) GetAccountBalance(address, asset string) (string, error) {
	var balance string
	err := ai.db.QueryRow(`
		SELECT balance FROM account_balances 
		WHERE address = $1 AND asset = $2
	`, address, asset).Scan(&balance)
	
	if err == sql.ErrNoRows {
		return "0", nil
	}
	return balance, err
}

// GetTopAccounts retrieves top accounts by balance
func (ai *AccountIndexer) GetTopAccounts(asset string, limit int) ([]*Account, error) {
	rows, err := ai.db.Query(`
		SELECT a.address, a.nonce, a.tx_count, ab.balance
		FROM accounts a
		JOIN account_balances ab ON a.address = ab.address
		WHERE ab.asset = $1
		ORDER BY CAST(ab.balance AS NUMERIC) DESC
		LIMIT $2
	`, asset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var accounts []*Account
	for rows.Next() {
		acc := &Account{Balances: make(map[string]string)}
		var balance string
		if err := rows.Scan(&acc.Address, &acc.Nonce, &acc.TxCount, &balance); err != nil {
			return nil, err
		}
		acc.Balances[asset] = balance
		accounts = append(accounts, acc)
	}
	
	return accounts, nil
}

// GetAccountTransactions retrieves transactions for an account
func (ai *AccountIndexer) GetAccountTransactions(address string, limit, offset int) ([]*TransactionRecord, error) {
	rows, err := ai.db.Query(`
		SELECT hash, block_number, tx_index, from_address, to_address, 
		       value, asset, fee, tx_type, status, created_at
		FROM transactions
		WHERE from_address = $1 OR to_address = $1
		ORDER BY block_number DESC, tx_index DESC
		LIMIT $2 OFFSET $3
	`, address, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var txs []*TransactionRecord
	for rows.Next() {
		txn := &TransactionRecord{}
		if err := rows.Scan(
			&txn.Hash, &txn.BlockNumber, &txn.TxIndex, &txn.From, &txn.To,
			&txn.Value, &txn.Asset, &txn.Fee, &txn.Type, &txn.Status, &txn.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, txn)
	}
	
	return txs, nil
}

// Account represents an indexed account
type Account struct {
	Address        string            `json:"address"`
	Nonce          uint64            `json:"nonce"`
	TxCount        uint64            `json:"tx_count"`
	FirstSeenBlock uint64            `json:"first_seen_block"`
	LastSeenBlock  uint64            `json:"last_seen_block"`
	Balances       map[string]string `json:"balances"`
}

// TransactionRecord represents a transaction record
type TransactionRecord struct {
	Hash        string `json:"hash"`
	BlockNumber uint64 `json:"block_number"`
	TxIndex     int    `json:"tx_index"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Asset       string `json:"asset"`
	Fee         string `json:"fee"`
	Type        string `json:"type"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"created_at"`
}
