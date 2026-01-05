package pow

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"sync"
	"time"
)

// CPUMiner implements CPU-based proof of work mining
type CPUMiner struct {
	mu          sync.RWMutex
	running     bool
	hashRate    uint64
	difficulty  *big.Int
	workers     int
	stopChan    chan struct{}
	resultChan  chan *MiningResult
}

// MiningResult contains the result of a successful mining operation
type MiningResult struct {
	Nonce      uint64 `json:"nonce"`
	Hash       string `json:"hash"`
	Difficulty uint64 `json:"difficulty"`
	Timestamp  int64  `json:"timestamp"`
	WorkerID   int    `json:"worker_id"`
}

// NewCPUMiner creates a new CPU miner
func NewCPUMiner(workers int) *CPUMiner {
	if workers <= 0 {
		workers = 1
	}
	
	return &CPUMiner{
		workers:    workers,
		difficulty: big.NewInt(1),
		stopChan:   make(chan struct{}),
		resultChan: make(chan *MiningResult, 1),
	}
}

// Start begins mining
func (m *CPUMiner) Start(blockData []byte, target *big.Int) <-chan *MiningResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.running {
		return m.resultChan
	}
	
	m.running = true
	m.difficulty = target
	m.stopChan = make(chan struct{})
	m.resultChan = make(chan *MiningResult, 1)
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < m.workers; i++ {
		wg.Add(1)
		go m.mine(blockData, target, uint64(i), &wg)
	}
	
	// Cleanup goroutine
	go func() {
		wg.Wait()
		close(m.resultChan)
	}()
	
	return m.resultChan
}

// Stop halts mining
func (m *CPUMiner) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.running {
		return
	}
	
	m.running = false
	close(m.stopChan)
}

// mine is the worker function that searches for a valid nonce
func (m *CPUMiner) mine(blockData []byte, target *big.Int, workerID uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	
	nonce := workerID
	startTime := time.Now()
	hashes := uint64(0)
	
	for {
		select {
		case <-m.stopChan:
			return
		default:
			// Calculate hash
			hash := m.calculateHash(blockData, nonce)
			hashes++
			
			// Check if meets target
			hashInt := new(big.Int).SetBytes(hash)
			if hashInt.Cmp(target) < 0 {
				m.mu.Lock()
				if m.running {
					m.resultChan <- &MiningResult{
						Nonce:      nonce,
						Hash:       hex.EncodeToString(hash),
						Difficulty: m.difficulty.Uint64(),
						Timestamp:  time.Now().Unix(),
						WorkerID:   int(workerID),
					}
					m.running = false
					close(m.stopChan)
				}
				m.mu.Unlock()
				return
			}
			
			// Update hash rate periodically
			if hashes%10000 == 0 {
				elapsed := time.Since(startTime).Seconds()
				if elapsed > 0 {
					m.mu.Lock()
					m.hashRate = uint64(float64(hashes) / elapsed)
					m.mu.Unlock()
				}
			}
			
			// Increment nonce by worker count to distribute work
			nonce += uint64(m.workers)
		}
	}
}

// calculateHash computes the hash for a given nonce
func (m *CPUMiner) calculateHash(blockData []byte, nonce uint64) []byte {
	// Append nonce to block data
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	
	data := append(blockData, nonceBytes...)
	
	// Double SHA256 (like Bitcoin)
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	
	return second[:]
}

// GetHashRate returns the current hash rate
func (m *CPUMiner) GetHashRate() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.hashRate
}

// IsRunning returns true if miner is active
func (m *CPUMiner) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// SetWorkers updates the number of workers
func (m *CPUMiner) SetWorkers(workers int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if workers > 0 {
		m.workers = workers
	}
}

// CalculateTarget converts difficulty to target
func CalculateTarget(difficulty uint64) *big.Int {
	if difficulty == 0 {
		difficulty = 1
	}
	
	// Max target / difficulty
	maxTarget := new(big.Int)
	maxTarget.SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	
	target := new(big.Int).Div(maxTarget, big.NewInt(int64(difficulty)))
	return target
}

// ValidatePoW verifies a proof of work
func ValidatePoW(blockData []byte, nonce uint64, target *big.Int) bool {
	// Append nonce to block data
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	
	data := append(blockData, nonceBytes...)
	
	// Double SHA256
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	
	hashInt := new(big.Int).SetBytes(second[:])
	return hashInt.Cmp(target) < 0
}

// DifficultyAdjustment calculates new difficulty
func DifficultyAdjustment(currentDiff uint64, blockTime, targetTime time.Duration) uint64 {
	if blockTime == 0 {
		return currentDiff
	}
	
	ratio := float64(targetTime) / float64(blockTime)
	
	// Limit adjustment to 4x in either direction
	if ratio > 4 {
		ratio = 4
	} else if ratio < 0.25 {
		ratio = 0.25
	}
	
	newDiff := uint64(float64(currentDiff) * ratio)
	if newDiff < 1 {
		newDiff = 1
	}
	
	return newDiff
}

// MinerStats contains mining statistics
type MinerStats struct {
	HashRate     uint64  `json:"hash_rate"`
	BlocksFound  uint64  `json:"blocks_found"`
	TotalHashes  uint64  `json:"total_hashes"`
	Workers      int     `json:"workers"`
	Difficulty   uint64  `json:"difficulty"`
	Uptime       float64 `json:"uptime_hours"`
	AvgBlockTime float64 `json:"avg_block_time"`
}

// Argon2Config for memory-hard hashing (optional alternative)
type Argon2Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

// DefaultArgon2Config returns default Argon2 parameters
func DefaultArgon2Config() *Argon2Config {
	return &Argon2Config{
		Time:    1,
		Memory:  64 * 1024, // 64MB
		Threads: 4,
		KeyLen:  32,
	}
}
