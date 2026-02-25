import App from './App.svelte';
import { mount } from 'svelte';

// Tell Telegram the app is ready
if (window.Telegram?.WebApp) {
  window.Telegram.WebApp.ready();
  window.Telegram.WebApp.expand();
}

const app = mount(App, { target: document.getElementById('app') });

export default app;
