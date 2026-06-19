/// <reference types="vitest" />
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [
    tailwindcss(),
    react()
  ],
  build: {
    outDir: "out",
    sourcemap: true,
  },
  // @ts-ignore
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/setupTests.ts",
  },
});
