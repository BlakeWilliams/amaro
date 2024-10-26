import { defineConfig } from 'vite'
// vite.config.js
export default defineConfig({
  build: {
    manifest: true,
    base: "/assets",
    rollupOptions: {
      input: 'public/index.js',
      wow: 'public/main.js',
    },
  },
})
