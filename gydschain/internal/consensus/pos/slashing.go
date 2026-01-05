package pos

import (
	"sync"
	"time"
)

// SlashingReason defines why a validator was slashed
type SlashingReason string

const (
	SlashReasonDoubleSign    SlashingReason = "double_sign"
	SlashReasonDowntime      SlashingReason = "downtime"
	SlashReasonMisbehavior   SlashingReason = "misbehavior"
	SlashReasonInvalidBlock  SlashingReason = "invalid_block"
)

// SlashingParams defines slashing parameters
type SlashingParams struct {
	DoubleSignPenalty   uint64        `json:"double_sign_penalty"`   // basis points
	DowntimePenalty     uint64        `json:"downtime_penalty"`      // basis points
	MisbehaviorPenalty  uint64        `json:"misbehavior_penalty"`   // basis points
	MinSignedPerWindow  uint64        `json:"min_signed_per_window"` // minimum blocks to sign
	SignedBlocksWindow  uint64        `json:"signed_blocks_window"`  // window size
	DowntimeJailDuration time.Duration `json:"downtime_jail_duration"`
	DoubleSignJailDuration time.Duration `json:"double_sign_jail_duration"`
}

// DefaultSlashingParams returns default slashing parameters
func DefaultSlashingParams() *SlashingParams {
	return &SlashingParams{
		DoubleSignPenalty:      500,  // 5%
		DowntimePenalty:        100,  // 1%
		MisbehaviorPenalty:     200,  // 2%
		MinSignedPerWindow:     50,   // 50%
		SignedBlocksWindow:     1000,
		DowntimeJailDuration:   time.Hour * 24,
		DoubleSignJailDuration: time.Hour * 24 * 30, // 30 days
	}
}

// SlashingKeeper manages slashing logic
type SlashingKeeper struct {
	mu                sync.RWMutex
	params            *SlashingParams
	engine            *Engine
	signingInfo       map[string]*ValidatorSigningInfo
	slashingEvents    []SlashingEvent
}

// ValidatorSigningInfo tracks validator signing history
type ValidatorSigningInfo struct {
	Address             string `json:"address"`
	StartHeight         uint64 `json:"start_height"`
	IndexOffset         uint64 `json:"index_offset"`
	JailedUntil         int64  `json:"jailed_until"`
	Tombstoned          bool   `json:"tombstoned"`
	MissedBlocksCounter uint64 `json:"missed_blocks_counter"`
	SignedBlocksBitmap  []bool `json:"signed_blocks_bitmap"`
}

// SlashingEvent records a slashing incident
type SlashingEvent struct {
	ValidatorAddress string         `json:"validator_address"`
	Height           uint64         `json:"height"`
	Reason           SlashingReason `json:"reason"`
	Amount           uint64         `json:"amount"`
	Timestamp        int64          `json:"timestamp"`
}

// NewSlashingKeeper creates a new slashing keeper
func NewSlashingKeeper(engine *Engine, params *SlashingParams) *SlashingKeeper {
	if params == nil {
		params = DefaultSlashingParams()
	}

	return &SlashingKeeper{
		params:         params,
		engine:         engine,
		signingInfo:    make(map[string]*ValidatorSigningInfo),
		slashingEvents: make([]SlashingEvent, 0),
	}
}

// HandleDoubleSign processes a double signing infraction
func (k *SlashingKeeper) HandleDoubleSign(address string, height uint64) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	validator, err := k.engine.GetValidator(address)
	if err != nil {
		return err
	}

	// Check if already tombstoned
	info := k.getOrCreateSigningInfo(address)
	if info.Tombstoned {
		return nil // Already permanently jailed
	}

	// Slash
	slashAmount := validator.Slash(k.params.DoubleSignPenalty/100, string(SlashReasonDoubleSign), height)

	// Jail
	validator.Jail(k.params.DoubleSignJailDuration)

	// Tombstone (permanent)
	info.Tombstoned = true

	// Record event
	k.slashingEvents = append(k.slashingEvents, SlashingEvent{
		ValidatorAddress: address,
		Height:           height,
		Reason:           SlashReasonDoubleSign,
		Amount:           slashAmount,
		Timestamp:        time.Now().Unix(),
	})

	return nil
}

// HandleDowntime processes a downtime infraction
func (k *SlashingKeeper) HandleDowntime(address string, height uint64) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	validator, err := k.engine.GetValidator(address)
	if err != nil {
		return err
	}

	info := k.getOrCreateSigningInfo(address)

	// Check if already jailed
	if info.JailedUntil > time.Now().Unix() {
		return nil
	}

	// Slash
	slashAmount := validator.Slash(k.params.DowntimePenalty/100, string(SlashReasonDowntime), height)

	// Jail
	validator.Jail(k.params.DowntimeJailDuration)
	info.JailedUntil = time.Now().Add(k.params.DowntimeJailDuration).Unix()

	// Record event
	k.slashingEvents = append(k.slashingEvents, SlashingEvent{
		ValidatorAddress: address,
		Height:           height,
		Reason:           SlashReasonDowntime,
		Amount:           slashAmount,
		Timestamp:        time.Now().Unix(),
	})

	return nil
}

// SignBlock records a validator signing a block
func (k *SlashingKeeper) SignBlock(address string, height uint64, signed bool) {
	k.mu.Lock()
	defer k.mu.Unlock()

	info := k.getOrCreateSigningInfo(address)

	// Update bitmap
	index := height % k.params.SignedBlocksWindow
	if uint64(len(info.SignedBlocksBitmap)) <= index {
		// Extend bitmap
		newBitmap := make([]bool, k.params.SignedBlocksWindow)
		copy(newBitmap, info.SignedBlocksBitmap)
		info.SignedBlocksBitmap = newBitmap
	}

	// If was previously missed at this index, decrement counter
	if !info.SignedBlocksBitmap[index] && info.MissedBlocksCounter > 0 {
		info.MissedBlocksCounter--
	}

	info.SignedBlocksBitmap[index] = signed

	if !signed {
		info.MissedBlocksCounter++
	}

	// Check for downtime
	minSigned := (k.params.SignedBlocksWindow * k.params.MinSignedPerWindow) / 100
	if info.MissedBlocksCounter > k.params.SignedBlocksWindow-minSigned {
		go k.HandleDowntime(address, height)
	}
}

// getOrCreateSigningInfo gets or creates signing info for a validator
func (k *SlashingKeeper) getOrCreateSigningInfo(address string) *ValidatorSigningInfo {
	info, exists := k.signingInfo[address]
	if !exists {
		info = &ValidatorSigningInfo{
			Address:            address,
			SignedBlocksBitmap: make([]bool, k.params.SignedBlocksWindow),
		}
		k.signingInfo[address] = info
	}
	return info
}

// GetSigningInfo returns signing info for a validator
func (k *SlashingKeeper) GetSigningInfo(address string) *ValidatorSigningInfo {
	k.mu.RLock()
	defer k.mu.RUnlock()

	info, exists := k.signingInfo[address]
	if !exists {
		return nil
	}

	// Copy to avoid race conditions
	copy := *info
	copy.SignedBlocksBitmap = append([]bool{}, info.SignedBlocksBitmap...)
	return &copy
}

// GetSlashingEvents returns recent slashing events
func (k *SlashingKeeper) GetSlashingEvents(limit int) []SlashingEvent {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if limit <= 0 || limit > len(k.slashingEvents) {
		limit = len(k.slashingEvents)
	}

	start := len(k.slashingEvents) - limit
	events := make([]SlashingEvent, limit)
	copy(events, k.slashingEvents[start:])

	return events
}

// IsTombstoned returns true if validator is permanently jailed
func (k *SlashingKeeper) IsTombstoned(address string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()

	info, exists := k.signingInfo[address]
	if !exists {
		return false
	}

	return info.Tombstoned
}

// Unjail attempts to unjail a validator
func (k *SlashingKeeper) Unjail(address string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	info, exists := k.signingInfo[address]
	if !exists {
		return ErrValidatorNotFound
	}

	if info.Tombstoned {
		return &SlashingError{"validator is tombstoned"}
	}

	if info.JailedUntil > time.Now().Unix() {
		return ErrStillJailed
	}

	validator, err := k.engine.GetValidator(address)
	if err != nil {
		return err
	}

	return validator.Unjail()
}

// UpdateParams updates slashing parameters
func (k *SlashingKeeper) UpdateParams(params *SlashingParams) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.params = params
}

// GetParams returns current slashing parameters
func (k *SlashingKeeper) GetParams() *SlashingParams {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.params
}

// SlashingError represents a slashing error
type SlashingError struct {
	msg string
}

func (e *SlashingError) Error() string {
	return e.msg
}
