-- Sync contract, slice D1 (OpenSpec add-offline-sync, design decisions 1-3):
-- per-entity revision + per-profile monotonic counter for cheap polling.
-- Every write bumps the profile counter INSIDE its transaction.

ALTER TABLE routines
    ADD COLUMN revision BIGINT NOT NULL DEFAULT 1;

CREATE TABLE profile_sync_state (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    rev        BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
