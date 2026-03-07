import { expect, test } from '@playwright/test';

test.describe('PWA install prompt', () => {
  test('shows the custom mobile install CTA and triggers the native prompt', async ({ page }) => {
    await page.addInitScript(() => {
      (window as Window & { __pwaInstallPromptCalls?: number }).__pwaInstallPromptCalls = 0;

      window.addEventListener('DOMContentLoaded', () => {
        const installEvent = new Event('beforeinstallprompt');

        Object.defineProperty(installEvent, 'prompt', {
          configurable: true,
          value: () => {
            const state = window as Window & { __pwaInstallPromptCalls?: number };
            state.__pwaInstallPromptCalls = (state.__pwaInstallPromptCalls ?? 0) + 1;
            return Promise.resolve();
          },
        });
        Object.defineProperty(installEvent, 'userChoice', {
          configurable: true,
          value: Promise.resolve({ outcome: 'accepted', platform: 'web' }),
        });

        window.dispatchEvent(installEvent);
      });
    });

    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto('/login');

    const banner = page.locator('.mobile-install-banner');
    await expect(banner).toBeVisible();
    await expect(banner).toContainText('Install Ovumcy');

    await page.getByRole('button', { name: 'Install app' }).click();

    await expect
      .poll(async () => {
        return page.evaluate(() => {
          return (window as Window & { __pwaInstallPromptCalls?: number }).__pwaInstallPromptCalls ?? 0;
        });
      })
      .toBe(1);
    await expect(banner).toBeHidden();
  });
});
