// agentUsage serve — thin client glue for the server-rendered dashboard.
// All navigation, refresh, layout, theme and usage-mode actions are htmx
// requests; this file only covers bearer auth, modal/keyboard ergonomics.
(() => {
  "use strict";
  const TOKEN_KEY = "au-serve-token";
  const $ = (id) => document.getElementById(id);
  const click = (id) => { const el = $(id); if (el) el.click(); };

  document.addEventListener("htmx:configRequest", (ev) => {
    const token = sessionStorage.getItem(TOKEN_KEY);
    if (token) ev.detail.headers.Authorization = "Bearer " + token;
  });

  const updateModalScrollLock = () => {
    const inspectOpen = $("inspect-modal") && !$("inspect-modal").hidden;
    const providersOpen = $("providers-panel") && !$("providers-panel").hidden;
    const filterOpen = $("filter-bar") && !$("filter-bar").hidden;
    const tokenOpen = $("token-modal") && !$("token-modal").hidden;
    const anyOpen = Boolean(inspectOpen || providersOpen || filterOpen || tokenOpen);
    document.body.classList.toggle("modal-open", anyOpen);
  };

  document.addEventListener("htmx:responseError", (ev) => {
    if (!ev.detail.xhr || ev.detail.xhr.status !== 401) return;
    const error = $("token-error");
    if (error) {
      error.textContent = sessionStorage.getItem(TOKEN_KEY) ? "Invalid token" : "Token required";
      error.hidden = false;
    }
    if ($("token-modal")) {
      $("token-modal").hidden = false;
      updateModalScrollLock();
    }
    if ($("token-input")) $("token-input").focus();
  });

  if ($("token-form")) {
    $("token-form").addEventListener("submit", (ev) => {
      ev.preventDefault();
      sessionStorage.setItem(TOKEN_KEY, $("token-input").value.trim());
      $("token-modal").hidden = true;
      updateModalScrollLock();
      htmx.ajax("GET", "partial/app", { target: "#app", swap: "outerHTML" });
    });
  }

  let lastFocus = null;
  const closeInspect = () => {
    if ($("inspect-modal")) $("inspect-modal").hidden = true;
    updateModalScrollLock();
    if (lastFocus && typeof lastFocus.focus === "function") {
      lastFocus.focus();
      lastFocus = null;
    }
  };
  const openInspectModal = () => {
    const modal = $("inspect-modal");
    if (!modal) return;
    lastFocus = document.activeElement;
    modal.hidden = false;
    updateModalScrollLock();
    const card = modal.querySelector(".inspect-card");
    if (card) card.scrollTop = 0;
    const closeBtn = $("inspect-close");
    if (closeBtn) closeBtn.focus();
  };

  let lastProvidersFocus = null;
  const closeProviders = () => {
    if ($("providers-panel")) $("providers-panel").hidden = true;
    updateModalScrollLock();
    if (lastProvidersFocus && typeof lastProvidersFocus.focus === "function") {
      lastProvidersFocus.focus();
      lastProvidersFocus = null;
    }
  };
  const openProvidersPanel = () => {
    const panel = $("providers-panel");
    if (!panel) return;
    if (panel.hidden) {
      lastProvidersFocus = document.activeElement;
      panel.hidden = false;
      const closeBtn = $("providers-close");
      if (closeBtn) closeBtn.focus();
    }
    updateModalScrollLock();
  };

  const syncTheme = () => {
    const sel = $("footer-theme-select");
    const studio = document.querySelector(".brand-studio");
    const themeName = (sel && sel.value) || (studio && studio.textContent.trim()) || "";
    if (themeName) {
      const slug = themeName.toLowerCase().replace(/\s+/g, "-");
      document.documentElement.setAttribute("data-theme", slug);
      document.documentElement.dataset.theme = slug;
      document.title = "agentUsage " + themeName;
    }
  };

  const syncFilter = () => {
    const input = $("filter-input");
    if (!input) return;
    const hasQuery = document.querySelector(".search-box.has-query");
    const textEl = document.querySelector(".search-box .search-text");
    if (hasQuery && textEl) {
      const val = textEl.textContent.trim();
      if (input.value !== val) input.value = val;
    } else if (!hasQuery && document.querySelector(".search-box")) {
      if (input.value !== "") input.value = "";
    }
  };

  let toastTimer = null;
  const scheduleToastDismiss = () => {
    const sb = $("status-bar");
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    if (sb) {
      if (!sb.textContent.trim()) {
        sb.hidden = true;
      } else if (!sb.hidden) {
        toastTimer = setTimeout(() => {
          if (sb) sb.hidden = true;
        }, 5000);
      }
    }
  };

  const showToast = (msg) => {
    const sb = $("status-bar");
    if (!sb) return;
    sb.textContent = msg;
    sb.classList.add("toast");
    sb.hidden = false;
    scheduleToastDismiss();
  };

  const ensureScrollContainer = () => {
    const panel = document.querySelector(".panel");
    if (panel && !panel.hasAttribute("tabindex")) {
      panel.setAttribute("tabindex", "0");
    }
    const nav = document.querySelector(".nav");
    if (nav && !nav.hasAttribute("tabindex")) {
      nav.setAttribute("tabindex", "0");
    }
  };

  const scrollSelectedIntoView = () => {
    const selected = document.querySelector(".nav-item.selected, .agent-card.selected, .agent.selected, .bento-tile.selected, .matrix-row.selected");
    if (!selected) return;
    if (typeof selected.scrollIntoView === "function") {
      selected.scrollIntoView({ block: "nearest", inline: "nearest" });
    }
    const container = selected.closest(".panel, .nav");
    if (container) {
      const selRect = selected.getBoundingClientRect();
      const contRect = container.getBoundingClientRect();
      const extraBottom = selRect.bottom - (contRect.bottom - 24);
      if (extraBottom > 0) {
        container.scrollTop += extraBottom;
      }
      const extraTop = (contRect.top + 16) - selRect.top;
      if (extraTop > 0) {
        container.scrollTop -= extraTop;
      }
    }
  };

  const layoutScrollPositions = {};
  let lastSwapLayout = "";
  let savedPanelScroll = { top: 0, left: 0 };
  let savedNavScroll = { top: 0, left: 0 };

  document.addEventListener("htmx:beforeSwap", (ev) => {
    const target = ev.detail.target;
    if (!target || (target.id !== "app" && !target.classList?.contains("shell"))) return;
    const app = $("app");
    if (!app) return;
    const layout = app.dataset.layout || "";
    lastSwapLayout = layout;
    const panel = app.querySelector(".panel");
    const nav = app.querySelector(".nav");
    savedPanelScroll = {
      top: panel ? panel.scrollTop : 0,
      left: panel ? panel.scrollLeft : 0
    };
    savedNavScroll = {
      top: nav ? nav.scrollTop : 0,
      left: nav ? nav.scrollLeft : 0
    };
    if (layout) {
      layoutScrollPositions[layout] = {
        panel: { ...savedPanelScroll },
        nav: { ...savedNavScroll }
      };
    }
  });

  const restoreScrollPosition = () => {
    const app = $("app");
    if (!app) return;
    const layout = app.dataset.layout || "";
    const panel = app.querySelector(".panel");
    const nav = app.querySelector(".nav");
    if (panel) {
      if (layout === lastSwapLayout) {
        panel.scrollTop = savedPanelScroll.top;
        panel.scrollLeft = savedPanelScroll.left;
      } else if (layoutScrollPositions[layout]?.panel) {
        panel.scrollTop = layoutScrollPositions[layout].panel.top;
        panel.scrollLeft = layoutScrollPositions[layout].panel.left;
      }
    }
    if (nav) {
      if (layout === lastSwapLayout) {
        nav.scrollTop = savedNavScroll.top;
        nav.scrollLeft = savedNavScroll.left;
      } else if (layoutScrollPositions[layout]?.nav) {
        nav.scrollTop = layoutScrollPositions[layout].nav.top;
        nav.scrollLeft = layoutScrollPositions[layout].nav.left;
      }
    }
  };

  document.addEventListener("htmx:afterSwap", (ev) => {
    if (ev.detail.target && ev.detail.target.id === "inspect-content") {
      openInspectModal();
    }
    if (ev.detail.target && ev.detail.target.id === "providers-content") {
      openProvidersPanel();
    }
    const target = ev.detail.target;
    if (target && (target.id === "app" || target.classList?.contains("shell"))) {
      restoreScrollPosition();
    }
    ensureScrollContainer();
    scrollSelectedIntoView();
    syncTheme();
    syncFilter();
    scheduleToastDismiss();
    updateModalScrollLock();
  });

  if ($("inspect-close")) $("inspect-close").addEventListener("click", closeInspect);
  if ($("inspect-modal")) {
    $("inspect-modal").addEventListener("click", (ev) => { if (ev.target === $("inspect-modal")) closeInspect(); });
  }
  if ($("providers-close")) $("providers-close").addEventListener("click", closeProviders);
  document.addEventListener("click", (ev) => {
    const panel = $("providers-panel");
    if (!panel || panel.hidden) return;
    const target = ev.target;
    if (target && target.closest && (target.closest("#providers-panel") || target.closest("#footer-btn-providers"))) return;
    closeProviders();
  });

  let filterBackup = "";
  const closeFilter = (apply) => {
    const bar = $("filter-bar"), input = $("filter-input");
    if (!bar || !input) return;
    if (apply) {
      htmx.trigger(input, "search");
    } else {
      input.value = filterBackup;
      htmx.trigger(input, "search");
    }
    bar.hidden = true;
    updateModalScrollLock();
    input.blur();
  };
  const openFilter = () => {
    const bar = $("filter-bar"), input = $("filter-input");
    if (!bar || !input) return;
    filterBackup = input.value;
    bar.hidden = false;
    updateModalScrollLock();
    input.focus();
    input.select();
  };
  const clearFilter = () => {
    const input = $("filter-input");
    if (input) {
      input.value = "";
      htmx.trigger(input, "search");
    }
    if (window.history && window.history.replaceState && window.location.search) {
      const url = new URL(window.location.href);
      if (url.searchParams.has("q")) {
        url.searchParams.delete("q");
        const newSearch = url.searchParams.toString();
        const newUrl = url.pathname + (newSearch ? "?" + newSearch : "") + url.hash;
        window.history.replaceState(null, "", newUrl);
      }
    }
    document.cookie = "au_filter=; Path=/; Max-Age=0; SameSite=Lax";
  };

  // Footer buttons live inside the swapped #app fragment: delegate.
  document.addEventListener("click", (ev) => {
    const target = ev.target;
    if (!target || !target.closest) return;
    if (target.closest(".btn-search-clear, .btn-clear-filter")) {
      ev.preventDefault();
      ev.stopPropagation();
      clearFilter();
      return;
    }
    if (target.closest(".search-box") || target.closest("#footer-btn-filter")) {
      openFilter();
      return;
    }
    if (target.closest(".layout-btn") && document.querySelector(".empty-filter-state")) {
      clearFilter();
    }
    if (ev.shiftKey && target.closest("#footer-btn-theme")) {
      ev.preventDefault();
      ev.stopPropagation();
      click("key-theme-back");
      return;
    }
    if (target.closest("#status-bar")) {
      const sb = $("status-bar");
      if (sb) sb.hidden = true;
      return;
    }
  }, true);

  document.addEventListener("contextmenu", (ev) => {
    if (ev.target && ev.target.closest && ev.target.closest("#footer-btn-theme")) {
      ev.preventDefault();
      click("key-theme-back");
    }
  });
  if ($("filter-close")) $("filter-close").addEventListener("click", () => closeFilter(true));
  if ($("filter-input")) {
    $("filter-input").addEventListener("keydown", (ev) => {
      if (ev.key === "Enter") {
        ev.preventDefault();
        closeFilter(true);
      }
    });
  }

  const selectedRow = (sel) => document.querySelector(sel);
  const rowAction = () => {
    const layout = ($("app") || {}).dataset ? $("app").dataset.layout : "";
    if (layout === "matrix") selectedRow(".matrix-row.selected")?.click();
    else if (layout === "bento") selectedRow(".bento-tile.selected")?.click();
    else if (layout === "split") {
      selectedRow(".nav-item.selected")?.click();
    }
    else selectedRow(".agent-card.selected, .agent.selected")?.click();
  };

  const trapFocus = (container, ev) => {
    if (ev.key !== "Tab" || !container || container.hidden) return;
    const focusables = container.querySelectorAll('button:not([disabled]):not([hidden]), [href], input:not([disabled]):not([hidden]), select:not([disabled]):not([hidden]), textarea:not([disabled]):not([hidden]), [tabindex]:not([tabindex="-1"])');
    if (!focusables || focusables.length === 0) return;
    const first = focusables[0];
    const last = focusables[focusables.length - 1];
    if (ev.shiftKey && document.activeElement === first) {
      ev.preventDefault();
      last.focus();
    } else if (!ev.shiftKey && document.activeElement === last) {
      ev.preventDefault();
      first.focus();
    }
  };

  document.addEventListener("keydown", (ev) => {
    if (ev.key === "Escape") {
      if ($("filter-bar") && !$("filter-bar").hidden) {
        ev.preventDefault();
        closeFilter(false);
        return;
      }
      if ($("providers-panel") && !$("providers-panel").hidden) {
        ev.preventDefault();
        closeProviders();
        return;
      }
      if ($("inspect-modal") && !$("inspect-modal").hidden) {
        ev.preventDefault();
        closeInspect();
        return;
      }
      if (document.querySelector(".search-box.has-query") || document.querySelector(".empty-filter-state")) {
        ev.preventDefault();
        clearFilter();
        return;
      }
      return;
    }

    const inspectModal = $("inspect-modal");
    if (inspectModal && !inspectModal.hidden) {
      trapFocus(inspectModal, ev);
    }
    const providersPanel = $("providers-panel");
    if (providersPanel && !providersPanel.hidden) {
      trapFocus(providersPanel, ev);
    }
    if (ev.key === "Enter" || ev.key === " ") {
      const active = document.activeElement;
      if (active && active.classList && (active.classList.contains("bento-tile") || active.classList.contains("matrix-row"))) {
        ev.preventDefault();
        active.click();
      }
    }
  });

  document.addEventListener("keydown", (ev) => {
    if ($("token-modal") && !$("token-modal").hidden) return;
    if ($("filter-bar") && !$("filter-bar").hidden) return;
    if (ev.ctrlKey || ev.metaKey || ev.altKey) return;
    const target = ev.target;
    if (target && target.matches && (target.matches("input, textarea, select") || target.isContentEditable)) return;
    switch (ev.key) {
      case "Tab": ev.preventDefault(); click("key-layout-main"); break;
      case "l": case "L": click("key-layout-main"); break;
      case "ArrowUp": case "k": case "K": click("key-prev"); break;
      case "ArrowDown": case "j": case "J": case "n": case "N": click("key-next"); break;
      case "/": ev.preventDefault(); openFilter(); break;
      case "?": ev.preventDefault(); showToast("Keys: j/k or n (nav), / (filter), u (mode), r/R (refresh), t/T (theme), v (layout), p (boxes)"); break;
      case "r": click("key-refresh"); break;
      case "R": click("key-refresh-all"); break;
      case "u": case "U": click("key-mode"); break;
      case "t": click("key-theme-fwd"); break;
      case "T": click("key-theme-back"); break;
      case "v": case "V": click("key-layout-cycle"); break;
      case "p": case "P":
        if ($("providers-panel") && !$("providers-panel").hidden) closeProviders();
        else click("footer-btn-providers");
        break;
      case "Enter": case " ": rowAction(); break;
    }
  });

  window.addEventListener("popstate", () => {
    const params = new URLSearchParams(window.location.search);
    const q = params.get("q") || "";
    const input = $("filter-input");
    if (input && input.value !== q) {
      input.value = q;
      htmx.trigger(input, "search");
    }
    syncFilter();
  });

  (function sheetDrag() {
    const modal = $("inspect-modal");
    const card = modal && modal.querySelector(".inspect-card");
    let startY = null;
    if (!modal || !card) return;
    modal.addEventListener("touchstart", (ev) => {
      if (ev.target.closest(".mobile-sheet-handle, .inspect-header")) startY = ev.touches[0].clientY;
    }, { passive: true });
    modal.addEventListener("touchmove", (ev) => {
      if (startY === null) return;
      card.style.transform = "translateY(" + Math.max(0, ev.touches[0].clientY - startY) + "px)";
    }, { passive: true });
    modal.addEventListener("touchend", (ev) => {
      if (startY !== null && (ev.changedTouches[0] || {}).clientY - startY > 60) closeInspect();
      card.style.transform = "";
      startY = null;
    });
  })();

  ensureScrollContainer();
  syncTheme();
  syncFilter();
  scheduleToastDismiss();
  updateModalScrollLock();
})();
