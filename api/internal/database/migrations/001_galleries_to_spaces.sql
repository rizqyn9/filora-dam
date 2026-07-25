-- ============================================================================
-- Migration: galleries/albums → spaces/folders
-- ============================================================================
-- Transforms the old gallery/album schema into the new space/folder model.
--
-- Run against the EXISTING database (that was created from the old schema.sql).
-- After running, the database matches the new schema.sql.
--
-- IMPORTANT: Back up the database before running this migration.
--
--   psql "$DATABASE_URL" -f internal/database/migrations/001_galleries_to_spaces.sql
--
-- Changes:
--   1. Rename galleries → spaces, gallery_members → space_members
--   2. Drop album tables (albums, album_members, album_assets)
--   3. Create folders table
--   4. Add folder_id to assets, rename gallery_id → space_id
--   5. Update invitations (drop album_id, rename gallery_id → space_id)
--   6. Update tags (rename gallery_id → space_id)
--   7. Update permissions seed data
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. Rename galleries → spaces
-- ============================================================================

-- Rename the table
ALTER TABLE galleries RENAME TO spaces;

-- Rename indexes
ALTER INDEX idx_galleries_owner_id RENAME TO idx_spaces_owner_id;
ALTER INDEX idx_galleries_one_default RENAME TO idx_spaces_one_default;

-- Rename trigger
ALTER TRIGGER trg_galleries_updated_at ON spaces RENAME TO trg_spaces_updated_at;

-- Rename gallery_members → space_members
ALTER TABLE gallery_members RENAME TO space_members;
ALTER TABLE space_members RENAME COLUMN gallery_id TO space_id;
ALTER INDEX idx_gallery_members_user_id RENAME TO idx_space_members_user_id;

-- ============================================================================
-- 2. Drop album tables
-- ============================================================================

-- Drop in dependency order (album_assets and album_members reference albums)
DROP TABLE IF EXISTS album_assets;
DROP TABLE IF EXISTS album_members;
DROP TABLE IF EXISTS albums;

-- ============================================================================
-- 3. Create folders table
-- ============================================================================

CREATE TABLE folders (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    space_id   BIGINT      NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    parent_id  BIGINT      REFERENCES folders(id) ON DELETE CASCADE,
    owner_id   BIGINT      NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    path       TEXT        NOT NULL DEFAULT '/',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (space_id, parent_id, name)
);

CREATE INDEX idx_folders_space_id  ON folders (space_id);
CREATE INDEX idx_folders_parent_id ON folders (parent_id);
CREATE INDEX idx_folders_path      ON folders USING btree (path text_pattern_ops);

CREATE TRIGGER trg_folders_updated_at
    BEFORE UPDATE ON folders
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================================
-- 4. Update assets table
-- ============================================================================

-- Rename gallery_id → space_id
ALTER TABLE assets RENAME COLUMN gallery_id TO space_id;

-- Add folder_id column (nullable, SET NULL on folder delete)
ALTER TABLE assets ADD COLUMN folder_id BIGINT REFERENCES folders(id) ON DELETE SET NULL;

-- Rename indexes
ALTER INDEX idx_assets_gallery_id RENAME TO idx_assets_space_id;
ALTER INDEX idx_assets_gallery_active RENAME TO idx_assets_space_active;
ALTER INDEX idx_assets_gallery_hash RENAME TO idx_assets_space_hash;

-- Add new folder-related indexes
CREATE INDEX idx_assets_folder_id ON assets (folder_id);
CREATE INDEX idx_assets_folder_active ON assets (space_id, folder_id) WHERE deleted_at IS NULL;

-- ============================================================================
-- 5. Update invitations table
-- ============================================================================

-- Drop album-related constraints and columns
ALTER TABLE invitations DROP CONSTRAINT IF EXISTS invitations_one_target;
DROP INDEX IF EXISTS idx_invitations_pending_album;

-- Drop album_id column
ALTER TABLE invitations DROP COLUMN IF EXISTS album_id;

-- Rename gallery_id → space_id and make it NOT NULL
ALTER TABLE invitations RENAME COLUMN gallery_id TO space_id;
ALTER TABLE invitations ALTER COLUMN space_id SET NOT NULL;

-- Rename index
ALTER INDEX idx_invitations_pending_gallery RENAME TO idx_invitations_pending_space;

-- ============================================================================
-- 6. Update tags table
-- ============================================================================

ALTER TABLE tags RENAME COLUMN gallery_id TO space_id;

-- The unique constraint on (gallery_id, name) needs to be recreated
-- Drop the old one and create new
ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_gallery_id_name_key;
ALTER TABLE tags ADD CONSTRAINT tags_space_id_name_key UNIQUE (space_id, name);

-- ============================================================================
-- 7. Update permissions seed data
-- ============================================================================

-- Replace gallery permissions with space permissions
UPDATE permissions SET resource = 'space' WHERE resource = 'gallery';

-- Replace album permissions with folder permissions
UPDATE permissions SET resource = 'folder', description = CASE action
    WHEN 'read' THEN 'View folders'
    WHEN 'create' THEN 'Create folders'
    WHEN 'update' THEN 'Rename / move folders'
    WHEN 'delete' THEN 'Delete folders'
    WHEN 'invite' THEN NULL
    ELSE description
END WHERE resource = 'album';

-- Remove album:invite (folders don't have separate invitations)
DELETE FROM permissions WHERE resource = 'folder' AND action = 'invite';

-- Update role descriptions
UPDATE roles SET description = 'Owns spaces/folders and manages their own assets.' WHERE slug = 'member';
UPDATE roles SET description = 'Read-only access to spaces they belong to.' WHERE slug = 'viewer';

COMMIT;
