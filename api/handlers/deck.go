package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"imperium/db"
	"imperium/models"

	"github.com/gorilla/mux"
)

func GetDeck(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT ud.user_id, ud.slot, ud.user_card_id,
		        uc.id, uc.user_id, uc.card_id, uc.quality, uc.current_hp, uc.current_durability, uc.created_at,
		        cd.id, cd.name, cd.base_hp, cd.base_damage, cd.base_durability, cd.rarity, cd.effects, cd.is_fuel, cd.spawns
		 FROM user_deck ud
		 JOIN user_cards uc ON uc.id = ud.user_card_id
		 JOIN card_definitions cd ON cd.id = uc.card_id
		 WHERE ud.user_id = $1
		 ORDER BY ud.slot`, userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type DeckEntry struct {
		Slot int             `json:"slot"`
		Card models.UserCard `json:"card"`
	}

	deck := []DeckEntry{}
	for rows.Next() {
		var ds models.DeckSlot
		var uc models.UserCard
		var cd models.CardDefinition
		var effectsRaw json.RawMessage
		var spawns *string
		err := rows.Scan(&ds.UserID, &ds.Slot, &ds.UserCardID,
			&uc.ID, &uc.UserID, &uc.CardID, &uc.Quality, &uc.CurrentHP, &uc.CurrentDurability, &uc.CreatedAt,
			&cd.ID, &cd.Name, &cd.BaseHP, &cd.BaseDamage, &cd.BaseDurability, &cd.Rarity, &effectsRaw, &cd.IsFuel, &spawns)
		if err != nil {
			http.Error(w, `{"error":"scan error"}`, http.StatusInternalServerError)
			return
		}
		cd.ParseEffects(effectsRaw)
		cd.Spawns = spawns
		uc.Definition = &cd
		deck = append(deck, DeckEntry{Slot: ds.Slot, Card: uc})
	}

	writeJSON(w, http.StatusOK, deck)
}

type SetDeckRequest struct {
	Slots []struct {
		Slot       int    `json:"slot"`
		UserCardID string `json:"user_card_id"`
	} `json:"slots"`
}

func SetDeck(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	var req SetDeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if len(req.Slots) > 5 {
		http.Error(w, `{"error":"max 5 deck slots"}`, http.StatusBadRequest)
		return
	}

	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		http.Error(w, `{"error":"tx error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `DELETE FROM user_deck WHERE user_id = $1`, userID)
	if err != nil {
		http.Error(w, `{"error":"clear deck error"}`, http.StatusInternalServerError)
		return
	}

	for _, s := range req.Slots {
		if s.Slot < 1 || s.Slot > 5 {
			http.Error(w, `{"error":"slot must be 1-5"}`, http.StatusBadRequest)
			return
		}
		// Verify card belongs to user
		var count int
		err := tx.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM user_cards WHERE id = $1 AND user_id = $2`, s.UserCardID, userID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, `{"error":"card not found in inventory"}`, http.StatusBadRequest)
			return
		}

		_, err = tx.Exec(context.Background(),
			`INSERT INTO user_deck (user_id, slot, user_card_id) VALUES ($1, $2, $3)`,
			userID, s.Slot, s.UserCardID)
		if err != nil {
			http.Error(w, `{"error":"insert deck error"}`, http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		http.Error(w, `{"error":"commit error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
