import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { fileURLToPath } from 'node:url';

// The express repo is served as a GitHub *project* page at
// https://malcolmston.github.io/express/, so assets must be based under /express/.
export default defineConfig({
  base: '/express/',
  plugins: [react()],
  resolve: {
    alias: {
      // Import the vendored shared library (a git submodule) from source.
      'go-ui': fileURLToPath(new URL('./vendor/go/ui/src/index.ts', import.meta.url)),
    },
  },
  build: { outDir: 'dist', emptyOutDir: true },
});
