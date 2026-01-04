-- GYDS Chain Indexer Database Schema
-- PostgreSQL

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Blocks table
CREATE TABLE IF NOT EXISTS blocks (
    id SERIAL PRIMARY KEY,
    number BIGINT NOT NULL UNIQUE,
    hash VARCHAR(66) NOT NULL UNIQUE,
    parent_hash VARCHAR(66) NOT NULL,
    state_root VARCHAR(66) NOT NULL,
    transactions_root VARCHAR(66) NOT NULL,
    receipts_root VARCHAR(66) NOT NULL,
    validator VARCHAR(42) NOT NULL,
    timestamp BIGINT NOT NULL,
    gas_used BIGINT NOT NULL DEFAULT 0,
    gas_limit BIGINT NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    tx_count INT NOT NULL DEFAULT 0,
    extra_data BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_blocks_timestamp (timestamp),
    INDEX idx_blocks_validator (validator)
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    hash VARCHAR(66) NOT NULL UNIQUE,
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    block_hash VARCHAR(66) NOT NULL,
    tx_index INT NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42),
    value VARCHAR(78) NOT NULL,
    asset VARCHAR(42) NOT NULL DEFAULT 'GYDS',
    fee VARCHAR(78) NOT NULL,
    nonce BIGINT NOT NULL,
    data BYTEA,
    signature VARCHAR(130) NOT NULL,
    tx_type VARCHAR(20) NOT NULL DEFAULT 'transfer',
    status SMALLINT NOT NULL DEFAULT 1,
    gas_used BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_tx_from (from_address),
    INDEX idx_tx_to (to_address),
    INDEX idx_tx_block (block_number),
    INDEX idx_tx_asset (asset),
    INDEX idx_tx_type (tx_type)
);

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE,
    nonce BIGINT NOT NULL DEFAULT 0,
    tx_count BIGINT NOT NULL DEFAULT 0,
    first_seen_block BIGINT NOT NULL,
    last_seen_block BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_accounts_last_seen (last_seen_block)
);

-- Account balances table
CREATE TABLE IF NOT EXISTS account_balances (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL,
    asset VARCHAR(42) NOT NULL,
    balance VARCHAR(78) NOT NULL DEFAULT '0',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(address, asset),
    INDEX idx_balances_address (address),
    INDEX idx_balances_asset (asset)
);

-- Assets table
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    asset_id VARCHAR(42) NOT NULL UNIQUE,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    decimals SMALLINT NOT NULL DEFAULT 18,
    total_supply VARCHAR(78) NOT NULL,
    max_supply VARCHAR(78),
    creator VARCHAR(42) NOT NULL,
    is_native BOOLEAN NOT NULL DEFAULT FALSE,
    is_stablecoin BOOLEAN NOT NULL DEFAULT FALSE,
    peg_target VARCHAR(10),
    mintable BOOLEAN NOT NULL DEFAULT FALSE,
    burnable BOOLEAN NOT NULL DEFAULT FALSE,
    created_block BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_assets_symbol (symbol),
    INDEX idx_assets_creator (creator)
);

-- Validators table
CREATE TABLE IF NOT EXISTS validators (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL UNIQUE,
    stake VARCHAR(78) NOT NULL,
    commission SMALLINT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    jailed BOOLEAN NOT NULL DEFAULT FALSE,
    jailed_until BIGINT,
    blocks_proposed BIGINT NOT NULL DEFAULT 0,
    blocks_signed BIGINT NOT NULL DEFAULT 0,
    slashing_events INT NOT NULL DEFAULT 0,
    delegator_count INT NOT NULL DEFAULT 0,
    total_delegations VARCHAR(78) NOT NULL DEFAULT '0',
    created_block BIGINT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_validators_active (active),
    INDEX idx_validators_stake (stake)
);

-- Delegations table
CREATE TABLE IF NOT EXISTS delegations (
    id SERIAL PRIMARY KEY,
    delegator VARCHAR(42) NOT NULL,
    validator VARCHAR(42) NOT NULL REFERENCES validators(address),
    amount VARCHAR(78) NOT NULL,
    rewards VARCHAR(78) NOT NULL DEFAULT '0',
    created_block BIGINT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(delegator, validator),
    INDEX idx_delegations_delegator (delegator),
    INDEX idx_delegations_validator (validator)
);

-- Slashing events table
CREATE TABLE IF NOT EXISTS slashing_events (
    id SERIAL PRIMARY KEY,
    validator VARCHAR(42) NOT NULL REFERENCES validators(address),
    block_number BIGINT NOT NULL,
    reason VARCHAR(50) NOT NULL,
    amount VARCHAR(78) NOT NULL,
    jailed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_slashing_validator (validator),
    INDEX idx_slashing_block (block_number)
);

-- Token transfers table (for detailed transfer history)
CREATE TABLE IF NOT EXISTS token_transfers (
    id SERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL REFERENCES transactions(hash),
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    asset VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL,
    block_number BIGINT NOT NULL,
    log_index INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_transfers_from (from_address),
    INDEX idx_transfers_to (to_address),
    INDEX idx_transfers_asset (asset),
    INDEX idx_transfers_block (block_number)
);

-- Mining rewards table
CREATE TABLE IF NOT EXISTS mining_rewards (
    id SERIAL PRIMARY KEY,
    block_number BIGINT NOT NULL REFERENCES blocks(number),
    miner VARCHAR(42) NOT NULL,
    reward VARCHAR(78) NOT NULL,
    fees VARCHAR(78) NOT NULL DEFAULT '0',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_rewards_miner (miner),
    INDEX idx_rewards_block (block_number)
);

-- Stablecoin peg history
CREATE TABLE IF NOT EXISTS stablecoin_peg_history (
    id SERIAL PRIMARY KEY,
    asset VARCHAR(42) NOT NULL,
    block_number BIGINT NOT NULL,
    price VARCHAR(78) NOT NULL,
    target VARCHAR(78) NOT NULL,
    deviation VARCHAR(78) NOT NULL,
    supply VARCHAR(78) NOT NULL,
    collateral_ratio VARCHAR(78),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    INDEX idx_peg_asset (asset),
    INDEX idx_peg_block (block_number)
);

-- Indexer state table
CREATE TABLE IF NOT EXISTS indexer_state (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert native assets
INSERT INTO assets (asset_id, symbol, name, decimals, total_supply, creator, is_native, is_stablecoin, created_block)
VALUES 
    ('GYDS', 'GYDS', 'GYDS Token', 18, '1000000000000000000000000000', '0x0000000000000000000000000000000000000000', TRUE, FALSE, 0),
    ('GYD', 'GYD', 'GYD Stablecoin', 18, '0', '0x0000000000000000000000000000000000000000', TRUE, TRUE, 0)
ON CONFLICT DO NOTHING;

-- Insert initial indexer state
INSERT INTO indexer_state (key, value)
VALUES 
    ('last_indexed_block', '0'),
    ('indexer_version', '1.0.0')
ON CONFLICT DO NOTHING;

-- Useful views

-- View: Account summary
CREATE OR REPLACE VIEW account_summary AS
SELECT 
    a.address,
    a.nonce,
    a.tx_count,
    a.first_seen_block,
    a.last_seen_block,
    COALESCE(
        json_object_agg(ab.asset, ab.balance) FILTER (WHERE ab.asset IS NOT NULL),
        '{}'::json
    ) as balances
FROM accounts a
LEFT JOIN account_balances ab ON a.address = ab.address
GROUP BY a.address, a.nonce, a.tx_count, a.first_seen_block, a.last_seen_block;

-- View: Validator summary
CREATE OR REPLACE VIEW validator_summary AS
SELECT 
    v.*,
    (v.blocks_signed::float / NULLIF(v.blocks_proposed, 0) * 100) as uptime_percentage
FROM validators v
WHERE v.active = TRUE
ORDER BY v.stake DESC;

-- View: Recent transactions
CREATE OR REPLACE VIEW recent_transactions AS
SELECT 
    t.*,
    b.timestamp as block_timestamp
FROM transactions t
JOIN blocks b ON t.block_number = b.number
ORDER BY t.id DESC
LIMIT 100;
