package p2p

import (
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"
)

// NodeConfig contains P2P node configuration
type NodeConfig struct {
	ListenAddr    string        `json:"listen_addr"`
	ExternalAddr  string        `json:"external_addr"`
	MaxPeers      int           `json:"max_peers"`
	DialTimeout   time.Duration `json:"dial_timeout"`
	PingInterval  time.Duration `json:"ping_interval"`
	Seeds         []string      `json:"seeds"`
	NetworkID     uint64        `json:"network_id"`
}

// DefaultNodeConfig returns default P2P configuration
func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		ListenAddr:   "0.0.0.0:26656",
		MaxPeers:     50,
		DialTimeout:  10 * time.Second,
		PingInterval: 30 * time.Second,
		NetworkID:    1,
	}
}

// Node represents a P2P network node
type Node struct {
	mu          sync.RWMutex
	config      *NodeConfig
	id          string
	listener    net.Listener
	peers       map[string]*Peer
	running     bool
	stopChan    chan struct{}
	
	// Callbacks
	onPeerConnect    func(*Peer)
	onPeerDisconnect func(*Peer)
	onMessage        func(*Peer, *Message)
}

// Peer represents a connected peer
type Peer struct {
	mu         sync.RWMutex
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	Version    string    `json:"version"`
	NetworkID  uint64    `json:"network_id"`
	Height     uint64    `json:"height"`
	Conn       net.Conn  `json:"-"`
	Connected  time.Time `json:"connected"`
	LastSeen   time.Time `json:"last_seen"`
	Inbound    bool      `json:"inbound"`
	MessagesSent uint64  `json:"messages_sent"`
	MessagesRecv uint64  `json:"messages_recv"`
	BytesSent  uint64    `json:"bytes_sent"`
	BytesRecv  uint64    `json:"bytes_recv"`
}

// Message represents a P2P message
type Message struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
	PeerID    string          `json:"peer_id"`
}

// MessageType identifies the message type
type MessageType uint8

const (
	MsgTypePing MessageType = iota
	MsgTypePong
	MsgTypeHandshake
	MsgTypeBlock
	MsgTypeTransaction
	MsgTypeBlockRequest
	MsgTypeTxRequest
	MsgTypePeers
)

// NewNode creates a new P2P node
func NewNode(config *NodeConfig) (*Node, error) {
	if config == nil {
		config = DefaultNodeConfig()
	}
	
	return &Node{
		config:   config,
		peers:    make(map[string]*Peer),
		stopChan: make(chan struct{}),
	}, nil
}

// Start starts the P2P node
func (n *Node) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if n.running {
		return errors.New("node already running")
	}
	
	listener, err := net.Listen("tcp", n.config.ListenAddr)
	if err != nil {
		return err
	}
	
	n.listener = listener
	n.running = true
	n.stopChan = make(chan struct{})
	
	// Accept incoming connections
	go n.acceptLoop()
	
	// Connect to seeds
	go n.connectToSeeds()
	
	// Start ping loop
	go n.pingLoop()
	
	return nil
}

// Stop stops the P2P node
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	
	if !n.running {
		return nil
	}
	
	close(n.stopChan)
	n.running = false
	
	if n.listener != nil {
		n.listener.Close()
	}
	
	// Disconnect all peers
	for _, peer := range n.peers {
		peer.Disconnect()
	}
	
	return nil
}

// acceptLoop accepts incoming connections
func (n *Node) acceptLoop() {
	for {
		select {
		case <-n.stopChan:
			return
		default:
			conn, err := n.listener.Accept()
			if err != nil {
				continue
			}
			
			go n.handleConnection(conn, true)
		}
	}
}

// handleConnection handles a new connection
func (n *Node) handleConnection(conn net.Conn, inbound bool) {
	peer := &Peer{
		Address:   conn.RemoteAddr().String(),
		Conn:      conn,
		Connected: time.Now(),
		LastSeen:  time.Now(),
		Inbound:   inbound,
	}
	
	// Perform handshake
	if err := n.handshake(peer); err != nil {
		conn.Close()
		return
	}
	
	n.mu.Lock()
	if len(n.peers) >= n.config.MaxPeers {
		n.mu.Unlock()
		conn.Close()
		return
	}
	n.peers[peer.ID] = peer
	n.mu.Unlock()
	
	if n.onPeerConnect != nil {
		n.onPeerConnect(peer)
	}
	
	// Start reading messages
	go n.readLoop(peer)
}

// handshake performs the P2P handshake
func (n *Node) handshake(peer *Peer) error {
	// Send our handshake
	hs := &Handshake{
		Version:   "1.0.0",
		NetworkID: n.config.NetworkID,
		NodeID:    n.id,
		Timestamp: time.Now().Unix(),
	}
	
	if err := n.sendMessage(peer, MsgTypeHandshake, hs); err != nil {
		return err
	}
	
	// Read peer's handshake
	msg, err := n.readMessage(peer)
	if err != nil {
		return err
	}
	
	if msg.Type != MsgTypeHandshake {
		return errors.New("expected handshake message")
	}
	
	var peerHs Handshake
	if err := json.Unmarshal(msg.Payload, &peerHs); err != nil {
		return err
	}
	
	if peerHs.NetworkID != n.config.NetworkID {
		return errors.New("network ID mismatch")
	}
	
	peer.ID = peerHs.NodeID
	peer.Version = peerHs.Version
	peer.NetworkID = peerHs.NetworkID
	
	return nil
}

// Handshake message
type Handshake struct {
	Version   string `json:"version"`
	NetworkID uint64 `json:"network_id"`
	NodeID    string `json:"node_id"`
	Height    uint64 `json:"height"`
	Timestamp int64  `json:"timestamp"`
}

// connectToSeeds connects to seed nodes
func (n *Node) connectToSeeds() {
	for _, seed := range n.config.Seeds {
		go n.Connect(seed)
	}
}

// Connect connects to a peer by address
func (n *Node) Connect(address string) error {
	conn, err := net.DialTimeout("tcp", address, n.config.DialTimeout)
	if err != nil {
		return err
	}
	
	go n.handleConnection(conn, false)
	return nil
}

// pingLoop periodically pings peers
func (n *Node) pingLoop() {
	ticker := time.NewTicker(n.config.PingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			n.mu.RLock()
			peers := make([]*Peer, 0, len(n.peers))
			for _, p := range n.peers {
				peers = append(peers, p)
			}
			n.mu.RUnlock()
			
			for _, peer := range peers {
				n.sendMessage(peer, MsgTypePing, nil)
			}
		}
	}
}

// readLoop reads messages from a peer
func (n *Node) readLoop(peer *Peer) {
	for {
		select {
		case <-n.stopChan:
			return
		default:
			msg, err := n.readMessage(peer)
			if err != nil {
				n.disconnectPeer(peer)
				return
			}
			
			peer.mu.Lock()
			peer.LastSeen = time.Now()
			peer.MessagesRecv++
			peer.mu.Unlock()
			
			n.handleMessage(peer, msg)
		}
	}
}

// handleMessage processes an incoming message
func (n *Node) handleMessage(peer *Peer, msg *Message) {
	switch msg.Type {
	case MsgTypePing:
		n.sendMessage(peer, MsgTypePong, nil)
	case MsgTypePong:
		// Update last seen (already done)
	default:
		if n.onMessage != nil {
			n.onMessage(peer, msg)
		}
	}
}

// sendMessage sends a message to a peer
func (n *Node) sendMessage(peer *Peer, msgType MessageType, payload interface{}) error {
	var payloadBytes json.RawMessage
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	
	msg := &Message{
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now().Unix(),
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	peer.mu.Lock()
	_, err = peer.Conn.Write(append(data, '\n'))
	if err == nil {
		peer.MessagesSent++
		peer.BytesSent += uint64(len(data))
	}
	peer.mu.Unlock()
	
	return err
}

// readMessage reads a message from a peer
func (n *Node) readMessage(peer *Peer) (*Message, error) {
	buf := make([]byte, 1024*1024) // 1MB max
	
	peer.Conn.SetReadDeadline(time.Now().Add(time.Minute))
	num, err := peer.Conn.Read(buf)
	if err != nil {
		return nil, err
	}
	
	peer.mu.Lock()
	peer.BytesRecv += uint64(num)
	peer.mu.Unlock()
	
	var msg Message
	if err := json.Unmarshal(buf[:num], &msg); err != nil {
		return nil, err
	}
	
	msg.PeerID = peer.ID
	return &msg, nil
}

// disconnectPeer removes a peer
func (n *Node) disconnectPeer(peer *Peer) {
	n.mu.Lock()
	delete(n.peers, peer.ID)
	n.mu.Unlock()
	
	peer.Disconnect()
	
	if n.onPeerDisconnect != nil {
		n.onPeerDisconnect(peer)
	}
}

// Disconnect closes the peer connection
func (p *Peer) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.Conn != nil {
		p.Conn.Close()
	}
}

// GetPeers returns all connected peers
func (n *Node) GetPeers() []*Peer {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	peers := make([]*Peer, 0, len(n.peers))
	for _, p := range n.peers {
		peers = append(peers, p)
	}
	return peers
}

// PeerCount returns the number of connected peers
func (n *Node) PeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.peers)
}

// Broadcast sends a message to all peers
func (n *Node) Broadcast(msgType MessageType, payload interface{}) {
	n.mu.RLock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, p := range n.peers {
		peers = append(peers, p)
	}
	n.mu.RUnlock()
	
	for _, peer := range peers {
		go n.sendMessage(peer, msgType, payload)
	}
}

// SetMessageHandler sets the message handler callback
func (n *Node) SetMessageHandler(handler func(*Peer, *Message)) {
	n.onMessage = handler
}

// SetPeerConnectHandler sets the peer connect callback
func (n *Node) SetPeerConnectHandler(handler func(*Peer)) {
	n.onPeerConnect = handler
}

// SetPeerDisconnectHandler sets the peer disconnect callback
func (n *Node) SetPeerDisconnectHandler(handler func(*Peer)) {
	n.onPeerDisconnect = handler
}
