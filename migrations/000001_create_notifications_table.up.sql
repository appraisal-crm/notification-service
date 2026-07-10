CREATE TABLE notification_channels (
    id TEXT PRIMARY KEY
);

CREATE TABLE notification_statuses (
    id TEXT PRIMARY KEY
);

INSERT INTO notification_channels (id) VALUES
    ('email'),
    ('sms'),
    ('in_app'),
    ('push');

INSERT INTO notification_statuses (id) VALUES
    ('pending'),
    ('sent'),
    ('failed');

CREATE TABLE notifications (
    id                UUID PRIMARY KEY,
    recipient_id      UUID,
    channel           TEXT NOT NULL REFERENCES notification_channels(id),
    recipient_address TEXT,
    event_type        TEXT NOT NULL,
    request_id        UUID,
    subject           TEXT,
    body              TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'pending' REFERENCES notification_statuses(id),
    created_at        TIMESTAMPTZ NOT NULL,
    updated_at        TIMESTAMPTZ NOT NULL,
    sent_at           TIMESTAMPTZ,
    read_at           TIMESTAMPTZ
);

CREATE INDEX idx_notifications_recipient_id ON notifications (recipient_id);
CREATE INDEX idx_notifications_status ON notifications (status);
CREATE INDEX idx_notifications_request_id ON notifications (request_id);
