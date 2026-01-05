package pos

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// ValidatorStatus represents validator state
type ValidatorStatus uint8

const (
	StatusInactive ValidatorStatus = iota
	StatusActive
	StatusJailed
	StatusUnbonding
)

// Validator represents a network validator
type Validator struct {
	mu           sync.RWMutex
	Address      string            `json:"address"`
	PubKey       string            `json:"pub_key"`
	SelfStake    uint64            `json:"self_stake"`
	TotalStake   uint64            `json:"total_stake"`
	Delegations  map[string]uint64 `json:"delegations"`
	Commission   uint64            `json:"commission"` // basis points (100 = 1%)
	Rewards      uint64            `json:"rewards"`
	Status       ValidatorStatus   `json:"status"`
	Active       bool              `json:"active"`
	JailedUntil  int64             `json:"jailed_until,omitempty"`
	UnbondingEnd int64             `json:"unbonding_end,omitempty"`
	SlashEvents  []SlashEvent      `json:"slash_events,omitempty"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
	
	// Performance metrics
	BlocksProduced   uint64 `json:"blocks_produced"`
	BlocksMissed     uint64 `json:"blocks_missed"`
	Uptime           float64 `json:"uptime"`
	
	// Metadata
	Name        string `json:"name,omitempty"`
	Website     string `json:"website,omitempty"`
	Description string `json:"description,omitempty"`
}

// SlashEvent records a slashing incident
type SlashEvent struct {
	Height    uint64 `json:"height"`
	Reason    string `json:"reason"`
	Amount    uint64 `json:"amount"`
	Timestamp int64  `json:"timestamp"`
}

// NewValidator creates a new validator
func NewValidator(address, pubKey string, stake uint64) *Validator {
	return &Validator{
		Address:     address,
		PubKey:      pubKey,
		SelfStake:   stake,
		TotalStake:  stake,
		Delegations: make(map[string]uint64),
		Commission:  500, // 5% default
		Status:      StatusActive,
		Active:      true,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		Uptime:      100.0,
	}
}

// AddDelegation adds a delegation to the validator
func (v *Validator) AddDelegation(delegator string, amount uint64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.Delegations[delegator] += amount
	v.TotalStake += amount
	v.UpdatedAt = time.Now().Unix()
}

// RemoveDelegation removes a delegation
func (v *Validator) RemoveDelegation(delegator string, amount uint64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.Delegations[delegator] < amount {
		return ErrInsufficientStake
	}
	
	v.Delegations[delegator] -= amount
	v.TotalStake -= amount
	
	if v.Delegations[delegator] == 0 {
		delete(v.Delegations, delegator)
	}
	
	v.UpdatedAt = time.Now().Unix()
	return nil
}

// GetDelegation returns delegation amount for an address
func (v *Validator) GetDelegation(delegator string) uint64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Delegations[delegator]
}

// AddReward adds rewards to the validator
func (v *Validator) AddReward(amount uint64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.Rewards += amount
}

// WithdrawRewards withdraws accumulated rewards
func (v *Validator) WithdrawRewards() uint64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	rewards := v.Rewards
	v.Rewards = 0
	return rewards
}

// Slash reduces validator stake as penalty
func (v *Validator) Slash(percentage uint64, reason string, height uint64) uint64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	slashAmount := (v.TotalStake * percentage) / 100
	
	// Slash from self-stake first
	if v.SelfStake >= slashAmount {
		v.SelfStake -= slashAmount
	} else {
		remaining := slashAmount - v.SelfStake
		v.SelfStake = 0
		
		// Slash proportionally from delegations
		for delegator, amount := range v.Delegations {
			delegatorSlash := (amount * remaining) / (v.TotalStake - v.SelfStake)
			v.Delegations[delegator] -= delegatorSlash
		}
	}
	
	v.TotalStake -= slashAmount
	v.SlashEvents = append(v.SlashEvents, SlashEvent{
		Height:    height,
		Reason:    reason,
		Amount:    slashAmount,
		Timestamp: time.Now().Unix(),
	})
	
	v.UpdatedAt = time.Now().Unix()
	return slashAmount
}

// Jail puts the validator in jail
func (v *Validator) Jail(duration time.Duration) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.Status = StatusJailed
	v.Active = false
	v.JailedUntil = time.Now().Add(duration).Unix()
	v.UpdatedAt = time.Now().Unix()
}

// Unjail releases the validator from jail
func (v *Validator) Unjail() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if v.Status != StatusJailed {
		return nil
	}
	
	if time.Now().Unix() < v.JailedUntil {
		return ErrStillJailed
	}
	
	v.Status = StatusActive
	v.Active = true
	v.JailedUntil = 0
	v.UpdatedAt = time.Now().Unix()
	
	return nil
}

// StartUnbonding begins the unbonding period
func (v *Validator) StartUnbonding(unbondingTime time.Duration) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	v.Status = StatusUnbonding
	v.Active = false
	v.UnbondingEnd = time.Now().Add(unbondingTime).Unix()
	v.UpdatedAt = time.Now().Unix()
}

// IsUnbonded returns true if unbonding is complete
func (v *Validator) IsUnbonded() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	return v.Status == StatusUnbonding && time.Now().Unix() >= v.UnbondingEnd
}

// RecordBlock records a produced or missed block
func (v *Validator) RecordBlock(produced bool) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if produced {
		v.BlocksProduced++
	} else {
		v.BlocksMissed++
	}
	
	total := v.BlocksProduced + v.BlocksMissed
	if total > 0 {
		v.Uptime = float64(v.BlocksProduced) / float64(total) * 100
	}
	
	v.UpdatedAt = time.Now().Unix()
}

// SetCommission updates the commission rate
func (v *Validator) SetCommission(commission uint64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if commission > 10000 { // 100%
		return ErrInvalidCommission
	}
	
	v.Commission = commission
	v.UpdatedAt = time.Now().Unix()
	return nil
}

// VerifySignature verifies a signature (placeholder)
func (v *Validator) VerifySignature(signature []byte) bool {
	// Placeholder: actual verification would use ed25519 or secp256k1
	return len(signature) > 0
}

// Sign signs data with the validator's key (placeholder)
func (v *Validator) Sign(data []byte) []byte {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	combined := append(data, []byte(v.PubKey)...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// Copy creates a deep copy of the validator
func (v *Validator) Copy() *Validator {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	copy := &Validator{
		Address:        v.Address,
		PubKey:         v.PubKey,
		SelfStake:      v.SelfStake,
		TotalStake:     v.TotalStake,
		Delegations:    make(map[string]uint64),
		Commission:     v.Commission,
		Rewards:        v.Rewards,
		Status:         v.Status,
		Active:         v.Active,
		JailedUntil:    v.JailedUntil,
		UnbondingEnd:   v.UnbondingEnd,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
		BlocksProduced: v.BlocksProduced,
		BlocksMissed:   v.BlocksMissed,
		Uptime:         v.Uptime,
		Name:           v.Name,
		Website:        v.Website,
		Description:    v.Description,
	}
	
	for k, val := range v.Delegations {
		copy.Delegations[k] = val
	}
	
	copy.SlashEvents = append(copy.SlashEvents, v.SlashEvents...)
	
	return copy
}

// AddressHash returns the hash of the validator address
func (v *Validator) AddressHash() string {
	hash := sha256.Sum256([]byte(v.Address))
	return hex.EncodeToString(hash[:8])
}

// Errors
var (
	ErrStillJailed       = &ValidatorError{"validator still jailed"}
	ErrInvalidCommission = &ValidatorError{"invalid commission rate"}
)

type ValidatorError struct {
	msg string
}

func (e *ValidatorError) Error() string {
	return e.msg
}
