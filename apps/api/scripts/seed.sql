-- Seed script for testing
-- Insert test user
INSERT INTO users (id, email, name, password_hash, storage_quota, storage_used)
VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'test@filora.com', 'Test User', '$2a$10$placeholder', 5368709120, 0)
ON CONFLICT (email) DO NOTHING;

-- Insert test storage provider
INSERT INTO storage_providers (id, user_id, name, type, credentials, quota, used, is_active)
VALUES
    ('660e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440000', 'Test Cloudinary', 'cloudinary', '{"api_key": "test"}', 10737418240, 0, true)
ON CONFLICT DO NOTHING;
