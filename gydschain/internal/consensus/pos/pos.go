package pos

import (
	"errors"
	"sort"
	"sync"
	"time"
)

// PoS consensus engine errors
var (
	ErrNoValidators       = errors.New("no validators available")
	ErrNotValidator       = errors.New("not a validator")
	ErrInsufficientStake  = errors.New("insufficient stake")
	ErrValidatorNotFound  = errors.New("validator not found")
	ErrAlreadyValidator   = errors.New("already a validator")
	ErrInvalidSignature   = errors.New("invalid block signature")
)

// Engine represents the PoS consensus engine
type Engine struct {
	mu            sync.RWMutex
	validators    map[string]*Validator
	validatorList []*Validator
	totalStake    uint64
	minStake      uint64
	maxValidators uint32
	blockTime     time.Duration
	currentRound  uint64
	currentLeader string
}

// NewEngine creates a new PoS consensus engine
func NewEngine(minStake uint64, maxValidators uint32, blockTime time.Duration) *Engine {
	return &Engine{
		validators:    make(map[string]*Validator),
		validatorList: make([]*Validator, 0),
		minStake:      minStake,
		maxValidators: maxValidators,
		blockTime:     blockTime,
	}
}

// RegisterValidator registers a new validator
func (e *Engine) RegisterValidator(address, pubKey string, stake uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if _, exists := e.validators[address]; exists {
		return ErrAlreadyValidator
	}
	
	if stake < e.minStake {
		return ErrInsufficientStake
	}
	
	if uint32(len(e.validators)) >= e.maxValidators {
		// Check if new stake is higher than lowest
		if len(e.validatorList) > 0 {
			lowest := e.validatorList[len(e.validatorList)-1]
			if stake <= lowest.TotalStake {
				return errors.New("stake too low to join validator set")
			}
		}
	}
	
	validator := NewValidator(address, pubKey, stake)
	e.validators[address] = validator
	e.totalStake += stake
	
	e.updateValidatorList()
	
	return nil
}

// UnregisterValidator removes a validator
func (e *Engine) UnregisterValidator(address string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	validator, exists := e.validators[address]
	if !exists {
		return ErrValidatorNotFound
	}
	
	e.totalStake -= validator.TotalStake
	delete(e.validators, address)
	e.updateValidatorList()
	
	return nil
}

// Delegate adds stake delegation to a validator
func (e *Engine) Delegate(delegator, validator string, amount uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	v, exists := e.validators[validator]
	if !exists {
		return ErrValidatorNotFound
	}
	
	v.AddDelegation(delegator, amount)
	e.totalStake += amount
	e.updateValidatorList()
	
	return nil
}

// Undelegate removes stake delegation from a validator
func (e *Engine) Undelegate(delegator, validator string, amount uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	v, exists := e.validators[validator]
	if !exists {
		return ErrValidatorNotFound
	}
	
	if err := v.RemoveDelegation(delegator, amount); err != nil {
		return err
	}
	
	e.totalStake -= amount
	e.updateValidatorList()
	
	return nil
}

// SelectLeader selects the block proposer for a round
func (e *Engine) SelectLeader(round uint64) (*Validator, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if len(e.validatorList) == 0 {
		return nil, ErrNoValidators
	}
	
	e.currentRound = round
	
	// Weighted random selection based on stake
	totalWeight := e.totalStake
	target := round % totalWeight
	
	var cumulative uint64
	for _, v := range e.validatorList {
		cumulative += v.TotalStake
		if cumulative > target {
			e.currentLeader = v.Address
			return v, nil
		}
	}
	
	// Fallback to first validator
	e.currentLeader = e.validatorList[0].Address
	return e.validatorList[0], nil
}

// VerifyBlock verifies a block was produced by a valid validator
func (e *Engine) VerifyBlock(proposer string, signature []byte) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	validator, exists := e.validators[proposer]
	if !exists {
		return ErrNotValidator
	}
	
	// Verify the block signature (simplified)
	if !validator.VerifySignature(signature) {
		return ErrInvalidSignature
	}
	
	return nil
}

// GetValidator returns a validator by address
func (e *Engine) GetValidator(address string) (*Validator, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	v, exists := e.validators[address]
	if !exists {
		return nil, ErrValidatorNotFound
	}
	
	return v.Copy(), nil
}

// GetValidators returns all active validators
func (e *Engine) GetValidators() []*Validator {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	validators := make([]*Validator, len(e.validatorList))
	for i, v := range e.validatorList {
		validators[i] = v.Copy()
	}
	
	return validators
}

// GetTotalStake returns the total staked amount
func (e *Engine) GetTotalStake() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.totalStake
}

// updateValidatorList updates and sorts the validator list
func (e *Engine) updateValidatorList() {
	e.validatorList = make([]*Validator, 0, len(e.validators))
	
	for _, v := range e.validators {
		if v.Active && v.TotalStake >= e.minStake {
			e.validatorList = append(e.validatorList, v)
		}
	}
	
	// Sort by stake (descending)
	sort.Slice(e.validatorList, func(i, j int) bool {
		return e.validatorList[i].TotalStake > e.validatorList[j].TotalStake
	})
	
	// Limit to max validators
	if uint32(len(e.validatorList)) > e.maxValidators {
		e.validatorList = e.validatorList[:e.maxValidators]
	}
}

// ProcessRewards distributes block rewards
func (e *Engine) ProcessRewards(blockReward uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if len(e.validatorList) == 0 || e.totalStake == 0 {
		return
	}
	
	for _, v := range e.validatorList {
		// Proportional reward based on stake
		reward := (blockReward * v.TotalStake) / e.totalStake
		v.AddReward(reward)
	}
}

// CurrentLeader returns the current round's leader
func (e *Engine) CurrentLeader() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.currentLeader
}

// ValidatorCount returns the number of active validators
func (e *Engine) ValidatorCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.validatorList)
}
