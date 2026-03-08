(() => {
  const COOKIE_NAME = "ovumcy_tz";
  const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365;

  function isSafeClientTimezone(value) {
    if (!value || value.length > 128) {
      return false;
    }
    return /^[A-Za-z0-9_+/-]+$/.test(value);
  }

  function detectClientTimezone() {
    try {
      const formatter = Intl && Intl.DateTimeFormat ? Intl.DateTimeFormat() : null;
      const options = formatter && formatter.resolvedOptions ? formatter.resolvedOptions() : null;
      const timezone = options && options.timeZone ? String(options.timeZone).trim() : "";
      if (!isSafeClientTimezone(timezone)) {
        return "";
      }
      return timezone;
    } catch {
      return "";
    }
  }

  function writeClientCookie(name, value, maxAgeSeconds) {
    if (!name || !value) {
      return;
    }

    let cookie = `${name}=${value}; Path=/; SameSite=Lax; Max-Age=${String(maxAgeSeconds || 0)}`;
    if (window.location && window.location.protocol === "https:") {
      cookie += "; Secure";
    }
    document.cookie = cookie;
  }

  const timezone = detectClientTimezone();
  if (!timezone) {
    return;
  }

  window.__ovumcyTimezone = timezone;
  writeClientCookie(COOKIE_NAME, timezone, COOKIE_MAX_AGE_SECONDS);
})();
