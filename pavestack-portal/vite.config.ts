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
    // jsdom's default about:blank URL has an opaque origin, which makes
    // window.localStorage unavailable (theme persistence tests need it).
    environmentOptions: {
      jsdom: { url: "http://localhost:3000/" },
    },
    setupFiles: "./src/setupTests.ts",
    coverage: {
      provider: "v8",
      reporter: ["text", "lcov"],
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.{ts,tsx}", "src/setupTests.ts", "src/vite-env.d.ts"],
      // Floor, not a target - ratchet up as more routes/components get
      // tests. Re-baselined after the platform-hardening branch merge added
      // less-tested code (~65/61/56/65 at that point); set a few points
      // below so small refactors don't flake CI.
      thresholds: {
        statements: 60,
        branches: 58,
        functions: 50,
        lines: 60,
      },
    },
  },
});
