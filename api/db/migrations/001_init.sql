CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS card_definitions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    base_hp INT NOT NULL,
    base_damage INT NOT NULL,
    base_durability INT NOT NULL,
    rarity TEXT NOT NULL,
    effects JSONB DEFAULT '[]',
    is_fuel BOOLEAN DEFAULT FALSE,
    spawns TEXT
);

CREATE TABLE IF NOT EXISTS user_cards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT REFERENCES users(id),
    card_id TEXT REFERENCES card_definitions(id),
    quality INT DEFAULT 1,
    current_hp INT,
    current_durability INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_deck (
    user_id BIGINT REFERENCES users(id),
    slot INT NOT NULL,
    user_card_id UUID REFERENCES user_cards(id),
    PRIMARY KEY (user_id, slot)
);

CREATE TABLE IF NOT EXISTS user_items (
    user_id BIGINT REFERENCES users(id),
    item_type TEXT NOT NULL,
    quantity INT DEFAULT 0,
    PRIMARY KEY (user_id, item_type)
);

CREATE TABLE IF NOT EXISTS battles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attacker_id BIGINT REFERENCES users(id),
    defender_id BIGINT,
    winner_id BIGINT,
    battle_log JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed card definitions
INSERT INTO card_definitions (id, name, base_hp, base_damage, base_durability, rarity, effects, is_fuel, spawns) VALUES
    ('venom', 'Venom', 2, 2, 3, 'common', '[]', false, NULL),
    ('thug', 'Thug', 3, 1, 3, 'common', '[]', false, NULL),
    ('goon', 'Goon', 2, 3, 2, 'common', '["deathrattle"]', false, 'cobblestone'),
    ('cobblestone', 'Cobblestone', 1, 1, 1, 'common', '["no_attack"]', false, NULL),
    ('enforcer', 'Enforcer', 4, 2, 4, 'uncommon', '[]', false, NULL),
    ('hitman', 'Hitman', 3, 4, 3, 'uncommon', '[]', false, NULL),
    ('spider-man', 'Spider-Man', 5, 3, 4, 'rare', '[]', false, NULL),
    ('capo', 'Capo', 4, 3, 5, 'rare', '[]', false, NULL),
    ('don', 'Don', 6, 5, 5, 'epic', '["rampage"]', false, NULL),
    ('mastermind', 'Mastermind', 5, 5, 6, 'epic', '[]', false, NULL),
    ('berserker', 'Berserker', 4, 6, 4, 'epic', '[]', false, NULL),
    ('godfather', 'Godfather', 8, 6, 8, 'legendary', '[]', false, NULL),
    ('fuel-card', 'Fuel Card', 1, 0, 1, 'common', '[]', true, NULL),
    ('pvp-assassin', 'PvP Assassin', 5, 3, 5, 'legendary', '[]', false, NULL),
    ('pvp-warlord', 'PvP Warlord', 4, 5, 6, 'legendary', '[]', false, NULL),
    ('pvp-champion', 'PvP Champion', 6, 4, 7, 'legendary', '[]', false, NULL)
ON CONFLICT (id) DO NOTHING;
