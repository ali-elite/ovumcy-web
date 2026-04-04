  function initAuthPanelTransitions() {
    var panel = document.querySelector("[data-auth-panel]");
    if (!panel) {
      return;
    }

    var prefersReducedMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (!prefersReducedMotion) {
      panel.classList.add("auth-panel-transition");
      panel.classList.add("auth-panel-enter");
      window.requestAnimationFrame(function () {
        panel.classList.remove("auth-panel-enter");
      });
    }

    document.addEventListener("click", function (event) {
      var link = closestFromEvent(event, "a[data-auth-switch]");
      if (!link) {
        return;
      }

      if (event.defaultPrevented || !isPrimaryClick(event)) {
        return;
      }
      if (link.getAttribute("target") === "_blank") {
        return;
      }

      var href = (link.getAttribute("href") || "").trim();
      if (!href || prefersReducedMotion) {
        return;
      }

      event.preventDefault();
      panel.classList.add("auth-panel-transition");
      panel.classList.add("auth-panel-exit");
      window.setTimeout(function () {
        window.location.href = href;
      }, 140);
    });
  }

