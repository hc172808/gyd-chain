package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidHeight    = errors.New("invalid block height")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
	ErrInvalidTxRoot    = errors.New("invalid transaction root")
	ErrInvalidStateRoot = errors.New("invalid state root")
)

// Header represents the block header
type Header struct {
	Version      uint32 `json:"version"`
	Height       uint64 `json:"height"`
	Timestamp    int64  `json:"timestamp"`
	ParentHash   string `json:"parent_hash"`
	TxRoot       string `json:"tx_root"`
	StateRoot    string `json:"state_root"`
	ReceiptRoot  string `json:"receipt_root"`
	ValidatorSet string `json:"validator_set"`
	Difficulty   uint64 `json:"difficulty"`
	Nonce        uint64 `json:"nonce"`
	ExtraData    []byte `json:"extra_data"`
	GasLimit     uint64 `json:"gas_limit"`
	GasUsed      uint64 `json:"gas_used"`
}

// NewHeader creates a new block header
func NewHeader(parentHash string, height uint64) *Header {
	return &Header{
		Version:    1,
		Height:     height,
		Timestamp:  time.Now().Unix(),
		ParentHash: parentHash,
		Difficulty: 1000,
		GasLimit:   10000000,
	}
}

// Hash computes the header hash
func (h *Header) Hash() (string, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// Validate checks the header fields
func (h *Header) Validate() error {
	// Check timestamp is not too far in the future
	if h.Timestamp > time.Now().Add(15*time.Second).Unix() {
		return ErrInvalidTimestamp
	}
	
	// Height must be positive for non-genesis blocks
	if h.Height > 0 && h.ParentHash == "" {
		return ErrInvalidHeight
	}
	
	return nil
}

// IsGenesis returns true if this is a genesis block header
func (h *Header) IsGenesis() bool {
	return h.Height == 0 && h.ParentHash == ""
}

// Size returns the approximate size of the header in bytes
func (h *Header) Size() int {
	data, _ := json.Marshal(h)
	return len(data)
}

// SetStateRoot updates the state root
func (h *Header) SetStateRoot(root string) {
	h.StateRoot = root
}

// SetReceiptRoot updates the receipt root
func (h *Header) SetReceiptRoot(root string) {
	h.ReceiptRoot = root
}

// IncrementNonce increases the nonce for mining
func (h *Header) IncrementNonce() {
	h.Nonce++
}

// MeetsTarget checks if the header hash meets the difficulty target
func (h *Header) MeetsTarget() bool {
	hash, err := h.Hash()
	if err != nil {
		return false
	}
	
	// Simple difficulty check - leading zeros
	target := h.Difficulty / 100
	for i := uint64(0); i < target && i < uint64(len(hash)); i++ {
		if hash[i] != '0' {
			return false
		}
	}
	
	return true
}

// HeaderWithProof includes PoW proof data
type HeaderWithProof struct {
	Header    *Header `json:"header"`
	ProofHash string  `json:"proof_hash"`
	WorkDone  uint64  `json:"work_done"`
}

// NewHeaderWithProof creates a header with PoW proof
func NewHeaderWithProof(header *Header) *HeaderWithProof {
	hash, _ := header.Hash()
	return &HeaderWithProof{
		Header:    header,
		ProofHash: hash,
		WorkDone:  header.Difficulty,
	}
}
