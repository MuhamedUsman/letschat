CREATE TABLE IF NOT EXISTS conversation (
    sender_id UUID REFERENCES users ON DELETE SET NULL, -- conversation initiator
    receiver_id UUID REFERENCES users ON DELETE SET NULL,
    PRIMARY KEY (sender_id, receiver_id) -- composite primary key
);

CREATE INDEX idx_sender_id ON conversation(sender_id);
CREATE INDEX idx_receiver_id ON conversation(receiver_id);