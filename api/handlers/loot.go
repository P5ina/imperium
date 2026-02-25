package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"

	"imperium/db"
)

type LootRequest struct {
	UserID int64 `json:"user_id"`
}

type DungeonRequest struct {
	UserID  int64  `json:"user_id"`
	Dungeon string `json:"dungeon"`
}

type LootResult struct {
	Type    string `json:"type"`
	CardID  string `json:"card_id,omitempty"`
	ItemID  string `json:"item_id,omitempty"`
	Rarity  string `json:"rarity,omitempty"`
	Quality int    `json:"quality,omitempty"`
}

func OpenCase(w http.ResponseWriter, r *http.Request) {
	var req LootRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	// Verify user exists
	var exists bool
	db.Pool.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, req.UserID).Scan(&exists)
	if !exists {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	results := []LootResult{}

	roll := rand.Float64()
	if roll < 0.70 {
		// Common cards: venom, thug, goon
		cards := []string{"venom", "thug", "goon"}
		cardID := cards[rand.Intn(len(cards))]
		result, err := giveCard(req.UserID, cardID)
		if err != nil {
			http.Error(w, `{"error":"give card error: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		results = append(results, *result)
	} else if roll < 0.85 {
		// Uncommon cards: enforcer, hitman
		cards := []string{"enforcer", "hitman"}
		cardID := cards[rand.Intn(len(cards))]
		result, err := giveCard(req.UserID, cardID)
		if err != nil {
			http.Error(w, `{"error":"give card error: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		results = append(results, *result)
	} else {
		// Bronze key
		err := giveItem(req.UserID, "bronze_key", 1)
		if err != nil {
			http.Error(w, `{"error":"give item error: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		results = append(results, LootResult{Type: "item", ItemID: "bronze_key"})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

func EnterDungeon(w http.ResponseWriter, r *http.Request) {
	var req DungeonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	// Determine required key
	var requiredKey string
	switch req.Dungeon {
	case "easy":
		requiredKey = "bronze_key"
	case "medium":
		requiredKey = "silver_key"
	case "hard":
		requiredKey = "gold_key"
	default:
		http.Error(w, `{"error":"invalid dungeon: easy, medium, hard"}`, http.StatusBadRequest)
		return
	}

	// Check and consume key
	var qty int
	err := db.Pool.QueryRow(context.Background(),
		`SELECT COALESCE((SELECT quantity FROM user_items WHERE user_id=$1 AND item_type=$2), 0)`,
		req.UserID, requiredKey).Scan(&qty)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	if qty < 1 {
		http.Error(w, `{"error":"not enough keys: need 1 `+requiredKey+`"}`, http.StatusBadRequest)
		return
	}

	_, err = db.Pool.Exec(context.Background(),
		`UPDATE user_items SET quantity = quantity - 1 WHERE user_id=$1 AND item_type=$2`,
		req.UserID, requiredKey)
	if err != nil {
		http.Error(w, `{"error":"consume key error"}`, http.StatusInternalServerError)
		return
	}

	results := []LootResult{}

	// Card drop based on dungeon
	roll := rand.Float64()
	switch req.Dungeon {
	case "easy":
		cards := []string{"enforcer", "hitman", "spider-man", "capo"}
		if roll < 0.80 {
			cardID := cards[rand.Intn(len(cards))]
			result, err := giveCard(req.UserID, cardID)
			if err != nil {
				http.Error(w, `{"error":"give card error"}`, http.StatusInternalServerError)
				return
			}
			results = append(results, *result)
		}
		// 100% bonus silver key
		if err := giveItem(req.UserID, "silver_key", 1); err != nil {
			http.Error(w, `{"error":"give item error"}`, http.StatusInternalServerError)
			return
		}
		results = append(results, LootResult{Type: "item", ItemID: "silver_key"})

	case "medium":
		cards := []string{"spider-man", "capo", "don", "mastermind", "berserker"}
		if roll < 0.80 {
			cardID := cards[rand.Intn(len(cards))]
			result, err := giveCard(req.UserID, cardID)
			if err != nil {
				http.Error(w, `{"error":"give card error"}`, http.StatusInternalServerError)
				return
			}
			results = append(results, *result)
		}
		// 100% bonus gold key
		if err := giveItem(req.UserID, "gold_key", 1); err != nil {
			http.Error(w, `{"error":"give item error"}`, http.StatusInternalServerError)
			return
		}
		results = append(results, LootResult{Type: "item", ItemID: "gold_key"})

	case "hard":
		if roll < 0.80 {
			cards := []string{"don", "mastermind", "godfather"}
			cardID := cards[rand.Intn(len(cards))]
			result, err := giveCard(req.UserID, cardID)
			if err != nil {
				http.Error(w, `{"error":"give card error"}`, http.StatusInternalServerError)
				return
			}
			results = append(results, *result)
		} else {
			pvpCards := []string{"pvp-assassin", "pvp-warlord", "pvp-champion"}
			cardID := pvpCards[rand.Intn(len(pvpCards))]
			result, err := giveCard(req.UserID, cardID)
			if err != nil {
				http.Error(w, `{"error":"give card error"}`, http.StatusInternalServerError)
				return
			}
			results = append(results, *result)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"results": results})
}

func giveCard(userID int64, cardID string) (*LootResult, error) {
	var rarity string
	var baseHP, baseDur int
	err := db.Pool.QueryRow(context.Background(),
		`SELECT rarity, base_hp, base_durability FROM card_definitions WHERE id=$1`, cardID).Scan(&rarity, &baseHP, &baseDur)
	if err != nil {
		return nil, err
	}

	quality := 1 + rand.Intn(3) // 1-3 for most, can be higher later

	_, err = db.Pool.Exec(context.Background(),
		`INSERT INTO user_cards (user_id, card_id, quality, current_hp, current_durability) VALUES ($1, $2, $3, $4, $5)`,
		userID, cardID, quality, baseHP, baseDur)
	if err != nil {
		return nil, err
	}

	return &LootResult{
		Type:    "card",
		CardID:  cardID,
		Rarity:  rarity,
		Quality: quality,
	}, nil
}

func giveItem(userID int64, itemType string, qty int) error {
	_, err := db.Pool.Exec(context.Background(),
		`INSERT INTO user_items (user_id, item_type, quantity) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, item_type) DO UPDATE SET quantity = user_items.quantity + $3`,
		userID, itemType, qty)
	return err
}
