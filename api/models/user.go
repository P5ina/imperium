package models

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

type UserItem struct {
	UserID   int64  `json:"user_id"`
	ItemType string `json:"item_type"`
	Quantity int    `json:"quantity"`
}

type DeckSlot struct {
	UserID     int64  `json:"user_id"`
	Slot       int    `json:"slot"`
	UserCardID string `json:"user_card_id"`
}
