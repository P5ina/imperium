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

type CreateUserRequest struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.Pool.QueryRow(context.Background(),
		`INSERT INTO users (id, username) VALUES ($1, $2)
		 ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username
		 RETURNING id, username, created_at`,
		req.ID, req.Username,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		http.Error(w, `{"error":"db error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func GetInventory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT uc.id, uc.user_id, uc.card_id, uc.quality, uc.current_hp, uc.current_durability, uc.created_at,
		        cd.id, cd.name, cd.base_hp, cd.base_damage, cd.base_durability, cd.rarity, cd.effects, cd.is_fuel, cd.spawns
		 FROM user_cards uc
		 JOIN card_definitions cd ON cd.id = uc.card_id
		 WHERE uc.user_id = $1
		 ORDER BY uc.created_at DESC`, userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cards := []models.UserCard{}
	for rows.Next() {
		var uc models.UserCard
		var cd models.CardDefinition
		var effectsRaw json.RawMessage
		var spawns *string
		err := rows.Scan(&uc.ID, &uc.UserID, &uc.CardID, &uc.Quality, &uc.CurrentHP, &uc.CurrentDurability, &uc.CreatedAt,
			&cd.ID, &cd.Name, &cd.BaseHP, &cd.BaseDamage, &cd.BaseDurability, &cd.Rarity, &effectsRaw, &cd.IsFuel, &spawns)
		if err != nil {
			http.Error(w, `{"error":"scan error"}`, http.StatusInternalServerError)
			return
		}
		cd.ParseEffects(effectsRaw)
		cd.Spawns = spawns
		uc.Definition = &cd
		cards = append(cards, uc)
	}

	writeJSON(w, http.StatusOK, cards)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
