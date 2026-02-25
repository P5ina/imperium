package models

import "encoding/json"

type CardDefinition struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	BaseHP         int      `json:"base_hp"`
	BaseDamage     int      `json:"base_damage"`
	BaseDurability int      `json:"base_durability"`
	Rarity         string   `json:"rarity"`
	Effects        []string `json:"effects"`
	IsFuel         bool     `json:"is_fuel"`
	Spawns         *string  `json:"spawns,omitempty"`
}

type UserCard struct {
	ID                string          `json:"id"`
	UserID            int64           `json:"user_id"`
	CardID            string          `json:"card_id"`
	Quality           int             `json:"quality"`
	CurrentHP         int             `json:"current_hp"`
	CurrentDurability int             `json:"current_durability"`
	CreatedAt         string          `json:"created_at"`
	Definition        *CardDefinition `json:"definition,omitempty"`
}

func (cd *CardDefinition) ParseEffects(raw json.RawMessage) error {
	return json.Unmarshal(raw, &cd.Effects)
}
