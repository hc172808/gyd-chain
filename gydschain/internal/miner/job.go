package miner

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/gydschain/gydschain/internal/chain"
	"github.com/gydschain/gydschain/internal/crypto"
)

// Job represents a mining job
type Job struct {
	ID          string
	Height      uint64
	BlockHeader []byte
	Target      []byte
	Difficulty  uint64
	Timestamp   uint64
	ExtraData   []byte
	
	// Template data
	PrevHash    []byte
	StateRoot   []byte
	TxRoot      []byte
	Coinbase    []byte
}

// JobManager manages mining jobs
type JobManager struct {
	jobs       map[string]*Job
	currentJob *Job
	mu         sync.RWMutex
	
	// Callbacks
	onNewBlock func(*chain.Block)
}

// NewJobManager creates a new job manager
func NewJobManager(onNewBlock func(*chain.Block)) *JobManager {
	return &JobManager{
		jobs:       make(map[string]*Job),
		onNewBlock: onNewBlock,
	}
}

// CreateJob creates a new mining job
func (jm *JobManager) CreateJob(template *BlockTemplate) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	
	job := &Job{
		ID:          generateJobID(),
		Height:      template.Height,
		BlockHeader: template.HeaderBytes,
		Target:      template.Target,
		Difficulty:  template.Difficulty,
		Timestamp:   uint64(time.Now().Unix()),
		ExtraData:   template.ExtraData,
		PrevHash:    template.PrevHash,
		StateRoot:   template.StateRoot,
		TxRoot:      template.TxRoot,
		Coinbase:    template.Coinbase,
	}
	
	jm.jobs[job.ID] = job
	jm.currentJob = job
	
	// Clean old jobs
	jm.cleanOldJobs()
	
	return job
}

// GetCurrentJob returns the current job
func (jm *JobManager) GetCurrentJob() *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.currentJob
}

// GetJob returns a job by ID
func (jm *JobManager) GetJob(id string) *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.jobs[id]
}

// cleanOldJobs removes old jobs
func (jm *JobManager) cleanOldJobs() {
	maxJobs := 10
	if len(jm.jobs) <= maxJobs {
		return
	}
	
	// Remove oldest jobs
	for id := range jm.jobs {
		if len(jm.jobs) <= maxJobs {
			break
		}
		if id != jm.currentJob.ID {
			delete(jm.jobs, id)
		}
	}
}

// BlockTemplate contains data for creating a mining job
type BlockTemplate struct {
	Height      uint64
	PrevHash    []byte
	StateRoot   []byte
	TxRoot      []byte
	Target      []byte
	Difficulty  uint64
	Coinbase    []byte
	HeaderBytes []byte
	ExtraData   []byte
}

// NewBlockTemplate creates a block template
func NewBlockTemplate(
	height uint64,
	prevHash []byte,
	stateRoot []byte,
	txRoot []byte,
	difficulty uint64,
	coinbase []byte,
) *BlockTemplate {
	target := difficultyToTarget(difficulty)
	
	template := &BlockTemplate{
		Height:     height,
		PrevHash:   prevHash,
		StateRoot:  stateRoot,
		TxRoot:     txRoot,
		Target:     target,
		Difficulty: difficulty,
		Coinbase:   coinbase,
	}
	
	// Build header bytes
	template.HeaderBytes = buildHeaderBytes(template)
	
	return template
}

// buildHeaderBytes builds the header bytes for mining
func buildHeaderBytes(template *BlockTemplate) []byte {
	// Concatenate header fields for hashing
	// This is a simplified version
	header := make([]byte, 0, 200)
	
	// Add height (8 bytes)
	header = append(header, uint64ToBytes(template.Height)...)
	
	// Add prev hash (32 bytes)
	header = append(header, template.PrevHash...)
	
	// Add state root (32 bytes)
	header = append(header, template.StateRoot...)
	
	// Add tx root (32 bytes)
	header = append(header, template.TxRoot...)
	
	// Add timestamp placeholder (8 bytes) - will be filled by miner
	header = append(header, make([]byte, 8)...)
	
	// Add nonce placeholder (8 bytes) - will be filled by miner
	header = append(header, make([]byte, 8)...)
	
	return header
}

// difficultyToTarget converts difficulty to target
func difficultyToTarget(difficulty uint64) []byte {
	// Target = MaxTarget / Difficulty
	// MaxTarget is 2^256 - 1
	// This is a simplified version
	target := make([]byte, 32)
	
	if difficulty == 0 {
		difficulty = 1
	}
	
	// Simple target calculation
	leadingZeros := 0
	d := difficulty
	for d > 0 {
		d >>= 4
		leadingZeros++
	}
	
	// Set target bytes
	for i := leadingZeros; i < 32; i++ {
		target[i] = 0xff
	}
	
	return target
}

// uint64ToBytes converts uint64 to bytes
func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		b[i] = byte(v & 0xff)
		v >>= 8
	}
	return b
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return hex.EncodeToString(crypto.RandomBytes(8))
}

// WorkResult represents a mining work result
type WorkResult struct {
	JobID     string
	Nonce     uint64
	Timestamp uint64
	Hash      []byte
	ExtraNonce []byte
}

// ValidateWork validates a work result
func (jm *JobManager) ValidateWork(result *WorkResult) bool {
	job := jm.GetJob(result.JobID)
	if job == nil {
		return false
	}
	
	// Rebuild header with nonce and timestamp
	header := make([]byte, len(job.BlockHeader))
	copy(header, job.BlockHeader)
	
	// Insert timestamp
	timestampOffset := 32 + 32 + 32 + 8 // height + prevhash + stateroot + txroot
	copy(header[timestampOffset:], uint64ToBytes(result.Timestamp))
	
	// Insert nonce
	nonceOffset := timestampOffset + 8
	copy(header[nonceOffset:], uint64ToBytes(result.Nonce))
	
	// Hash the header
	hash := crypto.Hash256(header)
	
	// Check against target
	return compareHash(hash, job.Target)
}

// compareHash checks if hash meets target
func compareHash(hash, target []byte) bool {
	for i := 0; i < 32; i++ {
		if hash[i] < target[i] {
			return true
		}
		if hash[i] > target[i] {
			return false
		}
	}
	return true
}
