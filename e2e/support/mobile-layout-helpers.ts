import { expect, type Locator, type Page } from '@playwright/test';

export function mobileTabbar(page: Page): Locator {
  return page.locator('nav.mobile-tabbar');
}

export async function expectElementAboveMobileTabbar(
  page: Page,
  element: Locator,
  options?: { minGap?: number }
): Promise<void> {
  const minGap = options?.minGap ?? 8;
  const tabbar = mobileTabbar(page);

  await expect(tabbar).toBeVisible();
  await expect(element).toBeVisible();

  const [elementBox, tabbarBox] = await Promise.all([element.boundingBox(), tabbar.boundingBox()]);

  expect(elementBox, 'expected target element to have a visible bounding box').not.toBeNull();
  expect(tabbarBox, 'expected mobile tabbar to have a visible bounding box').not.toBeNull();

  const elementBottom = elementBox!.y + elementBox!.height;
  const tabbarTop = tabbarBox!.y;

  expect(elementBottom).toBeLessThanOrEqual(tabbarTop - minGap);
}
