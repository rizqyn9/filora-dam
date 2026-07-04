-- ============================================================================
-- Filora DAM - Baseline RBAC seed
-- ============================================================================
-- Run after schema.sql:
--
--     psql "$DATABASE_URL" -f internal/database/seed.sql
--
-- Idempotent: safe to re-run. Grants are refreshed on conflict.
--
-- Note on layered authorization:
--   * These global RBAC grants gate app-wide capability (who may create
--     galleries, manage storage accounts, administer users, etc.).
--   * Day-to-day access to a specific gallery/album is governed by the local
--     membership role (owner/editor/viewer) in gallery_members / album_members.
--   * 'own' scope on gallery/album/asset/tag = resources the user owns or is a
--     member of; superuser ('*','*') bypasses everything.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Roles
-- ----------------------------------------------------------------------------
INSERT INTO roles (slug, name, description, is_system) VALUES
    ('superuser', 'Superuser', 'Full, unrestricted access to everything.',                     TRUE),
    ('admin',     'Admin',     'Manages assets, storage accounts, users and CLI sessions.',    TRUE),
    ('member',    'Member',    'Owns galleries/albums and manages their own assets.',          TRUE),
    ('viewer',    'Viewer',    'Read-only access to galleries/albums they belong to.',         TRUE)
ON CONFLICT (slug) DO UPDATE
    SET name = EXCLUDED.name,
        description = EXCLUDED.description,
        is_system = EXCLUDED.is_system;

-- ----------------------------------------------------------------------------
-- Permissions catalog (resource, action)
-- ----------------------------------------------------------------------------
INSERT INTO permissions (resource, action, description) VALUES
    ('*',        '*',        'Wildcard: full access to every resource and action'),

    ('asset',    'read',     'View assets'),
    ('asset',    'create',   'Upload / create assets'),
    ('asset',    'update',   'Update asset metadata'),
    ('asset',    'delete',   'Delete assets'),
    ('asset',    'download', 'Download asset files'),

    ('gallery',  'read',     'View galleries'),
    ('gallery',  'create',   'Create galleries'),
    ('gallery',  'update',   'Rename / edit galleries'),
    ('gallery',  'delete',   'Delete galleries'),
    ('gallery',  'invite',   'Invite / manage gallery members'),

    ('album',    'read',     'View albums'),
    ('album',    'create',   'Create albums'),
    ('album',    'update',   'Edit albums / manage album assets'),
    ('album',    'delete',   'Delete albums'),
    ('album',    'invite',   'Invite / manage album members'),

    ('tag',      'read',     'View tags'),
    ('tag',      'create',   'Create tags / tag assets'),
    ('tag',      'update',   'Rename tags'),
    ('tag',      'delete',   'Delete tags / untag assets'),

    ('storage',  'read',     'View storage accounts and usage'),
    ('storage',  'create',   'Register storage accounts'),
    ('storage',  'update',   'Update storage accounts'),
    ('storage',  'delete',   'Deactivate / remove storage accounts'),

    ('user',     'read',     'View users'),
    ('user',     'update',   'Update users'),
    ('user',     'delete',   'Deactivate / remove users'),

    ('role',     'read',     'View roles and permissions'),
    ('role',     'assign',   'Assign / revoke roles to users'),
    ('role',     'manage',   'Create / edit roles and their permissions'),

    ('session',  'read',     'View CLI sessions'),
    ('session',  'revoke',   'Revoke CLI sessions'),

    ('dashboard','read',     'View dashboard metrics')
ON CONFLICT (resource, action) DO UPDATE
    SET description = EXCLUDED.description;

-- ----------------------------------------------------------------------------
-- Grants: role_permissions (role -> permission @ scope)
-- ----------------------------------------------------------------------------

-- superuser: wildcard, scope all
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT r.id, p.id, 'all'::permission_scope
FROM roles r
JOIN permissions p ON p.resource = '*' AND p.action = '*'
WHERE r.slug = 'superuser'
ON CONFLICT (role_id, permission_id) DO UPDATE SET scope = EXCLUDED.scope;

-- admin: manage the whole workspace (scope all), except the raw wildcard and
-- 'role:manage' which are reserved for superuser.
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT r.id, p.id, 'all'::permission_scope
FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
    ('asset','read'), ('asset','create'), ('asset','update'), ('asset','delete'), ('asset','download'),
    ('gallery','read'), ('gallery','create'), ('gallery','update'), ('gallery','delete'), ('gallery','invite'),
    ('album','read'), ('album','create'), ('album','update'), ('album','delete'), ('album','invite'),
    ('tag','read'), ('tag','create'), ('tag','update'), ('tag','delete'),
    ('storage','read'), ('storage','create'), ('storage','update'), ('storage','delete'),
    ('user','read'), ('user','update'),
    ('role','read'), ('role','assign'),
    ('session','read'), ('session','revoke'),
    ('dashboard','read')
)
WHERE r.slug = 'admin'
ON CONFLICT (role_id, permission_id) DO UPDATE SET scope = EXCLUDED.scope;

-- member: owns galleries/albums/tags and manages their OWN assets. Access to a
-- specific gallery/album is further governed by membership role.
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT r.id, p.id, 'own'::permission_scope
FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
    ('asset','read'), ('asset','create'), ('asset','update'), ('asset','delete'), ('asset','download'),
    ('gallery','read'), ('gallery','create'), ('gallery','update'), ('gallery','delete'), ('gallery','invite'),
    ('album','read'), ('album','create'), ('album','update'), ('album','delete'), ('album','invite'),
    ('tag','read'), ('tag','create'), ('tag','update'), ('tag','delete'),
    ('storage','read'),
    ('dashboard','read')
)
WHERE r.slug = 'member'
ON CONFLICT (role_id, permission_id) DO UPDATE SET scope = EXCLUDED.scope;

-- viewer: read-only over galleries/albums/assets they belong to.
INSERT INTO role_permissions (role_id, permission_id, scope)
SELECT r.id, p.id, 'own'::permission_scope
FROM roles r
JOIN permissions p ON (p.resource, p.action) IN (
    ('asset','read'), ('asset','download'),
    ('gallery','read'),
    ('album','read'),
    ('tag','read'),
    ('dashboard','read')
)
WHERE r.slug = 'viewer'
ON CONFLICT (role_id, permission_id) DO UPDATE SET scope = EXCLUDED.scope;

-- ----------------------------------------------------------------------------
-- Superuser bootstrap
-- ----------------------------------------------------------------------------
-- Users are created from Clerk (via webhook / first request), so the owner row
-- does not exist until they sign in once. After the family owner has signed in,
-- grant them the superuser role by running (replace the email):
--
--   INSERT INTO user_roles (user_id, role_id)
--   SELECT u.id, r.id
--   FROM users u, roles r
--   WHERE u.email = 'owner@example.com' AND r.slug = 'superuser'
--   ON CONFLICT DO NOTHING;
