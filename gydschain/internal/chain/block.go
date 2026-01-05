package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gydschain/gydschain/internal/tx"
)

// Block represents a complete block in the GYDS blockchain
type Block struct {
	Header       *Header          `json:"header"`
	Transactions []*tx.Transaction `json:"transactions"`
	Validator    string           `json:"validator"`
	Signature    []byte           `json:"signature"`
}

// NewBlock creates a new block with the given transactions
func NewBlock(parentHash string, height uint64, transactions []*tx.Transaction, validator string) *Block {
	header := NewHeader(parentHash, height)
	
	block := &Block{
		Header:       header,
		Transactions: transactions,
		Validator:    validator,
	}
	
	// Calculate transaction root
	block.Header.TxRoot = block.CalculateTxRoot()
	
	return block
}

// CalculateTxRoot computes the merkle root of all transactions
func (b *Block) CalculateTxRoot() string {
	if len(b.Transactions) == 0 {
		return "0x0000000000000000000000000000000000000000000000000000000000000000"
	}
	
	var hashes [][]byte
	for _, tx := range b.Transactions {
		hash, _ := tx.Hash()
		hashes = append(hashes, hash)
	}
	
	return hex.EncodeToString(merkleRoot(hashes))
}

// Hash calculates the block hash
func (b *Block) Hash() (string, error) {
	headerBytes, err := json.Marshal(b.Header)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(headerBytes)
	return hex.EncodeToString(hash[:]), nil
}

// Verify validates the block structure and signatures
func (b *Block) Verify() error {
	// Verify header
	if err := b.Header.Validate(); err != nil {
		return err
	}
	
	// Verify all transactions
	for _, tx := range b.Transactions {
		if err := tx.Verify(); err != nil {
			return err
		}
	}
	
	// Verify transaction root
	calculatedRoot := b.CalculateTxRoot()
	if calculatedRoot != b.Header.TxRoot {
		return ErrInvalidTxRoot
	}
	
	return nil
}

// Size returns the approximate size of the block in bytes
func (b *Block) Size() int {
	data, _ := json.Marshal(b)
	return len(data)
}

// TxCount returns the number of transactions in the block
func (b *Block) TxCount() int {
	return len(b.Transactions)
}

// GetTransaction returns a transaction by index
func (b *Block) GetTransaction(index int) *tx.Transaction {
	if index < 0 || index >= len(b.Transactions) {
		return nil
	}
	return b.Transactions[index]
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(transaction *tx.Transaction) {
	b.Transactions = append(b.Transactions, transaction)
	b.Header.TxRoot = b.CalculateTxRoot()
}

// Finalize prepares the block for signing
func (b *Block) Finalize() {
	b.Header.Timestamp = time.Now().Unix()
	b.Header.TxRoot = b.CalculateTxRoot()
}

// merkleRoot calculates the merkle root from a list of hashes
func merkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return make([]byte, 32)
	}
	
	if len(hashes) == 1 {
		return hashes[0]
	}
	
	// Ensure even number of hashes
	if len(hashes)%2 != 0 {
		hashes = append(hashes, hashes[len(hashes)-1])
	}
	
	var newLevel [][]byte
	for i := 0; i < len(hashes); i += 2 {
		combined := append(hashes[i], hashes[i+1]...)
		hash := sha256.Sum256(combined)
		newLevel = append(newLevel, hash[:])
	}
	
	return merkleRoot(newLevel)
}

// BlockReward contains reward information for a block
type BlockReward struct {
	Validator    string `json:"validator"`
	GYDSReward   uint64 `json:"gyds_reward"`
	GYDReward    uint64 `json:"gyd_reward"`
	TotalFees    uint64 `json:"total_fees"`
	MinerReward  uint64 `json:"miner_reward"`
	BlockHeight  uint64 `json:"block_height"`
}

// CalculateReward computes the block reward
func (b *Block) CalculateReward() *BlockReward {
	baseReward := uint64(10 * 1e8) // 10 GYDS in smallest unit
	
	// Calculate total fees
	var totalFees uint64
	for _, tx := range b.Transactions {
		totalFees += tx.Fee
	}
	
	// 80% to validator, 20% to miners
	validatorReward := (totalFees * 80) / 100
	minerReward := totalFees - validatorReward
	
	return &BlockReward{
		Validator:   b.Validator,
		GYDSReward:  baseReward + validatorReward,
		TotalFees:   totalFees,
		MinerReward: minerReward,
		BlockHeight: b.Header.Height,
	}
}
