package crypto

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

// Hash256 returns SHA256 hash
func Hash256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Hash256Hex returns hex-encoded SHA256 hash
func Hash256Hex(data []byte) string {
	return hex.EncodeToString(Hash256(data))
}

// DoubleHash256 returns double SHA256 hash (like Bitcoin)
func DoubleHash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

// DoubleHash256Hex returns hex-encoded double SHA256 hash
func DoubleHash256Hex(data []byte) string {
	return hex.EncodeToString(DoubleHash256(data))
}

// Hash512 returns SHA512 hash
func Hash512(data []byte) []byte {
	hash := sha512.Sum512(data)
	return hash[:]
}

// Hash512Hex returns hex-encoded SHA512 hash
func Hash512Hex(data []byte) string {
	return hex.EncodeToString(Hash512(data))
}

// Keccak256 returns Keccak-256 hash (Ethereum-style)
func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Keccak256Hex returns hex-encoded Keccak-256 hash
func Keccak256Hex(data []byte) string {
	return hex.EncodeToString(Keccak256(data))
}

// SHA3_256 returns SHA3-256 hash
func SHA3_256(data []byte) []byte {
	hash := sha3.New256()
	hash.Write(data)
	return hash.Sum(nil)
}

// SHA3_256Hex returns hex-encoded SHA3-256 hash
func SHA3_256Hex(data []byte) string {
	return hex.EncodeToString(SHA3_256(data))
}

// RIPEMD160 returns RIPEMD-160 hash
func RIPEMD160(data []byte) []byte {
	hash := ripemd160.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// Hash160 returns RIPEMD160(SHA256(data)) - Bitcoin address hash
func Hash160(data []byte) []byte {
	sha := sha256.Sum256(data)
	return RIPEMD160(sha[:])
}

// Hash160Hex returns hex-encoded Hash160
func Hash160Hex(data []byte) string {
	return hex.EncodeToString(Hash160(data))
}

// HashMultiple hashes multiple byte slices together
func HashMultiple(data ...[]byte) []byte {
	hash := sha256.New()
	for _, d := range data {
		hash.Write(d)
	}
	return hash.Sum(nil)
}

// HashMultipleHex returns hex-encoded hash of multiple byte slices
func HashMultipleHex(data ...[]byte) string {
	return hex.EncodeToString(HashMultiple(data...))
}

// ComputeMerkleRoot computes merkle root from leaf hashes
func ComputeMerkleRoot(leaves [][]byte) []byte {
	if len(leaves) == 0 {
		return make([]byte, 32)
	}
	
	if len(leaves) == 1 {
		return leaves[0]
	}
	
	// Ensure even number of leaves
	if len(leaves)%2 != 0 {
		leaves = append(leaves, leaves[len(leaves)-1])
	}
	
	var nextLevel [][]byte
	for i := 0; i < len(leaves); i += 2 {
		combined := append(leaves[i], leaves[i+1]...)
		hash := Hash256(combined)
		nextLevel = append(nextLevel, hash)
	}
	
	return ComputeMerkleRoot(nextLevel)
}

// ComputeMerkleRootHex returns hex-encoded merkle root
func ComputeMerkleRootHex(leaves [][]byte) string {
	return hex.EncodeToString(ComputeMerkleRoot(leaves))
}

// VerifyMerkleProof verifies a merkle proof
func VerifyMerkleProof(leaf []byte, proof [][]byte, root []byte, index int) bool {
	current := leaf
	
	for i, sibling := range proof {
		var combined []byte
		if (index>>i)&1 == 0 {
			combined = append(current, sibling...)
		} else {
			combined = append(sibling, current...)
		}
		current = Hash256(combined)
	}
	
	return hex.EncodeToString(current) == hex.EncodeToString(root)
}

// Checksum calculates a 4-byte checksum
func Checksum(data []byte) []byte {
	hash := DoubleHash256(data)
	return hash[:4]
}

// VerifyChecksum verifies a checksum
func VerifyChecksum(data []byte, checksum []byte) bool {
	calculated := Checksum(data)
	if len(calculated) != len(checksum) {
		return false
	}
	
	for i := range calculated {
		if calculated[i] != checksum[i] {
			return false
		}
	}
	
	return true
}

// HMAC256 calculates HMAC-SHA256
func HMAC256(key, data []byte) []byte {
	// Simplified HMAC implementation
	blockSize := 64
	
	if len(key) > blockSize {
		key = Hash256(key)
	}
	
	if len(key) < blockSize {
		padding := make([]byte, blockSize-len(key))
		key = append(key, padding...)
	}
	
	ipad := make([]byte, blockSize)
	opad := make([]byte, blockSize)
	
	for i := range key {
		ipad[i] = key[i] ^ 0x36
		opad[i] = key[i] ^ 0x5c
	}
	
	inner := Hash256(append(ipad, data...))
	return Hash256(append(opad, inner...))
}

// HMAC256Hex returns hex-encoded HMAC-SHA256
func HMAC256Hex(key, data []byte) string {
	return hex.EncodeToString(HMAC256(key, data))
}
