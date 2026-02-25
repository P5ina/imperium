import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  base: '/app/',
  server: {
    port: 5173,
  },
  resolve: {
    alias: {
      'node:async_hooks': '/dev/null',
    },
  },
});
