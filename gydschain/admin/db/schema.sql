-- GYDS Chain Admin Database Schema
-- PostgreSQL 15+

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Admin Users
-- ============================================
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    role VARCHAR(20) DEFAULT 'admin',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

-- ============================================
-- Node Registry
-- ============================================
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id VARCHAR(64) UNIQUE NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    public_ip INET NOT NULL,
    wireguard_public_key VARCHAR(44),
    vpn_address INET,
    node_type VARCHAR(20) NOT NULL DEFAULT 'litenode',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    registered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID REFERENCES admin_users(id),
    last_seen TIMESTAMP WITH TIME ZONE,
    sync_height BIGINT DEFAULT 0,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_nodes_status ON nodes(status);
CREATE INDEX idx_nodes_type ON nodes(node_type);
CREATE INDEX idx_nodes_last_seen ON nodes(last_seen);

-- ============================================
-- Node Heartbeats
-- ============================================
CREATE TABLE IF NOT EXISTS node_heartbeats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id VARCHAR(64) REFERENCES nodes(node_id) ON DELETE CASCADE,
    sync_height BIGINT,
    peer_count INT,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    disk_usage DECIMAL(5,2),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_heartbeats_node_id ON node_heartbeats(node_id);
CREATE INDEX idx_heartbeats_timestamp ON node_heartbeats(timestamp);

-- Partition by month for performance
-- CREATE TABLE node_heartbeats_y2024m01 PARTITION OF node_heartbeats
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- ============================================
-- VPN Configuration
-- ============================================
CREATE TABLE IF NOT EXISTS vpn_peers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    node_id VARCHAR(64) REFERENCES nodes(node_id) ON DELETE CASCADE,
    public_key VARCHAR(44) NOT NULL,
    private_key_encrypted BYTEA,
    allowed_ips INET[] NOT NULL,
    endpoint INET,
    last_handshake TIMESTAMP WITH TIME ZONE,
    rx_bytes BIGINT DEFAULT 0,
    tx_bytes BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_vpn_peers_node_id ON vpn_peers(node_id);

-- ============================================
-- System Logs
-- ============================================
CREATE TABLE IF NOT EXISTS system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(10) NOT NULL,
    source VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_logs_level ON system_logs(level);
CREATE INDEX idx_logs_source ON system_logs(source);
CREATE INDEX idx_logs_timestamp ON system_logs(timestamp);

-- ============================================
-- Deployment History
-- ============================================
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deploy_type VARCHAR(20) NOT NULL, -- 'full', 'frontend', 'backend'
    git_commit VARCHAR(40),
    git_branch VARCHAR(100),
    status VARCHAR(20) NOT NULL, -- 'started', 'success', 'failed'
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    triggered_by UUID REFERENCES admin_users(id),
    logs TEXT,
    error_message TEXT
);

CREATE INDEX idx_deployments_status ON deployments(status);

-- ============================================
-- Security Events
-- ============================================
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(10) NOT NULL, -- 'info', 'warning', 'critical'
    source_ip INET,
    target VARCHAR(255),
    description TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_security_events_type ON security_events(event_type);
CREATE INDEX idx_security_events_severity ON security_events(severity);
CREATE INDEX idx_security_events_timestamp ON security_events(timestamp);

-- ============================================
-- Settings
-- ============================================
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_by UUID REFERENCES admin_users(id)
);

-- Insert default settings
INSERT INTO settings (key, value, description) VALUES
    ('vpn_network', '"10.100.0.0/24"', 'VPN network CIDR'),
    ('vpn_server_address', '"10.100.0.1"', 'VPN server address'),
    ('max_pending_nodes', '100', 'Maximum pending node registrations'),
    ('node_heartbeat_interval', '60', 'Expected heartbeat interval in seconds'),
    ('auto_approve_enabled', 'false', 'Automatically approve new nodes'),
    ('github_repo', '"https://github.com/gydschain/gydschain.git"', 'GitHub repository URL'),
    ('github_branch', '"main"', 'Default branch to deploy')
ON CONFLICT (key) DO NOTHING;

-- ============================================
-- Views
-- ============================================

-- Active nodes summary
CREATE OR REPLACE VIEW active_nodes_summary AS
SELECT 
    node_type,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE last_seen > NOW() - INTERVAL '5 minutes') as online,
    COUNT(*) FILTER (WHERE last_seen <= NOW() - INTERVAL '5 minutes' OR last_seen IS NULL) as offline,
    MAX(sync_height) as max_sync_height,
    MIN(sync_height) as min_sync_height
FROM nodes
WHERE status = 'approved'
GROUP BY node_type;

-- Recent node registrations
CREATE OR REPLACE VIEW recent_registrations AS
SELECT 
    n.node_id,
    n.hostname,
    n.public_ip,
    n.node_type,
    n.status,
    n.registered_at,
    a.username as approved_by_username
FROM nodes n
LEFT JOIN admin_users a ON n.approved_by = a.id
ORDER BY n.registered_at DESC
LIMIT 50;

-- ============================================
-- Functions
-- ============================================

-- Allocate next VPN address
CREATE OR REPLACE FUNCTION allocate_vpn_address()
RETURNS INET AS $$
DECLARE
    base_network INET := '10.100.0.0/24';
    next_host INT;
BEGIN
    SELECT COALESCE(MAX(
        (regexp_match(vpn_address::text, '\.(\d+)/'))[1]::int
    ), 1) + 1 INTO next_host
    FROM nodes
    WHERE vpn_address IS NOT NULL;
    
    IF next_host > 254 THEN
        RAISE EXCEPTION 'No more VPN addresses available';
    END IF;
    
    RETURN ('10.100.0.' || next_host || '/24')::INET;
END;
$$ LANGUAGE plpgsql;

-- Approve node
CREATE OR REPLACE FUNCTION approve_node(
    p_node_id VARCHAR(64),
    p_admin_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_vpn_address INET;
BEGIN
    -- Allocate VPN address
    v_vpn_address := allocate_vpn_address();
    
    -- Update node
    UPDATE nodes SET
        status = 'approved',
        vpn_address = v_vpn_address,
        approved_at = NOW(),
        approved_by = p_admin_id
    WHERE node_id = p_node_id AND status = 'pending';
    
    -- Log the event
    INSERT INTO system_logs (level, source, message, metadata)
    VALUES ('info', 'node_approval', 'Node approved', 
            jsonb_build_object('node_id', p_node_id, 'vpn_address', v_vpn_address::text));
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Clean old heartbeats (keep 7 days)
CREATE OR REPLACE FUNCTION clean_old_heartbeats()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM node_heartbeats
    WHERE timestamp < NOW() - INTERVAL '7 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;
