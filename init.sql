-- Initialize the smart_mail_relay database
-- This script runs when the MySQL container starts for the first time

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS smart_mail_relay CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the database
USE smart_mail_relay;

-- Create tables (these will be auto-migrated by GORM, but we can add initial data here)

-- Insert some sample forwarding rules (optional)
-- INSERT INTO forward_rules (keyword, target_email, enabled, created_at, updated_at) VALUES
-- ('urgent', 'admin@company.com', true, NOW(), NOW()),
-- ('support', 'support@company.com', true, NOW(), NOW()),
-- ('sales', 'sales@company.com', true, NOW(), NOW());

-- Grant permissions to the application user
GRANT ALL PRIVILEGES ON smart_mail_relay.* TO 'smart_mail_relay'@'%';
FLUSH PRIVILEGES; 