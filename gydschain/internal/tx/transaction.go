package tx

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// Transaction types
const (
	TxTypeTransfer     = "transfer"
	TxTypeStake        = "stake"
	TxTypeUnstake      = "unstake"
	TxTypeMint         = "mint"
	TxTypeBurn         = "burn"
	TxTypeCreateAsset  = "create_asset"
	TxTypeUpdateOracle = "update_oracle"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint64 `json:"amount"`
	Asset     string `json:"asset"`
	Fee       uint64 `json:"fee"`
	Nonce     uint64 `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data,omitempty"`
	Signature []byte `json:"signature"`
	PubKey    []byte `json:"pub_key"`
}

// NewTransaction creates a new transaction
func NewTransaction(txType, from, to string, amount uint64, asset string) *Transaction {
	return &Transaction{
		Type:      txType,
		From:      from,
		To:        to,
		Amount:    amount,
		Asset:     asset,
		Timestamp: time.Now().Unix(),
	}
}

// NewTransfer creates a new transfer transaction
func NewTransfer(from, to string, amount uint64, asset string) *Transaction {
	return NewTransaction(TxTypeTransfer, from, to, amount, asset)
}

// NewStake creates a new staking transaction
func NewStake(from string, amount uint64, validatorAddr string) *Transaction {
	tx := NewTransaction(TxTypeStake, from, validatorAddr, amount, "GYDS")
	return tx
}

// NewUnstake creates a new unstaking transaction
func NewUnstake(from string, amount uint64, validatorAddr string) *Transaction {
	return NewTransaction(TxTypeUnstake, from, validatorAddr, amount, "GYDS")
}

// Hash computes the transaction hash
func (t *Transaction) Hash() ([]byte, error) {
	// Create a copy without signature for hashing
	hashTx := *t
	hashTx.Signature = nil
	
	data, err := json.Marshal(hashTx)
	if err != nil {
		return nil, err
	}
	
	hash := sha256.Sum256(data)
	return hash[:], nil
}

// HashHex returns the hex-encoded transaction hash
func (t *Transaction) HashHex() (string, error) {
	hash, err := t.Hash()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash), nil
}

// SetFee sets the transaction fee
func (t *Transaction) SetFee(fee uint64) {
	t.Fee = fee
}

// SetNonce sets the transaction nonce
func (t *Transaction) SetNonce(nonce uint64) {
	t.Nonce = nonce
}

// SetData sets additional transaction data
func (t *Transaction) SetData(data []byte) {
	t.Data = data
}

// Sign signs the transaction (placeholder - actual signing in crypto package)
func (t *Transaction) Sign(privateKey []byte) error {
	hash, err := t.Hash()
	if err != nil {
		return err
	}
	
	// Placeholder: actual signature would use ed25519 or secp256k1
	combined := append(hash, privateKey...)
	sig := sha256.Sum256(combined)
	t.Signature = sig[:]
	
	return nil
}

// Verify validates the transaction
func (t *Transaction) Verify() error {
	// Validate required fields
	if t.From == "" {
		return ErrMissingFrom
	}
	
	if t.Type != TxTypeBurn && t.To == "" {
		return ErrMissingTo
	}
	
	if t.Amount == 0 && t.Type == TxTypeTransfer {
		return ErrZeroAmount
	}
	
	if t.Asset == "" {
		return ErrMissingAsset
	}
	
	if t.Asset != "GYDS" && t.Asset != "GYD" {
		return ErrInvalidAsset
	}
	
	if len(t.Signature) == 0 {
		return ErrMissingSignature
	}
	
	// Verify signature (placeholder)
	// In production, verify using public key cryptography
	
	return nil
}

// Size returns the transaction size in bytes
func (t *Transaction) Size() int {
	data, _ := json.Marshal(t)
	return len(data)
}

// IsTransfer returns true if this is a transfer transaction
func (t *Transaction) IsTransfer() bool {
	return t.Type == TxTypeTransfer
}

// IsStaking returns true if this is a staking-related transaction
func (t *Transaction) IsStaking() bool {
	return t.Type == TxTypeStake || t.Type == TxTypeUnstake
}

// Errors
var (
	ErrMissingFrom      = errors.New("missing sender address")
	ErrMissingTo        = errors.New("missing recipient address")
	ErrZeroAmount       = errors.New("amount cannot be zero")
	ErrMissingAsset     = errors.New("missing asset type")
	ErrInvalidAsset     = errors.New("invalid asset type")
	ErrMissingSignature = errors.New("missing signature")
	ErrInvalidSignature = errors.New("invalid signature")
)

// TransactionReceipt represents a transaction receipt
type TransactionReceipt struct {
	TxHash      string `json:"tx_hash"`
	BlockHash   string `json:"block_hash"`
	BlockHeight uint64 `json:"block_height"`
	Index       uint32 `json:"index"`
	Status      uint8  `json:"status"` // 0 = failed, 1 = success
	GasUsed     uint64 `json:"gas_used"`
	Logs        []Log  `json:"logs"`
}

// Log represents a transaction log entry
type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    []byte   `json:"data"`
}

// NewReceipt creates a new transaction receipt
func NewReceipt(txHash, blockHash string, height uint64, status uint8) *TransactionReceipt {
	return &TransactionReceipt{
		TxHash:      txHash,
		BlockHash:   blockHash,
		BlockHeight: height,
		Status:      status,
		Logs:        make([]Log, 0),
	}
}
