package chain

import (
	"encoding/json"
	"os"
	"time"
)

// GenesisConfig represents the genesis block configuration
type GenesisConfig struct {
	ChainID     string            `json:"chain_id"`
	Timestamp   int64             `json:"timestamp"`
	Validators  []ValidatorConfig `json:"validators"`
	Alloc       []AllocConfig     `json:"alloc"`
	GYDSConfig  TokenConfig       `json:"gyds_config"`
	GYDConfig   TokenConfig       `json:"gyd_config"`
	Params      ChainParams       `json:"params"`
}

// ValidatorConfig represents a genesis validator
type ValidatorConfig struct {
	Address     string `json:"address"`
	PubKey      string `json:"pub_key"`
	Power       uint64 `json:"power"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AllocConfig represents a genesis account allocation
type AllocConfig struct {
	Address     string `json:"address"`
	GYDSBalance uint64 `json:"gyds_balance"`
	GYDBalance  uint64 `json:"gyd_balance"`
	Vesting     *VestingConfig `json:"vesting,omitempty"`
}

// VestingConfig represents token vesting configuration
type VestingConfig struct {
	StartTime    int64  `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	CliffTime    int64  `json:"cliff_time"`
	TotalAmount  uint64 `json:"total_amount"`
	VestedAmount uint64 `json:"vested_amount"`
}

// TokenConfig represents token configuration
type TokenConfig struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    uint8  `json:"decimals"`
	TotalSupply uint64 `json:"total_supply"`
	MaxSupply   uint64 `json:"max_supply"`
	Mintable    bool   `json:"mintable"`
	Burnable    bool   `json:"burnable"`
}

// ChainParams represents chain-wide parameters
type ChainParams struct {
	BlockTime           uint64 `json:"block_time"`
	MaxValidators       uint32 `json:"max_validators"`
	MinStake            uint64 `json:"min_stake"`
	UnbondingTime       uint64 `json:"unbonding_time"`
	SlashingPenalty     uint64 `json:"slashing_penalty"`
	InflationRate       uint64 `json:"inflation_rate"`
	StablecoinReserve   uint64 `json:"stablecoin_reserve"`
	OracleUpdateFreq    uint64 `json:"oracle_update_freq"`
}

// DefaultGenesis returns a default genesis configuration
func DefaultGenesis() *GenesisConfig {
	return &GenesisConfig{
		ChainID:   "gydschain-1",
		Timestamp: time.Now().Unix(),
		Validators: []ValidatorConfig{
			{
				Address:     "gyds1validator000000000000000000000000000001",
				PubKey:      "genesis_validator_pubkey",
				Power:       1000000,
				Name:        "Genesis Validator",
				Description: "Initial network validator",
			},
		},
		Alloc: []AllocConfig{
			{
				Address:     "gyds1foundation00000000000000000000000000001",
				GYDSBalance: 100000000 * 1e8, // 100M GYDS
				GYDBalance:  10000000 * 1e8,  // 10M GYD
			},
			{
				Address:     "gyds1treasury0000000000000000000000000000001",
				GYDSBalance: 50000000 * 1e8,  // 50M GYDS
				GYDBalance:  5000000 * 1e8,   // 5M GYD
			},
		},
		GYDSConfig: TokenConfig{
			Name:        "GYDS Token",
			Symbol:      "GYDS",
			Decimals:    8,
			TotalSupply: 1000000000 * 1e8, // 1B GYDS
			MaxSupply:   2000000000 * 1e8, // 2B GYDS max
			Mintable:    true,
			Burnable:    true,
		},
		GYDConfig: TokenConfig{
			Name:        "GYD Stablecoin",
			Symbol:      "GYD",
			Decimals:    8,
			TotalSupply: 100000000 * 1e8, // 100M GYD
			MaxSupply:   0,               // No max (collateral-backed)
			Mintable:    true,
			Burnable:    true,
		},
		Params: ChainParams{
			BlockTime:         5,
			MaxValidators:     100,
			MinStake:          10000 * 1e8, // 10,000 GYDS
			UnbondingTime:     21 * 24 * 60 * 60, // 21 days
			SlashingPenalty:   5, // 5%
			InflationRate:     5, // 5% annual
			StablecoinReserve: 150, // 150% collateralization
			OracleUpdateFreq:  60, // 60 seconds
		},
	}
}

// LoadGenesis loads genesis config from a file
func LoadGenesis(path string) (*GenesisConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var genesis GenesisConfig
	if err := json.Unmarshal(data, &genesis); err != nil {
		return nil, err
	}
	
	return &genesis, nil
}

// Save writes the genesis config to a file
func (g *GenesisConfig) Save(path string) error {
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}

// ToBlock converts genesis config to the genesis block
func (g *GenesisConfig) ToBlock() *Block {
	header := &Header{
		Version:    1,
		Height:     0,
		Timestamp:  g.Timestamp,
		ParentHash: "",
		TxRoot:     "0x0000000000000000000000000000000000000000000000000000000000000000",
		StateRoot:  "0x0000000000000000000000000000000000000000000000000000000000000000",
		Difficulty: 1,
		GasLimit:   10000000,
	}
	
	return &Block{
		Header:       header,
		Transactions: nil,
		Validator:    "genesis",
	}
}

// Validate checks the genesis configuration
func (g *GenesisConfig) Validate() error {
	if g.ChainID == "" {
		return ErrInvalidChainID
	}
	
	if len(g.Validators) == 0 {
		return ErrNoValidators
	}
	
	if g.GYDSConfig.TotalSupply == 0 {
		return ErrInvalidTokenConfig
	}
	
	return nil
}

// Errors
var (
	ErrInvalidChainID    = ErrInvalidBlock
	ErrNoValidators      = ErrInvalidBlock
	ErrInvalidTokenConfig = ErrInvalidBlock
)
