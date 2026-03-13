import { expect, type Locator, type Page } from '@playwright/test';

type NoteScope = Locator | Page;

export async function ensureNotesFieldVisible(
  scope: NoteScope,
  fieldSelector: string
): Promise<Locator> {
  const field = scope.locator(fieldSelector).first();
  if ((await field.count()) > 0 && (await field.isVisible())) {
    return field;
  }

  const disclosure = scope.locator('details.note-disclosure').first();
  await expect(disclosure).toHaveCount(1);

  const isOpen = await disclosure.evaluate((element) => element.hasAttribute('open'));
  if (!isOpen) {
    await disclosure.locator('summary').click();
  }

  await expect(field).toBeVisible();
  return field;
}
