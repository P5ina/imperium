package main

import (
	"log"
	"net/http"

	"imperium/config"
	"imperium/db"
	"imperium/handlers"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	if err := db.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrated successfully")

	r := mux.NewRouter()

	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Users
	r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id}/inventory", handlers.GetInventory).Methods("GET")
	r.HandleFunc("/users/{id}/deck", handlers.GetDeck).Methods("GET")
	r.HandleFunc("/users/{id}/deck", handlers.SetDeck).Methods("PUT")
	r.HandleFunc("/users/{id}/items", handlers.GetItems).Methods("GET")

	// Cards
	r.HandleFunc("/cards", handlers.GetCards).Methods("GET")

	// Loot
	r.HandleFunc("/loot/case", handlers.OpenCase).Methods("POST")
	r.HandleFunc("/loot/dungeon", handlers.EnterDungeon).Methods("POST")

	// Battle
	r.HandleFunc("/battle/pve", handlers.BattlePvE).Methods("POST")
	r.HandleFunc("/battle/pvp", handlers.BattlePvP).Methods("POST")
	r.HandleFunc("/battle/{id}", handlers.GetBattle).Methods("GET")

	log.Printf("Imperium API starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
