CREATE TABLE IF NOT EXISTS message (
    id UUID PRIMARY KEY,
    sender_id UUID REFERENCES users,
    receiver_id UUID REFERENCES users,
    body TEXT NOT NULL,
    sent_at TIMESTAMP(0) WITH TIME ZONE,
    delivered_at TIMESTAMP(0) WITH TIME ZONE,
    read_at TIMESTAMP(0) WITH TIME ZONE,
    operation INT,
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_message_sender_receiver_sent_at ON message(sender_id, receiver_id, timestamp DESC);