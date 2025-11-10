-- ============================================
-- Add DeepSeek Agent
-- ============================================

-- Insert DeepSeek agent
INSERT INTO agents (name, model, status, initial_balance, current_balance) VALUES
('DeepSeek Chat', 'deepseek-chat', 'active', 100000.00, 100000.00)
ON CONFLICT DO NOTHING;

-- Initialize agent metrics for DeepSeek agent
INSERT INTO agent_metrics (agent_id)
SELECT id FROM agents WHERE name = 'DeepSeek Chat'
ON CONFLICT (agent_id) DO NOTHING;

-- Verify agent was created
SELECT id, name, model, status, current_balance FROM agents WHERE name = 'DeepSeek Chat';

