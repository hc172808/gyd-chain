package miner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Pool represents a mining pool server
type Pool struct {
	addr     string
	router   *mux.Router
	upgrader websocket.Upgrader
	
	// Connected miners
	miners   map[string]*PoolMiner
	minersMu sync.RWMutex
	
	// Current work
	currentJob *Job
	jobMu      sync.RWMutex
	
	// Statistics
	stats    PoolStats
	statsMu  sync.RWMutex
	
	// Configuration
	config   PoolConfig
	
	// Channels
	newJobs  chan *Job
	shares   chan *Share
	stop     chan struct{}
}

// PoolConfig contains pool configuration
type PoolConfig struct {
	MinDifficulty    uint64  `json:"min_difficulty"`
	MaxDifficulty    uint64  `json:"max_difficulty"`
	VarDiffTarget    float64 `json:"vardiff_target"`    // Target shares per minute
	VarDiffRetarget  int     `json:"vardiff_retarget"`  // Retarget interval in seconds
	PayoutThreshold  string  `json:"payout_threshold"`
	PoolFee          float64 `json:"pool_fee"`          // Percentage
	BlockReward      string  `json:"block_reward"`
}

// PoolMiner represents a connected miner
type PoolMiner struct {
	ID            string
	Address       string
	Conn          *websocket.Conn
	Difficulty    uint64
	Hashrate      float64
	SharesValid   uint64
	SharesInvalid uint64
	LastShare     time.Time
	ConnectedAt   time.Time
	mu            sync.Mutex
}

// PoolStats contains pool statistics
type PoolStats struct {
	TotalMiners     int     `json:"total_miners"`
	TotalHashrate   float64 `json:"total_hashrate"`
	BlocksFound     uint64  `json:"blocks_found"`
	SharesValid     uint64  `json:"shares_valid"`
	SharesInvalid   uint64  `json:"shares_invalid"`
	LastBlockTime   uint64  `json:"last_block_time"`
	CurrentHeight   uint64  `json:"current_height"`
}

// Share represents a submitted share
type Share struct {
	MinerID    string
	JobID      string
	Nonce      uint64
	Hash       []byte
	Difficulty uint64
	Timestamp  time.Time
}

// NewPool creates a new mining pool
func NewPool(addr string, config PoolConfig) *Pool {
	p := &Pool{
		addr:     addr,
		router:   mux.NewRouter(),
		miners:   make(map[string]*PoolMiner),
		config:   config,
		newJobs:  make(chan *Job, 10),
		shares:   make(chan *Share, 1000),
		stop:     make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	p.setupRoutes()
	return p
}

// setupRoutes configures HTTP routes
func (p *Pool) setupRoutes() {
	p.router.HandleFunc("/", p.handleMiner)
	p.router.HandleFunc("/stats", p.handleStats).Methods("GET")
	p.router.HandleFunc("/miners", p.handleMiners).Methods("GET")
}

// Start starts the pool server
func (p *Pool) Start() error {
	// Start share processor
	go p.processShares()
	
	// Start vardiff adjuster
	go p.adjustDifficulty()
	
	// Start HTTP server
	fmt.Printf("Mining pool starting on %s\n", p.addr)
	return http.ListenAndServe(p.addr, p.router)
}

// Stop stops the pool server
func (p *Pool) Stop() {
	close(p.stop)
}

// handleMiner handles WebSocket connections from miners
func (p *Pool) handleMiner(w http.ResponseWriter, r *http.Request) {
	conn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	
	miner := &PoolMiner{
		ID:          generateMinerID(),
		Conn:        conn,
		Difficulty:  p.config.MinDifficulty,
		ConnectedAt: time.Now(),
	}
	
	p.minersMu.Lock()
	p.miners[miner.ID] = miner
	p.minersMu.Unlock()
	
	defer func() {
		p.minersMu.Lock()
		delete(p.miners, miner.ID)
		p.minersMu.Unlock()
	}()
	
	// Send current job
	p.sendJob(miner)
	
	// Handle messages
	for {
		var msg StratumMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		
		switch msg.Method {
		case "mining.subscribe":
			p.handleSubscribe(miner, msg)
		case "mining.authorize":
			p.handleAuthorize(miner, msg)
		case "mining.submit":
			p.handleSubmit(miner, msg)
		}
	}
}

// StratumMessage represents a Stratum protocol message
type StratumMessage struct {
	ID     interface{}     `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// handleSubscribe handles miner subscription
func (p *Pool) handleSubscribe(miner *PoolMiner, msg StratumMessage) {
	response := map[string]interface{}{
		"id":     msg.ID,
		"result": []interface{}{miner.ID, "00000000"},
		"error":  nil,
	}
	miner.Conn.WriteJSON(response)
}

// handleAuthorize handles miner authorization
func (p *Pool) handleAuthorize(miner *PoolMiner, msg StratumMessage) {
	var params []string
	json.Unmarshal(msg.Params, &params)
	
	if len(params) > 0 {
		miner.Address = params[0]
	}
	
	response := map[string]interface{}{
		"id":     msg.ID,
		"result": true,
		"error":  nil,
	}
	miner.Conn.WriteJSON(response)
}

// handleSubmit handles share submission
func (p *Pool) handleSubmit(miner *PoolMiner, msg StratumMessage) {
	var params []interface{}
	json.Unmarshal(msg.Params, &params)
	
	share := &Share{
		MinerID:   miner.ID,
		Timestamp: time.Now(),
	}
	
	// Parse share data from params
	if len(params) >= 3 {
		share.JobID = params[1].(string)
		// Parse nonce and other data
	}
	
	// Submit share for processing
	select {
	case p.shares <- share:
	default:
		// Channel full, drop share
	}
	
	response := map[string]interface{}{
		"id":     msg.ID,
		"result": true,
		"error":  nil,
	}
	miner.Conn.WriteJSON(response)
}

// sendJob sends a job to a miner
func (p *Pool) sendJob(miner *PoolMiner) {
	p.jobMu.RLock()
	job := p.currentJob
	p.jobMu.RUnlock()
	
	if job == nil {
		return
	}
	
	notification := map[string]interface{}{
		"id":     nil,
		"method": "mining.notify",
		"params": []interface{}{
			job.ID,
			job.BlockHeader,
			job.Target,
			true, // Clean jobs
		},
	}
	miner.Conn.WriteJSON(notification)
}

// BroadcastJob sends a new job to all miners
func (p *Pool) BroadcastJob(job *Job) {
	p.jobMu.Lock()
	p.currentJob = job
	p.jobMu.Unlock()
	
	p.minersMu.RLock()
	for _, miner := range p.miners {
		go p.sendJob(miner)
	}
	p.minersMu.RUnlock()
}

// processShares processes submitted shares
func (p *Pool) processShares() {
	for {
		select {
		case share := <-p.shares:
			p.processShare(share)
		case <-p.stop:
			return
		}
	}
}

// processShare processes a single share
func (p *Pool) processShare(share *Share) {
	p.minersMu.RLock()
	miner, exists := p.miners[share.MinerID]
	p.minersMu.RUnlock()
	
	if !exists {
		return
	}
	
	// Validate share (simplified)
	valid := true // TODO: Actual validation
	
	miner.mu.Lock()
	if valid {
		miner.SharesValid++
		miner.LastShare = share.Timestamp
	} else {
		miner.SharesInvalid++
	}
	miner.mu.Unlock()
	
	p.statsMu.Lock()
	if valid {
		p.stats.SharesValid++
	} else {
		p.stats.SharesInvalid++
	}
	p.statsMu.Unlock()
}

// adjustDifficulty adjusts miner difficulties
func (p *Pool) adjustDifficulty() {
	ticker := time.NewTicker(time.Duration(p.config.VarDiffRetarget) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			p.minersMu.RLock()
			for _, miner := range p.miners {
				p.adjustMinerDifficulty(miner)
			}
			p.minersMu.RUnlock()
		case <-p.stop:
			return
		}
	}
}

// adjustMinerDifficulty adjusts difficulty for a single miner
func (p *Pool) adjustMinerDifficulty(miner *PoolMiner) {
	// Calculate shares per minute
	// Adjust difficulty to target shares/minute
	// TODO: Implement vardiff algorithm
}

// handleStats returns pool statistics
func (p *Pool) handleStats(w http.ResponseWriter, r *http.Request) {
	p.statsMu.RLock()
	stats := p.stats
	p.statsMu.RUnlock()
	
	p.minersMu.RLock()
	stats.TotalMiners = len(p.miners)
	for _, miner := range p.miners {
		stats.TotalHashrate += miner.Hashrate
	}
	p.minersMu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleMiners returns connected miners
func (p *Pool) handleMiners(w http.ResponseWriter, r *http.Request) {
	p.minersMu.RLock()
	miners := make([]map[string]interface{}, 0, len(p.miners))
	for _, miner := range p.miners {
		miners = append(miners, map[string]interface{}{
			"id":            miner.ID,
			"address":       miner.Address,
			"difficulty":    miner.Difficulty,
			"hashrate":      miner.Hashrate,
			"shares_valid":  miner.SharesValid,
			"shares_invalid": miner.SharesInvalid,
			"connected_at":  miner.ConnectedAt,
		})
	}
	p.minersMu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(miners)
}

// generateMinerID generates a unique miner ID
func generateMinerID() string {
	return fmt.Sprintf("miner_%d", time.Now().UnixNano())
}
