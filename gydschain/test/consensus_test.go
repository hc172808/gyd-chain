package test

import (
	"math/big"
	"testing"
	"time"

	"github.com/gydschain/gydschain/internal/chain"
	"github.com/gydschain/gydschain/internal/consensus/pos"
	"github.com/gydschain/gydschain/internal/crypto"
)

func TestValidatorSet(t *testing.T) {
	vs := pos.NewValidatorSet()
	
	// Add validators
	v1 := &pos.Validator{
		Address:    "gyds1validator1",
		Stake:      big.NewInt(100000),
		Commission: 500,
		Active:     true,
	}
	v2 := &pos.Validator{
		Address:    "gyds1validator2",
		Stake:      big.NewInt(200000),
		Commission: 300,
		Active:     true,
	}
	
	vs.Add(v1)
	vs.Add(v2)
	
	if vs.Len() != 2 {
		t.Errorf("expected 2 validators, got %d", vs.Len())
	}
	
	// Get validator
	got := vs.Get("gyds1validator1")
	if got == nil {
		t.Error("expected validator, got nil")
	}
	if got.Stake.Cmp(big.NewInt(100000)) != 0 {
		t.Errorf("expected stake 100000, got %s", got.Stake.String())
	}
}

func TestValidatorSelection(t *testing.T) {
	vs := pos.NewValidatorSet()
	
	// Add validators with different stakes
	for i := 1; i <= 10; i++ {
		v := &pos.Validator{
			Address:    crypto.GenerateAddress(),
			Stake:      big.NewInt(int64(i * 10000)),
			Commission: 500,
			Active:     true,
		}
		vs.Add(v)
	}
	
	// Select proposer
	proposer := vs.SelectProposer(1, time.Now().Unix())
	if proposer == nil {
		t.Error("expected proposer, got nil")
	}
	
	// Proposer should be deterministic for same inputs
	proposer2 := vs.SelectProposer(1, time.Now().Unix())
	if proposer.Address != proposer2.Address {
		t.Error("proposer selection should be deterministic")
	}
}

func TestStaking(t *testing.T) {
	staking := pos.NewStakingManager(big.NewInt(10000))
	
	// Stake
	err := staking.Stake("gyds1user1", big.NewInt(50000))
	if err != nil {
		t.Errorf("stake failed: %v", err)
	}
	
	stake := staking.GetStake("gyds1user1")
	if stake.Cmp(big.NewInt(50000)) != 0 {
		t.Errorf("expected stake 50000, got %s", stake.String())
	}
	
	// Stake more
	err = staking.Stake("gyds1user1", big.NewInt(25000))
	if err != nil {
		t.Errorf("additional stake failed: %v", err)
	}
	
	stake = staking.GetStake("gyds1user1")
	if stake.Cmp(big.NewInt(75000)) != 0 {
		t.Errorf("expected stake 75000, got %s", stake.String())
	}
	
	// Unstake
	err = staking.Unstake("gyds1user1", big.NewInt(25000))
	if err != nil {
		t.Errorf("unstake failed: %v", err)
	}
	
	stake = staking.GetStake("gyds1user1")
	if stake.Cmp(big.NewInt(50000)) != 0 {
		t.Errorf("expected stake 50000 after unstake, got %s", stake.String())
	}
}

func TestMinimumStake(t *testing.T) {
	staking := pos.NewStakingManager(big.NewInt(10000))
	
	// Try to stake below minimum
	err := staking.Stake("gyds1user1", big.NewInt(5000))
	if err == nil {
		t.Error("expected error for stake below minimum")
	}
}

func TestSlashing(t *testing.T) {
	slasher := pos.NewSlasher()
	
	// Record signed block
	slasher.RecordSignedBlock("gyds1validator1", 1)
	slasher.RecordSignedBlock("gyds1validator1", 2)
	slasher.RecordSignedBlock("gyds1validator1", 3)
	
	// Check uptime
	uptime := slasher.GetUptime("gyds1validator1")
	if uptime < 0.9 {
		t.Errorf("expected high uptime, got %f", uptime)
	}
	
	// Record missed block
	slasher.RecordMissedBlock("gyds1validator1", 4)
	slasher.RecordMissedBlock("gyds1validator1", 5)
	
	// Check if should slash for downtime
	shouldSlash := slasher.ShouldSlashDowntime("gyds1validator1", 100, 0.5)
	if shouldSlash {
		t.Error("should not slash with only 2 missed blocks")
	}
}

func TestDoubleSignDetection(t *testing.T) {
	slasher := pos.NewSlasher()
	
	// Record first block
	block1 := &chain.Block{
		Number: 100,
	}
	slasher.RecordProposedBlock("gyds1validator1", block1)
	
	// Record different block at same height (double sign)
	block2 := &chain.Block{
		Number: 100,
	}
	
	isDoubleSign := slasher.IsDoubleSign("gyds1validator1", block2)
	if !isDoubleSign {
		t.Error("expected double sign detection")
	}
}

func TestSlashAmount(t *testing.T) {
	slasher := pos.NewSlasher()
	
	stake := big.NewInt(100000)
	
	// Double sign slash (5%)
	doubleSignSlash := slasher.CalculateSlashAmount(stake, pos.SlashReasonDoubleSign)
	expected := big.NewInt(5000)
	if doubleSignSlash.Cmp(expected) != 0 {
		t.Errorf("expected slash %s, got %s", expected.String(), doubleSignSlash.String())
	}
	
	// Downtime slash (1%)
	downtimeSlash := slasher.CalculateSlashAmount(stake, pos.SlashReasonDowntime)
	expected = big.NewInt(1000)
	if downtimeSlash.Cmp(expected) != 0 {
		t.Errorf("expected slash %s, got %s", expected.String(), downtimeSlash.String())
	}
}

func TestProposerRotation(t *testing.T) {
	engine := pos.NewEngine(nil)
	vs := pos.NewValidatorSet()
	
	// Add validators
	for i := 1; i <= 5; i++ {
		v := &pos.Validator{
			Address: crypto.GenerateAddress(),
			Stake:   big.NewInt(int64(i * 10000)),
			Active:  true,
		}
		vs.Add(v)
	}
	
	// Check that different heights get different proposers
	proposers := make(map[string]bool)
	for height := uint64(1); height <= 100; height++ {
		proposer := engine.SelectProposer(vs, height)
		proposers[proposer.Address] = true
	}
	
	// Should have multiple different proposers
	if len(proposers) < 2 {
		t.Error("expected multiple proposers in rotation")
	}
}
