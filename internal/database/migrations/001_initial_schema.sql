-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- System metadata table
CREATE TABLE IF NOT EXISTS system_info (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version VARCHAR(50) NOT NULL,
    initialized_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert initial system info
INSERT INTO system_info (version) VALUES ('0.1.0');

-- Create index
CREATE INDEX idx_system_info_version ON system_info(version);
