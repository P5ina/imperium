import App from './App.svelte';

if (window.Telegram?.WebApp) {
  window.Telegram.WebApp.ready();
  window.Telegram.WebApp.expand();
}

new App({ target: document.getElementById('app') });
