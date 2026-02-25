package handlers

import (
	"context"
	"net/http"
	"strconv"

	"imperium/db"
	"imperium/models"

	"github.com/gorilla/mux"
)

func GetItems(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusBadRequest)
		return
	}

	rows, err := db.Pool.Query(context.Background(),
		`SELECT user_id, item_type, quantity FROM user_items WHERE user_id = $1`, userID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := []models.UserItem{}
	for rows.Next() {
		var item models.UserItem
		if err := rows.Scan(&item.UserID, &item.ItemType, &item.Quantity); err != nil {
			http.Error(w, `{"error":"scan error"}`, http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, items)
}
