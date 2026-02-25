package engine

import (
	"imperium/models"
	"strconv"
	"strings"
	"time"
)

var spawnIDCounter int64 = 10000

func nextSpawnID() int64 {
	spawnIDCounter++
	return spawnIDCounter
}

func hasEffect(card *models.BattleCard, effect string) bool {
	for _, e := range card.Effects {
		if e == effect || strings.HasPrefix(e, effect+":") {
			return true
		}
	}
	return false
}

func getThornsDamage(card *models.BattleCard) int16 {
	for _, e := range card.Effects {
		if strings.HasPrefix(e, "thorns:") {
			parts := strings.SplitN(e, ":", 2)
			if len(parts) == 2 {
				v, err := strconv.Atoi(parts[1])
				if err == nil {
					return int16(v)
				}
			}
		}
	}
	return 0
}

func findTauntTarget(deck []models.BattleCard) int {
	for i := range deck {
		if hasEffect(&deck[i], "taunt") {
			return i
		}
	}
	return 0
}

func snapshotDeck(deck []models.BattleCard) []models.BattleCard {
	snap := make([]models.BattleCard, len(deck))
	for i, c := range deck {
		snap[i] = c
		snap[i].Effects = make([]string, len(c.Effects))
		copy(snap[i].Effects, c.Effects)
	}
	return snap
}

func applyRampage(deck []models.BattleCard) {
	for i := range deck {
		if hasEffect(&deck[i], "rampage") {
			deck[i].CurrentHP++
			deck[i].MaxHP++
		}
	}
}

func processDeathrattle(deck *[]models.BattleCard, idx int) *models.BattleCard {
	dead := (*deck)[idx]
	if !hasEffect(&dead, "deathrattle") {
		return nil
	}

	// Find spawns info from effects
	var spawnCardID string
	for _, e := range dead.Effects {
		if strings.HasPrefix(e, "spawns:") {
			spawnCardID = strings.TrimPrefix(e, "spawns:")
			break
		}
	}
	if spawnCardID == "" {
		// Default cobblestone spawn for deathrattle
		spawnCardID = "cobblestone"
	}

	spawned := models.BattleCard{
		ID:        nextSpawnID(),
		CardID:    spawnCardID,
		Name:      spawnCardID,
		CurrentHP: 1,
		MaxHP:     1,
		Attack:    1,
		Rarity:    "common",
		Effects:   []string{"no_attack"},
	}

	// Insert spawned card at front (replacing the dead card's position)
	newDeck := make([]models.BattleCard, 0, len(*deck))
	newDeck = append(newDeck, (*deck)[:idx]...)
	newDeck = append(newDeck, spawned)
	newDeck = append(newDeck, (*deck)[idx+1:]...)
	*deck = newDeck

	return &spawned
}

func removeCard(deck *[]models.BattleCard, idx int) {
	*deck = append((*deck)[:idx], (*deck)[idx+1:]...)
}

func RunBattle(attackerDeck, defenderDeck []models.BattleCard) models.BattleLog {
	log := models.BattleLog{}
	isDefenderTurn := true
	startTime := time.Now()

	for round := 1; round <= 2000; round++ {
		if len(attackerDeck) == 0 || len(defenderDeck) == 0 {
			break
		}

		var actions []models.BattleLogAction

		// a. Apply rampage to all cards on both sides
		applyRampage(attackerDeck)
		applyRampage(defenderDeck)

		// b. Determine active and passive sides
		var activeDeck *[]models.BattleCard
		var passiveDeck *[]models.BattleCard
		var turnSide string
		var activeSide, passiveSide string

		if isDefenderTurn {
			activeDeck = &defenderDeck
			passiveDeck = &attackerDeck
			turnSide = "defender"
			activeSide = "defender"
			passiveSide = "attacker"
		} else {
			activeDeck = &attackerDeck
			passiveDeck = &defenderDeck
			turnSide = "attacker"
			activeSide = "attacker"
			passiveSide = "defender"
		}

		// c. Active card = front card of active side
		activeCard := &(*activeDeck)[0]

		// d. Find target (taunt check)
		targetIdx := findTauntTarget(*passiveDeck)
		targetCard := &(*passiveDeck)[targetIdx]

		// f. Calculate damage
		damage := activeCard.Attack
		if hasEffect(activeCard, "no_attack") {
			damage = 0
		}

		// g. Apply damage
		if damage > 0 {
			targetCard.CurrentHP -= damage
			actions = append(actions, models.BattleLogAction{
				Type:       "attack",
				AttackerID: &activeCard.ID,
				DefenderID: &targetCard.ID,
				Damage:     &damage,
			})
		}

		// h. Check thorns on target
		thornsDmg := getThornsDamage(targetCard)
		if thornsDmg > 0 && damage > 0 {
			activeCard.CurrentHP -= thornsDmg
			actions = append(actions, models.BattleLogAction{
				Type:       "attack",
				AttackerID: &targetCard.ID,
				DefenderID: &activeCard.ID,
				Damage:     &thornsDmg,
			})
		}

		// i. Check if target died
		targetDied := targetCard.CurrentHP <= 0
		activeDied := activeCard.CurrentHP <= 0

		if targetDied {
			diedID := targetCard.ID
			actions = append(actions, models.BattleLogAction{
				Type:       "card_died",
				DiedCardID: &diedID,
				DiedSide:   &passiveSide,
			})

			spawned := processDeathrattle(passiveDeck, targetIdx)
			if spawned != nil {
				spawnSide := passiveSide
				actions = append(actions, models.BattleLogAction{
					Type:        "spawn_card",
					Side:        &spawnSide,
					SpawnedCard: spawned,
				})
			} else {
				removeCard(passiveDeck, targetIdx)
			}
		}

		// j. Check if active card died (from thorns)
		if activeDied {
			diedID := activeCard.ID
			actions = append(actions, models.BattleLogAction{
				Type:       "card_died",
				DiedCardID: &diedID,
				DiedSide:   &activeSide,
			})

			spawned := processDeathrattle(activeDeck, 0)
			if spawned != nil {
				spawnSide := activeSide
				actions = append(actions, models.BattleLogAction{
					Type:        "spawn_card",
					Side:        &spawnSide,
					SpawnedCard: spawned,
				})
			} else {
				removeCard(activeDeck, 0)
			}
		}

		// k. Snapshot both decks after deaths/spawns
		entry := models.BattleLogEntry{
			Round:        round,
			TurnSide:     turnSide,
			Timestamp:    startTime.Add(time.Duration(round-1) * 800 * time.Millisecond),
			DurationMs:   800,
			Actions:      actions,
			AttackerDeck: snapshotDeck(attackerDeck),
			DefenderDeck: snapshotDeck(defenderDeck),
		}

		log.Entries = append(log.Entries, entry)

		// m. Toggle turn
		isDefenderTurn = !isDefenderTurn
	}

	log.TotalRounds = len(log.Entries)
	log.AttackerRemaining = len(attackerDeck)
	log.DefenderRemaining = len(defenderDeck)

	if len(attackerDeck) > 0 && len(defenderDeck) == 0 {
		log.Winner = "attacker"
	} else if len(defenderDeck) > 0 && len(attackerDeck) == 0 {
		log.Winner = "defender"
	} else if len(attackerDeck) > 0 && len(defenderDeck) > 0 {
		// Both have cards — compare total HP
		atkHP := totalHP(attackerDeck)
		defHP := totalHP(defenderDeck)
		if atkHP > defHP {
			log.Winner = "attacker"
		} else if defHP > atkHP {
			log.Winner = "defender"
		} else {
			log.Winner = "tie"
		}
	} else {
		log.Winner = "tie"
	}

	return log
}

func totalHP(deck []models.BattleCard) int16 {
	var total int16
	for _, c := range deck {
		total += c.CurrentHP
	}
	return total
}
