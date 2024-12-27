-- Create extension for UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Version table
CREATE TABLE whatsmeow_version (
    version INTEGER
);

-- Device table
CREATE TABLE whatsmeow_device (
    jid TEXT PRIMARY KEY,
    registration_id BIGINT NOT NULL CHECK (registration_id >= 0 AND registration_id < 4294967296),
    noise_key BYTEA NOT NULL CHECK (length(noise_key) = 32),
    identity_key BYTEA NOT NULL CHECK (length(identity_key) = 32),
    signed_pre_key BYTEA NOT NULL CHECK (length(signed_pre_key) = 32),
    signed_pre_key_id INTEGER NOT NULL CHECK (signed_pre_key_id >= 0 AND signed_pre_key_id < 16777216),
    signed_pre_key_sig BYTEA NOT NULL CHECK (length(signed_pre_key_sig) = 64),
    adv_key BYTEA NOT NULL,
    adv_details BYTEA NOT NULL,
    adv_account_sig BYTEA NOT NULL CHECK (length(adv_account_sig) = 64),
    adv_device_sig BYTEA NOT NULL CHECK (length(adv_device_sig) = 64),
    platform TEXT NOT NULL DEFAULT '',
    business_name TEXT NOT NULL DEFAULT '',
    push_name TEXT NOT NULL DEFAULT '',
    adv_account_sig_key BYTEA CHECK (length(adv_account_sig_key) = 32),
    facebook_uuid UUID
);

-- Identity keys table
CREATE TABLE whatsmeow_identity_keys (
    our_jid TEXT,
    their_id TEXT,
    identity BYTEA NOT NULL CHECK (length(identity) = 32),
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Pre keys table
CREATE TABLE whatsmeow_pre_keys (
    jid TEXT,
    key_id INTEGER CHECK (key_id >= 0 AND key_id < 16777216),
    key BYTEA NOT NULL CHECK (length(key) = 32),
    uploaded BOOLEAN NOT NULL,
    PRIMARY KEY (jid, key_id),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Sessions table
CREATE TABLE whatsmeow_sessions (
    our_jid TEXT,
    their_id TEXT,
    session BYTEA,
    PRIMARY KEY (our_jid, their_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Sender keys table
CREATE TABLE whatsmeow_sender_keys (
    our_jid TEXT,
    chat_id TEXT,
    sender_id TEXT,
    sender_key BYTEA NOT NULL,
    PRIMARY KEY (our_jid, chat_id, sender_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- App state sync keys table
CREATE TABLE whatsmeow_app_state_sync_keys (
    jid TEXT,
    key_id BYTEA,
    key_data BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    fingerprint BYTEA NOT NULL,
    PRIMARY KEY (jid, key_id),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- App state version table
CREATE TABLE whatsmeow_app_state_version (
    jid TEXT,
    name TEXT,
    version BIGINT NOT NULL,
    hash BYTEA NOT NULL CHECK (length(hash) = 128),
    PRIMARY KEY (jid, name),
    FOREIGN KEY (jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- App state mutation macs table
CREATE TABLE whatsmeow_app_state_mutation_macs (
    jid TEXT,
    name TEXT,
    version BIGINT,
    index_mac BYTEA CHECK (length(index_mac) = 32),
    value_mac BYTEA NOT NULL CHECK (length(value_mac) = 32),
    PRIMARY KEY (jid, name, version, index_mac),
    FOREIGN KEY (jid, name) REFERENCES whatsmeow_app_state_version(jid, name) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Contacts table
CREATE TABLE whatsmeow_contacts (
    our_jid TEXT,
    their_jid TEXT,
    first_name TEXT,
    full_name TEXT,
    push_name TEXT,
    business_name TEXT,
    PRIMARY KEY (our_jid, their_jid),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Chat settings table
CREATE TABLE whatsmeow_chat_settings (
    our_jid TEXT,
    chat_jid TEXT,
    muted_until BIGINT NOT NULL DEFAULT 0,
    pinned BOOLEAN NOT NULL DEFAULT false,
    archived BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (our_jid, chat_jid),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Message secrets table
CREATE TABLE whatsmeow_message_secrets (
    our_jid TEXT,
    chat_jid TEXT,
    sender_jid TEXT,
    message_id TEXT,
    key BYTEA NOT NULL,
    PRIMARY KEY (our_jid, chat_jid, sender_jid, message_id),
    FOREIGN KEY (our_jid) REFERENCES whatsmeow_device(jid) ON DELETE CASCADE ON UPDATE CASCADE
);

-- Privacy tokens table
CREATE TABLE whatsmeow_privacy_tokens (
    our_jid TEXT,
    their_jid TEXT,
    token BYTEA NOT NULL,
    timestamp BIGINT NOT NULL,
    PRIMARY KEY (our_jid, their_jid)
);
