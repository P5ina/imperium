<script>
  import BattlePlayer from './BattlePlayer.svelte';

  const params = new URLSearchParams(window.location.search);
  const battleId = params.get('battle_id');
</script>

<main>
  {#if battleId}
    <BattlePlayer {battleId} />
  {:else}
    <div class="center">
      <h1>Imperium</h1>
      <p>Не указан battle_id</p>
    </div>
  {/if}
</main>

<style>
  main {
    padding: 16px;
    max-width: 480px;
    margin: 0 auto;
  }
  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    text-align: center;
  }
  h1 {
    font-size: 2rem;
    margin-bottom: 8px;
  }
</style>
