package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// AdminServer manages node registrations and VPN configuration
type AdminServer struct {
	mu           sync.RWMutex
	port         int
	registryFile string
	vpnConfigDir string
	registry     *NodeRegistry
}

// NodeRegistry tracks all registered nodes
type NodeRegistry struct {
	Pending  []NodeInfo `json:"pending"`
	Approved []NodeInfo `json:"approved"`
	Rejected []NodeInfo `json:"rejected"`
}

// NodeInfo represents a registered node
type NodeInfo struct {
	NodeID           string    `json:"node_id"`
	Hostname         string    `json:"hostname"`
	PublicIP         string    `json:"public_ip"`
	WireGuardPubKey  string    `json:"wireguard_public_key"`
	RegisteredAt     time.Time `json:"registered_at"`
	ApprovedAt       time.Time `json:"approved_at,omitempty"`
	Status           string    `json:"status"`
	Type             string    `json:"type"`
	VPNAddress       string    `json:"vpn_address,omitempty"`
	LastSeen         time.Time `json:"last_seen,omitempty"`
	SyncHeight       uint64    `json:"sync_height,omitempty"`
}

func main() {
	port := flag.Int("port", 9000, "Admin API port")
	registryFile := flag.String("registry", "/opt/gydschain/config/node_registry.json", "Node registry file")
	vpnConfigDir := flag.String("vpn-dir", "/etc/wireguard", "WireGuard config directory")
	flag.Parse()

	server := &AdminServer{
		port:         *port,
		registryFile: *registryFile,
		vpnConfigDir: *vpnConfigDir,
	}

	// Load existing registry
	if err := server.loadRegistry(); err != nil {
		log.Printf("Creating new registry: %v", err)
		server.registry = &NodeRegistry{
			Pending:  []NodeInfo{},
			Approved: []NodeInfo{},
			Rejected: []NodeInfo{},
		}
		server.saveRegistry()
	}

	// Setup routes
	http.HandleFunc("/nodes/register", server.handleRegister)
	http.HandleFunc("/nodes/pending", server.handleGetPending)
	http.HandleFunc("/nodes/approved", server.handleGetApproved)
	http.HandleFunc("/nodes/approve/", server.handleApprove)
	http.HandleFunc("/nodes/reject/", server.handleReject)
	http.HandleFunc("/nodes/remove/", server.handleRemove)
	http.HandleFunc("/nodes/", server.handleGetNodeConfig)
	http.HandleFunc("/system/update", server.handleSystemUpdate)
	http.HandleFunc("/system/rebuild", server.handleRebuildFrontend)
	http.HandleFunc("/system/status", server.handleSystemStatus)
	http.HandleFunc("/health", server.handleHealth)

	fmt.Printf("🔧 Admin API Server starting on port %d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func (s *AdminServer) loadRegistry() error {
	data, err := ioutil.ReadFile(s.registryFile)
	if err != nil {
		return err
	}

	s.registry = &NodeRegistry{}
	return json.Unmarshal(data, s.registry)
}

func (s *AdminServer) saveRegistry() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.registry, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.registryFile, data, 0644)
}

// Handle node registration requests
func (s *AdminServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var node NodeInfo
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	node.Status = "pending"
	node.RegisteredAt = time.Now()

	s.mu.Lock()
	// Check if node already exists
	for _, existing := range s.registry.Pending {
		if existing.NodeID == node.NodeID {
			s.mu.Unlock()
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "success",
				"message": "Node already registered, pending approval",
			})
			return
		}
	}
	s.registry.Pending = append(s.registry.Pending, node)
	s.mu.Unlock()

	s.saveRegistry()

	log.Printf("New node registered: %s (%s)", node.NodeID[:16], node.Hostname)

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Node registered, pending approval",
		"node_id": node.NodeID,
	})
}

// Get pending nodes
func (s *AdminServer) handleGetPending(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	json.NewEncoder(w).Encode(s.registry.Pending)
}

// Get approved nodes
func (s *AdminServer) handleGetApproved(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	json.NewEncoder(w).Encode(s.registry.Approved)
}

// Approve a node
func (s *AdminServer) handleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Path[len("/nodes/approve/"):]
	if nodeID == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	var approvedNode *NodeInfo
	var newPending []NodeInfo

	for _, node := range s.registry.Pending {
		if node.NodeID == nodeID {
			node.Status = "approved"
			node.ApprovedAt = time.Now()
			node.VPNAddress = s.allocateVPNAddress()
			approvedNode = &node
			s.registry.Approved = append(s.registry.Approved, node)
		} else {
			newPending = append(newPending, node)
		}
	}
	s.registry.Pending = newPending
	s.mu.Unlock()

	if approvedNode == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Generate VPN config for the node
	s.generateVPNConfig(approvedNode)
	s.saveRegistry()

	log.Printf("Node approved: %s (%s)", approvedNode.NodeID[:16], approvedNode.Hostname)

	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"message":     "Node approved and VPN configured",
		"vpn_address": approvedNode.VPNAddress,
	})
}

// Reject a node
func (s *AdminServer) handleReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Path[len("/nodes/reject/"):]
	if nodeID == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	var rejectedNode *NodeInfo
	var newPending []NodeInfo

	for _, node := range s.registry.Pending {
		if node.NodeID == nodeID {
			node.Status = "rejected"
			rejectedNode = &node
			s.registry.Rejected = append(s.registry.Rejected, node)
		} else {
			newPending = append(newPending, node)
		}
	}
	s.registry.Pending = newPending
	s.mu.Unlock()

	if rejectedNode == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	s.saveRegistry()

	log.Printf("Node rejected: %s", rejectedNode.NodeID[:16])

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Node rejected",
	})
}

// Remove an approved node
func (s *AdminServer) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Path[len("/nodes/remove/"):]
	if nodeID == "" {
		http.Error(w, "Node ID required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	var removedNode *NodeInfo
	var newApproved []NodeInfo

	for _, node := range s.registry.Approved {
		if node.NodeID == nodeID {
			removedNode = &node
		} else {
			newApproved = append(newApproved, node)
		}
	}
	s.registry.Approved = newApproved
	s.mu.Unlock()

	if removedNode == nil {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Remove from VPN config
	s.removeFromVPN(removedNode)
	s.saveRegistry()

	log.Printf("Node removed: %s", removedNode.NodeID[:16])

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Node removed from network",
	})
}

// Get node config (for lite nodes to retrieve their VPN config)
func (s *AdminServer) handleGetNodeConfig(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Path[len("/nodes/"):]
	if len(nodeID) > 6 && nodeID[len(nodeID)-7:] == "/config" {
		nodeID = nodeID[:len(nodeID)-7]
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, node := range s.registry.Approved {
		if node.NodeID == nodeID {
			// Generate VPN client config
			vpnConfig := s.generateClientVPNConfig(&node)
			bootstrapNodes := s.getBootstrapNodes()

			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":          "approved",
				"vpn_config":      vpnConfig,
				"bootstrap_nodes": bootstrapNodes,
				"vpn_address":     node.VPNAddress,
			})
			return
		}
	}

	for _, node := range s.registry.Pending {
		if node.NodeID == nodeID {
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "pending",
				"message": "Node awaiting approval",
			})
			return
		}
	}

	http.Error(w, "Node not found", http.StatusNotFound)
}

// System update - pull from GitHub and rebuild
func (s *AdminServer) handleSystemUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		log.Println("Starting system update from GitHub...")

		// Run update script
		cmd := exec.Command("bash", "/opt/gydschain/scripts/setup-ubuntu.sh", "--update")
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Update failed: %v\nOutput: %s", err, output)
		} else {
			log.Printf("Update completed successfully\nOutput: %s", output)
		}
	}()

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Update started in background",
	})
}

// Rebuild frontend only
func (s *AdminServer) handleRebuildFrontend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		log.Println("Rebuilding frontend...")

		// Pull latest and rebuild
		cmds := [][]string{
			{"git", "-C", "/opt/gydschain", "pull", "origin", "main"},
			{"npm", "--prefix", "/opt/gydschain", "install"},
			{"npm", "--prefix", "/opt/gydschain", "run", "build"},
		}

		for _, args := range cmds {
			cmd := exec.Command(args[0], args[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("Command failed: %v\nOutput: %s", err, output)
				return
			}
		}

		// Copy to web directory
		exec.Command("cp", "-r", "/opt/gydschain/dist/.", "/var/www/gydschain/").Run()
		exec.Command("chown", "-R", "www-data:www-data", "/var/www/gydschain").Run()

		log.Println("Frontend rebuild completed")
	}()

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Frontend rebuild started",
	})
}

// Get system status
func (s *AdminServer) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check service statuses
	services := []string{"gydschain-node", "gydschain-indexer", "nginx"}
	serviceStatus := make(map[string]string)

	for _, service := range services {
		cmd := exec.Command("systemctl", "is-active", service)
		output, _ := cmd.Output()
		serviceStatus[service] = string(output)
	}

	status := map[string]interface{}{
		"pending_nodes":  len(s.registry.Pending),
		"approved_nodes": len(s.registry.Approved),
		"rejected_nodes": len(s.registry.Rejected),
		"services":       serviceStatus,
		"uptime":         getUptime(),
	}

	json.NewEncoder(w).Encode(status)
}

func (s *AdminServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Helper functions
func (s *AdminServer) allocateVPNAddress() string {
	// Allocate next available VPN address
	baseIP := "10.100.0."
	nextID := len(s.registry.Approved) + 2 // Start from .2, .1 is server
	return fmt.Sprintf("%s%d/24", baseIP, nextID)
}

func (s *AdminServer) generateVPNConfig(node *NodeInfo) {
	// Add peer to WireGuard server config
	peerConfig := fmt.Sprintf(`
# Node: %s (%s)
[Peer]
PublicKey = %s
AllowedIPs = %s
`, node.NodeID[:16], node.Hostname, node.WireGuardPubKey, node.VPNAddress)

	// Append to wg0.conf
	f, err := os.OpenFile(s.vpnConfigDir+"/wg0.conf", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Error opening VPN config: %v", err)
		return
	}
	defer f.Close()

	f.WriteString(peerConfig)

	// Reload WireGuard
	exec.Command("wg", "syncconf", "wg0", s.vpnConfigDir+"/wg0.conf").Run()
}

func (s *AdminServer) generateClientVPNConfig(node *NodeInfo) string {
	// Read server public key
	serverPubKey, _ := ioutil.ReadFile(s.vpnConfigDir + "/server_public.key")

	return fmt.Sprintf(`[Interface]
PrivateKey = <YOUR_PRIVATE_KEY>
Address = %s

[Peer]
PublicKey = %s
Endpoint = <SERVER_IP>:51820
AllowedIPs = 10.100.0.0/24
PersistentKeepalive = 25
`, node.VPNAddress, string(serverPubKey))
}

func (s *AdminServer) getBootstrapNodes() []map[string]string {
	nodes := []map[string]string{}

	for _, node := range s.registry.Approved {
		if node.Type == "fullnode" || node.Type == "validator" {
			nodes = append(nodes, map[string]string{
				"address":   node.VPNAddress[:len(node.VPNAddress)-3] + ":30303",
				"node_id":   node.NodeID,
				"public_ip": node.PublicIP,
			})
		}
	}

	return nodes
}

func (s *AdminServer) removeFromVPN(node *NodeInfo) {
	// Remove peer from WireGuard (would need to rewrite config)
	log.Printf("Removing node %s from VPN", node.NodeID[:16])
	// Reload WireGuard
	exec.Command("wg", "syncconf", "wg0", s.vpnConfigDir+"/wg0.conf").Run()
}

func getUptime() string {
	data, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}
	var uptime float64
	fmt.Sscanf(string(data), "%f", &uptime)
	hours := int(uptime) / 3600
	mins := (int(uptime) % 3600) / 60
	return fmt.Sprintf("%dh %dm", hours, mins)
}
