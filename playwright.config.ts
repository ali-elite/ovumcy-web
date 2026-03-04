import { defineConfig, devices } from '@playwright/test';

const envWorkers = Number.parseInt(process.env.PLAYWRIGHT_WORKERS ?? '', 10);
const workers = Number.isInteger(envWorkers) && envWorkers > 0 ? envWorkers : undefined;

export default defineConfig({
  testDir: 'e2e',
  timeout: 30 * 1000,
  workers,
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:8080',
    headless: true,
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
