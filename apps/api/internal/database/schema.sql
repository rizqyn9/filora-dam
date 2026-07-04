-- ============================================================================
-- Filora DAM - Canonical Database Schema
-- ============================================================================
-- Target: PostgreSQL (Neon)
-- This file is the single source of truth for the database.
-- There are no migrations. Apply this file manually:
--
--     psql "$DATABASE_URL" -f internal/database/schema.sql
--     psql "$DATABASE_URL" -f internal/database/seed.sql
--
-- ID strategy:
--   * Control / lookup tables (users, roles, permissions, galleries, albums,
--     tags, storage_providers) use incremental BIGINT identity columns.
--   * High-volume / externally-exposed rows (assets, storage_locations,
--     cli_sessions) use UUID v7 (time-ordered) via uuid_generate_v7().
--   * Pure join tables use composite primary keys.
--
-- Domains:
--   * Identity & access : users, roles, permissions, role_permissions,
--                         user_roles, cli_sessions, clerk_webhook_events
--   * Organization      : galleries, gallery_members, albums, album_members,
--                         invitations, album_assets, tags, asset_tags
--   * Assets            : assets (soft-deletable via deleted_at)
--   * Storage           : storage_providers (2 layers), storage_locations,
--                         archive_sync_jobs
--
-- Auth model:
--   * Web login + sessions are handled by Clerk (https://clerk.com); we mirror
--     the Clerk user via clerk_user_id.
--   * Terminal (CLI) login uses our own opaque tokens (cli_sessions); a user
--     may hold many concurrent sessions.
--   * Global authorization is RBAC: users -> roles -> permissions, each grant
--     carrying a scope ('own' | 'all'); the 'superuser' role holds ('*','*').
--   * Per-resource sharing of galleries/albums uses a local membership role
--     (owner | editor | viewer).
--
-- Storage model:
--   * Two layers. 'serving' = free-tier hot providers (Cloudinary/ImageKit).
--     'archive' = cheap cold storage class (GCS Archive / R2 / etc.).
--   * Every asset is stored in BOTH layers: it has >= 1 serving location and
--     >= 1 archive location (tracked in storage_locations, with per-location
--     status). Storage accounts are global and managed by superadmin / users
--     holding the 'storage' permission.
--   * BACKLOG: per-layer account election (which account a new upload lands on)
--     is not modeled yet; it will be a selection strategy over active accounts.
-- ============================================================================

-- ----------------------------------------------------------------------------
-- Extensions & helper functions
-- ----------------------------------------------------------------------------

-- gen_random_uuid() is built into PostgreSQL core (>= 13), no extension needed.

-- UUID v7 (time-ordered). PostgreSQL has no native uuidv7() until PG18, so we
-- provide a portable implementation. It reuses gen_random_uuid() for entropy
-- and the correct variant bits, then overlays a millisecond timestamp and sets
-- the version nibble to 7.
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
AS $$
BEGIN
    RETURN encode(
        set_bit(
            set_bit(
                overlay(
                    uuid_send(gen_random_uuid())
                    PLACING substring(
                        int8send(floor(extract(epoch FROM clock_timestamp()) * 1000)::bigint)
                        FROM 3
                    )
                    FROM 1 FOR 6
                ),
                52, 1
            ),
            53, 1
        ),
        'hex'
    )::uuid;
END;
$$;

-- Keeps updated_at fresh on UPDATE.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

-- ----------------------------------------------------------------------------
-- Enums
-- ----------------------------------------------------------------------------

-- Scope qualifier for a global RBAC grant.
--   own : the permission applies only to rows the user owns.
--   all : the permission applies to every row in the family workspace.
CREATE TYPE permission_scope AS ENUM ('own', 'all');

-- Local membership role for a gallery or album.
CREATE TYPE member_role AS ENUM ('owner', 'editor', 'viewer');

-- Storage layer.
--   serving : hot, publicly-servable free-tier providers (Cloudinary/ImageKit).
--   archive : cold, cheap archive-class storage (GCS Archive / R2 / etc.).
CREATE TYPE storage_layer AS ENUM ('serving', 'archive');

-- Lifecycle of a single physical copy (one asset, one provider).
CREATE TYPE location_status AS ENUM ('pending', 'stored', 'failed');

-- Lifecycle of a gallery/album invitation.
CREATE TYPE invitation_status AS ENUM ('pending', 'accepted', 'revoked', 'expired');

-- Lifecycle of a background replication job.
CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');

-- ----------------------------------------------------------------------------
-- users  (mirrors Clerk identities)
-- ----------------------------------------------------------------------------
CREATE TABLE users (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    clerk_user_id TEXT        NOT NULL UNIQUE,   -- e.g. "user_2abc..." from Clerk
    email         TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    avatar_url    TEXT,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    last_seen_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email ON users (email);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- roles
-- ----------------------------------------------------------------------------
-- A user may hold multiple roles; that collection is the user's "role group".
CREATE TABLE roles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug        TEXT        NOT NULL UNIQUE,     -- superuser, admin, member, viewer
    name        TEXT        NOT NULL,
    description TEXT,
    is_system   BOOLEAN     NOT NULL DEFAULT FALSE, -- system roles must not be deleted
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- permissions  (catalog of resource + action pairs)
-- ----------------------------------------------------------------------------
-- Wildcard rows use '*' for resource and/or action. ('*','*') = full access.
CREATE TABLE permissions (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource    TEXT        NOT NULL,            -- asset, gallery, album, tag, storage, user, role, session, dashboard, *
    action      TEXT        NOT NULL,            -- read, create, update, delete, download, invite, assign, manage, revoke, *
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (resource, action)
);

-- ----------------------------------------------------------------------------
-- role_permissions  (grant: role -> permission @ scope)
-- ----------------------------------------------------------------------------
CREATE TABLE role_permissions (
    role_id       BIGINT           NOT NULL REFERENCES roles(id)       ON DELETE CASCADE,
    permission_id BIGINT           NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope         permission_scope NOT NULL DEFAULT 'own',
    created_at    TIMESTAMPTZ      NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_permission_id ON role_permissions (permission_id);

-- ----------------------------------------------------------------------------
-- user_roles  (assignment: user -> role)
-- ----------------------------------------------------------------------------
CREATE TABLE user_roles (
    user_id    BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    BIGINT      NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_by BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);

-- ----------------------------------------------------------------------------
-- cli_sessions  (terminal login; multiple concurrent sessions per user)
-- ----------------------------------------------------------------------------
-- The raw token is shown to the CLI once; only its SHA-256 hash is stored.
CREATE TABLE cli_sessions (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT        NOT NULL UNIQUE,    -- sha256(raw token)
    label        TEXT,                            -- device / terminal name
    ip_address   INET,
    user_agent   TEXT,
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,                     -- NULL = never expires
    revoked_at   TIMESTAMPTZ,                     -- NULL = active
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_cli_sessions_user_id ON cli_sessions (user_id);
CREATE INDEX idx_cli_sessions_active  ON cli_sessions (user_id) WHERE revoked_at IS NULL;

-- ----------------------------------------------------------------------------
-- clerk_webhook_events  (idempotency for Clerk user sync)
-- ----------------------------------------------------------------------------
-- Clerk may deliver a webhook more than once or out of order. We record each
-- delivery by its unique event id (Svix message id) so it is processed once.
CREATE TABLE clerk_webhook_events (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id     TEXT        NOT NULL UNIQUE,   -- Svix message / Clerk event id (idempotency key)
    event_type   TEXT        NOT NULL,          -- e.g. user.created, user.updated, user.deleted
    payload      JSONB       NOT NULL,
    processed_at TIMESTAMPTZ,                    -- NULL until handled
    error        TEXT,                           -- last processing error, if any
    received_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_clerk_webhook_events_unprocessed ON clerk_webhook_events (received_at) WHERE processed_at IS NULL;

-- ----------------------------------------------------------------------------
-- galleries  (top-level asset space; each user gets one default gallery)
-- ----------------------------------------------------------------------------
-- Quota is tracked per gallery. Physical capacity is spread across multiple
-- storage accounts per layer (see storage_providers) to work around per-account
-- free-tier limits.
CREATE TABLE galleries (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT        NOT NULL,
    description   TEXT,
    is_default    BOOLEAN     NOT NULL DEFAULT FALSE, -- the auto-created gallery for a user
    storage_quota BIGINT      NOT NULL DEFAULT 5368709120, -- 5 GB; per-gallery limit
    storage_used  BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_galleries_owner_id ON galleries (owner_id);
-- at most one default gallery per user
CREATE UNIQUE INDEX idx_galleries_one_default ON galleries (owner_id) WHERE is_default;

CREATE TRIGGER trg_galleries_updated_at
    BEFORE UPDATE ON galleries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- gallery_members  (who can access a gallery, and at what local role)
-- ----------------------------------------------------------------------------
-- The gallery owner also gets a row here with role 'owner' (created by the app)
-- so access checks can be a single membership lookup.
CREATE TABLE gallery_members (
    gallery_id BIGINT      NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    role       member_role NOT NULL DEFAULT 'viewer',
    invited_by BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (gallery_id, user_id)
);

CREATE INDEX idx_gallery_members_user_id ON gallery_members (user_id);

-- ----------------------------------------------------------------------------
-- storage_providers  (GLOBAL storage accounts, managed by admins)
-- ----------------------------------------------------------------------------
-- Not owned by end users. Access to manage is gated by the 'storage' RBAC
-- permission. Each account belongs to exactly one layer.
CREATE TABLE storage_providers (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    layer       storage_layer NOT NULL,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL CHECK (type IN ('cloudinary', 'imagekit', 'r2', 'gcs')),
    credentials JSONB       NOT NULL,
    quota       BIGINT,                           -- NULL = provider has no fixed quota
    used        BIGINT      NOT NULL DEFAULT 0,
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
    created_by  BIGINT      REFERENCES users(id) ON DELETE SET NULL, -- audit: who added it
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_storage_providers_layer     ON storage_providers (layer);
CREATE INDEX idx_storage_providers_is_active ON storage_providers (is_active);

CREATE TRIGGER trg_storage_providers_updated_at
    BEFORE UPDATE ON storage_providers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- assets  (logical asset records; metadata is the source of truth)
-- ----------------------------------------------------------------------------
-- An asset lives in exactly one gallery. uploaded_by is the contributor and is
-- used for the RBAC 'own' scope. Deduplication is per gallery. Soft delete
-- (deleted_at) moves an asset to the trash without losing its rows.
CREATE TABLE assets (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v7(),
    gallery_id  BIGINT      NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
    uploaded_by BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    name        TEXT        NOT NULL,
    type        TEXT        NOT NULL CHECK (type IN ('image', 'video', 'document', 'archive', 'file')),
    mime_type   TEXT        NOT NULL,
    size        BIGINT      NOT NULL,
    hash        TEXT        NOT NULL,             -- SHA-256, used for dedup
    metadata    JSONB,
    deleted_at  TIMESTAMPTZ,                       -- NULL = active; set = in trash
    deleted_by  BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_assets_gallery_id  ON assets (gallery_id);
CREATE INDEX idx_assets_uploaded_by ON assets (uploaded_by);
CREATE INDEX idx_assets_type        ON assets (type);
CREATE INDEX idx_assets_created_at  ON assets (created_at DESC);
-- fast listing of live (non-trashed) assets in a gallery
CREATE INDEX idx_assets_gallery_active ON assets (gallery_id) WHERE deleted_at IS NULL;
-- dedup is scoped per gallery and ignores trashed rows, so a re-upload after
-- deletion is allowed
CREATE UNIQUE INDEX idx_assets_gallery_hash ON assets (gallery_id, hash) WHERE deleted_at IS NULL;

CREATE TRIGGER trg_assets_updated_at
    BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- albums  (grouping of assets within a gallery)
-- ----------------------------------------------------------------------------
CREATE TABLE albums (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gallery_id     BIGINT      NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
    owner_id       BIGINT      NOT NULL REFERENCES users(id)     ON DELETE CASCADE,
    name           TEXT        NOT NULL,
    description    TEXT,
    cover_asset_id UUID        REFERENCES assets(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_albums_gallery_id ON albums (gallery_id);
CREATE INDEX idx_albums_owner_id   ON albums (owner_id);

CREATE TRIGGER trg_albums_updated_at
    BEFORE UPDATE ON albums
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- album_members  (album owner can invite users to the album)
-- ----------------------------------------------------------------------------
CREATE TABLE album_members (
    album_id   BIGINT      NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    user_id    BIGINT      NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    role       member_role NOT NULL DEFAULT 'viewer',
    invited_by BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (album_id, user_id)
);

CREATE INDEX idx_album_members_user_id ON album_members (user_id);

-- ----------------------------------------------------------------------------
-- invitations  (invite a user by email to a gallery or album)
-- ----------------------------------------------------------------------------
-- Targets exactly one of gallery_id / album_id (enforced by CHECK). The invitee
-- may not be a user yet: they receive a link with `token`, and on acceptance
-- (after Clerk sign-in) a membership row is created and accepted_user_id is set.
CREATE TABLE invitations (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gallery_id       BIGINT      REFERENCES galleries(id) ON DELETE CASCADE,
    album_id         BIGINT      REFERENCES albums(id)    ON DELETE CASCADE,
    email            TEXT        NOT NULL,
    role             member_role NOT NULL DEFAULT 'viewer',
    token            TEXT        NOT NULL UNIQUE,   -- opaque token for the invite link
    status           invitation_status NOT NULL DEFAULT 'pending',
    invited_by       BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    accepted_user_id BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    expires_at       TIMESTAMPTZ,
    accepted_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- exactly one target
    CONSTRAINT invitations_one_target CHECK (
        (gallery_id IS NOT NULL)::int + (album_id IS NOT NULL)::int = 1
    )
);

CREATE INDEX idx_invitations_email ON invitations (email);
-- at most one pending invite per (target, email)
CREATE UNIQUE INDEX idx_invitations_pending_gallery ON invitations (gallery_id, email)
    WHERE status = 'pending' AND gallery_id IS NOT NULL;
CREATE UNIQUE INDEX idx_invitations_pending_album ON invitations (album_id, email)
    WHERE status = 'pending' AND album_id IS NOT NULL;

CREATE TRIGGER trg_invitations_updated_at
    BEFORE UPDATE ON invitations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- album_assets  (many-to-many: asset <-> album)
-- ----------------------------------------------------------------------------
CREATE TABLE album_assets (
    album_id   BIGINT      NOT NULL REFERENCES albums(id) ON DELETE CASCADE,
    asset_id   UUID        NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    added_by   BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    sort_order INTEGER     NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (album_id, asset_id)
);

CREATE INDEX idx_album_assets_asset_id ON album_assets (asset_id);

-- ----------------------------------------------------------------------------
-- tags  (per-gallery tag vocabulary)
-- ----------------------------------------------------------------------------
CREATE TABLE tags (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    gallery_id BIGINT      NOT NULL REFERENCES galleries(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    created_by BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (gallery_id, name)
);

CREATE TRIGGER trg_tags_updated_at
    BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- asset_tags  (many-to-many: asset <-> tag)
-- ----------------------------------------------------------------------------
CREATE TABLE asset_tags (
    asset_id   UUID        NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    tag_id     BIGINT      NOT NULL REFERENCES tags(id)   ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (asset_id, tag_id)
);

CREATE INDEX idx_asset_tags_tag_id ON asset_tags (tag_id);

-- ----------------------------------------------------------------------------
-- storage_locations  (physical copies of an asset across layers)
-- ----------------------------------------------------------------------------
-- Each asset should end up with at least one 'serving' and one 'archive' copy.
-- provider_id uses ON DELETE RESTRICT: an account still hosting files cannot be
-- deleted (deactivate it via is_active instead).
CREATE TABLE storage_locations (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v7(),
    asset_id     UUID          NOT NULL REFERENCES assets(id)            ON DELETE CASCADE,
    provider_id  BIGINT        NOT NULL REFERENCES storage_providers(id) ON DELETE RESTRICT,
    layer        storage_layer NOT NULL,          -- mirrors the provider's layer
    provider_key TEXT          NOT NULL,          -- key/path within the provider
    url          TEXT,                            -- public URL (may be NULL for archive)
    status       location_status NOT NULL DEFAULT 'pending',
    metadata     JSONB,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_storage_locations_asset_id    ON storage_locations (asset_id);
CREATE INDEX idx_storage_locations_provider_id ON storage_locations (provider_id);
CREATE INDEX idx_storage_locations_asset_layer ON storage_locations (asset_id, layer);
CREATE INDEX idx_storage_locations_status      ON storage_locations (status);
-- one copy of an asset per provider account
CREATE UNIQUE INDEX idx_storage_locations_asset_provider ON storage_locations (asset_id, provider_id);

CREATE TRIGGER trg_storage_locations_updated_at
    BEFORE UPDATE ON storage_locations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- archive_sync_jobs  (async replication of an asset into a target layer)
-- ----------------------------------------------------------------------------
-- Uploads land on the serving layer synchronously; replication to the archive
-- layer runs in the background. One job drives one asset->layer replication,
-- with retry bookkeeping. On success it creates/updates the storage_location
-- and marks itself 'completed'. provider_id is filled once an account is chosen
-- (account election is a backlog concern).
CREATE TABLE archive_sync_jobs (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    asset_id     UUID          NOT NULL REFERENCES assets(id)            ON DELETE CASCADE,
    target_layer storage_layer NOT NULL DEFAULT 'archive',
    provider_id  BIGINT        REFERENCES storage_providers(id) ON DELETE SET NULL,
    status       job_status    NOT NULL DEFAULT 'pending',
    attempts     INTEGER       NOT NULL DEFAULT 0,
    max_attempts INTEGER       NOT NULL DEFAULT 5,
    last_error   TEXT,
    next_retry_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_archive_sync_jobs_asset_id ON archive_sync_jobs (asset_id);
-- worker queue: pick up due, runnable jobs in order
CREATE INDEX idx_archive_sync_jobs_queue ON archive_sync_jobs (next_retry_at)
    WHERE status IN ('pending', 'failed');
-- at most one open job per (asset, target layer)
CREATE UNIQUE INDEX idx_archive_sync_jobs_open ON archive_sync_jobs (asset_id, target_layer)
    WHERE status IN ('pending', 'running');

CREATE TRIGGER trg_archive_sync_jobs_updated_at
    BEFORE UPDATE ON archive_sync_jobs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ----------------------------------------------------------------------------
-- storage_account_usage  (summary per storage account)
-- ----------------------------------------------------------------------------
-- Convenience view for the storage-management UI: usage + live copy counts.
CREATE VIEW storage_account_usage AS
SELECT
    sp.id,
    sp.name,
    sp.layer,
    sp.type,
    sp.is_active,
    sp.quota,
    sp.used,
    CASE WHEN sp.quota IS NULL OR sp.quota = 0
         THEN NULL
         ELSE round((sp.used::numeric / sp.quota) * 100, 2)
    END AS used_percent,
    count(sl.id)                                            AS location_count,
    count(sl.id) FILTER (WHERE sl.status = 'stored')        AS stored_count,
    count(sl.id) FILTER (WHERE sl.status = 'pending')       AS pending_count,
    count(sl.id) FILTER (WHERE sl.status = 'failed')        AS failed_count
FROM storage_providers sp
LEFT JOIN storage_locations sl ON sl.provider_id = sp.id
GROUP BY sp.id;
