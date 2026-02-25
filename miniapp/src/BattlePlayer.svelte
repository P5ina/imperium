<script>
  import { onMount } from 'svelte';

  let { battleId } = $props();

  const API_URL = window.location.origin.includes('localhost')
    ? 'http://localhost:8090'
    : window.location.origin.replace(/:\d+$/, ':8090');

  let battleLog = $state(null);
  let currentEntryIdx = $state(-1);
  let playing = $state(false);
  let error = $state(null);
  let done = $state(false);
  let flashCardId = $state(null);
  let dyingCardId = $state(null);
  let spawningCardId = $state(null);

  let attackerDeck = $state([]);
  let defenderDeck = $state([]);

  const RARITY_COLORS = {
    common: '#888',
    uncommon: '#4caf50',
    rare: '#2196f3',
    epic: '#9c27b0',
    legendary: '#ff9800',
  };

  onMount(async () => {
    try {
      const resp = await fetch(`${API_URL}/battle/${battleId}`);
      if (!resp.ok) throw new Error('Battle not found');
      const data = await resp.json();
      battleLog = data.battle_log;
      if (battleLog && battleLog.entries && battleLog.entries.length > 0) {
        // Initialize decks from first entry
        attackerDeck = battleLog.entries[0].attacker_deck || [];
        defenderDeck = battleLog.entries[0].defender_deck || [];
        startPlayback();
      }
    } catch (e) {
      error = e.message;
    }
  });

  function startPlayback() {
    if (!battleLog || !battleLog.entries.length) return;
    playing = true;
    currentEntryIdx = 0;
    processEntry(0);
  }

  function processEntry(idx) {
    if (idx >= battleLog.entries.length) {
      done = true;
      playing = false;
      return;
    }

    currentEntryIdx = idx;
    const entry = battleLog.entries[idx];

    // Process actions for visual effects
    for (const action of entry.actions) {
      if (action.type === 'attack' && action.defender_id) {
        flashCardId = action.defender_id;
        setTimeout(() => { flashCardId = null; }, 400);
      }
      if (action.type === 'card_died' && action.died_card_id) {
        dyingCardId = action.died_card_id;
        setTimeout(() => { dyingCardId = null; }, 600);
      }
      if (action.type === 'spawn_card' && action.spawned_card) {
        spawningCardId = action.spawned_card.id;
        setTimeout(() => { spawningCardId = null; }, 500);
      }
    }

    // Update deck snapshots
    attackerDeck = entry.attacker_deck || [];
    defenderDeck = entry.defender_deck || [];

    setTimeout(() => {
      processEntry(idx + 1);
    }, 800);
  }

  function getHPPercent(hp, maxHP) {
    return Math.max(0, Math.min(100, (hp / Math.max(maxHP, 1)) * 100));
  }

  function currentEntry() {
    if (!battleLog || currentEntryIdx < 0 || currentEntryIdx >= battleLog.entries.length) return null;
    return battleLog.entries[currentEntryIdx];
  }
</script>

<div class="battle-player">
  {#if error}
    <div class="error">{error}</div>
  {:else if !battleLog}
    <div class="loading">Loading battle...</div>
  {:else}
    <div class="header">
      <h2>Round {currentEntry()?.round || 0} / {battleLog.total_rounds}</h2>
      {#if currentEntry()}
        <div class="turn-indicator" class:defender-turn={currentEntry().turn_side === 'defender'}>
          {currentEntry().turn_side === 'defender' ? 'Defender attacks' : 'Attacker attacks'}
        </div>
      {/if}
    </div>

    <div class="arena">
      <div class="side attacker-side">
        <div class="label">Attacker</div>
        <div class="card-stack">
          {#each attackerDeck as card, i (card.id)}
            <div
              class="card"
              class:flash={flashCardId === card.id}
              class:dying={dyingCardId === card.id}
              class:spawning={spawningCardId === card.id}
              class:front={i === 0}
              style="border-color: {RARITY_COLORS[card.rarity] || '#888'}"
            >
              <div class="card-name">{card.name}</div>
              <div class="hp-bar">
                <div class="hp-fill" style="width: {getHPPercent(card.current_hp, card.max_hp)}%"></div>
              </div>
              <div class="stats">
                <span>HP: {card.current_hp}/{card.max_hp}</span>
                <span>ATK: {card.attack}</span>
              </div>
              {#if card.effects && card.effects.length > 0}
                <div class="effects">
                  {#each card.effects.filter(e => !e.startsWith('spawns:')) as effect}
                    <span class="effect-tag">{effect}</span>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
          {#if attackerDeck.length === 0}
            <div class="card empty">No cards</div>
          {/if}
        </div>
      </div>

      <div class="vs">VS</div>

      <div class="side defender-side">
        <div class="label">Defender</div>
        <div class="card-stack">
          {#each defenderDeck as card, i (card.id)}
            <div
              class="card"
              class:flash={flashCardId === card.id}
              class:dying={dyingCardId === card.id}
              class:spawning={spawningCardId === card.id}
              class:front={i === 0}
              style="border-color: {RARITY_COLORS[card.rarity] || '#888'}"
            >
              <div class="card-name">{card.name}</div>
              <div class="hp-bar">
                <div class="hp-fill defender" style="width: {getHPPercent(card.current_hp, card.max_hp)}%"></div>
              </div>
              <div class="stats">
                <span>HP: {card.current_hp}/{card.max_hp}</span>
                <span>ATK: {card.attack}</span>
              </div>
              {#if card.effects && card.effects.length > 0}
                <div class="effects">
                  {#each card.effects.filter(e => !e.startsWith('spawns:')) as effect}
                    <span class="effect-tag">{effect}</span>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
          {#if defenderDeck.length === 0}
            <div class="card empty">No cards</div>
          {/if}
        </div>
      </div>
    </div>

    {#if currentEntry()}
      <div class="actions-log">
        {#each currentEntry().actions as action}
          <span class="action-tag {action.type}">
            {#if action.type === 'attack'}
              Attack: {action.damage} dmg
            {:else if action.type === 'card_died'}
              Card died
            {:else if action.type === 'spawn_card'}
              Spawned: {action.spawned_card?.name}
            {/if}
          </span>
        {/each}
      </div>
    {/if}

    {#if done}
      <div class="winner-banner" class:win={battleLog.winner === 'attacker'} class:lose={battleLog.winner === 'defender'}>
        {#if battleLog.winner === 'attacker'}
          Attacker wins!
        {:else if battleLog.winner === 'defender'}
          Defender wins!
        {:else}
          Tie!
        {/if}
        <div class="remaining">
          Cards remaining: {battleLog.attacker_remaining} vs {battleLog.defender_remaining}
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .battle-player {
    max-width: 520px;
    margin: 0 auto;
    padding: 8px;
  }
  .header {
    text-align: center;
    margin-bottom: 12px;
  }
  .header h2 {
    color: var(--tg-theme-text-color, #eee);
    margin: 0 0 4px 0;
  }
  .turn-indicator {
    font-size: 0.85rem;
    padding: 4px 12px;
    border-radius: 12px;
    display: inline-block;
    background: rgba(33, 150, 243, 0.2);
    color: #64b5f6;
  }
  .turn-indicator.defender-turn {
    background: rgba(244, 67, 54, 0.2);
    color: #ef9a9a;
  }
  .loading, .error {
    text-align: center;
    padding: 40px;
    font-size: 1.2rem;
  }
  .error {
    color: #f44;
  }
  .arena {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 12px;
  }
  .side {
    flex: 1;
    text-align: center;
    min-width: 0;
  }
  .label {
    font-size: 0.8rem;
    opacity: 0.7;
    margin-bottom: 8px;
    text-transform: uppercase;
    letter-spacing: 1px;
  }
  .vs {
    font-size: 1.5rem;
    font-weight: bold;
    opacity: 0.5;
    padding: 0 4px;
    margin-top: 40px;
  }
  .card-stack {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .card {
    background: var(--tg-theme-secondary-bg-color, #2a2a4a);
    border: 2px solid #888;
    border-radius: 10px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    transition: all 0.3s ease;
    opacity: 0.6;
    transform: scale(0.92);
  }
  .card.front {
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 12px rgba(255, 255, 255, 0.1);
  }
  .card.empty {
    opacity: 0.3;
    min-height: 60px;
    justify-content: center;
  }
  .card.flash {
    background: rgba(244, 67, 54, 0.4) !important;
    transition: background 0.1s ease;
  }
  .card.dying {
    opacity: 0.15;
    transform: scale(0.8);
  }
  .card.spawning {
    animation: spawn-in 0.4s ease-out;
  }
  @keyframes spawn-in {
    from { opacity: 0; transform: scale(0.5); }
    to { opacity: 1; transform: scale(1); }
  }
  .card-name {
    font-weight: bold;
    font-size: 0.85rem;
    text-transform: capitalize;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 100%;
  }
  .hp-bar {
    width: 100%;
    height: 6px;
    background: #333;
    border-radius: 3px;
    overflow: hidden;
  }
  .hp-fill {
    height: 100%;
    background: #4caf50;
    border-radius: 3px;
    transition: width 0.3s ease;
  }
  .hp-fill.defender {
    background: #f44336;
  }
  .stats {
    font-size: 0.7rem;
    opacity: 0.8;
    display: flex;
    justify-content: space-between;
    width: 100%;
  }
  .effects {
    display: flex;
    flex-wrap: wrap;
    gap: 2px;
    justify-content: center;
  }
  .effect-tag {
    font-size: 0.6rem;
    background: rgba(255, 255, 255, 0.1);
    padding: 1px 5px;
    border-radius: 4px;
    opacity: 0.7;
  }
  .actions-log {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    justify-content: center;
    margin-bottom: 12px;
  }
  .action-tag {
    padding: 3px 8px;
    border-radius: 8px;
    font-size: 0.75rem;
    font-weight: 500;
  }
  .action-tag.attack {
    background: rgba(255, 152, 0, 0.2);
    color: #ffb74d;
  }
  .action-tag.card_died {
    background: rgba(244, 67, 54, 0.2);
    color: #ef9a9a;
  }
  .action-tag.spawn_card {
    background: rgba(76, 175, 80, 0.2);
    color: #81c784;
  }
  .winner-banner {
    text-align: center;
    padding: 20px;
    border-radius: 12px;
    font-size: 1.3rem;
    font-weight: bold;
    margin-top: 12px;
    background: rgba(158, 158, 158, 0.2);
    border: 2px solid #888;
  }
  .winner-banner.win {
    background: rgba(76, 175, 80, 0.2);
    border: 2px solid #4caf50;
  }
  .winner-banner.lose {
    background: rgba(244, 67, 54, 0.2);
    border: 2px solid #f44336;
  }
  .remaining {
    font-size: 0.9rem;
    font-weight: normal;
    margin-top: 8px;
    opacity: 0.8;
  }
</style>
