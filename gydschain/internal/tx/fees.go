package tx

import (
	"sync"
)

// FeeConfig contains fee configuration
type FeeConfig struct {
	MinGasPrice      uint64 `json:"min_gas_price"`
	MaxGasPrice      uint64 `json:"max_gas_price"`
	BaseFee          uint64 `json:"base_fee"`
	GasPerByte       uint64 `json:"gas_per_byte"`
	GasPerSignature  uint64 `json:"gas_per_signature"`
	TransferGas      uint64 `json:"transfer_gas"`
	StakeGas         uint64 `json:"stake_gas"`
	UnstakeGas       uint64 `json:"unstake_gas"`
	CreateAssetGas   uint64 `json:"create_asset_gas"`
}

// DefaultFeeConfig returns default fee configuration
func DefaultFeeConfig() *FeeConfig {
	return &FeeConfig{
		MinGasPrice:     1,
		MaxGasPrice:     1000000,
		BaseFee:         1000,
		GasPerByte:      10,
		GasPerSignature: 5000,
		TransferGas:     21000,
		StakeGas:        50000,
		UnstakeGas:      50000,
		CreateAssetGas:  100000,
	}
}

// FeeEstimator estimates transaction fees
type FeeEstimator struct {
	mu         sync.RWMutex
	config     *FeeConfig
	recentFees []uint64
	avgGasPrice uint64
}

// NewFeeEstimator creates a new fee estimator
func NewFeeEstimator(config *FeeConfig) *FeeEstimator {
	if config == nil {
		config = DefaultFeeConfig()
	}

	return &FeeEstimator{
		config:     config,
		recentFees: make([]uint64, 0, 100),
		avgGasPrice: config.MinGasPrice,
	}
}

// EstimateFee estimates the fee for a transaction
func (e *FeeEstimator) EstimateFee(tx *Transaction) uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	gas := e.EstimateGas(tx)
	return gas * e.avgGasPrice
}

// EstimateGas estimates gas needed for a transaction
func (e *FeeEstimator) EstimateGas(tx *Transaction) uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var gas uint64

	// Base gas by transaction type
	switch tx.Type {
	case TxTypeTransfer:
		gas = e.config.TransferGas
	case TxTypeStake:
		gas = e.config.StakeGas
	case TxTypeUnstake:
		gas = e.config.UnstakeGas
	case TxTypeCreateAsset:
		gas = e.config.CreateAssetGas
	default:
		gas = e.config.TransferGas
	}

	// Add gas for data
	gas += uint64(len(tx.Data)) * e.config.GasPerByte

	// Add gas for signature
	gas += e.config.GasPerSignature

	return gas
}

// SuggestGasPrice suggests a gas price based on recent transactions
func (e *FeeEstimator) SuggestGasPrice(priority string) uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	base := e.avgGasPrice
	if base < e.config.MinGasPrice {
		base = e.config.MinGasPrice
	}

	switch priority {
	case "low":
		return base
	case "medium":
		return base * 120 / 100 // 20% higher
	case "high":
		return base * 150 / 100 // 50% higher
	case "urgent":
		return base * 200 / 100 // 100% higher
	default:
		return base * 120 / 100
	}
}

// RecordFee records a fee from a confirmed transaction
func (e *FeeEstimator) RecordFee(fee, gasUsed uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if gasUsed == 0 {
		return
	}

	gasPrice := fee / gasUsed

	e.recentFees = append(e.recentFees, gasPrice)

	// Keep only last 100 fees
	if len(e.recentFees) > 100 {
		e.recentFees = e.recentFees[1:]
	}

	// Recalculate average
	e.recalculateAverage()
}

// recalculateAverage updates the average gas price
func (e *FeeEstimator) recalculateAverage() {
	if len(e.recentFees) == 0 {
		e.avgGasPrice = e.config.MinGasPrice
		return
	}

	var sum uint64
	for _, fee := range e.recentFees {
		sum += fee
	}

	e.avgGasPrice = sum / uint64(len(e.recentFees))

	// Clamp to bounds
	if e.avgGasPrice < e.config.MinGasPrice {
		e.avgGasPrice = e.config.MinGasPrice
	}
	if e.avgGasPrice > e.config.MaxGasPrice {
		e.avgGasPrice = e.config.MaxGasPrice
	}
}

// GetAverageGasPrice returns the current average gas price
func (e *FeeEstimator) GetAverageGasPrice() uint64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.avgGasPrice
}

// UpdateConfig updates the fee configuration
func (e *FeeEstimator) UpdateConfig(config *FeeConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = config
}

// GetConfig returns the current fee configuration
func (e *FeeEstimator) GetConfig() *FeeConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy
	config := *e.config
	return &config
}

// FeeEstimate contains fee estimate details
type FeeEstimate struct {
	GasUsed     uint64 `json:"gas_used"`
	GasPrice    uint64 `json:"gas_price"`
	TotalFee    uint64 `json:"total_fee"`
	GYDSFee     uint64 `json:"gyds_fee"` // Fee in GYDS
	Priority    string `json:"priority"`
	EstimatedTime string `json:"estimated_time"`
}

// GetFeeEstimate returns a detailed fee estimate
func (e *FeeEstimator) GetFeeEstimate(tx *Transaction, priority string) *FeeEstimate {
	gas := e.EstimateGas(tx)
	gasPrice := e.SuggestGasPrice(priority)
	totalFee := gas * gasPrice

	var estimatedTime string
	switch priority {
	case "low":
		estimatedTime = "~5 minutes"
	case "medium":
		estimatedTime = "~1 minute"
	case "high":
		estimatedTime = "~30 seconds"
	case "urgent":
		estimatedTime = "~10 seconds"
	default:
		estimatedTime = "~1 minute"
	}

	return &FeeEstimate{
		GasUsed:       gas,
		GasPrice:      gasPrice,
		TotalFee:      totalFee,
		GYDSFee:       totalFee, // Assuming fees are in GYDS smallest unit
		Priority:      priority,
		EstimatedTime: estimatedTime,
	}
}

// CalculateBurnAmount calculates the amount to burn from fees
func CalculateBurnAmount(totalFees, burnRate uint64) uint64 {
	return (totalFees * burnRate) / 10000 // burnRate in basis points
}

// CalculateValidatorShare calculates validator's share of fees
func CalculateValidatorShare(totalFees, burnRate uint64) uint64 {
	burnAmount := CalculateBurnAmount(totalFees, burnRate)
	return totalFees - burnAmount
}
