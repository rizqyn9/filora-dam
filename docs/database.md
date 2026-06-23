# Database Design

## Principles

* Metadata first
* Provider agnostic
* Backup aware
* Simple MVP design

---

# Tables

## users

User accounts.

---

## folders

Hierarchical asset organization.

Supports nested folders.

---

## assets

Logical asset record.

Stores:

* ownership
* metadata
* checksum
* mime type
* size

---

## asset_versions

Asset version history.

Future-proofing for versioning.

---

## storage_accounts

Represents physical storage accounts.

Examples:

* ImageKit #1
* ImageKit #2
* Cloudinary #1
* R2 Main

Fields:

* provider
* quota
* usage
* status

---

## asset_objects

Physical file locations.

Maps assets to storage providers.

Contains:

* object key
* provider
* storage account
* checksum

---

## backup_objects

Backup metadata.

Contains:

* backup provider
* backup key
* backup status
* storage class

---

## upload_sessions

Tracks multipart uploads.

---

# Source of Truth

PostgreSQL is always the source of truth.

Cloud providers are not authoritative data sources.
