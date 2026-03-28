import { expect, test } from '@playwright/test';
import { expectNoSensitiveAuthParams } from './support/auth-helpers';

test.describe('Auth: OIDC login entry', () => {
  test.skip(process.env.OIDC_ENABLED !== 'true', 'Requires OIDC_ENABLED=true');

  test('shows SSO CTA and falls back to login with safe error UX', async ({ page }) => {
    await page.goto('/login');
    await expect(page).toHaveURL(/\/login(?:\?.*)?$/);

    await expect(page.locator('#login-form')).toBeVisible();
    const ssoCTA = page.locator('[data-auth-sso-cta]');
    await expect(ssoCTA).toBeVisible();
    await expect(ssoCTA).toContainText('Sign in with SSO');

    await ssoCTA.click();

    await expect(page).toHaveURL(/\/login$/);
    expectNoSensitiveAuthParams(page.url());
    await expect(page.locator('[data-auth-server-error]')).toContainText(
      'SSO sign-in is currently unavailable.'
    );
    await expect(page.locator('#login-form')).toBeVisible();
  });
});
