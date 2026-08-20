-- Filora Database Schema
-- Canonical source of truth. All domain terms match CONTEXT.md.
-- Design decisions documented in docs/adr/.

-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUMS
-- ============================================================================

CREATE TYPE storage_provider AS ENUM ('cloudinary', 'imagekit', 'r2', 'gcs');
CREATE TYPE storage_layer AS ENUM ('serving', 'archive');
CREATE TYPE location_status AS ENUM ('pending', 'stored', 'failed');
CREATE TYPE membership_role AS ENUM ('owner', 'editor', 'viewer');
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'expired', 'revoked');
CREATE TYPE client_type AS ENUM ('web', 'cli');

-- ============================================================================
-- USERS (self-managed auth, invite-only)
-- ============================================================================

CREATE TABLE users (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    name          text NOT NULL DEFAULT '',
    avatar_url    text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- ============================================================================
-- RBAC (roles hardcoded in code; DB stores assignments only)
-- ============================================================================

CREATE TABLE user_roles (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id   bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_name text NOT NULL, -- superuser, admin, member, viewer
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, role_name)
);

-- ============================================================================
-- SESSIONS (unified: web + CLI, opaque token with sliding window TTL)
-- ============================================================================

CREATE TABLE sessions (
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash   text NOT NULL UNIQUE,
    client       client_type NOT NULL,
    label        text NOT NULL DEFAULT '',
    last_used_at timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user ON sessions (user_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_sessions_token ON sessions (token_hash) WHERE revoked_at IS NULL;

-- ============================================================================
-- SPACES
-- ============================================================================

CREATE TABLE spaces (
    id                 uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name               text NOT NULL,
    owner_id           bigint NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    storage_quota_bytes bigint NOT NULL DEFAULT 0, -- 0 = unlimited
    storage_used_bytes  bigint NOT NULL DEFAULT 0,
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_spaces_owner ON spaces (owner_id);

-- ============================================================================
-- SPACE MEMBERS
-- ============================================================================

CREATE TABLE space_members (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    space_id  uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    user_id   bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role      membership_role NOT NULL DEFAULT 'viewer',
    joined_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (space_id, user_id)
);

-- ============================================================================
-- INVITATIONS (opaque token, manual link share)
-- ============================================================================

CREATE TABLE invitations (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    space_id    uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    email       text NOT NULL,
    role        membership_role NOT NULL DEFAULT 'viewer',
    status      invitation_status NOT NULL DEFAULT 'pending',
    token_hash  text NOT NULL UNIQUE, -- SHA-256 of opaque invitation token
    invited_by  bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (space_id, email)
);

-- ============================================================================
-- FOLDERS (adjacency list, unlimited nesting)
-- ============================================================================

CREATE TABLE folders (
    id         uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    space_id   uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    parent_id  uuid REFERENCES folders (id) ON DELETE CASCADE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz, -- soft delete
    UNIQUE (space_id, parent_id, name)
);

CREATE INDEX idx_folders_space ON folders (space_id);
CREATE INDEX idx_folders_parent ON folders (parent_id);

-- ============================================================================
-- ASSETS
-- ============================================================================

CREATE TABLE assets (
    id                uuid NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    original_filename text NOT NULL,
    name              text NOT NULL, -- display name, mutable
    mime_type         text NOT NULL,
    size_bytes        bigint NOT NULL,
    checksum_sha256   text NOT NULL,
    width             int,    -- nullable, images only
    height            int,    -- nullable, images only
    uploaded_by       bigint NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_assets_checksum ON assets (checksum_sha256);

-- ============================================================================
-- ASSET REFERENCES (many-to-many: asset ↔ folder/space)
-- ============================================================================

CREATE TABLE asset_references (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    asset_id   uuid NOT NULL REFERENCES assets (id) ON DELETE RESTRICT,
    space_id   uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    folder_id  uuid REFERENCES folders (id) ON DELETE CASCADE, -- NULL = space root
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz -- soft delete
);

CREATE UNIQUE INDEX idx_asset_refs_in_folder
    ON asset_references (asset_id, space_id, folder_id)
    WHERE folder_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX idx_asset_refs_at_root
    ON asset_references (asset_id, space_id)
    WHERE folder_id IS NULL AND deleted_at IS NULL;

CREATE INDEX idx_asset_refs_space ON asset_references (space_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_asset_refs_folder ON asset_references (folder_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_asset_refs_asset ON asset_references (asset_id) WHERE deleted_at IS NULL;

-- ============================================================================
-- TAGS (flat labels, scoped per space)
-- ============================================================================

CREATE TABLE tags (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    space_id  uuid NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    name      text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (space_id, name)
);

CREATE TABLE asset_tags (
    asset_id uuid NOT NULL REFERENCES assets (id) ON DELETE CASCADE,
    tag_id   bigint NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (asset_id, tag_id)
);

-- ============================================================================
-- STORAGE ACCOUNTS
-- ============================================================================

CREATE TABLE storage_accounts (
    id                    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    provider              storage_provider NOT NULL,
    label                 text NOT NULL,
    layer                 storage_layer NOT NULL,
    credentials_encrypted bytea NOT NULL,
    is_active             boolean NOT NULL DEFAULT true,
    quota_bytes           bigint NOT NULL DEFAULT 0,
    used_bytes            bigint NOT NULL DEFAULT 0,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);

-- ============================================================================
-- STORAGE LOCATIONS (physical copy records)
-- ============================================================================

CREATE TABLE storage_locations (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    asset_id    uuid NOT NULL REFERENCES assets (id) ON DELETE RESTRICT,
    account_id  bigint NOT NULL REFERENCES storage_accounts (id) ON DELETE RESTRICT,
    layer       storage_layer NOT NULL,
    status      location_status NOT NULL DEFAULT 'pending',
    remote_path text,
    remote_url  text,
    error       text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_storage_loc_stored
    ON storage_locations (asset_id, account_id)
    WHERE status = 'stored';

CREATE INDEX idx_storage_loc_asset ON storage_locations (asset_id);
CREATE INDEX idx_storage_loc_account ON storage_locations (account_id);

-- ============================================================================
-- ARCHIVE SYNC JOBS
-- ============================================================================

CREATE TABLE archive_sync_jobs (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    asset_id      uuid NOT NULL REFERENCES assets (id) ON DELETE CASCADE,
    status        text NOT NULL DEFAULT 'pending',
    attempts      int NOT NULL DEFAULT 0,
    max_attempts  int NOT NULL DEFAULT 5,
    last_error    text,
    next_retry_at timestamptz,
    completed_at  timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_archive_jobs_pending
    ON archive_sync_jobs (next_retry_at)
    WHERE status IN ('pending', 'failed') AND attempts < max_attempts;
