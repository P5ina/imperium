<script>
  import { onMount } from 'svelte';

  export let battleId;

  const API_URL = window.location.origin.includes('localhost')
    ? 'http://localhost:8090'
    : `${window.location.origin}/imperium-api`;

  let battleLog = null;
  let currentEntryIdx = -1;
  let error = null;
  let done = false;
  let flashCardId = null;
  let dyingCardId = null;
  let attackerDeck = [];
  let defenderDeck = [];

  const RARITY_COLORS = {
    common: '#888', uncommon: '#4caf50', rare: '#2196f3',
    epic: '#9c27b0', legendary: '#ff9800',
  };

  onMount(async () => {
    try {
      const resp = await fetch(`${API_URL}/battle/${battleId}`);
      if (!resp.ok) throw new Error(`Battle not found (${resp.status})`);
      const data = await resp.json();
      battleLog = data.battle_log;
      if (battleLog?.entries?.length > 0) {
        attackerDeck = battleLog.entries[0].attacker_deck || [];
        defenderDeck = battleLog.entries[0].defender_deck || [];
        setTimeout(() => processEntry(0), 500);
      }
    } catch (e) {
      error = e.message;
    }
  });

  function processEntry(idx) {
    if (!battleLog || idx >= battleLog.entries.length) {
      done = true;
      return;
    }
    currentEntryIdx = idx;
    const entry = battleLog.entries[idx];

    for (const action of entry.actions) {
      if (action.type === 'attack' && action.defender_id != null) {
        flashCardId = action.defender_id;
        setTimeout(() => { flashCardId = null; }, 350);
      }
      if (action.type === 'card_died' && action.died_card_id != null) {
        dyingCardId = action.died_card_id;
        setTimeout(() => { dyingCardId = null; }, 500);
      }
    }

    attackerDeck = entry.attacker_deck || [];
    defenderDeck = entry.defender_deck || [];

    setTimeout(() => processEntry(idx + 1), 900);
  }

  function hpPct(hp, max) {
    return Math.max(0, Math.min(100, (hp / Math.max(max, 1)) * 100));
  }

  $: entry = battleLog && currentEntryIdx >= 0 ? battleLog.entries[currentEntryIdx] : null;
</script>

<div class="player">
  {#if error}
    <div class="msg error">⚠️ {error}</div>
  {:else if !battleLog}
    <div class="msg">Загрузка...</div>
  {:else}
    <div class="header">
      <span class="round">Раунд {entry?.round ?? 0} / {battleLog.total_rounds}</span>
      {#if entry}
        <span class="turn" class:def={entry.turn_side === 'defender'}>
          {entry.turn_side === 'defender' ? '🛡 Защитник атакует' : '⚔️ Атакующий атакует'}
        </span>
      {/if}
    </div>

    <div class="arena">
      <div class="side">
        <div class="label">⚔️ Атакующий</div>
        {#each attackerDeck as card, i (card.id)}
          <div class="card"
            class:front={i === 0}
            class:flash={flashCardId === card.id}
            class:dying={dyingCardId === card.id}
            style="border-color:{RARITY_COLORS[card.rarity]||'#888'}">
            <div class="name">{card.name}</div>
            <div class="bar"><div class="fill" style="width:{hpPct(card.current_hp,card.max_hp)}%"></div></div>
            <div class="stats">❤️{card.current_hp} ⚔️{card.attack}</div>
          </div>
        {/each}
        {#if attackerDeck.length === 0}<div class="card empty">—</div>{/if}
      </div>

      <div class="vs">VS</div>

      <div class="side">
        <div class="label">🛡 Защитник</div>
        {#each defenderDeck as card, i (card.id)}
          <div class="card"
            class:front={i === 0}
            class:flash={flashCardId === card.id}
            class:dying={dyingCardId === card.id}
            style="border-color:{RARITY_COLORS[card.rarity]||'#888'}">
            <div class="name">{card.name}</div>
            <div class="bar"><div class="fill def" style="width:{hpPct(card.current_hp,card.max_hp)}%"></div></div>
            <div class="stats">❤️{card.current_hp} ⚔️{card.attack}</div>
          </div>
        {/each}
        {#if defenderDeck.length === 0}<div class="card empty">—</div>{/if}
      </div>
    </div>

    {#if done}
      <div class="banner" class:win={battleLog.winner==='attacker'} class:lose={battleLog.winner==='defender'}>
        {#if battleLog.winner === 'attacker'}🏆 Победа!
        {:else if battleLog.winner === 'defender'}💀 Поражение
        {:else}🤝 Ничья{/if}
      </div>
    {/if}
  {/if}
</div>

<style>
  .player { max-width: 480px; margin: 0 auto; padding: 12px; font-family: system-ui, sans-serif; }
  .msg { text-align: center; padding: 40px; font-size: 1.1rem; color: var(--tg-theme-text-color, #eee); }
  .msg.error { color: #f44; }
  .header { text-align: center; margin-bottom: 12px; display: flex; flex-direction: column; gap: 6px; }
  .round { font-size: 1rem; font-weight: bold; color: var(--tg-theme-text-color, #eee); }
  .turn { font-size: 0.82rem; padding: 3px 12px; border-radius: 12px; display: inline-block; background: rgba(33,150,243,.2); color: #64b5f6; }
  .turn.def { background: rgba(244,67,54,.2); color: #ef9a9a; }
  .arena { display: flex; gap: 8px; align-items: flex-start; }
  .side { flex: 1; display: flex; flex-direction: column; gap: 6px; }
  .label { font-size: 0.75rem; text-align: center; opacity: .6; margin-bottom: 2px; }
  .vs { padding: 0 4px; font-size: 1.3rem; opacity: .5; margin-top: 32px; }
  .card { background: var(--tg-theme-secondary-bg-color, #1e1e2e); border: 2px solid #888; border-radius: 10px; padding: 8px; opacity: .55; transform: scale(.93); transition: all .25s; }
  .card.front { opacity: 1; transform: scale(1); box-shadow: 0 0 10px rgba(255,255,255,.12); }
  .card.empty { opacity: .25; min-height: 50px; }
  .card.flash { background: rgba(244,67,54,.35) !important; }
  .card.dying { opacity: .1; transform: scale(.75); }
  .name { font-weight: 600; font-size: .82rem; text-align: center; color: var(--tg-theme-text-color, #eee); }
  .bar { width: 100%; height: 5px; background: #333; border-radius: 3px; overflow: hidden; margin: 4px 0; }
  .fill { height: 100%; background: #4caf50; border-radius: 3px; transition: width .3s; }
  .fill.def { background: #f44336; }
  .stats { font-size: .7rem; opacity: .8; text-align: center; color: var(--tg-theme-text-color, #eee); }
  .banner { text-align: center; padding: 18px; border-radius: 12px; font-size: 1.4rem; font-weight: bold; margin-top: 16px; background: rgba(158,158,158,.15); border: 2px solid #888; color: var(--tg-theme-text-color, #eee); }
  .banner.win { background: rgba(76,175,80,.2); border-color: #4caf50; }
  .banner.lose { background: rgba(244,67,54,.2); border-color: #f44336; }
</style>
