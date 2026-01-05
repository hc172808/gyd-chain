package state

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// MerkleTree represents a Merkle tree for state verification
type MerkleTree struct {
	Root   *MerkleNode
	Leaves []*MerkleNode
}

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash   []byte
	Left   *MerkleNode
	Right  *MerkleNode
	Parent *MerkleNode
	Data   []byte
}

// NewMerkleTree creates a new Merkle tree from data
func NewMerkleTree(data [][]byte) *MerkleTree {
	if len(data) == 0 {
		return &MerkleTree{}
	}
	
	// Create leaf nodes
	leaves := make([]*MerkleNode, len(data))
	for i, d := range data {
		hash := sha256.Sum256(d)
		leaves[i] = &MerkleNode{
			Hash: hash[:],
			Data: d,
		}
	}
	
	// Build tree
	root := buildTree(leaves)
	
	return &MerkleTree{
		Root:   root,
		Leaves: leaves,
	}
}

// buildTree recursively builds the Merkle tree
func buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 0 {
		return nil
	}
	
	if len(nodes) == 1 {
		return nodes[0]
	}
	
	// Ensure even number of nodes
	if len(nodes)%2 != 0 {
		nodes = append(nodes, nodes[len(nodes)-1])
	}
	
	var parents []*MerkleNode
	for i := 0; i < len(nodes); i += 2 {
		parent := combineNodes(nodes[i], nodes[i+1])
		parents = append(parents, parent)
	}
	
	return buildTree(parents)
}

// combineNodes creates a parent node from two children
func combineNodes(left, right *MerkleNode) *MerkleNode {
	combined := append(left.Hash, right.Hash...)
	hash := sha256.Sum256(combined)
	
	parent := &MerkleNode{
		Hash:  hash[:],
		Left:  left,
		Right: right,
	}
	
	left.Parent = parent
	right.Parent = parent
	
	return parent
}

// RootHash returns the root hash of the tree
func (t *MerkleTree) RootHash() []byte {
	if t.Root == nil {
		return make([]byte, 32)
	}
	return t.Root.Hash
}

// RootHashHex returns the hex-encoded root hash
func (t *MerkleTree) RootHashHex() string {
	return hex.EncodeToString(t.RootHash())
}

// GetProof generates a Merkle proof for a leaf
func (t *MerkleTree) GetProof(index int) [][]byte {
	if index < 0 || index >= len(t.Leaves) {
		return nil
	}
	
	var proof [][]byte
	node := t.Leaves[index]
	
	for node.Parent != nil {
		parent := node.Parent
		if parent.Left == node && parent.Right != nil {
			proof = append(proof, parent.Right.Hash)
		} else if parent.Left != nil {
			proof = append(proof, parent.Left.Hash)
		}
		node = parent
	}
	
	return proof
}

// VerifyProof verifies a Merkle proof
func VerifyProof(data []byte, proof [][]byte, root []byte, index int) bool {
	hash := sha256.Sum256(data)
	current := hash[:]
	
	for i, sibling := range proof {
		var combined []byte
		if (index>>i)&1 == 0 {
			combined = append(current, sibling...)
		} else {
			combined = append(sibling, current...)
		}
		hash := sha256.Sum256(combined)
		current = hash[:]
	}
	
	return hex.EncodeToString(current) == hex.EncodeToString(root)
}

// CalculateMerkleRoot calculates the merkle root from raw data
func CalculateMerkleRoot(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// StateProof represents a state inclusion proof
type StateProof struct {
	Key     string   `json:"key"`
	Value   []byte   `json:"value"`
	Proof   [][]byte `json:"proof"`
	Root    string   `json:"root"`
	Height  uint64   `json:"height"`
}

// AccountStateProof represents proof for account state
type AccountStateProof struct {
	Address    string        `json:"address"`
	Account    *Account      `json:"account"`
	Proof      *StateProof   `json:"proof"`
}

// PatriciaTrie represents a Patricia Merkle Trie for efficient state storage
type PatriciaTrie struct {
	root *TrieNode
}

// TrieNode represents a node in the Patricia Trie
type TrieNode struct {
	Key      []byte
	Value    []byte
	Hash     []byte
	Children map[byte]*TrieNode
}

// NewPatriciaTrie creates a new Patricia Trie
func NewPatriciaTrie() *PatriciaTrie {
	return &PatriciaTrie{
		root: &TrieNode{
			Children: make(map[byte]*TrieNode),
		},
	}
}

// Insert adds a key-value pair to the trie
func (t *PatriciaTrie) Insert(key, value []byte) {
	node := t.root
	
	for _, b := range key {
		if node.Children[b] == nil {
			node.Children[b] = &TrieNode{
				Children: make(map[byte]*TrieNode),
			}
		}
		node = node.Children[b]
	}
	
	node.Key = key
	node.Value = value
	t.updateHashes(t.root)
}

// Get retrieves a value by key
func (t *PatriciaTrie) Get(key []byte) []byte {
	node := t.root
	
	for _, b := range key {
		if node.Children[b] == nil {
			return nil
		}
		node = node.Children[b]
	}
	
	return node.Value
}

// Delete removes a key from the trie
func (t *PatriciaTrie) Delete(key []byte) bool {
	return t.deleteRecursive(t.root, key, 0)
}

func (t *PatriciaTrie) deleteRecursive(node *TrieNode, key []byte, depth int) bool {
	if depth == len(key) {
		if node.Value == nil {
			return false
		}
		node.Value = nil
		return len(node.Children) == 0
	}
	
	b := key[depth]
	child := node.Children[b]
	if child == nil {
		return false
	}
	
	shouldDelete := t.deleteRecursive(child, key, depth+1)
	if shouldDelete {
		delete(node.Children, b)
		return len(node.Children) == 0 && node.Value == nil
	}
	
	return false
}

// RootHash returns the root hash of the trie
func (t *PatriciaTrie) RootHash() []byte {
	if t.root == nil {
		return make([]byte, 32)
	}
	return t.root.Hash
}

// updateHashes updates hashes from a node to root
func (t *PatriciaTrie) updateHashes(node *TrieNode) {
	if node == nil {
		return
	}
	
	// Collect child hashes
	var childHashes [][]byte
	
	// Sort keys for deterministic ordering
	keys := make([]byte, 0, len(node.Children))
	for k := range node.Children {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	
	for _, k := range keys {
		t.updateHashes(node.Children[k])
		childHashes = append(childHashes, node.Children[k].Hash)
	}
	
	// Calculate node hash
	var data []byte
	data = append(data, node.Key...)
	data = append(data, node.Value...)
	for _, h := range childHashes {
		data = append(data, h...)
	}
	
	hash := sha256.Sum256(data)
	node.Hash = hash[:]
}
