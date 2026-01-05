package crypto

import (
	"errors"
	"strings"
)

const (
	// AddressPrefix is the prefix for GYDS addresses
	AddressPrefix = "gyds1"
	
	// AddressLength is the length of the address without prefix
	AddressLength = 38
	
	// Bech32Charset is the character set for bech32 encoding
	Bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
)

// DeriveAddress derives an address from a public key
func DeriveAddress(publicKey []byte) string {
	// Hash the public key
	hash := Hash160(publicKey)
	
	// Convert to bech32
	converted := convertBits(hash, 8, 5, true)
	
	// Add checksum
	checksum := bech32Checksum(AddressPrefix, converted)
	combined := append(converted, checksum...)
	
	// Encode
	var address strings.Builder
	address.WriteString(AddressPrefix)
	for _, b := range combined {
		address.WriteByte(Bech32Charset[b])
	}
	
	return address.String()
}

// ValidateAddress checks if an address is valid
func ValidateAddress(address string) error {
	if !strings.HasPrefix(address, AddressPrefix) {
		return errors.New("invalid address prefix")
	}
	
	if len(address) != len(AddressPrefix)+AddressLength {
		return errors.New("invalid address length")
	}
	
	// Decode and verify checksum
	data := address[len(AddressPrefix):]
	decoded := make([]byte, len(data))
	
	for i, c := range data {
		idx := strings.IndexByte(Bech32Charset, byte(c))
		if idx < 0 {
			return errors.New("invalid character in address")
		}
		decoded[i] = byte(idx)
	}
	
	// Verify checksum
	if !verifyBech32Checksum(AddressPrefix, decoded) {
		return errors.New("invalid address checksum")
	}
	
	return nil
}

// IsValidAddress returns true if address is valid
func IsValidAddress(address string) bool {
	return ValidateAddress(address) == nil
}

// AddressFromHash creates an address from a hash
func AddressFromHash(hash []byte) string {
	converted := convertBits(hash, 8, 5, true)
	checksum := bech32Checksum(AddressPrefix, converted)
	combined := append(converted, checksum...)
	
	var address strings.Builder
	address.WriteString(AddressPrefix)
	for _, b := range combined {
		address.WriteByte(Bech32Charset[b])
	}
	
	return address.String()
}

// DecodeAddress decodes an address to its hash
func DecodeAddress(address string) ([]byte, error) {
	if err := ValidateAddress(address); err != nil {
		return nil, err
	}
	
	data := address[len(AddressPrefix):]
	decoded := make([]byte, len(data)-6) // Remove checksum
	
	for i := 0; i < len(decoded); i++ {
		idx := strings.IndexByte(Bech32Charset, data[i])
		decoded[i] = byte(idx)
	}
	
	// Convert from 5-bit to 8-bit
	result, err := convertBits(decoded, 5, 8, false)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// convertBits converts between bit sizes
func convertBits(data []byte, fromBits, toBits int, pad bool) []byte {
	acc := 0
	bits := 0
	var result []byte
	maxv := (1 << toBits) - 1
	
	for _, value := range data {
		acc = (acc << fromBits) | int(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxv))
		}
	}
	
	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxv))
		}
	}
	
	return result
}

// bech32Checksum calculates bech32 checksum
func bech32Checksum(hrp string, data []byte) []byte {
	values := hrpExpand(hrp)
	values = append(values, data...)
	values = append(values, make([]byte, 6)...)
	
	polymod := bech32Polymod(values) ^ 1
	
	checksum := make([]byte, 6)
	for i := 0; i < 6; i++ {
		checksum[i] = byte((polymod >> (5 * (5 - i))) & 31)
	}
	
	return checksum
}

// verifyBech32Checksum verifies bech32 checksum
func verifyBech32Checksum(hrp string, data []byte) bool {
	values := hrpExpand(hrp)
	values = append(values, data...)
	return bech32Polymod(values) == 1
}

// hrpExpand expands the human-readable part
func hrpExpand(hrp string) []byte {
	result := make([]byte, len(hrp)*2+1)
	for i, c := range hrp {
		result[i] = byte(c >> 5)
		result[i+len(hrp)+1] = byte(c & 31)
	}
	result[len(hrp)] = 0
	return result
}

// bech32Polymod calculates the polymod for bech32
func bech32Polymod(values []byte) int {
	gen := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ int(v)
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	
	return chk
}

// GenerateValidatorAddress generates a validator address
func GenerateValidatorAddress(pubKey []byte) string {
	hash := Hash160(pubKey)
	
	// Use different prefix for validators
	prefix := "gydsvaloper1"
	converted := convertBits(hash, 8, 5, true)
	checksum := bech32Checksum(prefix, converted)
	combined := append(converted, checksum...)
	
	var address strings.Builder
	address.WriteString(prefix)
	for _, b := range combined {
		address.WriteByte(Bech32Charset[b])
	}
	
	return address.String()
}

// GenerateContractAddress generates a contract address from deployer and nonce
func GenerateContractAddress(deployer string, nonce uint64) string {
	data := append([]byte(deployer), byte(nonce))
	hash := Hash160(data)
	return AddressFromHash(hash)
}

// ShortAddress returns a shortened address for display
func ShortAddress(address string) string {
	if len(address) <= 16 {
		return address
	}
	return address[:10] + "..." + address[len(address)-6:]
}
