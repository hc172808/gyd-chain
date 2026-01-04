package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gydschain/gydschain/internal/chain"
	"github.com/gydschain/gydschain/internal/rpc"
)

// Indexer processes blocks and indexes data
type Indexer struct {
	db        *sql.DB
	rpcClient *rpc.Client
	
	// State
	lastBlock   uint64
	isRunning   bool
	mu          sync.RWMutex
	
	// Sub-services
	accounts    *AccountIndexer
	assets      *AssetIndexer
	txs         *TransactionIndexer
	validators  *ValidatorIndexer
	
	// Channels
	blocks      chan *chain.Block
	stop        chan struct{}
	
	// Configuration
	config      IndexerConfig
}

// IndexerConfig contains indexer configuration
type IndexerConfig struct {
	BatchSize       int           `json:"batch_size"`
	PollInterval    time.Duration `json:"poll_interval"`
	ConfirmBlocks   int           `json:"confirm_blocks"`
	StartBlock      uint64        `json:"start_block"`
	ReorgDepth      int           `json:"reorg_depth"`
}

// DefaultIndexerConfig returns default configuration
func DefaultIndexerConfig() IndexerConfig {
	return IndexerConfig{
		BatchSize:     100,
		PollInterval:  time.Second,
		ConfirmBlocks: 6,
		StartBlock:    0,
		ReorgDepth:    100,
	}
}

// NewIndexer creates a new indexer
func NewIndexer(db *sql.DB, rpcClient *rpc.Client, config IndexerConfig) *Indexer {
	idx := &Indexer{
		db:        db,
		rpcClient: rpcClient,
		config:    config,
		blocks:    make(chan *chain.Block, 100),
		stop:      make(chan struct{}),
	}
	
	// Initialize sub-services
	idx.accounts = NewAccountIndexer(db)
	idx.assets = NewAssetIndexer(db)
	idx.txs = NewTransactionIndexer(db)
	idx.validators = NewValidatorIndexer(db)
	
	return idx
}

// Start starts the indexer
func (idx *Indexer) Start(ctx context.Context) error {
	idx.mu.Lock()
	if idx.isRunning {
		idx.mu.Unlock()
		return fmt.Errorf("indexer already running")
	}
	idx.isRunning = true
	idx.mu.Unlock()
	
	// Load last indexed block
	if err := idx.loadState(); err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	
	fmt.Printf("Starting indexer from block %d\n", idx.lastBlock)
	
	// Start block processor
	go idx.processBlocks(ctx)
	
	// Start block fetcher
	go idx.fetchBlocks(ctx)
	
	return nil
}

// Stop stops the indexer
func (idx *Indexer) Stop() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	if !idx.isRunning {
		return
	}
	
	close(idx.stop)
	idx.isRunning = false
}

// loadState loads the indexer state from database
func (idx *Indexer) loadState() error {
	var lastBlock string
	err := idx.db.QueryRow(
		"SELECT value FROM indexer_state WHERE key = 'last_indexed_block'",
	).Scan(&lastBlock)
	
	if err == sql.ErrNoRows {
		idx.lastBlock = idx.config.StartBlock
		return nil
	}
	if err != nil {
		return err
	}
	
	fmt.Sscanf(lastBlock, "%d", &idx.lastBlock)
	return nil
}

// saveState saves the indexer state to database
func (idx *Indexer) saveState() error {
	_, err := idx.db.Exec(
		"UPDATE indexer_state SET value = $1, updated_at = NOW() WHERE key = 'last_indexed_block'",
		fmt.Sprintf("%d", idx.lastBlock),
	)
	return err
}

// fetchBlocks fetches blocks from the node
func (idx *Indexer) fetchBlocks(ctx context.Context) {
	ticker := time.NewTicker(idx.config.PollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-idx.stop:
			return
		case <-ticker.C:
			idx.fetchNewBlocks()
		}
	}
}

// fetchNewBlocks fetches new blocks
func (idx *Indexer) fetchNewBlocks() {
	// Get current chain height
	height, err := idx.rpcClient.GetBlockHeight()
	if err != nil {
		fmt.Printf("Error getting block height: %v\n", err)
		return
	}
	
	// Calculate safe height (accounting for reorgs)
	safeHeight := height - uint64(idx.config.ConfirmBlocks)
	
	idx.mu.RLock()
	lastBlock := idx.lastBlock
	idx.mu.RUnlock()
	
	if safeHeight <= lastBlock {
		return
	}
	
	// Fetch blocks in batches
	for blockNum := lastBlock + 1; blockNum <= safeHeight; blockNum++ {
		block, err := idx.rpcClient.GetBlockByNumber(blockNum)
		if err != nil {
			fmt.Printf("Error fetching block %d: %v\n", blockNum, err)
			return
		}
		
		select {
		case idx.blocks <- block:
		case <-idx.stop:
			return
		}
	}
}

// processBlocks processes blocks from the channel
func (idx *Indexer) processBlocks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-idx.stop:
			return
		case block := <-idx.blocks:
			if err := idx.processBlock(block); err != nil {
				fmt.Printf("Error processing block %d: %v\n", block.Number, err)
				continue
			}
		}
	}
}

// processBlock processes a single block
func (idx *Indexer) processBlock(block *chain.Block) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Index block
	if err := idx.indexBlock(tx, block); err != nil {
		return fmt.Errorf("index block: %w", err)
	}
	
	// Index transactions
	for i, txn := range block.Transactions {
		if err := idx.txs.IndexTransaction(tx, block, txn, i); err != nil {
			return fmt.Errorf("index transaction: %w", err)
		}
		
		// Update accounts
		if err := idx.accounts.UpdateFromTransaction(tx, txn, block.Number); err != nil {
			return fmt.Errorf("update accounts: %w", err)
		}
		
		// Update assets
		if err := idx.assets.UpdateFromTransaction(tx, txn); err != nil {
			return fmt.Errorf("update assets: %w", err)
		}
	}
	
	// Update validator stats
	if err := idx.validators.UpdateFromBlock(tx, block); err != nil {
		return fmt.Errorf("update validators: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}
	
	// Update state
	idx.mu.Lock()
	idx.lastBlock = block.Number
	idx.mu.Unlock()
	
	// Save state periodically
	if block.Number%100 == 0 {
		idx.saveState()
	}
	
	fmt.Printf("Indexed block %d with %d transactions\n", block.Number, len(block.Transactions))
	return nil
}

// indexBlock indexes a block
func (idx *Indexer) indexBlock(tx *sql.Tx, block *chain.Block) error {
	_, err := tx.Exec(`
		INSERT INTO blocks (number, hash, parent_hash, state_root, transactions_root, 
		                    receipts_root, validator, timestamp, gas_used, gas_limit, 
		                    size, tx_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (number) DO UPDATE SET
			hash = EXCLUDED.hash,
			parent_hash = EXCLUDED.parent_hash,
			state_root = EXCLUDED.state_root,
			validator = EXCLUDED.validator
	`,
		block.Number,
		block.Hash(),
		block.Header.ParentHash,
		block.Header.StateRoot,
		block.Header.TxRoot,
		block.Header.ReceiptsRoot,
		block.Header.Validator,
		block.Header.Timestamp,
		block.Header.GasUsed,
		block.Header.GasLimit,
		block.Size(),
		len(block.Transactions),
	)
	return err
}

// GetLastIndexedBlock returns the last indexed block number
func (idx *Indexer) GetLastIndexedBlock() uint64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.lastBlock
}

// HandleReorg handles chain reorganizations
func (idx *Indexer) HandleReorg(fromBlock uint64) error {
	tx, err := idx.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Delete blocks from the reorg point
	if _, err := tx.Exec("DELETE FROM blocks WHERE number >= $1", fromBlock); err != nil {
		return err
	}
	
	// Reset state
	idx.mu.Lock()
	idx.lastBlock = fromBlock - 1
	idx.mu.Unlock()
	
	return tx.Commit()
}
