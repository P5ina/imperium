package models

import "time"

type Battle struct {
	ID         string     `json:"id"`
	AttackerID int64      `json:"attacker_id"`
	DefenderID *int64     `json:"defender_id,omitempty"`
	WinnerID   *int64     `json:"winner_id,omitempty"`
	BattleLog  *BattleLog `json:"battle_log,omitempty"`
	CreatedAt  string     `json:"created_at"`
}

type BattleLog struct {
	Entries           []BattleLogEntry `json:"entries"`
	Winner            string           `json:"winner"`
	TotalRounds       int              `json:"total_rounds"`
	AttackerRemaining int              `json:"attacker_remaining"`
	DefenderRemaining int              `json:"defender_remaining"`
}

type BattleLogEntry struct {
	Round        int               `json:"round"`
	TurnSide     string            `json:"turn_side"`
	Timestamp    time.Time         `json:"timestamp"`
	DurationMs   int64             `json:"duration_ms"`
	Actions      []BattleLogAction `json:"actions"`
	AttackerDeck []BattleCard      `json:"attacker_deck"`
	DefenderDeck []BattleCard      `json:"defender_deck"`
}

type BattleLogAction struct {
	Type string `json:"type"`

	// attack fields
	AttackerID *int64 `json:"attacker_id,omitempty"`
	DefenderID *int64 `json:"defender_id,omitempty"`
	Damage     *int16 `json:"damage,omitempty"`

	// spawn_card fields
	Side        *string     `json:"side,omitempty"`
	SpawnedCard *BattleCard `json:"spawned_card,omitempty"`

	// card_died fields
	DiedCardID *int64  `json:"died_card_id,omitempty"`
	DiedSide   *string `json:"died_side,omitempty"`
}

type BattleCard struct {
	ID        int64    `json:"id"`
	CardID    string   `json:"card_id"`
	Name      string   `json:"name"`
	CurrentHP int16    `json:"current_hp"`
	MaxHP     int16    `json:"max_hp"`
	Attack    int16    `json:"attack"`
	Rarity    string   `json:"rarity"`
	Effects   []string `json:"effects"`
}
