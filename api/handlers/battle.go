package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"imperium/db"
	"imperium/engine"
	"imperium/models"

	"github.com/gorilla/mux"
)

type PvERequest struct {
	UserID  int64  `json:"user_id"`
	Dungeon string `json:"dungeon"`
}

type PvPRequest struct {
	AttackerID int64 `json:"attacker_id"`
	DefenderID int64 `json:"defender_id"`
}

var botDecks = map[string][]string{
	"easy":   {"thug", "thug", "goon", "enforcer", "cobblestone"},
	"medium": {"enforcer", "hitman", "spider-man", "capo", "don"},
	"hard":   {"don", "mastermind", "berserker", "godfather", "pvp-warlord"},
}

func BattlePvE(w http.ResponseWriter, r *http.Request) {
	var req PvERequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	botDeck, ok := botDecks[req.Dungeon]
	if !ok {
		http.Error(w, `{"error":"invalid dungeon"}`, http.StatusBadRequest)
		return
	}

	attackerDeck, err := loadUserDeck(req.UserID)
	if err != nil {
		http.Error(w, `{"error":"load deck error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if len(attackerDeck) == 0 {
		http.Error(w, `{"error":"deck is empty, set your deck first"}`, http.StatusBadRequest)
		return
	}

	defenderDeck, err := buildBotDeck(botDeck)
	if err != nil {
		http.Error(w, `{"error":"build bot deck error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	battleLog := engine.RunBattle(attackerDeck, defenderDeck)

	var winnerID *int64
	if battleLog.Winner == "attacker" {
		winnerID = &req.UserID
	}

	var pveID int64 = -1
	var battleID string
	logJSON, _ := json.Marshal(battleLog)
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO battles (attacker_id, defender_id, winner_id, battle_log)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		req.UserID, pveID, winnerID, logJSON).Scan(&battleID)
	if err != nil {
		http.Error(w, `{"error":"save battle error: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"battle_id":  battleID,
		"winner":     battleLog.Winner,
		"rounds":     battleLog.TotalRounds,
		"battle_log": battleLog,
	})
}

func BattlePvP(w http.ResponseWriter, r *http.Request) {
	var req PvPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if req.AttackerID == req.DefenderID {
		http.Error(w, `{"error":"cannot fight yourself"}`, http.StatusBadRequest)
		return
	}

	attackerDeck, err := loadUserDeck(req.AttackerID)
	if err != nil {
		http.Error(w, `{"error":"load attacker deck: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if len(attackerDeck) == 0 {
		http.Error(w, `{"error":"attacker deck is empty"}`, http.StatusBadRequest)
		return
	}

	defenderDeck, err := loadUserDeck(req.DefenderID)
	if err != nil {
		http.Error(w, `{"error":"load defender deck: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if len(defenderDeck) == 0 {
		http.Error(w, `{"error":"defender deck is empty"}`, http.StatusBadRequest)
		return
	}

	battleLog := engine.RunBattle(attackerDeck, defenderDeck)

	var winnerID *int64
	switch battleLog.Winner {
	case "attacker":
		winnerID = &req.AttackerID
	case "defender":
		winnerID = &req.DefenderID
	}

	var battleID string
	logJSON, _ := json.Marshal(battleLog)
	err = db.Pool.QueryRow(context.Background(),
		`INSERT INTO battles (attacker_id, defender_id, winner_id, battle_log)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		req.AttackerID, req.DefenderID, winnerID, logJSON).Scan(&battleID)
	if err != nil {
		http.Error(w, `{"error":"save battle error"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"battle_id":  battleID,
		"winner":     battleLog.Winner,
		"rounds":     battleLog.TotalRounds,
		"battle_log": battleLog,
	})
}

func GetBattle(w http.ResponseWriter, r *http.Request) {
	battleID := mux.Vars(r)["id"]

	var battle models.Battle
	var logRaw json.RawMessage
	err := db.Pool.QueryRow(context.Background(),
		`SELECT id, attacker_id, defender_id, winner_id, battle_log, created_at
		 FROM battles WHERE id = $1`, battleID).Scan(
		&battle.ID, &battle.AttackerID, &battle.DefenderID, &battle.WinnerID, &logRaw, &battle.CreatedAt)
	if err != nil {
		http.Error(w, `{"error":"battle not found"}`, http.StatusNotFound)
		return
	}

	if logRaw != nil {
		var bl models.BattleLog
		json.Unmarshal(logRaw, &bl)
		battle.BattleLog = &bl
	}

	writeJSON(w, http.StatusOK, battle)
}

func loadUserDeck(userID int64) ([]models.BattleCard, error) {
	rows, err := db.Pool.Query(context.Background(),
		`SELECT uc.id, uc.card_id, cd.name, cd.base_hp, cd.base_damage, cd.rarity, cd.effects, cd.spawns
		 FROM user_deck ud
		 JOIN user_cards uc ON uc.id = ud.user_card_id
		 JOIN card_definitions cd ON cd.id = uc.card_id
		 WHERE ud.user_id = $1
		 ORDER BY ud.slot`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deck []models.BattleCard
	var idCounter int64
	for rows.Next() {
		idCounter++
		var dbID int64
		var cardID, name, rarity string
		var baseHP, baseDamage int
		var effectsRaw json.RawMessage
		var spawns *string

		err := rows.Scan(&dbID, &cardID, &name, &baseHP, &baseDamage, &rarity, &effectsRaw, &spawns)
		if err != nil {
			return nil, err
		}

		var effects []string
		json.Unmarshal(effectsRaw, &effects)

		if spawns != nil {
			effects = append(effects, "spawns:"+*spawns)
		}

		deck = append(deck, models.BattleCard{
			ID:        idCounter,
			CardID:    cardID,
			Name:      name,
			CurrentHP: int16(baseHP),
			MaxHP:     int16(baseHP),
			Attack:    int16(baseDamage),
			Rarity:    rarity,
			Effects:   effects,
		})
	}
	return deck, nil
}

func buildBotDeck(cardIDs []string) ([]models.BattleCard, error) {
	var deck []models.BattleCard
	for i, cardID := range cardIDs {
		var name, rarity string
		var baseHP, baseDamage int
		var effectsRaw json.RawMessage
		var spawns *string

		err := db.Pool.QueryRow(context.Background(),
			`SELECT name, base_hp, base_damage, rarity, effects, spawns FROM card_definitions WHERE id=$1`, cardID).Scan(
			&name, &baseHP, &baseDamage, &rarity, &effectsRaw, &spawns)
		if err != nil {
			return nil, err
		}

		var effects []string
		json.Unmarshal(effectsRaw, &effects)

		if spawns != nil {
			effects = append(effects, "spawns:"+*spawns)
		}

		deck = append(deck, models.BattleCard{
			ID:        int64(100 + i),
			CardID:    cardID,
			Name:      name,
			CurrentHP: int16(baseHP),
			MaxHP:     int16(baseHP),
			Attack:    int16(baseDamage),
			Rarity:    rarity,
			Effects:   effects,
		})
	}
	return deck, nil
}
