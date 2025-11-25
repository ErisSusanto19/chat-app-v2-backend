CREATE TABLE users (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    hashed_password TEXT NOT NULL,
    phone_number VARCHAR(20) UNIQUE,
    image TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contacts (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    alias_name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('registered', 'unregistered')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_user_id, email)
);

CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(100),
    image TEXT,
    description TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    last_message_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_conversations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) CHECK (role IN ('admin', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, conversation_id)
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'delivered', 'read')),
    status_changed_at JSONB,
    content JSONB NOT NULL,
    disappear_for_all BOOLEAN NOT NULL DEFAULT FALSE,
    is_edited BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE conversations ADD CONSTRAINT fk_last_message
FOREIGN KEY (last_message_id) REFERENCES messages(id) ON DELETE SET NULL;

CREATE INDEX ON contacts (owner_user_id);
CREATE INDEX ON contacts (contact_user_id);
CREATE INDEX ON user_conversations (user_id);
CREATE INDEX ON user_conversations (conversation_id);
CREATE INDEX ON messages (sender_id);
CREATE INDEX ON messages (conversation_id);

CREATE INDEX ON messages (conversation_id, created_at DESC);