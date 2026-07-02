/// <reference types="vitest" />
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [tailwindcss(), react()],
  build: {
    outDir: "out",
    sourcemap: true,
  },
  // @ts-expect-error - vitest's `test` config key isn't part of Vite's UserConfig type
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./src/setupTests.ts",
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/setupTests.ts", "src/vite-env.d.ts"],
      // Floor, not a target - ratchet up as more routes/components get
      // tests. Set a few points below the actual baseline (~76/84/66/76 at
      // the time this was added) so small refactors don't flake CI.
      thresholds: {
        statements: 65,
        branches: 75,
        functions: 55,
        lines: 65,
      },
    },
  },
});
