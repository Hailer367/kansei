-- C&C Server Database Schema for Supabase

-- Clients table to store registered clients
CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    ip TEXT NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status TEXT DEFAULT 'connected',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Registration tokens table for initial client registration
CREATE TABLE IF NOT EXISTS registration_tokens (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    token TEXT UNIQUE NOT NULL,
    created_by TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_used BOOLEAN DEFAULT FALSE
);

-- Commands table to store commands sent to clients
CREATE TABLE IF NOT EXISTS commands (
    id TEXT PRIMARY KEY,
    client_id TEXT REFERENCES clients(id),
    command TEXT NOT NULL,
    status TEXT DEFAULT 'pending', -- pending, executing, completed, failed
    result TEXT,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_clients_status ON clients(status);
CREATE INDEX IF NOT EXISTS idx_clients_last_seen ON clients(last_seen);
CREATE INDEX IF NOT EXISTS idx_commands_client_id ON commands(client_id);
CREATE INDEX IF NOT EXISTS idx_commands_status ON commands(status);
CREATE INDEX IF NOT EXISTS idx_commands_created_at ON commands(created_at);

-- RLS (Row Level Security) policies - adjust as needed for your security model
ALTER TABLE clients ENABLE ROW LEVEL SECURITY;
ALTER TABLE commands ENABLE ROW LEVEL SECURITY;
ALTER TABLE registration_tokens ENABLE ROW LEVEL SECURITY;

-- Example policies (you may need to adjust based on your auth model)
-- For now, we're setting permissive policies for testing
CREATE POLICY "Allow all operations for authenticated users" ON clients
    FOR ALL USING (true);

CREATE POLICY "Allow all operations for authenticated users" ON commands
    FOR ALL USING (true);

CREATE POLICY "Allow all operations for authenticated users" ON registration_tokens
    FOR ALL USING (true);