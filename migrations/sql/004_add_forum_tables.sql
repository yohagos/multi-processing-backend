CREATE TABLE IF NOT EXISTS forum_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT,
    avatar_url TEXT,
    is_online BOOLEAN DEFAULT false,
    last_seen TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS forum_channels(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT,
    is_private BOOLEAN DEFAULT false,
    is_direct_message BOOLEAN DEFAULT false,
    created_by UUID REFERENCES forum_users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channel_members(
    channel_id UUID REFERENCES forum_channels(id) ON DELETE CASCADE,
    user_id UUID REFERENCES forum_users(id) ON DELETE CASCADE,
    role TEXT DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (channel_id, user_id)
);

CREATE TABLE IF NOT EXISTS forum_messages(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES forum_channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES forum_users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type TEXT DEFAULT 'text',
    parent_message_id UUID REFERENCES forum_messages(id),
    is_edited BOOLEAN DEFAULT false,
    is_deleted BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS message_reactions(
    message_id UUID REFERENCES forum_messages(id) ON DELETE CASCADE,
    user_id UUID REFERENCES forum_users(id) ON DELETE CASCADE,
    emoji TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id, emoji)
);

CREATE TABLE IF NOT EXISTS message_read_status(
    message_id UUID REFERENCES forum_messages(id) ON DELETE CASCADE,
    user_id UUID REFERENCES forum_users(id) ON DELETE CASCADE,
    read_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

CREATE INDEX idx_forum_messages_channel ON forum_messages(channel_id, created_at DESC);
CREATE INDEX idx_forum_messages_user ON forum_messages(user_id);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);
CREATE INDEX idx_forum_users_online ON forum_users(is_online, last_seen);

CREATE OR REPLACE FUNCTION create_direct_message_channel(user1_id UUID, user2_id UUID)
RETURNS UUID AS $$
DECLARE 
    channel_id UUID;
    channel_name TEXT;
BEGIN
    SELECT STRING_AGG(u.username, ' & ' ORDER BY u.username)
    INTO channel_name
    FROM forum_users u
    WHERE u.id IN (user1_id, user2_id);
    
    INSERT INTO forum_channels (name, is_private, is_direct_message, created_by, created_at)
    VALUES (channel_name, true, true, user1_id, NOW())
    RETURNING id INTO channel_id;
    
    INSERT INTO channel_members (channel_id, user_id, role, joined_at) VALUES
    (channel_id, user1_id, 'member', NOW()),
    (channel_id, user2_id, 'member', NOW());
    
    RETURN channel_id;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE forum_users ALTER COLUMN avatar_url SET DEFAULT '';
