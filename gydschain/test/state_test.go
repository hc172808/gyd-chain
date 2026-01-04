package test

import (
	"math/big"
	"testing"

	"github.com/gydschain/gydschain/internal/state"
)

func TestAccountCreation(t *testing.T) {
	acc := state.NewAccount("gyds1test123")
	
	if acc.Address != "gyds1test123" {
		t.Errorf("expected address gyds1test123, got %s", acc.Address)
	}
	
	if acc.Nonce != 0 {
		t.Errorf("expected nonce 0, got %d", acc.Nonce)
	}
}

func TestAccountBalance(t *testing.T) {
	acc := state.NewAccount("gyds1test123")
	
	// Get balance for non-existent asset
	balance := acc.GetBalance("GYDS")
	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance 0, got %s", balance.String())
	}
	
	// Set balance
	acc.SetBalance("GYDS", big.NewInt(1000))
	balance = acc.GetBalance("GYDS")
	if balance.Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("expected balance 1000, got %s", balance.String())
	}
	
	// Add balance
	acc.AddBalance("GYDS", big.NewInt(500))
	balance = acc.GetBalance("GYDS")
	if balance.Cmp(big.NewInt(1500)) != 0 {
		t.Errorf("expected balance 1500, got %s", balance.String())
	}
	
	// Subtract balance
	err := acc.SubBalance("GYDS", big.NewInt(300))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	balance = acc.GetBalance("GYDS")
	if balance.Cmp(big.NewInt(1200)) != 0 {
		t.Errorf("expected balance 1200, got %s", balance.String())
	}
}

func TestAccountInsufficientBalance(t *testing.T) {
	acc := state.NewAccount("gyds1test123")
	acc.SetBalance("GYDS", big.NewInt(100))
	
	err := acc.SubBalance("GYDS", big.NewInt(200))
	if err == nil {
		t.Error("expected insufficient balance error")
	}
}

func TestAssetCreation(t *testing.T) {
	asset := state.NewAsset(
		"TEST",
		"Test Token",
		18,
		big.NewInt(1000000),
		nil,
		"gyds1creator",
		true,
		true,
	)
	
	if asset.Symbol != "TEST" {
		t.Errorf("expected symbol TEST, got %s", asset.Symbol)
	}
	
	if asset.Decimals != 18 {
		t.Errorf("expected decimals 18, got %d", asset.Decimals)
	}
	
	if !asset.Mintable {
		t.Error("expected mintable to be true")
	}
}

func TestAssetMinting(t *testing.T) {
	maxSupply := big.NewInt(2000000)
	asset := state.NewAsset(
		"TEST",
		"Test Token",
		18,
		big.NewInt(1000000),
		maxSupply,
		"gyds1creator",
		true,
		true,
	)
	
	// Mint tokens
	err := asset.Mint(big.NewInt(500000))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if asset.TotalSupply.Cmp(big.NewInt(1500000)) != 0 {
		t.Errorf("expected supply 1500000, got %s", asset.TotalSupply.String())
	}
	
	// Try to mint beyond max supply
	err = asset.Mint(big.NewInt(600000))
	if err == nil {
		t.Error("expected max supply error")
	}
}

func TestAssetBurning(t *testing.T) {
	asset := state.NewAsset(
		"TEST",
		"Test Token",
		18,
		big.NewInt(1000000),
		nil,
		"gyds1creator",
		true,
		true,
	)
	
	// Burn tokens
	err := asset.Burn(big.NewInt(300000))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if asset.TotalSupply.Cmp(big.NewInt(700000)) != 0 {
		t.Errorf("expected supply 700000, got %s", asset.TotalSupply.String())
	}
	
	// Try to burn more than supply
	err = asset.Burn(big.NewInt(800000))
	if err == nil {
		t.Error("expected insufficient supply error")
	}
}

func TestNonMintableAsset(t *testing.T) {
	asset := state.NewAsset(
		"FIXED",
		"Fixed Supply Token",
		18,
		big.NewInt(1000000),
		nil,
		"gyds1creator",
		false, // Not mintable
		false, // Not burnable
	)
	
	err := asset.Mint(big.NewInt(100))
	if err == nil {
		t.Error("expected non-mintable error")
	}
	
	err = asset.Burn(big.NewInt(100))
	if err == nil {
		t.Error("expected non-burnable error")
	}
}

func TestStateDB(t *testing.T) {
	db := state.NewMemoryStateDB()
	
	// Create account
	acc := state.NewAccount("gyds1test123")
	acc.SetBalance("GYDS", big.NewInt(1000))
	
	// Save account
	err := db.SetAccount(acc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	// Get account
	got, err := db.GetAccount("gyds1test123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	if got.GetBalance("GYDS").Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("expected balance 1000, got %s", got.GetBalance("GYDS").String())
	}
}

func TestMerkleRoot(t *testing.T) {
	db := state.NewMemoryStateDB()
	
	// Add some accounts
	for i := 1; i <= 10; i++ {
		acc := state.NewAccount(state.GenerateTestAddress(i))
		acc.SetBalance("GYDS", big.NewInt(int64(i*1000)))
		db.SetAccount(acc)
	}
	
	// Calculate state root
	root1 := db.StateRoot()
	if len(root1) != 32 {
		t.Errorf("expected 32 byte root, got %d bytes", len(root1))
	}
	
	// Modify state
	acc, _ := db.GetAccount(state.GenerateTestAddress(1))
	acc.AddBalance("GYDS", big.NewInt(100))
	db.SetAccount(acc)
	
	// New root should be different
	root2 := db.StateRoot()
	if string(root1) == string(root2) {
		t.Error("state root should change after modification")
	}
}

func TestStateSnapshot(t *testing.T) {
	db := state.NewMemoryStateDB()
	
	// Add account
	acc := state.NewAccount("gyds1test123")
	acc.SetBalance("GYDS", big.NewInt(1000))
	db.SetAccount(acc)
	
	// Take snapshot
	snapshot := db.Snapshot()
	
	// Modify state
	acc.SetBalance("GYDS", big.NewInt(2000))
	db.SetAccount(acc)
	
	// Revert to snapshot
	db.Revert(snapshot)
	
	// Check balance reverted
	acc, _ = db.GetAccount("gyds1test123")
	if acc.GetBalance("GYDS").Cmp(big.NewInt(1000)) != 0 {
		t.Errorf("expected reverted balance 1000, got %s", acc.GetBalance("GYDS").String())
	}
}
