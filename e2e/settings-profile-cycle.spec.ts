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

async function isoDaysAgoInBrowser(page: Page, days: number): Promise<string> {
  return page.evaluate((offset) => {
    const date = new Date();
    date.setHours(0, 0, 0, 0);
    date.setDate(date.getDate() - offset);

    const yyyy = date.getFullYear();
    const mm = String(date.getMonth() + 1).padStart(2, '0');
    const dd = String(date.getDate()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd}`;
  }, days);
}

async function browserTimezone(page: Page): Promise<string> {
  return page.evaluate(() => {
    try {
      return String(Intl.DateTimeFormat().resolvedOptions().timeZone || '').trim();
    } catch {
      return '';
    }
  });
}

function shiftISODate(iso: string, days: number): string {
  const [y, m, d] = iso.split('-').map((part) => Number(part));
  const date = new Date(y, m - 1, d);
  date.setDate(date.getDate() + days);
  return toISODate(date);
}

async function setRangeValue(locator: Locator, value: number): Promise<void> {
  await locator.evaluate((element, rawValue) => {
    const input = element as HTMLInputElement;
    input.value = String(rawValue);
    input.dispatchEvent(new Event('input', { bubbles: true }));
    input.dispatchEvent(new Event('change', { bubbles: true }));
  }, value);
}

async function assertNoHorizontalOverflow(page: Page): Promise<void> {
  const hasOverflow = await page.evaluate(() => {
    const root = document.documentElement;
    return root.scrollWidth > root.clientWidth + 1;
  });
  expect(hasOverflow).toBe(false);
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

async function completeOnboardingWithStartDate(page: Page, startDate: string): Promise<void> {
  const beginButton = page.locator('[data-onboarding-action="begin"]');
  if (await beginButton.isVisible().catch(() => false)) {
    await beginButton.click();
  }

  const startDateInput = page.locator('#last-period-start');
  await expect(startDateInput).toBeVisible();

  const startDateOption = page.locator(`[data-onboarding-day-option][data-onboarding-day-value="${startDate}"]`);
  if ((await startDateOption.count()) > 0) {
    await startDateOption.first().click();
  } else {
    await startDateInput.fill(startDate);
  }

  await page.locator('form[hx-post="/onboarding/step1"] button[type="submit"]').click();

  const stepTwoForm = page.locator('form[hx-post="/onboarding/step2"]');
  await expect(stepTwoForm).toBeVisible();
  await stepTwoForm.locator('button[type="submit"]').click();

  const stepThreeForm = page.locator('form[hx-post="/onboarding/complete"]');
  await expect(stepThreeForm).toBeVisible();
  await stepThreeForm.locator('button[type="submit"]').click();
  await expect(page).toHaveURL(/\/dashboard(?:\?.*)?$/);
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

  test('onboarding selected start date persists into settings cycle field', async ({ page }) => {
    const creds = createCredentials('settings-onboarding-date');

    await registerOwnerViaUI(page, creds);
    await expect(page).toHaveURL(/\/recovery-code$/);

    await readRecoveryCode(page);
    await continueFromRecoveryCode(page);
    await expect(page).toHaveURL(/\/onboarding(?:\?.*)?$/);

    const selectedStart = await isoDaysAgoInBrowser(page, 9);
    await completeOnboardingWithStartDate(page, selectedStart);

    const expectedTimezone = await browserTimezone(page);
    const timezoneCookie = (await page.context().cookies()).find((cookie) => cookie.name === 'ovumcy_tz');
    expect(timezoneCookie?.value || '').toBe(expectedTimezone);

    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings$/);
    await expect(page.locator('#settings-last-period-start')).toHaveValue(selectedStart);
  });

  test('custom symptoms can be created, hidden, restored, and renamed without losing old entries', async ({
    page,
  }) => {
    await registerOwnerAndOpenSettings(page, 'settings-custom-symptoms');

    const symptomSection = page.locator('#settings-symptoms-section');
    await expect(symptomSection).toBeVisible();

    const createForm = symptomSection.locator('[data-symptom-create-form]');
    await createForm.locator('#settings-new-symptom-name').fill('Joint stiffness');
    await createForm.locator('#settings-new-symptom-icon').fill('J');
    await createForm.locator('[data-color-preset="#64748B"]').click();
    await expect(createForm.locator('#settings-new-symptom-color')).toHaveValue('#64748B');
    await createForm.locator('button[type="submit"]').click();

    await expect(symptomSection.locator('.status-ok')).toBeVisible();
    await expect(
      symptomSection.locator('[data-custom-symptom-row][data-symptom-name="Joint stiffness"][data-symptom-state="active"]')
    ).toBeVisible();

    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard$/);

    const periodToggle = page.locator('input[name="is_period"]');
    const customSymptom = page.locator('input[name="symptom_ids"][data-symptom-name="Joint stiffness"]');
    await periodToggle.check();
    await expect(customSymptom).toBeVisible();
    await customSymptom.check({ force: true });
    await page.locator('button[data-save-button]').first().click();
    await expect(page.locator('#save-status .status-ok')).toBeVisible();

    const todayAction = await page.locator('form[hx-post^="/api/days/"]').first().getAttribute('hx-post');
    expect(todayAction).toMatch(/^\/api\/days\/\d{4}-\d{2}-\d{2}$/);
    const todayISO = String(todayAction).replace('/api/days/', '');
    const otherISO = shiftISODate(todayISO, 3);
    const otherMonth = otherISO.slice(0, 7);

    await page.goto('/settings');
    await expect(page).toHaveURL(/\/settings$/);

    const activeRow = page.locator(
      '[data-custom-symptom-row][data-symptom-name="Joint stiffness"][data-symptom-state="active"]'
    );
    await activeRow.locator('form[action$="/archive"] button[type="submit"]').click();
    await expect(symptomSection.locator('.status-ok')).toBeVisible();
    await expect(
      symptomSection.locator('[data-custom-symptom-row][data-symptom-name="Joint stiffness"][data-symptom-state="archived"]')
    ).toBeVisible();

    await page.goto('/dashboard');
    await expect(page.locator('input[name="symptom_ids"][data-symptom-name="Joint stiffness"]')).toBeVisible();
    await expect(page.locator('input[name="symptom_ids"][data-symptom-name="Joint stiffness"]')).toBeChecked();

    await page.goto(`/calendar?month=${otherMonth}&day=${otherISO}`);
    await expect(page).toHaveURL(new RegExp(`/calendar\\?month=${otherMonth}&day=${otherISO}`));
    await expect(page.locator(`form.calendar-day-editor-form[hx-post="/api/days/${otherISO}"]`)).toBeVisible();
    await expect(
      page.locator(`form.calendar-day-editor-form[hx-post="/api/days/${otherISO}"] input[name="symptom_ids"][data-symptom-name="Joint stiffness"]`)
    ).toHaveCount(0);

    await page.goto('/settings');
    const archivedRow = page.locator(
      '[data-custom-symptom-row][data-symptom-name="Joint stiffness"][data-symptom-state="archived"]'
    );
    await archivedRow.locator('input[name="name"]').fill('Joint support');
    await archivedRow.locator('input[name="icon"]').fill('S');
    await archivedRow.locator('input[name="color"]').fill('#556677');
    await archivedRow.locator('[data-symptom-edit-form] button[type="submit"]').click();
    await expect(symptomSection.locator('.status-ok')).toBeVisible();

    const renamedArchivedRow = page.locator(
      '[data-custom-symptom-row][data-symptom-name="Joint support"][data-symptom-state="archived"]'
    );
    await expect(renamedArchivedRow).toBeVisible();
    await renamedArchivedRow.locator('form[action$="/restore"] button[type="submit"]').click();
    await expect(symptomSection.locator('.status-ok')).toBeVisible();
    await expect(
      symptomSection.locator('[data-custom-symptom-row][data-symptom-name="Joint support"][data-symptom-state="active"]')
    ).toBeVisible();

    await page.goto(`/calendar?month=${otherMonth}&day=${otherISO}`);
    await expect(
      page.locator(`form.calendar-day-editor-form[hx-post="/api/days/${otherISO}"] input[name="symptom_ids"][data-symptom-name="Joint support"]`)
    ).toBeVisible();
  });

  test('custom symptom validation blocks duplicate, built-in, and markup names without layout overflow', async ({
    page,
  }) => {
    await registerOwnerAndOpenSettings(page, 'settings-custom-symptom-validation');

    const symptomSection = page.locator('#settings-symptoms-section');
    const createForm = symptomSection.locator('[data-symptom-create-form]');

    await createForm.locator('#settings-new-symptom-name').fill('Joint stiffness');
    await createForm.locator('#settings-new-symptom-icon').fill('J');
    await createForm.locator('[data-color-preset="#8B5CF6"]').click();
    await expect(createForm.locator('#settings-new-symptom-color')).toHaveValue('#8B5CF6');
    await createForm.locator('button[type="submit"]').click();
    await expect(symptomSection.locator('.status-ok')).toBeVisible();
    await expect(
      symptomSection.locator('[data-custom-symptom-row][data-symptom-name="Joint stiffness"][data-symptom-state="active"]')
    ).toBeVisible();

    await createForm.locator('#settings-new-symptom-name').fill(' joint STIFFNESS ');
    await createForm.locator('#settings-new-symptom-icon').fill('K');
    await createForm.locator('#settings-new-symptom-color').fill('#334455');
    await createForm.locator('button[type="submit"]').click();
    await expect(symptomSection.locator('.status-error')).toBeVisible();
    await expect(
      symptomSection.locator('[data-custom-symptom-row][data-symptom-name="Joint stiffness"]')
    ).toHaveCount(1);

    await createForm.locator('#settings-new-symptom-name').fill('Усталость');
    await createForm.locator('button[type="submit"]').click();
    await expect(symptomSection.locator('.status-error')).toBeVisible();

    await createForm.locator('#settings-new-symptom-name').fill('<script>alert(1)</script>');
    await createForm.locator('button[type="submit"]').click();
    await expect(symptomSection.locator('.status-error')).toBeVisible();

    const longName = `${'А'.repeat(18)} ${'Головокружение'.repeat(2)} ${'Б'.repeat(18)}`;
    await createForm.locator('#settings-new-symptom-name').fill(longName);
    await createForm.locator('#settings-new-symptom-icon').fill('L');
    await createForm.locator('#settings-new-symptom-color').fill('#556677');
    await createForm.locator('button[type="submit"]').click();
    await expect(symptomSection.locator('.status-ok')).toBeVisible();
    await expect(
      symptomSection.locator(`[data-custom-symptom-row][data-symptom-name="${longName}"][data-symptom-state="active"]`)
    ).toBeVisible();

    await assertNoHorizontalOverflow(page);

    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard$/);
    await page.locator('input[name="is_period"]').check();
    await expect(
      page.locator(`input[name="symptom_ids"][data-symptom-name="${longName}"]`)
    ).toBeVisible();
    await assertNoHorizontalOverflow(page);

    await page.goto('/calendar');
    await expect(page).toHaveURL(/\/calendar(?:\?.*)?$/);
    await expect(page.locator('#calendar-grid-panel')).toBeVisible();
    await assertNoHorizontalOverflow(page);
  });
});
