package state

import (
	"encoding/json"
	"time"
)

// AssetType represents the type of asset
type AssetType uint8

const (
	AssetTypeFungible AssetType = iota
	AssetTypeNFT
	AssetTypeStablecoin
)

// Asset represents a token or NFT
type Asset struct {
	ID          string    `json:"id"`
	Type        AssetType `json:"type"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"`
	Decimals    uint8     `json:"decimals"`
	TotalSupply uint64    `json:"total_supply"`
	MaxSupply   uint64    `json:"max_supply"`
	Owner       string    `json:"owner"`
	Mintable    bool      `json:"mintable"`
	Burnable    bool      `json:"burnable"`
	Pausable    bool      `json:"pausable"`
	Paused      bool      `json:"paused"`
	Metadata    *AssetMetadata `json:"metadata,omitempty"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
}

// AssetMetadata contains additional asset information
type AssetMetadata struct {
	Description string            `json:"description,omitempty"`
	Image       string            `json:"image,omitempty"`
	ExternalURL string            `json:"external_url,omitempty"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// NewFungibleAsset creates a new fungible token
func NewFungibleAsset(id, name, symbol string, decimals uint8, owner string) *Asset {
	return &Asset{
		ID:        id,
		Type:      AssetTypeFungible,
		Name:      name,
		Symbol:    symbol,
		Decimals:  decimals,
		Owner:     owner,
		Mintable:  true,
		Burnable:  true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
}

// NewNFT creates a new NFT
func NewNFT(id, name, owner string, metadata *AssetMetadata) *Asset {
	return &Asset{
		ID:          id,
		Type:        AssetTypeNFT,
		Name:        name,
		Symbol:      "NFT",
		Decimals:    0,
		TotalSupply: 1,
		MaxSupply:   1,
		Owner:       owner,
		Mintable:    false,
		Burnable:    true,
		Metadata:    metadata,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
}

// NewStablecoin creates a new stablecoin asset
func NewStablecoin(id, name, symbol string, owner string) *Asset {
	return &Asset{
		ID:        id,
		Type:      AssetTypeStablecoin,
		Name:      name,
		Symbol:    symbol,
		Decimals:  8,
		Owner:     owner,
		Mintable:  true,
		Burnable:  true,
		Pausable:  true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
}

// Mint increases the total supply
func (a *Asset) Mint(amount uint64) error {
	if !a.Mintable {
		return ErrNotMintable
	}
	
	if a.Paused {
		return ErrAssetPaused
	}
	
	if a.MaxSupply > 0 && a.TotalSupply+amount > a.MaxSupply {
		return ErrExceedsMaxSupply
	}
	
	a.TotalSupply += amount
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// Burn decreases the total supply
func (a *Asset) Burn(amount uint64) error {
	if !a.Burnable {
		return ErrNotBurnable
	}
	
	if a.Paused {
		return ErrAssetPaused
	}
	
	if a.TotalSupply < amount {
		return ErrInsufficientSupply
	}
	
	a.TotalSupply -= amount
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// Pause pauses the asset
func (a *Asset) Pause() error {
	if !a.Pausable {
		return ErrNotPausable
	}
	
	a.Paused = true
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// Unpause unpauses the asset
func (a *Asset) Unpause() error {
	if !a.Pausable {
		return ErrNotPausable
	}
	
	a.Paused = false
	a.UpdatedAt = time.Now().Unix()
	return nil
}

// TransferOwnership transfers asset ownership
func (a *Asset) TransferOwnership(newOwner string) {
	a.Owner = newOwner
	a.UpdatedAt = time.Now().Unix()
}

// SetMetadata updates asset metadata
func (a *Asset) SetMetadata(metadata *AssetMetadata) {
	a.Metadata = metadata
	a.UpdatedAt = time.Now().Unix()
}

// IsFungible returns true if asset is fungible
func (a *Asset) IsFungible() bool {
	return a.Type == AssetTypeFungible || a.Type == AssetTypeStablecoin
}

// IsNFT returns true if asset is an NFT
func (a *Asset) IsNFT() bool {
	return a.Type == AssetTypeNFT
}

// IsStablecoin returns true if asset is a stablecoin
func (a *Asset) IsStablecoin() bool {
	return a.Type == AssetTypeStablecoin
}

// Copy creates a deep copy of the asset
func (a *Asset) Copy() *Asset {
	copy := *a
	if a.Metadata != nil {
		metadata := *a.Metadata
		if a.Metadata.Properties != nil {
			metadata.Properties = make(map[string]string)
			for k, v := range a.Metadata.Properties {
				metadata.Properties[k] = v
			}
		}
		copy.Metadata = &metadata
	}
	return &copy
}

// Serialize converts asset to bytes
func (a *Asset) Serialize() ([]byte, error) {
	return json.Marshal(a)
}

// DeserializeAsset creates an asset from bytes
func DeserializeAsset(data []byte) (*Asset, error) {
	var asset Asset
	if err := json.Unmarshal(data, &asset); err != nil {
		return nil, err
	}
	return &asset, nil
}

// Asset errors
var (
	ErrNotMintable       = &AssetError{"asset is not mintable"}
	ErrNotBurnable       = &AssetError{"asset is not burnable"}
	ErrNotPausable       = &AssetError{"asset is not pausable"}
	ErrAssetPaused       = &AssetError{"asset is paused"}
	ErrExceedsMaxSupply  = &AssetError{"exceeds max supply"}
	ErrInsufficientSupply = &AssetError{"insufficient supply"}
)

type AssetError struct {
	msg string
}

func (e *AssetError) Error() string {
	return e.msg
}

// StablecoinOracle manages stablecoin price feeds
type StablecoinOracle struct {
	AssetID     string  `json:"asset_id"`
	PegCurrency string  `json:"peg_currency"`
	Price       float64 `json:"price"`
	LastUpdate  int64   `json:"last_update"`
	Sources     []string `json:"sources"`
}

// NewStablecoinOracle creates a new oracle
func NewStablecoinOracle(assetID, pegCurrency string) *StablecoinOracle {
	return &StablecoinOracle{
		AssetID:     assetID,
		PegCurrency: pegCurrency,
		Price:       1.0,
		LastUpdate:  time.Now().Unix(),
		Sources:     make([]string, 0),
	}
}

// UpdatePrice updates the oracle price
func (o *StablecoinOracle) UpdatePrice(price float64) {
	o.Price = price
	o.LastUpdate = time.Now().Unix()
}

// IsStale returns true if the price is stale
func (o *StablecoinOracle) IsStale(maxAge int64) bool {
	return time.Now().Unix()-o.LastUpdate > maxAge
}
