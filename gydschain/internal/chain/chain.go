package chain

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/gydschain/gydschain/internal/state"
	"github.com/gydschain/gydschain/internal/tx"
)

var (
	ErrBlockNotFound     = errors.New("block not found")
	ErrInvalidBlock      = errors.New("invalid block")
	ErrInvalidParent     = errors.New("invalid parent block")
	ErrDuplicateBlock    = errors.New("duplicate block")
	ErrChainNotReady     = errors.New("chain not initialized")
)

// Chain represents the blockchain state manager
type Chain struct {
	mu           sync.RWMutex
	blocks       map[string]*Block
	heights      map[uint64]string
	latestHash   string
	latestHeight uint64
	genesis      *Block
	stateDB      *state.StateDB
	config       *ChainConfig
}

// ChainConfig holds chain configuration
type ChainConfig struct {
	ChainID          string `json:"chain_id"`
	NetworkID        uint64 `json:"network_id"`
	BlockTime        uint64 `json:"block_time"`
	MaxBlockSize     uint64 `json:"max_block_size"`
	MaxTxPerBlock    uint64 `json:"max_tx_per_block"`
	GYDSDecimals     uint8  `json:"gyds_decimals"`
	GYDDecimals      uint8  `json:"gyd_decimals"`
	StablecoinPeg    string `json:"stablecoin_peg"`
}

// DefaultConfig returns the default chain configuration
func DefaultConfig() *ChainConfig {
	return &ChainConfig{
		ChainID:       "gydschain-1",
		NetworkID:     1,
		BlockTime:     5,
		MaxBlockSize:  1024 * 1024, // 1MB
		MaxTxPerBlock: 1000,
		GYDSDecimals:  8,
		GYDDecimals:   8,
		StablecoinPeg: "USD",
	}
}

// NewChain creates a new blockchain instance
func NewChain(config *ChainConfig, stateDB *state.StateDB) (*Chain, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	chain := &Chain{
		blocks:  make(map[string]*Block),
		heights: make(map[uint64]string),
		stateDB: stateDB,
		config:  config,
	}
	
	return chain, nil
}

// InitGenesis initializes the chain with the genesis block
func (c *Chain) InitGenesis(genesis *GenesisConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	block := genesis.ToBlock()
	hash, err := block.Hash()
	if err != nil {
		return err
	}
	
	c.genesis = block
	c.blocks[hash] = block
	c.heights[0] = hash
	c.latestHash = hash
	c.latestHeight = 0
	
	// Initialize genesis accounts
	for _, alloc := range genesis.Alloc {
		account := state.NewAccount(alloc.Address)
		account.SetBalance("GYDS", alloc.GYDSBalance)
		account.SetBalance("GYD", alloc.GYDBalance)
		c.stateDB.SetAccount(alloc.Address, account)
	}
	
	return nil
}

// AddBlock adds a validated block to the chain
func (c *Chain) AddBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Verify block
	if err := block.Verify(); err != nil {
		return err
	}
	
	// Verify parent exists
	if block.Header.Height > 0 {
		if _, exists := c.blocks[block.Header.ParentHash]; !exists {
			return ErrInvalidParent
		}
	}
	
	// Check for duplicate
	hash, err := block.Hash()
	if err != nil {
		return err
	}
	
	if _, exists := c.blocks[hash]; exists {
		return ErrDuplicateBlock
	}
	
	// Process transactions
	for _, transaction := range block.Transactions {
		if err := c.processTransaction(transaction); err != nil {
			return err
		}
	}
	
	// Store block
	c.blocks[hash] = block
	c.heights[block.Header.Height] = hash
	
	// Update latest
	if block.Header.Height > c.latestHeight {
		c.latestHeight = block.Header.Height
		c.latestHash = hash
	}
	
	return nil
}

// processTransaction executes a transaction and updates state
func (c *Chain) processTransaction(transaction *tx.Transaction) error {
	// Get sender account
	sender := c.stateDB.GetAccount(transaction.From)
	if sender == nil {
		return errors.New("sender account not found")
	}
	
	// Check balance
	balance := sender.GetBalance(transaction.Asset)
	if balance < transaction.Amount+transaction.Fee {
		return errors.New("insufficient balance")
	}
	
	// Get or create receiver account
	receiver := c.stateDB.GetAccount(transaction.To)
	if receiver == nil {
		receiver = state.NewAccount(transaction.To)
	}
	
	// Update balances
	sender.SetBalance(transaction.Asset, balance-transaction.Amount-transaction.Fee)
	receiver.SetBalance(transaction.Asset, receiver.GetBalance(transaction.Asset)+transaction.Amount)
	
	// Increment sender nonce
	sender.IncrementNonce()
	
	// Save accounts
	c.stateDB.SetAccount(transaction.From, sender)
	c.stateDB.SetAccount(transaction.To, receiver)
	
	return nil
}

// GetBlock returns a block by hash
func (c *Chain) GetBlock(hash string) (*Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	block, exists := c.blocks[hash]
	if !exists {
		return nil, ErrBlockNotFound
	}
	
	return block, nil
}

// GetBlockByHeight returns a block by height
func (c *Chain) GetBlockByHeight(height uint64) (*Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	hash, exists := c.heights[height]
	if !exists {
		return nil, ErrBlockNotFound
	}
	
	return c.blocks[hash], nil
}

// LatestBlock returns the most recent block
func (c *Chain) LatestBlock() (*Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if c.latestHash == "" {
		return nil, ErrChainNotReady
	}
	
	return c.blocks[c.latestHash], nil
}

// Height returns the current chain height
func (c *Chain) Height() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latestHeight
}

// Genesis returns the genesis block
func (c *Chain) Genesis() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.genesis
}

// Config returns the chain configuration
func (c *Chain) Config() *ChainConfig {
	return c.config
}

// Export exports the chain data for backup
func (c *Chain) Export() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	export := struct {
		Config *ChainConfig   `json:"config"`
		Blocks []*Block       `json:"blocks"`
	}{
		Config: c.config,
		Blocks: make([]*Block, 0, len(c.blocks)),
	}
	
	for i := uint64(0); i <= c.latestHeight; i++ {
		if hash, exists := c.heights[i]; exists {
			export.Blocks = append(export.Blocks, c.blocks[hash])
		}
	}
	
	return json.Marshal(export)
}

// Stats returns chain statistics
type ChainStats struct {
	Height       uint64 `json:"height"`
	TotalBlocks  int    `json:"total_blocks"`
	LatestHash   string `json:"latest_hash"`
	TotalTxCount int    `json:"total_tx_count"`
}

// Stats returns current chain statistics
func (c *Chain) Stats() *ChainStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	totalTx := 0
	for _, block := range c.blocks {
		totalTx += len(block.Transactions)
	}
	
	return &ChainStats{
		Height:       c.latestHeight,
		TotalBlocks:  len(c.blocks),
		LatestHash:   c.latestHash,
		TotalTxCount: totalTx,
	}
}
