  function configureHTMXForCSP() {
    if (!window.htmx || !window.htmx.config) {
      return;
    }

    window.htmx.config.allowEval = false;
    window.htmx.config.includeIndicatorStyles = false;
  }

  configureHTMXForCSP();
  initClientTimezone();
  initPWAInstallPrompt();

  onDocumentReady(function () {
    initThemePreference();
    initAuthPanelTransitions();
    initPasswordToggles();
    initLoginValidation();
    initForgotPasswordValidation();
    initRegisterValidation();
    initSettingsPasswordValidation();
    initResetPasswordValidation();
    initLoginErrorFocus();
    initConfirmModal();
    initClearDataPasswordConfirmation();
    bindCycleStartConfirmForms();
    initPushSubscription();
    initToastAPI();
    initHTMXHooks();
    initCSPFriendlyComponents();

    document.body.addEventListener("htmx:afterSwap", function () {
      initCSPFriendlyComponents();
    });
  });
})();
