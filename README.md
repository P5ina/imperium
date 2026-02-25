# Imperium — Telegram Card Battle Game

A Telegram card battle game with a Go REST API backend, Python Telegram bot, and Svelte Mini App.

## Architecture

- **api/** — Go REST API (gorilla/mux + pgx + PostgreSQL)
- **bot/** — Python Telegram bot (aiogram 3)
- **miniapp/** — Svelte + Vite Mini App for battle replays

## Quick Start

### 1. Configure environment

```bash
cp bot/.env.example bot/.env
# Edit bot/.env and set your BOT_TOKEN
```

### 2. Run with Docker Compose

```bash
# Set your bot token
export BOT_TOKEN=your_telegram_bot_token

# Start all services
docker compose up --build -d
```

This starts:
- PostgreSQL on port 5432
- Go API on port 8090
- Telegram bot

### 3. Mini App (development)

```bash
cd miniapp
npm install
npm run dev
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /users | Register/update user |
| GET | /users/:id/inventory | Get user's cards |
| GET | /users/:id/deck | Get user's deck |
| PUT | /users/:id/deck | Set deck (max 5 slots) |
| GET | /users/:id/items | Get user's keys/items |
| POST | /loot/case | Open a free case |
| POST | /loot/dungeon | Enter dungeon (requires key) |
| POST | /battle/pve | Fight PvE bot |
| POST | /battle/pvp | Fight another player |
| GET | /battle/:id | Get battle result + log |

## Game Mechanics

- **Cards** have HP, Damage, Durability, Rarity, and Effects
- **Deck** holds up to 5 cards
- **Battle**: front cards attack simultaneously each round
- **Effects**: deathrattle (spawn card on death), rampage (HP = round number), no_attack
- **Loot cases** drop common/uncommon cards and bronze keys
- **Dungeons** require keys and drop better cards + higher-tier keys

## Card Rarities

- ⚪ Common
- 🟢 Uncommon
- 🔵 Rare
- 🟣 Epic
- 🟠 Legendary
