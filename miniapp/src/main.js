import { mount } from 'svelte';
import App from './App.svelte';

if (window.Telegram?.WebApp) {
  window.Telegram.WebApp.ready();
  window.Telegram.WebApp.expand();
}

mount(App, { target: document.getElementById('app') });
