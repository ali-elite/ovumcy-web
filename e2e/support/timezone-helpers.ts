import { type Page } from '@playwright/test';

const TIMEZONE_COOKIE_NAME = 'ovumcy_tz';
const TIMEZONE_HEADER_NAME = 'X-Ovumcy-Timezone';

export async function browserTimezone(page: Page): Promise<string> {
  return page.evaluate(() => {
    try {
      return String(Intl.DateTimeFormat().resolvedOptions().timeZone || '').trim();
    } catch {
      return '';
    }
  });
}

export async function setRequestTimezoneFromBrowser(page: Page): Promise<string> {
  const timezone = await browserTimezone(page);
  if (!timezone) {
    return '';
  }

  await page.context().setExtraHTTPHeaders({
    [TIMEZONE_HEADER_NAME]: timezone,
  });

  const origin = new URL(page.url()).origin;
  await page.context().addCookies([
    {
      name: TIMEZONE_COOKIE_NAME,
      value: timezone,
      url: origin,
      sameSite: 'Lax',
    },
  ]);

  return timezone;
}
