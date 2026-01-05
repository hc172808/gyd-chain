package pow

import (
	"sync"
	"time"
)

// RewardDistributor handles mining reward distribution
type RewardDistributor struct {
	mu            sync.RWMutex
	baseReward    uint64
	halving       uint64 // blocks between halvings
	halvingCount  uint64
	minReward     uint64
	minerShare    uint64 // basis points (e.g., 2000 = 20%)
	validatorShare uint64
	totalDistributed uint64
	lastHeight    uint64
}

// RewardConfig contains reward configuration
type RewardConfig struct {
	BaseReward     uint64 `json:"base_reward"`
	HalvingBlocks  uint64 `json:"halving_blocks"`
	MinReward      uint64 `json:"min_reward"`
	MinerShare     uint64 `json:"miner_share"`     // basis points
	ValidatorShare uint64 `json:"validator_share"` // basis points
}

// DefaultRewardConfig returns default reward configuration
func DefaultRewardConfig() *RewardConfig {
	return &RewardConfig{
		BaseReward:     10 * 1e8,    // 10 GYDS
		HalvingBlocks:  2100000,     // ~4 years at 1 block/minute
		MinReward:      1e6,         // 0.01 GYDS
		MinerShare:     2000,        // 20%
		ValidatorShare: 8000,        // 80%
	}
}

// NewRewardDistributor creates a new reward distributor
func NewRewardDistributor(config *RewardConfig) *RewardDistributor {
	if config == nil {
		config = DefaultRewardConfig()
	}
	
	return &RewardDistributor{
		baseReward:     config.BaseReward,
		halving:        config.HalvingBlocks,
		minReward:      config.MinReward,
		minerShare:     config.MinerShare,
		validatorShare: config.ValidatorShare,
	}
}

// CalculateBlockReward calculates the block reward for a given height
func (d *RewardDistributor) CalculateBlockReward(height uint64) uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	halvings := height / d.halving
	reward := d.baseReward
	
	for i := uint64(0); i < halvings && reward > d.minReward; i++ {
		reward /= 2
	}
	
	if reward < d.minReward {
		reward = d.minReward
	}
	
	return reward
}

// DistributeReward calculates reward distribution
func (d *RewardDistributor) DistributeReward(height uint64, fees uint64) *BlockReward {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	blockReward := d.CalculateBlockReward(height)
	totalReward := blockReward + fees
	
	minerReward := (totalReward * d.minerShare) / 10000
	validatorReward := totalReward - minerReward
	
	d.totalDistributed += totalReward
	d.lastHeight = height
	
	return &BlockReward{
		Height:          height,
		BlockReward:     blockReward,
		Fees:            fees,
		TotalReward:     totalReward,
		MinerReward:     minerReward,
		ValidatorReward: validatorReward,
		Timestamp:       time.Now().Unix(),
	}
}

// BlockReward contains reward distribution details
type BlockReward struct {
	Height          uint64 `json:"height"`
	BlockReward     uint64 `json:"block_reward"`
	Fees            uint64 `json:"fees"`
	TotalReward     uint64 `json:"total_reward"`
	MinerReward     uint64 `json:"miner_reward"`
	ValidatorReward uint64 `json:"validator_reward"`
	Timestamp       int64  `json:"timestamp"`
}

// GetTotalDistributed returns total rewards distributed
func (d *RewardDistributor) GetTotalDistributed() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.totalDistributed
}

// GetCurrentReward returns the current block reward
func (d *RewardDistributor) GetCurrentReward(height uint64) uint64 {
	return d.CalculateBlockReward(height)
}

// NextHalving returns the block height of the next halving
func (d *RewardDistributor) NextHalving(currentHeight uint64) uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	currentHalving := currentHeight / d.halving
	return (currentHalving + 1) * d.halving
}

// HalvingsOccurred returns the number of halvings that have occurred
func (d *RewardDistributor) HalvingsOccurred(height uint64) uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return height / d.halving
}

// EstimatedSupply calculates estimated GYDS supply at a given height
func (d *RewardDistributor) EstimatedSupply(height uint64) uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	var supply uint64
	reward := d.baseReward
	
	for h := uint64(0); h < height; {
		blocksUntilHalving := d.halving - (h % d.halving)
		blocksToCount := blocksUntilHalving
		
		if h+blocksToCount > height {
			blocksToCount = height - h
		}
		
		supply += reward * blocksToCount
		h += blocksToCount
		
		if h%d.halving == 0 && reward > d.minReward {
			reward /= 2
		}
	}
	
	return supply
}

// UpdateShares updates the miner/validator share percentages
func (d *RewardDistributor) UpdateShares(minerShare, validatorShare uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if minerShare+validatorShare != 10000 {
		return &RewardError{"shares must sum to 10000"}
	}
	
	d.minerShare = minerShare
	d.validatorShare = validatorShare
	return nil
}

// Stats returns reward distribution statistics
type RewardStats struct {
	TotalDistributed uint64  `json:"total_distributed"`
	CurrentReward    uint64  `json:"current_reward"`
	NextHalving      uint64  `json:"next_halving"`
	Halvings         uint64  `json:"halvings"`
	EstimatedSupply  uint64  `json:"estimated_supply"`
	MinerShare       float64 `json:"miner_share_percent"`
	ValidatorShare   float64 `json:"validator_share_percent"`
}

// GetStats returns current reward statistics
func (d *RewardDistributor) GetStats(height uint64) *RewardStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	return &RewardStats{
		TotalDistributed: d.totalDistributed,
		CurrentReward:    d.CalculateBlockReward(height),
		NextHalving:      d.NextHalving(height),
		Halvings:         height / d.halving,
		EstimatedSupply:  d.EstimatedSupply(height),
		MinerShare:       float64(d.minerShare) / 100,
		ValidatorShare:   float64(d.validatorShare) / 100,
	}
}

// MinerPayout represents a payout to a miner
type MinerPayout struct {
	Address   string `json:"address"`
	Amount    uint64 `json:"amount"`
	BlockHash string `json:"block_hash"`
	Height    uint64 `json:"height"`
	Timestamp int64  `json:"timestamp"`
}

// RewardError represents a reward calculation error
type RewardError struct {
	msg string
}

func (e *RewardError) Error() string {
	return e.msg
}
