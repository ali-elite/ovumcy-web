  function configureHTMXForCSP() {
    if (!window.htmx || !window.htmx.config) {
      return;
    }

    window.htmx.config.allowEval = false;
    window.htmx.config.includeIndicatorStyles = false;
  }

  configureHTMXForCSP();
  initPWAInstallPrompt();

  onDocumentReady(function () {
    initThemePreference();
    initAuthPanelTransitions();
    initLanguageSwitcher();
    initClientTimezone();
    initPasswordToggles();
    initLoginValidation();
    initRegisterValidation();
    initLoginPasswordPersistence();
    initConfirmModal();
    initToastAPI();
    initHTMXHooks();
    initCSPFriendlyComponents();

    document.body.addEventListener("htmx:afterSwap", function () {
      initCSPFriendlyComponents();
    });
  });
})();
