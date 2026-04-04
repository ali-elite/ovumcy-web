import { expect, type Page } from '@playwright/test';

export async function switchPublicLanguage(page: Page, code: string): Promise<void> {
  const form = page.locator('[data-language-switch-form]');
  const button = form.locator(`[data-language-switch-option="${code}"]`);

  await expect(form).toBeVisible();
  await expect(button).toBeVisible();

  await Promise.all([
    page.waitForNavigation({ waitUntil: 'domcontentloaded' }),
    button.click(),
  ]);

  await expect(button).toHaveAttribute('aria-pressed', 'true');
}

export async function saveSettingsLanguage(page: Page, code: string): Promise<void> {
  const form = page.locator('[data-settings-interface-form]');
  const option = form.locator(`[data-settings-interface-language-option="${code}"]`);

  await expect(form).toBeVisible();
  await expect(option).toBeVisible();

  if ((await option.getAttribute('data-selected')) !== 'true') {
    await option.locator('.radio-tile').click();
    await form.locator('[data-settings-interface-save]').click();
  }

  await expect(option).toHaveAttribute('data-selected', 'true');
}
