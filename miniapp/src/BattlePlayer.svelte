<script>
  import { onMount } from 'svelte';

  let { battleId } = $props();

  const API_URL = window.location.hostname === 'localhost'
    ? 'http://localhost:8090'
    : window.location.origin + '/imperium-api';

  // --- Constants ---

  const CARD_EMOJIS = {
    venom: '🐍', thug: '😤', goon: '🤜', cobblestone: '🪨', enforcer: '💪',
    hitman: '🎯', 'spider-man': '🕷️', capo: '🎩', don: '👔', mastermind: '🧠',
    berserker: '😡', godfather: '🤵', 'fuel-card': '⛽', 'pvp-assassin': '🗡️',
    'pvp-warlord': '⚔️', 'pvp-champion': '🏆',
  };

  const RARITY_COLORS = {
    common: '#aaa', uncommon: '#4caf50', rare: '#2196f3',
    epic: '#9c27b0', legendary: '#ff9800',
  };

  const EFFECT_EMOJIS = {
    rampage: '🔥', deathrattle: '💀', taunt: '🛡️', thorns: '🌵', dungeon_key: '🔑',
  };

  // --- State ---

  let battleLog = $state(null);
  let error = $state(null);
  let loading = $state(true);

  let currentRound = $state(0);
  let totalRounds = $state(0);
  let turnSide = $state('attacker');

  let attackerDeck = $state([]);
  let defenderDeck = $state([]);

  // Card HP tracked separately for smooth animation
  let cardHp = $state({});

  // Animation state
  let activeCardKey = $state(null);
  let shakingCardKeys = $state(new Set());
  let dyingCardKeys = $state(new Set());
  let damageNumbers = $state({});   // key -> { amount, isAttackerSide }
  let spawningCardKeys = $state(new Set());
  let screenShake = $state(false);
  let showResult = $state(false);
  let playing = $state(false);
  let cancelled = $state(false);

  let currentActions = $state([]);

  // --- Helpers ---

  function cardEmoji(card) {
    return CARD_EMOJIS[card.card_id] || card.card_id || '❓';
  }

  function rarityColor(rarity) {
    return RARITY_COLORS[rarity] || '#aaa';
  }

  function hpPercent(card) {
    const hp = cardHp[card.id] ?? card.current_hp;
    const max = card.max_hp || 1;
    return Math.max(0, Math.min(100, (hp / max) * 100));
  }

  function hpBarColor(card) {
    const hasRampage = card.effects?.some(e => e === 'rampage' || e?.effect_type === 'rampage');
    if (hasRampage) return '#ff9800';
    const pct = hpPercent(card);
    if (pct > 60) return '#4caf50';
    if (pct > 30) return '#ffb300';
    return '#f44336';
  }

  function cardKey(cardId, side) {
    return `${cardId}-${side}`;
  }

  function sleep(ms) {
    return new Promise(r => setTimeout(r, ms));
  }

  function currentHp(card) {
    return cardHp[card.id] ?? card.current_hp;
  }

  // --- Data fetch ---

  onMount(async () => {
    try {
      const resp = await fetch(`${API_URL}/battle/${battleId}`);
      if (!resp.ok) throw new Error(`Бой не найден (${resp.status})`);
      const data = await resp.json();
      battleLog = data.battle_log;
      if (!battleLog?.entries?.length) throw new Error('Нет данных боя');

      totalRounds = battleLog.total_rounds;
      // Init decks from first entry's starting state
      const first = battleLog.entries[0];
      attackerDeck = [...first.attacker_deck];
      defenderDeck = [...first.defender_deck];

      // Init HP map
      const hp = {};
      for (const c of first.attacker_deck) hp[c.id] = c.current_hp;
      for (const c of first.defender_deck) hp[c.id] = c.current_hp;
      cardHp = hp;

      loading = false;

      await sleep(600);
      await playBattle();
    } catch (e) {
      error = e.message;
      loading = false;
    }
  });

  // --- Animation Sequencing ---

  async function playBattle() {
    if (playing || cancelled) return;
    playing = true;

    for (let i = 0; i < battleLog.entries.length; i++) {
      if (cancelled) break;
      const entry = battleLog.entries[i];
      await animateEntry(entry);
    }

    if (!cancelled) {
      await sleep(400);
      showResult = true;
    }

    playing = false;
  }

  async function animateEntry(entry) {
    currentRound = entry.round;
    turnSide = entry.turn_side;

    const attacks = entry.actions.filter(a => a.type === 'attack');
    const deaths = entry.actions.filter(a => a.type === 'card_died');
    const spawns = entry.actions.filter(a => a.type === 'spawn_card');

    currentActions = entry.actions;

    // Pre-entry pause
    await sleep(100);

    // --- Phase 1: Attacks ---
    if (attacks.length > 0) {
      const attackerIsOnAttackerSide = entry.turn_side === 'attacker';

      for (const action of attacks) {
        if (cancelled) return;
        const atkSide = attackerIsOnAttackerSide ? 'attacker' : 'defender';
        const defSide = attackerIsOnAttackerSide ? 'defender' : 'attacker';
        const atkKey = cardKey(action.attacker_id, atkSide);
        const defKey = cardKey(action.defender_id, defSide);

        // Trigger charge on attacker
        activeCardKey = atkKey;

        // After 80ms: hit flash + shake on target
        await sleep(80);

        if (action.damage > 0) {
          shakingCardKeys = new Set([...shakingCardKeys, defKey]);
          screenShake = true;

          // Update HP
          const newHp = { ...cardHp };
          newHp[action.defender_id] = Math.max(0, (newHp[action.defender_id] ?? 0) - action.damage);
          cardHp = newHp;
        }

        // After 200ms from start: show damage number
        await sleep(120);

        if (action.damage > 0) {
          damageNumbers = { ...damageNumbers, [defKey]: { amount: action.damage, isAttackerSide: !attackerIsOnAttackerSide } };
        }

        // Clear screen shake
        await sleep(100);
        screenShake = false;
      }

      // After 300ms from last attack: deaths
      await sleep(100);

      for (const action of deaths) {
        if (cancelled) return;
        const side = action.died_side;
        const dKey = cardKey(action.died_card_id, side);
        dyingCardKeys = new Set([...dyingCardKeys, dKey]);

        const newHp = { ...cardHp };
        newHp[action.died_card_id] = 0;
        cardHp = newHp;
      }

      if (deaths.length > 0) {
        // Wait for death animation
        await sleep(500);

        // Remove dead cards from decks
        for (const action of deaths) {
          if (action.died_side === 'attacker') {
            attackerDeck = attackerDeck.filter(c => c.id !== action.died_card_id);
          } else {
            defenderDeck = defenderDeck.filter(c => c.id !== action.died_card_id);
          }
        }
        dyingCardKeys = new Set();
      }
    } else if (deaths.length > 0) {
      // Deaths without attacks
      for (const action of deaths) {
        const side = action.died_side;
        const dKey = cardKey(action.died_card_id, side);
        dyingCardKeys = new Set([...dyingCardKeys, dKey]);
        const newHp = { ...cardHp };
        newHp[action.died_card_id] = 0;
        cardHp = newHp;
      }
      await sleep(500);
      for (const action of deaths) {
        if (action.died_side === 'attacker') {
          attackerDeck = attackerDeck.filter(c => c.id !== action.died_card_id);
        } else {
          defenderDeck = defenderDeck.filter(c => c.id !== action.died_card_id);
        }
      }
      dyingCardKeys = new Set();
    }

    // --- Phase 2: Spawns ---
    for (const action of spawns) {
      if (cancelled) return;
      if (action.spawned_card && action.side) {
        const sc = action.spawned_card;
        const sKey = cardKey(sc.id, action.side);
        spawningCardKeys = new Set([...spawningCardKeys, sKey]);

        const newHp = { ...cardHp };
        newHp[sc.id] = sc.current_hp;
        cardHp = newHp;

        if (action.side === 'attacker') {
          attackerDeck = [sc, ...attackerDeck];
        } else {
          defenderDeck = [sc, ...defenderDeck];
        }

        await sleep(300);
        spawningCardKeys = new Set();
      }
    }

    // --- Phase 3: Sync with server deck state ---
    await sleep(100);

    activeCardKey = null;
    shakingCardKeys = new Set();
    damageNumbers = {};
    currentActions = [];

    // Sync decks and HP from entry's end state
    attackerDeck = [...entry.attacker_deck];
    defenderDeck = [...entry.defender_deck];

    const syncHp = {};
    for (const c of entry.attacker_deck) syncHp[c.id] = c.current_hp;
    for (const c of entry.defender_deck) syncHp[c.id] = c.current_hp;
    cardHp = syncHp;

    // Total entry duration ~900ms, pad remaining
    await sleep(200);
  }

  function closeResult() {
    if (window.Telegram?.WebApp) {
      window.Telegram.WebApp.close();
    }
  }
</script>

<div class="battle-player" class:screen-shake={screenShake}>
  {#if error}
    <div class="error-screen">
      <div class="error-icon">⚠️</div>
      <div class="error-text">{error}</div>
      <button class="error-btn" onclick={closeResult}>Закрыть</button>
    </div>
  {:else if loading}
    <div class="loading-screen">
      <div class="loading-spinner"></div>
      <div class="loading-text">Подготовка боя...</div>
    </div>
  {:else}
    <!-- Header -->
    <div class="header">
      <div class="round-info">Раунд {currentRound} / {totalRounds}</div>
      <div class="turn-indicator" class:defender={turnSide === 'defender'}>
        {turnSide === 'attacker' ? '⚔️ Атакует' : '🛡 Защищается'}
      </div>
    </div>

    <!-- Arena -->
    <div class="arena">
      <!-- Attacker column -->
      <div class="column">
        <div class="side-label attacker-label">⚔️ АТК</div>
        {#each attackerDeck as card, i (card.id)}
          {@const isFront = i === 0}
          {@const key = cardKey(card.id, 'attacker')}
          {@const isActive = activeCardKey === key}
          {@const isShaking = shakingCardKeys.has(key)}
          {@const isDying = dyingCardKeys.has(key)}
          {@const isSpawning = spawningCardKeys.has(key)}
          {@const dmg = damageNumbers[key]}
          <div
            class="card-wrapper"
            class:frontline={isFront}
            class:backline={!isFront}
            class:attack-charge-right={isActive}
            class:hit-shake={isShaking}
            class:death-fly-left={isDying}
            class:spawn-in={isSpawning}
          >
            <div class="battle-card"
              style="--rarity-color: {rarityColor(card.rarity)}; --border-width: {isFront ? '3px' : '2px'}">
              <!-- Top stats -->
              <div class="card-top">
                <span class="stat hp-stat">❤️ <b style="color: {hpBarColor(card)}">{currentHp(card)}</b></span>
                <span class="stat atk-stat">
                  {card.attack === 0 ? '🔒' : '⚔️'}
                  <b style="color: {card.attack === 0 ? '#888' : '#ffb300'}">{card.attack}</b>
                </span>
              </div>

              <!-- Effects -->
              {#if card.effects?.length}
                <div class="effects-row">
                  {#each card.effects as eff}
                    {@const et = typeof eff === 'string' ? eff : eff.effect_type}
                    <span class="effect-badge">{EFFECT_EMOJIS[et] || '✨'}</span>
                  {/each}
                </div>
              {/if}

              <!-- Center emoji -->
              <div class="card-emoji" style="text-shadow: 0 0 16px {rarityColor(card.rarity)}88">
                {cardEmoji(card)}
              </div>

              <!-- HP bar -->
              <div class="hp-bar-track">
                <div class="hp-bar-fill" style="width: {hpPercent(card)}%; background: {hpBarColor(card)}"></div>
              </div>

              <!-- Bottom info -->
              <div class="card-name">{card.name}</div>
              <div class="rarity-badge" style="color: {rarityColor(card.rarity)}; background: {rarityColor(card.rarity)}22">
                {card.rarity?.toUpperCase()}
              </div>
            </div>

            <!-- Damage number -->
            {#if dmg}
              <div class="damage-number" class:defender-dmg={dmg.isAttackerSide}>-{dmg.amount}</div>
            {/if}
          </div>
        {/each}
        {#if attackerDeck.length === 0}
          <div class="empty-slot">—</div>
        {/if}
      </div>

      <!-- VS divider -->
      <div class="vs-divider">
        <div class="vs-line"></div>
        <div class="vs-text">VS</div>
        <div class="vs-line"></div>
      </div>

      <!-- Defender column -->
      <div class="column">
        <div class="side-label defender-label">🛡 ЗАЩ</div>
        {#each defenderDeck as card, i (card.id)}
          {@const isFront = i === 0}
          {@const key = cardKey(card.id, 'defender')}
          {@const isActive = activeCardKey === key}
          {@const isShaking = shakingCardKeys.has(key)}
          {@const isDying = dyingCardKeys.has(key)}
          {@const isSpawning = spawningCardKeys.has(key)}
          {@const dmg = damageNumbers[key]}
          <div
            class="card-wrapper"
            class:frontline={isFront}
            class:backline={!isFront}
            class:attack-charge-left={isActive}
            class:hit-shake={isShaking}
            class:death-fly-right={isDying}
            class:spawn-in={isSpawning}
          >
            <div class="battle-card"
              style="--rarity-color: {rarityColor(card.rarity)}; --border-width: {isFront ? '3px' : '2px'}">
              <div class="card-top">
                <span class="stat hp-stat">❤️ <b style="color: {hpBarColor(card)}">{currentHp(card)}</b></span>
                <span class="stat atk-stat">
                  {card.attack === 0 ? '🔒' : '⚔️'}
                  <b style="color: {card.attack === 0 ? '#888' : '#ffb300'}">{card.attack}</b>
                </span>
              </div>

              {#if card.effects?.length}
                <div class="effects-row">
                  {#each card.effects as eff}
                    {@const et = typeof eff === 'string' ? eff : eff.effect_type}
                    <span class="effect-badge">{EFFECT_EMOJIS[et] || '✨'}</span>
                  {/each}
                </div>
              {/if}

              <div class="card-emoji" style="text-shadow: 0 0 16px {rarityColor(card.rarity)}88">
                {cardEmoji(card)}
              </div>

              <div class="hp-bar-track">
                <div class="hp-bar-fill" style="width: {hpPercent(card)}%; background: {hpBarColor(card)}"></div>
              </div>

              <div class="card-name">{card.name}</div>
              <div class="rarity-badge" style="color: {rarityColor(card.rarity)}; background: {rarityColor(card.rarity)}22">
                {card.rarity?.toUpperCase()}
              </div>
            </div>

            {#if dmg}
              <div class="damage-number" class:defender-dmg={dmg.isAttackerSide}>-{dmg.amount}</div>
            {/if}
          </div>
        {/each}
        {#if defenderDeck.length === 0}
          <div class="empty-slot">—</div>
        {/if}
      </div>
    </div>

    <!-- Action log -->
    {#if currentActions.length > 0}
      <div class="action-log">
        {#each currentActions as action}
          {#if action.type === 'attack'}
            <div class="log-line">⚔️ Урон: <b>{action.damage}</b></div>
          {:else if action.type === 'card_died'}
            <div class="log-line death-log">💀 Карта уничтожена</div>
          {:else if action.type === 'spawn_card'}
            <div class="log-line spawn-log">✨ Призыв: {action.spawned_card?.name}</div>
          {/if}
        {/each}
      </div>
    {/if}

    <!-- Result overlay -->
    {#if showResult}
      <div class="result-overlay" class:victory={battleLog.winner === 'attacker'} class:defeat={battleLog.winner === 'defender'}>
        <div class="result-content">
          <div class="result-glow"></div>
          <div class="result-emoji">
            {#if battleLog.winner === 'attacker'}🏆{:else if battleLog.winner === 'defender'}💀{:else}🤝{/if}
          </div>
          <div class="result-title">
            {#if battleLog.winner === 'attacker'}ПОБЕДА!
            {:else if battleLog.winner === 'defender'}ПОРАЖЕНИЕ
            {:else}НИЧЬЯ{/if}
          </div>
          <div class="result-divider">
            <span class="divider-line"></span>
            <span class="divider-icon">{battleLog.winner === 'attacker' ? '⭐' : '✕'}</span>
            <span class="divider-line"></span>
          </div>
          <div class="result-stats">
            Осталось карт: ⚔️ {battleLog.attacker_remaining} — 🛡 {battleLog.defender_remaining}
          </div>
          <button class="result-btn" onclick={closeResult}>
            {battleLog.winner === 'attacker' ? 'Забрать победу →' : 'Продолжить →'}
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  /* ===== Base ===== */
  .battle-player {
    position: fixed;
    inset: 0;
    background: var(--tg-theme-bg-color, #0d0d1a);
    color: var(--tg-theme-text-color, #eee);
    font-family: system-ui, -apple-system, sans-serif;
    overflow-y: auto;
    overflow-x: hidden;
    display: flex;
    flex-direction: column;
    -webkit-overflow-scrolling: touch;
  }

  /* ===== Screen Shake ===== */
  @keyframes screen-shake {
    0%, 100% { transform: translateX(0); }
    12.5% { transform: translateX(-4px); }
    25% { transform: translateX(4px); }
    37.5% { transform: translateX(-4px); }
    50% { transform: translateX(4px); }
    62.5% { transform: translateX(-3px); }
    75% { transform: translateX(3px); }
    87.5% { transform: translateX(-2px); }
  }
  .screen-shake {
    animation: screen-shake 0.3s ease-out;
  }

  /* ===== Loading ===== */
  .loading-screen, .error-screen {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    min-height: 100vh;
  }
  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255,255,255,0.15);
    border-top-color: #64b5f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  .loading-text { opacity: 0.7; font-size: 0.95rem; }
  .error-icon { font-size: 3rem; }
  .error-text { font-size: 1rem; opacity: 0.8; text-align: center; padding: 0 24px; }
  .error-btn {
    margin-top: 12px;
    background: rgba(255,255,255,0.1);
    border: 1px solid rgba(255,255,255,0.2);
    color: #eee;
    padding: 10px 32px;
    border-radius: 20px;
    font-size: 0.95rem;
    cursor: pointer;
  }

  /* ===== Header ===== */
  .header {
    text-align: center;
    padding: 12px 16px 8px;
    flex-shrink: 0;
  }
  .round-info {
    font-size: 1.1rem;
    font-weight: 800;
    letter-spacing: 0.5px;
  }
  .turn-indicator {
    display: inline-block;
    margin-top: 6px;
    padding: 4px 16px;
    border-radius: 14px;
    font-size: 0.82rem;
    font-weight: 600;
    background: rgba(33,150,243,0.2);
    color: #64b5f6;
  }
  .turn-indicator.defender {
    background: rgba(244,67,54,0.2);
    color: #ef9a9a;
  }

  /* ===== Arena ===== */
  .arena {
    flex: 1;
    display: flex;
    align-items: flex-start;
    padding: 8px 10px;
    gap: 4px;
    min-height: 0;
  }
  .column {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  /* ===== Side labels ===== */
  .side-label {
    font-size: 0.7rem;
    font-weight: 900;
    letter-spacing: 1px;
    padding: 3px 12px;
    border-radius: 10px;
    text-transform: uppercase;
  }
  .attacker-label {
    background: rgba(0,188,212,0.15);
    color: #4dd0e1;
    border: 1px solid rgba(0,188,212,0.3);
  }
  .defender-label {
    background: rgba(244,67,54,0.15);
    color: #ef9a9a;
    border: 1px solid rgba(244,67,54,0.3);
  }

  /* ===== VS divider ===== */
  .vs-divider {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding-top: 36px;
    flex-shrink: 0;
  }
  .vs-line {
    width: 2px;
    height: 28px;
    background: linear-gradient(to bottom, transparent, rgba(255,255,255,0.15), transparent);
  }
  .vs-text {
    font-size: 0.8rem;
    font-weight: 900;
    opacity: 0.35;
    letter-spacing: 2px;
  }

  /* ===== Card wrapper ===== */
  .card-wrapper {
    position: relative;
    transition: transform 0.25s ease, opacity 0.25s ease;
  }
  .card-wrapper.frontline {
    transform: scale(1);
    z-index: 10;
  }
  .card-wrapper.backline {
    transform: scale(0.85);
    opacity: 0.5;
  }

  /* ===== Battle card ===== */
  .battle-card {
    width: 120px;
    background: #fff;
    border-radius: 12px;
    border: var(--border-width, 2px) solid var(--rarity-color, #aaa);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 6px 6px 8px;
    position: relative;
    box-sizing: border-box;
  }
  .frontline .battle-card {
    box-shadow: 0 0 14px var(--rarity-color, #aaa)66, 0 4px 12px rgba(0,0,0,0.4);
  }
  .backline .battle-card {
    box-shadow: 0 2px 8px rgba(0,0,0,0.3);
  }

  /* ===== Card internals ===== */
  .card-top {
    display: flex;
    justify-content: space-between;
    width: 100%;
    padding: 0 2px;
    margin-bottom: 2px;
  }
  .stat {
    font-size: 0.72rem;
    display: flex;
    align-items: center;
    gap: 2px;
  }
  .stat b { font-size: 0.82rem; }

  .effects-row {
    display: flex;
    gap: 3px;
    margin-bottom: 2px;
  }
  .effect-badge {
    font-size: 0.6rem;
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0,0,0,0.06);
    border-radius: 50%;
  }

  .card-emoji {
    font-size: 48px;
    line-height: 1;
    margin: 4px 0;
  }

  .hp-bar-track {
    width: 100%;
    height: 4px;
    background: rgba(0,0,0,0.12);
    border-radius: 2px;
    overflow: hidden;
    margin: 4px 0 3px;
  }
  .hp-bar-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  .card-name {
    font-size: 0.72rem;
    font-weight: 700;
    color: #222;
    text-align: center;
    line-height: 1.1;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .rarity-badge {
    font-size: 0.5rem;
    font-weight: 800;
    letter-spacing: 0.5px;
    padding: 1px 6px;
    border-radius: 8px;
    margin-top: 2px;
  }

  /* ===== Attack animations ===== */
  @keyframes charge-right {
    0%   { transform: scale(1) translateX(0) rotate(0deg); }
    16%  { transform: scale(1.12) translateX(60px) rotate(-3deg); }
    42%  { transform: scale(1.05) translateX(60px) rotate(-3deg); }
    58%  { transform: scale(0.92) translateX(-8px) rotate(2deg); }
    100% { transform: scale(1) translateX(0) rotate(0deg); }
  }
  @keyframes charge-left {
    0%   { transform: scale(1) translateX(0) rotate(0deg); }
    16%  { transform: scale(1.12) translateX(-60px) rotate(3deg); }
    42%  { transform: scale(1.05) translateX(-60px) rotate(3deg); }
    58%  { transform: scale(0.92) translateX(8px) rotate(-2deg); }
    100% { transform: scale(1) translateX(0) rotate(0deg); }
  }
  .attack-charge-right {
    animation: charge-right 0.62s cubic-bezier(0.25, 0.46, 0.45, 0.94);
    z-index: 20 !important;
  }
  .attack-charge-left {
    animation: charge-left 0.62s cubic-bezier(0.25, 0.46, 0.45, 0.94);
    z-index: 20 !important;
  }

  /* ===== Hit shake ===== */
  @keyframes hit-shake {
    0%, 100% { transform: translateX(0); }
    10% { transform: translateX(-6px); }
    20% { transform: translateX(6px); }
    30% { transform: translateX(-5px); }
    40% { transform: translateX(5px); }
    50% { transform: translateX(-3px); }
    60% { transform: translateX(3px); }
  }
  .hit-shake {
    animation: hit-shake 0.3s ease-out;
  }
  .hit-shake .battle-card {
    box-shadow: 0 0 20px rgba(244,67,54,0.5) !important;
    border-color: #f44336 !important;
    transition: box-shadow 0.15s, border-color 0.15s;
  }

  /* ===== Death animations ===== */
  @keyframes death-fly-left {
    0%   { transform: scale(1) translateX(0) rotate(0deg); opacity: 1; }
    15%  { transform: scale(1.1) translateX(10px) rotate(0deg); opacity: 1; }
    100% { transform: scale(0.3) translateX(-300px) rotate(-45deg); opacity: 0; }
  }
  @keyframes death-fly-right {
    0%   { transform: scale(1) translateX(0) rotate(0deg); opacity: 1; }
    15%  { transform: scale(1.1) translateX(-10px) rotate(0deg); opacity: 1; }
    100% { transform: scale(0.3) translateX(300px) rotate(45deg); opacity: 0; }
  }
  .death-fly-left {
    animation: death-fly-left 0.5s cubic-bezier(0.55, 0, 1, 0.45) forwards;
    pointer-events: none;
  }
  .death-fly-right {
    animation: death-fly-right 0.5s cubic-bezier(0.55, 0, 1, 0.45) forwards;
    pointer-events: none;
  }

  /* ===== Spawn animation ===== */
  @keyframes spawn-in {
    0%   { transform: scale(0.4); opacity: 0; }
    100% { transform: scale(1); opacity: 1; }
  }
  .spawn-in {
    animation: spawn-in 0.3s ease-out;
  }

  /* ===== Damage number ===== */
  @keyframes damage-float {
    0%   { transform: translateY(0) scale(1.3); opacity: 1; }
    30%  { transform: translateY(-30px) scale(1); opacity: 1; }
    70%  { transform: translateY(-50px) scale(0.9); opacity: 0.8; }
    100% { transform: translateY(-60px) scale(0.8); opacity: 0; }
  }
  .damage-number {
    position: absolute;
    top: 30%;
    left: 50%;
    transform: translateX(-50%);
    font-size: 24px;
    font-weight: 900;
    background: linear-gradient(to bottom, #ff4444, #ff8800);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    text-shadow: none;
    filter: drop-shadow(0 1px 2px rgba(0,0,0,0.5));
    animation: damage-float 0.6s ease-out forwards;
    pointer-events: none;
    z-index: 30;
    white-space: nowrap;
  }
  .damage-number.defender-dmg {
    background: linear-gradient(to bottom, #00bcd4, #2196f3);
    -webkit-background-clip: text;
    background-clip: text;
  }

  /* ===== Action log ===== */
  .action-log {
    padding: 6px 16px 10px;
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    justify-content: center;
    flex-shrink: 0;
  }
  .log-line {
    font-size: 0.72rem;
    padding: 3px 10px;
    border-radius: 8px;
    background: rgba(255,255,255,0.06);
    color: rgba(255,255,255,0.7);
  }
  .log-line b { color: #ffb300; }
  .death-log { color: #ef5350; }
  .spawn-log { color: #66bb6a; }

  /* ===== Empty slot ===== */
  .empty-slot {
    width: 120px;
    height: 160px;
    border: 2px dashed rgba(255,255,255,0.1);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.3;
    font-size: 1.5rem;
  }

  /* ===== Result overlay ===== */
  .result-overlay {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    animation: result-fade-in 0.4s ease-out;
  }
  @keyframes result-fade-in {
    0% { opacity: 0; }
    100% { opacity: 1; }
  }

  .result-overlay.victory {
    background: radial-gradient(circle at center, rgba(0,188,212,0.3) 0%, rgba(0,0,0,0.95) 70%);
  }
  .result-overlay.defeat {
    background: radial-gradient(circle at center, rgba(244,67,54,0.25) 0%, rgba(0,0,0,0.95) 70%);
  }
  .result-overlay:not(.victory):not(.defeat) {
    background: radial-gradient(circle at center, rgba(158,158,158,0.2) 0%, rgba(0,0,0,0.95) 70%);
  }

  .result-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    position: relative;
    animation: result-content-in 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
  }
  @keyframes result-content-in {
    0% { transform: scale(0.3); opacity: 0; }
    100% { transform: scale(1); opacity: 1; }
  }

  .result-glow {
    position: absolute;
    width: 200px;
    height: 200px;
    border-radius: 50%;
    top: -20px;
    z-index: -1;
  }
  .victory .result-glow {
    background: radial-gradient(circle, rgba(0,188,212,0.35) 0%, transparent 70%);
    box-shadow: 0 0 80px rgba(0,188,212,0.3);
  }
  .defeat .result-glow {
    background: radial-gradient(circle, rgba(244,67,54,0.3) 0%, transparent 70%);
    box-shadow: 0 0 80px rgba(244,67,54,0.25);
  }

  .result-emoji {
    font-size: 80px;
    animation: result-emoji-pulse 1.5s ease-in-out infinite alternate;
  }
  @keyframes result-emoji-pulse {
    0% { transform: scale(1); }
    100% { transform: scale(1.1); }
  }

  .result-title {
    font-size: 2.8rem;
    font-weight: 900;
    letter-spacing: 3px;
  }
  .victory .result-title {
    background: linear-gradient(to right, #00bcd4, #2196f3);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    filter: drop-shadow(0 0 20px rgba(0,188,212,0.5));
  }
  .defeat .result-title {
    background: linear-gradient(to right, #f44336, #ff9800);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
    filter: drop-shadow(0 0 20px rgba(244,67,54,0.5));
  }

  .result-divider {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 200px;
  }
  .divider-line {
    flex: 1;
    height: 2px;
  }
  .victory .divider-line {
    background: linear-gradient(to right, transparent, #00bcd4);
  }
  .victory .divider-line:last-child {
    background: linear-gradient(to left, transparent, #00bcd4);
  }
  .defeat .divider-line {
    background: linear-gradient(to right, transparent, #f44336);
  }
  .defeat .divider-line:last-child {
    background: linear-gradient(to left, transparent, #f44336);
  }
  .divider-icon {
    font-size: 0.75rem;
    opacity: 0.7;
  }

  .result-stats {
    font-size: 0.9rem;
    opacity: 0.75;
    margin-top: 4px;
  }

  .result-btn {
    margin-top: 24px;
    padding: 12px 36px;
    border-radius: 24px;
    border: none;
    font-size: 1rem;
    font-weight: 700;
    color: #fff;
    cursor: pointer;
    animation: result-btn-in 0.5s ease-out 0.6s both;
  }
  @keyframes result-btn-in {
    0% { transform: scale(0.8); opacity: 0; }
    100% { transform: scale(1); opacity: 1; }
  }
  .victory .result-btn {
    background: linear-gradient(to right, #00bcd4, #2196f3);
    box-shadow: 0 4px 16px rgba(0,188,212,0.4);
  }
  .defeat .result-btn {
    background: linear-gradient(to right, #f44336, #ff9800);
    box-shadow: 0 4px 16px rgba(244,67,54,0.4);
  }
  .result-overlay:not(.victory):not(.defeat) .result-btn {
    background: linear-gradient(to right, #666, #888);
    box-shadow: 0 4px 16px rgba(100,100,100,0.3);
  }
  .result-overlay:not(.victory):not(.defeat) .result-title {
    color: #ccc;
  }
  .result-overlay:not(.victory):not(.defeat) .divider-line {
    background: linear-gradient(to right, transparent, #888);
  }
  .result-overlay:not(.victory):not(.defeat) .divider-line:last-child {
    background: linear-gradient(to left, transparent, #888);
  }
</style>
