package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"imperium/db"
	"imperium/models"
)

func GetCards(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT id, name, base_hp, base_damage, base_durability, rarity, effects, is_fuel, spawns
		 FROM card_definitions ORDER BY rarity, name`)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cards := []models.CardDefinition{}
	for rows.Next() {
		var cd models.CardDefinition
		var effectsRaw json.RawMessage
		var spawns *string
		err := rows.Scan(&cd.ID, &cd.Name, &cd.BaseHP, &cd.BaseDamage, &cd.BaseDurability, &cd.Rarity, &effectsRaw, &cd.IsFuel, &spawns)
		if err != nil {
			http.Error(w, `{"error":"scan error"}`, http.StatusInternalServerError)
			return
		}
		cd.ParseEffects(effectsRaw)
		cd.Spawns = spawns
		cards = append(cards, cd)
	}

	writeJSON(w, http.StatusOK, cards)
}
