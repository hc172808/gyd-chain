package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// KeyType represents the type of cryptographic key
type KeyType uint8

const (
	KeyTypeEd25519 KeyType = iota
	KeyTypeSecp256k1
)

// KeyPair represents a cryptographic key pair
type KeyPair struct {
	Type       KeyType
	PublicKey  []byte
	PrivateKey []byte
}

// NewKeyPair generates a new Ed25519 key pair
func NewKeyPair() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	
	return &KeyPair{
		Type:       KeyTypeEd25519,
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// NewKeyPairFromSeed generates a key pair from a seed
func NewKeyPairFromSeed(seed []byte) (*KeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, errors.New("invalid seed size")
	}
	
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	
	return &KeyPair{
		Type:       KeyTypeEd25519,
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// NewKeyPairFromPrivateKey creates a key pair from an existing private key
func NewKeyPairFromPrivateKey(privateKey []byte) (*KeyPair, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	
	priv := ed25519.PrivateKey(privateKey)
	pub := priv.Public().(ed25519.PublicKey)
	
	return &KeyPair{
		Type:       KeyTypeEd25519,
		PublicKey:  pub,
		PrivateKey: priv,
	}, nil
}

// Sign signs a message with the private key
func (kp *KeyPair) Sign(message []byte) ([]byte, error) {
	if kp.PrivateKey == nil {
		return nil, errors.New("private key not available")
	}
	
	return ed25519.Sign(kp.PrivateKey, message), nil
}

// Verify verifies a signature
func (kp *KeyPair) Verify(message, signature []byte) bool {
	return ed25519.Verify(kp.PublicKey, message, signature)
}

// PublicKeyHex returns the hex-encoded public key
func (kp *KeyPair) PublicKeyHex() string {
	return hex.EncodeToString(kp.PublicKey)
}

// PrivateKeyHex returns the hex-encoded private key
func (kp *KeyPair) PrivateKeyHex() string {
	return hex.EncodeToString(kp.PrivateKey)
}

// Address returns the address derived from the public key
func (kp *KeyPair) Address() string {
	return DeriveAddress(kp.PublicKey)
}

// Seed returns the seed portion of the private key
func (kp *KeyPair) Seed() []byte {
	if len(kp.PrivateKey) < ed25519.SeedSize {
		return nil
	}
	return kp.PrivateKey[:ed25519.SeedSize]
}

// VerifySignature verifies a signature given a public key, message, and signature
func VerifySignature(publicKey, message, signature []byte) bool {
	if len(publicKey) != ed25519.PublicKeySize {
		return false
	}
	
	if len(signature) != ed25519.SignatureSize {
		return false
	}
	
	return ed25519.Verify(publicKey, message, signature)
}

// ParsePublicKey parses a hex-encoded public key
func ParsePublicKey(hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	
	if len(key) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key length")
	}
	
	return key, nil
}

// ParsePrivateKey parses a hex-encoded private key
func ParsePrivateKey(hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}
	
	if len(key) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key length")
	}
	
	return key, nil
}

// GenerateMnemonic generates a random mnemonic phrase (simplified)
func GenerateMnemonic() (string, error) {
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		return "", err
	}
	
	// Simplified: in production, use BIP39
	return hex.EncodeToString(entropy), nil
}

// MnemonicToSeed converts a mnemonic to a seed (simplified)
func MnemonicToSeed(mnemonic, password string) []byte {
	// Simplified: in production, use BIP39 PBKDF2
	data := []byte(mnemonic + password)
	return Hash256(data)
}

// Wallet represents a simple wallet
type Wallet struct {
	KeyPair *KeyPair
	Name    string
}

// NewWallet creates a new wallet with a new key pair
func NewWallet(name string) (*Wallet, error) {
	kp, err := NewKeyPair()
	if err != nil {
		return nil, err
	}
	
	return &Wallet{
		KeyPair: kp,
		Name:    name,
	}, nil
}

// NewWalletFromMnemonic creates a wallet from a mnemonic
func NewWalletFromMnemonic(name, mnemonic, password string) (*Wallet, error) {
	seed := MnemonicToSeed(mnemonic, password)
	
	kp, err := NewKeyPairFromSeed(seed[:ed25519.SeedSize])
	if err != nil {
		return nil, err
	}
	
	return &Wallet{
		KeyPair: kp,
		Name:    name,
	}, nil
}

// Address returns the wallet address
func (w *Wallet) Address() string {
	return w.KeyPair.Address()
}

// Sign signs a message with the wallet's private key
func (w *Wallet) Sign(message []byte) ([]byte, error) {
	return w.KeyPair.Sign(message)
}

// Verify verifies a signature
func (w *Wallet) Verify(message, signature []byte) bool {
	return w.KeyPair.Verify(message, signature)
}
