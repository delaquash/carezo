CREATE TABLE IF NOT EXISTS notifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title             VARCHAR(255) NOT NULL,
    message           TEXT NOT NULL,

    type              VARCHAR(50) NOT NULL,
 

    data              JSONB,
 
    is_read           BOOLEAN NOT NULL DEFAULT false,
 
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
 
-- Index on user_id so fetching a user's notifications is fast
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
 
-- Index on is_read so unread count query is fast
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read);