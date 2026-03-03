import { expect, test, type Locator, type Page } from '@playwright/test';
import {
  completeOnboardingIfPresent,
  continueFromRecoveryCode,
  createCredentials,
  readRecoveryCode,
  registerOwnerViaUI,
} from './support/auth-helpers';

function toISODate(date: Date): string {
  const copy = new Date(date);
  copy.setHours(0, 0, 0, 0);
  const yyyy = copy.getFullYear();
  const mm = String(copy.getMonth() + 1).padStart(2, '0');
  const dd = String(copy.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

function isoDaysAgo(days: number): string {
  return toISODate(new Date(Date.now() - days * 24 * 60 * 60 * 1000));
}

function isoDaysFromNow(days: number): string {
  return toISODate(new Date(Date.now() + days * 24 * 60 * 60 * 1000));
}

async function setRangeValue(locator: Locator, value: number): Promise<void> {
  await locator.evaluate((element, rawValue) => {
    const input = element as HTMLInputElement;
    input.value = String(rawValue);
    input.dispatchEvent(new Event('input', { bubbles: true }));
    input.dispatchEvent(new Event('change', { bubbles: true }));
  }, value);
}

async function registerOwnerAndOpenSettings(page: Page, prefix: string) {
  const creds = createCredentials(prefix);

  await registerOwnerViaUI(page, creds);
  await expect(page).toHaveURL(/\/recovery-code$/);

  await readRecoveryCode(page);
  await continueFromRecoveryCode(page);
  await completeOnboardingIfPresent(page);

  await page.goto('/settings');
  await expect(page).toHaveURL(/\/settings$/);

  return creds;
}

async function currentNextPeriodText(page: Page): Promise<string> {
  const value = await page
    .locator('article.stat-card')
    .nth(2)
    .locator('.stat-value')
    .textContent();

  return String(value || '').trim();
}

test.describe('Settings: profile and cycle', () => {
  test('profile name updates nav identity, long value is rejected, empty clears to email fallback', async ({
    page,
  }) => {
    const creds = await registerOwnerAndOpenSettings(page, 'settings-profile');

    const profileEmail = page.locator('#settings-profile-email');
    await expect(profileEmail).toHaveAttribute('readonly', '');

    const displayNameInput = page.locator('#settings-display-name');
    const saveProfileButton = page.locator(
      'form[action="/api/settings/profile"] button[data-save-button]'
    );

    const newName = `Profile-${Date.now()}`;
    await displayNameInput.fill(newName);
    await saveProfileButton.click();
    await expect(page.locator('#settings-profile-status .status-ok')).toBeVisible();

    await page.reload();
    await expect(page).toHaveURL(/\/settings$/);
    await expect(page.locator('.nav-user-chip')).toContainText(newName);

    await displayNameInput.evaluate((el) => {
      (el as HTMLInputElement).value = 'X'.repeat(80);
    });
    await saveProfileButton.click();
    await expect(page.locator('#settings-profile-status .status-error')).toBeVisible();

    await displayNameInput.fill('');
    await saveProfileButton.click();
    await expect(page.locator('#settings-profile-status .status-ok')).toBeVisible();

    const fallbackIdentity = creds.email.split('@')[0];
    await page.reload();
    await expect(page.locator('.nav-user-chip')).toContainText(fallbackIdentity);
  });

  test('cycle settings persist, affect dashboard predictions, and reject future last-period date', async ({
    page,
  }) => {
    await registerOwnerAndOpenSettings(page, 'settings-cycle');

    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard$/);
    const nextPeriodBefore = await currentNextPeriodText(page);

    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings$/);

    const cycleForm = page.locator('section#settings-cycle form[action="/settings/cycle"]');
    await expect(cycleForm).toBeVisible();

    const cycleLength = cycleForm.locator('#settings-cycle-length');
    const periodLength = cycleForm.locator('#settings-period-length');
    const lastPeriodStart = cycleForm.locator('#settings-last-period-start');
    const autoFill = cycleForm.locator('input[name="auto_period_fill"]');

    const targetCycleLength = 35;
    const targetPeriodLength = 6;
    const targetStart = isoDaysAgo(20);

    await setRangeValue(cycleLength, targetCycleLength);
    await setRangeValue(periodLength, targetPeriodLength);
    await lastPeriodStart.fill(targetStart);
    await autoFill.uncheck();

    await cycleForm.locator('button[data-save-button]').click();
    await expect(page.locator('#settings-cycle-status .status-ok')).toBeVisible();

    await page.reload();
    await expect(page).toHaveURL(/\/settings$/);

    await expect(page.locator('#settings-cycle-length')).toHaveValue(String(targetCycleLength));
    await expect(page.locator('#settings-period-length')).toHaveValue(String(targetPeriodLength));
    await expect(page.locator('#settings-last-period-start')).toHaveValue(targetStart);
    await expect(page.locator('section#settings-cycle input[name="auto_period_fill"]')).not.toBeChecked();

    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard$/);
    const nextPeriodAfter = await currentNextPeriodText(page);
    expect(nextPeriodAfter).not.toBe(nextPeriodBefore);

    await page.goto('/calendar');
    await expect(page).toHaveURL(/\/calendar(?:\?.*)?$/);
    await expect(page.locator('#calendar-grid-panel')).toBeVisible();

    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings$/);

    await page.locator('#settings-last-period-start').fill(isoDaysFromNow(1));
    await page
      .locator('section#settings-cycle form[action="/settings/cycle"] button[data-save-button]')
      .click();

    await expect(page.locator('#settings-cycle-status .status-error')).toBeVisible();
  });
});
