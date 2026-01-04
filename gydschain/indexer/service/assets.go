package service

import (
	"database/sql"

	"github.com/gydschain/gydschain/internal/tx"
)

// AssetIndexer indexes asset data
type AssetIndexer struct {
	db *sql.DB
}

// NewAssetIndexer creates a new asset indexer
func NewAssetIndexer(db *sql.DB) *AssetIndexer {
	return &AssetIndexer{db: db}
}

// UpdateFromTransaction updates asset data from a transaction
func (ai *AssetIndexer) UpdateFromTransaction(dbTx *sql.Tx, txn *tx.Transaction) error {
	// Handle asset creation transactions
	if txn.Type == tx.TxTypeAssetCreate {
		return ai.indexNewAsset(dbTx, txn)
	}
	
	// Handle mint transactions
	if txn.Type == tx.TxTypeAssetMint {
		return ai.updateSupply(dbTx, txn.Asset, txn.Value.String(), true)
	}
	
	// Handle burn transactions
	if txn.Type == tx.TxTypeAssetBurn {
		return ai.updateSupply(dbTx, txn.Asset, txn.Value.String(), false)
	}
	
	return nil
}

// indexNewAsset indexes a newly created asset
func (ai *AssetIndexer) indexNewAsset(dbTx *sql.Tx, txn *tx.Transaction) error {
	// Parse asset data from transaction data
	// This is simplified - in production you'd parse the actual asset data
	_, err := dbTx.Exec(`
		INSERT INTO assets (asset_id, symbol, name, decimals, total_supply, creator, 
		                    is_native, is_stablecoin, mintable, burnable, created_block)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (asset_id) DO NOTHING
	`,
		txn.Asset,
		txn.Asset, // Symbol - would parse from data
		txn.Asset, // Name - would parse from data
		18,        // Decimals - would parse from data
		"0",       // Initial supply
		txn.From,
		false,
		false,
		true,
		true,
		0, // Block number - would get from context
	)
	return err
}

// updateSupply updates asset total supply
func (ai *AssetIndexer) updateSupply(dbTx *sql.Tx, assetID, amount string, isMint bool) error {
	var operator string
	if isMint {
		operator = "+"
	} else {
		operator = "-"
	}
	
	_, err := dbTx.Exec(`
		UPDATE assets 
		SET total_supply = (CAST(total_supply AS NUMERIC) `+operator+` CAST($1 AS NUMERIC))::TEXT
		WHERE asset_id = $2
	`, amount, assetID)
	return err
}

// GetAsset retrieves an asset by ID
func (ai *AssetIndexer) GetAsset(assetID string) (*Asset, error) {
	asset := &Asset{}
	
	err := ai.db.QueryRow(`
		SELECT asset_id, symbol, name, decimals, total_supply, max_supply,
		       creator, is_native, is_stablecoin, peg_target, mintable, burnable, created_block
		FROM assets WHERE asset_id = $1
	`, assetID).Scan(
		&asset.ID, &asset.Symbol, &asset.Name, &asset.Decimals,
		&asset.TotalSupply, &asset.MaxSupply, &asset.Creator,
		&asset.IsNative, &asset.IsStablecoin, &asset.PegTarget,
		&asset.Mintable, &asset.Burnable, &asset.CreatedBlock,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return asset, err
}

// GetAllAssets retrieves all assets
func (ai *AssetIndexer) GetAllAssets() ([]*Asset, error) {
	rows, err := ai.db.Query(`
		SELECT asset_id, symbol, name, decimals, total_supply, max_supply,
		       creator, is_native, is_stablecoin, peg_target, mintable, burnable, created_block
		FROM assets
		ORDER BY is_native DESC, symbol ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var assets []*Asset
	for rows.Next() {
		asset := &Asset{}
		if err := rows.Scan(
			&asset.ID, &asset.Symbol, &asset.Name, &asset.Decimals,
			&asset.TotalSupply, &asset.MaxSupply, &asset.Creator,
			&asset.IsNative, &asset.IsStablecoin, &asset.PegTarget,
			&asset.Mintable, &asset.Burnable, &asset.CreatedBlock,
		); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	
	return assets, nil
}

// GetAssetHolders retrieves holders of an asset
func (ai *AssetIndexer) GetAssetHolders(assetID string, limit, offset int) ([]*AssetHolder, error) {
	rows, err := ai.db.Query(`
		SELECT address, balance
		FROM account_balances
		WHERE asset = $1 AND CAST(balance AS NUMERIC) > 0
		ORDER BY CAST(balance AS NUMERIC) DESC
		LIMIT $2 OFFSET $3
	`, assetID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var holders []*AssetHolder
	for rows.Next() {
		holder := &AssetHolder{}
		if err := rows.Scan(&holder.Address, &holder.Balance); err != nil {
			return nil, err
		}
		holders = append(holders, holder)
	}
	
	return holders, nil
}

// GetAssetTransfers retrieves transfers for an asset
func (ai *AssetIndexer) GetAssetTransfers(assetID string, limit, offset int) ([]*TokenTransfer, error) {
	rows, err := ai.db.Query(`
		SELECT tx_hash, from_address, to_address, amount, block_number, created_at
		FROM token_transfers
		WHERE asset = $1
		ORDER BY block_number DESC, log_index DESC
		LIMIT $2 OFFSET $3
	`, assetID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transfers []*TokenTransfer
	for rows.Next() {
		transfer := &TokenTransfer{}
		if err := rows.Scan(
			&transfer.TxHash, &transfer.From, &transfer.To,
			&transfer.Amount, &transfer.BlockNumber, &transfer.CreatedAt,
		); err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	
	return transfers, nil
}

// RecordTransfer records a token transfer
func (ai *AssetIndexer) RecordTransfer(dbTx *sql.Tx, txHash, from, to, asset, amount string, blockNumber uint64, logIndex int) error {
	_, err := dbTx.Exec(`
		INSERT INTO token_transfers (tx_hash, from_address, to_address, asset, amount, block_number, log_index)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, txHash, from, to, asset, amount, blockNumber, logIndex)
	return err
}

// GetStablecoinPegHistory retrieves peg history for a stablecoin
func (ai *AssetIndexer) GetStablecoinPegHistory(assetID string, limit int) ([]*PegRecord, error) {
	rows, err := ai.db.Query(`
		SELECT block_number, price, target, deviation, supply, collateral_ratio, created_at
		FROM stablecoin_peg_history
		WHERE asset = $1
		ORDER BY block_number DESC
		LIMIT $2
	`, assetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []*PegRecord
	for rows.Next() {
		record := &PegRecord{}
		if err := rows.Scan(
			&record.BlockNumber, &record.Price, &record.Target,
			&record.Deviation, &record.Supply, &record.CollateralRatio, &record.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	
	return records, nil
}

// Asset represents an indexed asset
type Asset struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Decimals     uint8   `json:"decimals"`
	TotalSupply  string  `json:"total_supply"`
	MaxSupply    *string `json:"max_supply,omitempty"`
	Creator      string  `json:"creator"`
	IsNative     bool    `json:"is_native"`
	IsStablecoin bool    `json:"is_stablecoin"`
	PegTarget    *string `json:"peg_target,omitempty"`
	Mintable     bool    `json:"mintable"`
	Burnable     bool    `json:"burnable"`
	CreatedBlock uint64  `json:"created_block"`
}

// AssetHolder represents an asset holder
type AssetHolder struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

// TokenTransfer represents a token transfer record
type TokenTransfer struct {
	TxHash      string `json:"tx_hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	BlockNumber uint64 `json:"block_number"`
	CreatedAt   string `json:"created_at"`
}

// PegRecord represents a stablecoin peg history record
type PegRecord struct {
	BlockNumber     uint64  `json:"block_number"`
	Price           string  `json:"price"`
	Target          string  `json:"target"`
	Deviation       string  `json:"deviation"`
	Supply          string  `json:"supply"`
	CollateralRatio *string `json:"collateral_ratio,omitempty"`
	CreatedAt       string  `json:"created_at"`
}
