package util

import (
	"errors"
	"fmt"
)

// Common blockchain errors
var (
	// Block errors
	ErrBlockNotFound      = errors.New("block not found")
	ErrInvalidBlockHash   = errors.New("invalid block hash")
	ErrInvalidBlockNumber = errors.New("invalid block number")
	ErrInvalidParentHash  = errors.New("invalid parent hash")
	ErrBlockTooOld        = errors.New("block is too old")
	ErrBlockTooNew        = errors.New("block timestamp is in the future")
	ErrDuplicateBlock     = errors.New("duplicate block")

	// Transaction errors
	ErrTxNotFound          = errors.New("transaction not found")
	ErrInvalidTxHash       = errors.New("invalid transaction hash")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrInvalidNonce        = errors.New("invalid nonce")
	ErrNonceTooLow         = errors.New("nonce too low")
	ErrNonceTooHigh        = errors.New("nonce too high")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInsufficientFee     = errors.New("insufficient fee")
	ErrGasLimitExceeded    = errors.New("gas limit exceeded")
	ErrTxPoolFull          = errors.New("transaction pool is full")
	ErrDuplicateTx         = errors.New("duplicate transaction")
	ErrTxTooLarge          = errors.New("transaction too large")

	// Account errors
	ErrAccountNotFound = errors.New("account not found")
	ErrInvalidAddress  = errors.New("invalid address")

	// Validator errors
	ErrValidatorNotFound  = errors.New("validator not found")
	ErrNotValidator       = errors.New("not a validator")
	ErrAlreadyValidator   = errors.New("already a validator")
	ErrInsufficientStake  = errors.New("insufficient stake")
	ErrValidatorJailed    = errors.New("validator is jailed")
	ErrSlashingViolation  = errors.New("slashing violation detected")
	ErrDoubleSign         = errors.New("double signing detected")
	ErrMissedBlocks       = errors.New("too many missed blocks")

	// Consensus errors
	ErrInvalidConsensus   = errors.New("invalid consensus")
	ErrNotMyTurn          = errors.New("not validator's turn")
	ErrInvalidProposer    = errors.New("invalid block proposer")
	ErrInvalidVote        = errors.New("invalid vote")
	ErrQuorumNotReached   = errors.New("quorum not reached")

	// State errors
	ErrStateNotFound      = errors.New("state not found")
	ErrInvalidStateRoot   = errors.New("invalid state root")
	ErrStateCorrupted     = errors.New("state is corrupted")

	// Asset errors
	ErrAssetNotFound      = errors.New("asset not found")
	ErrInvalidAsset       = errors.New("invalid asset")
	ErrAssetAlreadyExists = errors.New("asset already exists")
	ErrNotAssetOwner      = errors.New("not asset owner")

	// Network errors
	ErrPeerNotFound       = errors.New("peer not found")
	ErrConnectionFailed   = errors.New("connection failed")
	ErrMaxPeersReached    = errors.New("maximum peers reached")
	ErrInvalidProtocol    = errors.New("invalid protocol")

	// Database errors
	ErrDatabaseClosed     = errors.New("database is closed")
	ErrKeyNotFound        = errors.New("key not found")
	ErrDatabaseCorrupted  = errors.New("database is corrupted")

	// Crypto errors
	ErrInvalidPrivateKey  = errors.New("invalid private key")
	ErrInvalidPublicKey   = errors.New("invalid public key")
	ErrDecryptionFailed   = errors.New("decryption failed")
)

// ChainError represents a blockchain-specific error with context
type ChainError struct {
	Op      string // Operation that failed
	Kind    error  // Category of error
	Err     error  // Underlying error
	Context map[string]interface{}
}

// NewChainError creates a new ChainError
func NewChainError(op string, kind error, err error) *ChainError {
	return &ChainError{
		Op:      op,
		Kind:    kind,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *ChainError) WithContext(key string, value interface{}) *ChainError {
	e.Context[key] = value
	return e
}

// Error implements the error interface
func (e *ChainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Kind.Error(), e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Kind.Error())
}

// Unwrap returns the underlying error
func (e *ChainError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches a target
func (e *ChainError) Is(target error) bool {
	return errors.Is(e.Kind, target) || errors.Is(e.Err, target)
}

// IsKind checks if the error is of a specific kind
func (e *ChainError) IsKind(kind error) bool {
	return errors.Is(e.Kind, kind)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithOp wraps an error with operation context
func WrapWithOp(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrBlockNotFound) ||
		errors.Is(err, ErrTxNotFound) ||
		errors.Is(err, ErrAccountNotFound) ||
		errors.Is(err, ErrValidatorNotFound) ||
		errors.Is(err, ErrAssetNotFound) ||
		errors.Is(err, ErrPeerNotFound) ||
		errors.Is(err, ErrKeyNotFound)
}

// IsValidation checks if error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrInvalidBlockHash) ||
		errors.Is(err, ErrInvalidSignature) ||
		errors.Is(err, ErrInvalidNonce) ||
		errors.Is(err, ErrInvalidAddress) ||
		errors.Is(err, ErrInvalidAsset)
}

// IsInsufficientFunds checks if error is an insufficient funds error
func IsInsufficientFunds(err error) bool {
	return errors.Is(err, ErrInsufficientBalance) ||
		errors.Is(err, ErrInsufficientFee) ||
		errors.Is(err, ErrInsufficientStake)
}
