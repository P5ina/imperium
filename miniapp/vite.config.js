import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [
    svelte({
      compilerOptions: {
        generate: "client",
      },
    }),
  ],
  base: "/app/",
  build: {
    rollupOptions: {
      external: [],
    },
  },
  resolve: {
    conditions: ["browser", "import"],
  },
});
