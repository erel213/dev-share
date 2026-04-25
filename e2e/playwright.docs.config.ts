import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./docs-capture",
  globalSetup: "./docs-capture/global-setup.ts",
  fullyParallel: false,
  workers: 1,
  reporter: [["list"]],
  use: {
    baseURL: process.env.DOCS_BASE_URL || "http://localhost:3000",
    viewport: { width: 1280, height: 720 },
    screenshot: "off",
    video: "off",
    trace: "off",
  },
  projects: [
    {
      name: "chromium",
      use: { browserName: "chromium" },
    },
  ],
});
