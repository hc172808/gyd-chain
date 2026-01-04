package miner

import (
	"errors"
	"sync"
	"time"
)

// Errors
var (
	ErrInvalidJob       = errors.New("invalid job")
	ErrJobNotFound      = errors.New("job not found")
	ErrDuplicateShare   = errors.New("duplicate share")
	ErrLowDifficulty    = errors.New("share difficulty too low")
	ErrInvalidNonce     = errors.New("invalid nonce")
	ErrStaleShare       = errors.New("stale share")
)

// SubmissionHandler handles share submissions
type SubmissionHandler struct {
	jobManager *JobManager
	
	// Share tracking
	submissions map[string]map[uint64]bool // jobID -> nonce -> submitted
	subMu       sync.RWMutex
	
	// Statistics
	stats       SubmissionStats
	statsMu     sync.RWMutex
	
	// Callbacks
	onBlockFound func(block *BlockSubmission)
}

// SubmissionStats tracks submission statistics
type SubmissionStats struct {
	TotalSubmissions   uint64
	ValidShares        uint64
	InvalidShares      uint64
	StaleShares        uint64
	DuplicateShares    uint64
	BlocksFound        uint64
	LastSubmissionTime time.Time
}

// BlockSubmission represents a found block
type BlockSubmission struct {
	JobID     string
	Height    uint64
	Nonce     uint64
	Timestamp uint64
	Hash      []byte
	MinerID   string
	FoundAt   time.Time
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler(jm *JobManager, onBlockFound func(*BlockSubmission)) *SubmissionHandler {
	return &SubmissionHandler{
		jobManager:   jm,
		submissions:  make(map[string]map[uint64]bool),
		onBlockFound: onBlockFound,
	}
}

// Submit processes a share submission
func (sh *SubmissionHandler) Submit(minerID string, submission *ShareSubmission) (*SubmissionResult, error) {
	sh.statsMu.Lock()
	sh.stats.TotalSubmissions++
	sh.stats.LastSubmissionTime = time.Now()
	sh.statsMu.Unlock()
	
	// Validate job exists
	job := sh.jobManager.GetJob(submission.JobID)
	if job == nil {
		sh.statsMu.Lock()
		sh.stats.StaleShares++
		sh.statsMu.Unlock()
		return nil, ErrJobNotFound
	}
	
	// Check for duplicate
	if sh.isDuplicate(submission.JobID, submission.Nonce) {
		sh.statsMu.Lock()
		sh.stats.DuplicateShares++
		sh.statsMu.Unlock()
		return nil, ErrDuplicateShare
	}
	
	// Mark as submitted
	sh.markSubmitted(submission.JobID, submission.Nonce)
	
	// Validate the work
	workResult := &WorkResult{
		JobID:     submission.JobID,
		Nonce:     submission.Nonce,
		Timestamp: submission.Timestamp,
		Hash:      submission.Hash,
	}
	
	if !sh.jobManager.ValidateWork(workResult) {
		sh.statsMu.Lock()
		sh.stats.InvalidShares++
		sh.statsMu.Unlock()
		return &SubmissionResult{
			Valid:  false,
			Reason: "invalid proof of work",
		}, nil
	}
	
	// Valid share
	sh.statsMu.Lock()
	sh.stats.ValidShares++
	sh.statsMu.Unlock()
	
	result := &SubmissionResult{
		Valid:      true,
		Difficulty: submission.Difficulty,
	}
	
	// Check if this is a block
	if sh.meetsBlockDifficulty(submission, job) {
		sh.statsMu.Lock()
		sh.stats.BlocksFound++
		sh.statsMu.Unlock()
		
		result.IsBlock = true
		
		// Notify block found
		if sh.onBlockFound != nil {
			blockSub := &BlockSubmission{
				JobID:     submission.JobID,
				Height:    job.Height,
				Nonce:     submission.Nonce,
				Timestamp: submission.Timestamp,
				Hash:      submission.Hash,
				MinerID:   minerID,
				FoundAt:   time.Now(),
			}
			go sh.onBlockFound(blockSub)
		}
	}
	
	return result, nil
}

// ShareSubmission represents a submitted share
type ShareSubmission struct {
	JobID      string
	Nonce      uint64
	Timestamp  uint64
	Hash       []byte
	Difficulty uint64
}

// SubmissionResult represents the result of a submission
type SubmissionResult struct {
	Valid      bool
	IsBlock    bool
	Difficulty uint64
	Reason     string
}

// isDuplicate checks if a share is a duplicate
func (sh *SubmissionHandler) isDuplicate(jobID string, nonce uint64) bool {
	sh.subMu.RLock()
	defer sh.subMu.RUnlock()
	
	if jobSubs, ok := sh.submissions[jobID]; ok {
		return jobSubs[nonce]
	}
	return false
}

// markSubmitted marks a share as submitted
func (sh *SubmissionHandler) markSubmitted(jobID string, nonce uint64) {
	sh.subMu.Lock()
	defer sh.subMu.Unlock()
	
	if _, ok := sh.submissions[jobID]; !ok {
		sh.submissions[jobID] = make(map[uint64]bool)
	}
	sh.submissions[jobID][nonce] = true
}

// meetsBlockDifficulty checks if a share meets block difficulty
func (sh *SubmissionHandler) meetsBlockDifficulty(submission *ShareSubmission, job *Job) bool {
	return submission.Difficulty >= job.Difficulty
}

// GetStats returns submission statistics
func (sh *SubmissionHandler) GetStats() SubmissionStats {
	sh.statsMu.RLock()
	defer sh.statsMu.RUnlock()
	return sh.stats
}

// CleanOldSubmissions removes old submission records
func (sh *SubmissionHandler) CleanOldSubmissions(maxJobs int) {
	sh.subMu.Lock()
	defer sh.subMu.Unlock()
	
	if len(sh.submissions) <= maxJobs {
		return
	}
	
	currentJob := sh.jobManager.GetCurrentJob()
	
	// Remove oldest entries
	for jobID := range sh.submissions {
		if len(sh.submissions) <= maxJobs {
			break
		}
		if currentJob == nil || jobID != currentJob.ID {
			delete(sh.submissions, jobID)
		}
	}
}

// ShareValidator validates shares
type ShareValidator struct {
	minDifficulty uint64
	maxTimeDrift  time.Duration
}

// NewShareValidator creates a new share validator
func NewShareValidator(minDiff uint64, maxDrift time.Duration) *ShareValidator {
	return &ShareValidator{
		minDifficulty: minDiff,
		maxTimeDrift:  maxDrift,
	}
}

// Validate validates a share submission
func (sv *ShareValidator) Validate(submission *ShareSubmission) error {
	// Check difficulty
	if submission.Difficulty < sv.minDifficulty {
		return ErrLowDifficulty
	}
	
	// Check timestamp
	now := uint64(time.Now().Unix())
	drift := sv.maxTimeDrift.Milliseconds() / 1000
	
	if submission.Timestamp > now+uint64(drift) {
		return errors.New("timestamp too far in the future")
	}
	
	if submission.Timestamp < now-uint64(drift) {
		return ErrStaleShare
	}
	
	return nil
}
