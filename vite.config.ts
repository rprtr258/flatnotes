import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";

// Backend serves flatnotes/dist/index.html (SPA) and flatnotes/dist/ statically,
// so Vite root is flatnotes/ and output goes to flatnotes/dist.
export default defineConfig({
  root: "flatnotes",
  base: "/",
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./flatnotes/src", import.meta.url)),
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": "http://localhost:8080",
      "/attachments": "http://localhost:8080",
      "/static": "http://localhost:8080",
    },
  },
});
