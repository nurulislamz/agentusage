// agentUsage serve — thin client glue for the server-rendered dashboard.
// All navigation, refresh, layout, theme and usage-mode actions are htmx
// requests; this file only covers bearer auth, modal/keyboard ergonomics.
(() => {
  "use strict";
  const TOKEN_KEY = "au-serve-token";
  const KEYS_HINT = "keys: j/k or n navigate · / filter · u usage mode · r refresh · R refresh all · t/T theme · v layout · p providers · ? shortcuts";
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
    if (!ev.isTrusted) return;
    const panel = $("providers-panel");
    if (!panel || panel.hidden) return;
    const target = ev.target;
    if (target && target.closest && (target.closest("#providers-panel") || target.closest("#footer-btn-providers"))) return;
    if (ev.composedPath && ev.composedPath().some((el) => el === panel || el.id === "footer-btn-providers")) return;
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
    if (target.closest(".search-box")) {
      openFilter();
      return;
    }
    if (target.closest("#footer-btn-keys")) {
      showToast(KEYS_HINT);
      return;
    }
    if (target.closest(".layout-btn") && document.querySelector(".empty-filter-state")) {
      clearFilter();
    }
    if (target.closest("#status-bar")) {
      const sb = $("status-bar");
      if (sb) sb.hidden = true;
      return;
    }
  }, true);
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
      case "?": ev.preventDefault(); showToast(KEYS_HINT); break;
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
  (function pullToRefresh() {
    if (!("ontouchstart" in window) || !window.matchMedia("(pointer: coarse)").matches) return;
    const TRIGGER = 96;
    const MAX_SHIFT = 110;
    const RESIST = 0.35;
    const indicator = document.createElement("div");
    indicator.className = "pull-indicator";
    indicator.setAttribute("aria-hidden", "true");
    indicator.innerHTML = '<span class="pull-spin">↻</span><span class="pull-label">Pull to refresh</span>';
    document.body.appendChild(indicator);
    const label = indicator.querySelector(".pull-label");
    let panel = null, startY = 0, startX = 0, engaged = false, pulled = 0;

    const reset = () => {
      indicator.classList.remove("busy", "show", "over");
      label.textContent = "Pull to refresh";
    };
    document.addEventListener("htmx:afterSwap", reset);
    document.addEventListener("htmx:responseError", reset);

    const onMove = (ev) => {
      if (!panel) return;
      const t = ev.touches[0];
      const dy = t.clientY - startY;
      if (!engaged) {
        if (dy <= 8 || Math.abs(t.clientX - startX) > dy) return;
        engaged = true;
        indicator.classList.add("show");
      }
      ev.preventDefault();
      pulled = dy;
      panel.style.transform = "translateY(" + Math.min(Math.max(dy * RESIST, 0), MAX_SHIFT).toFixed(1) + "px)";
      const over = dy >= TRIGGER;
      indicator.classList.toggle("over", over);
      label.textContent = over ? "Release to refresh" : "Pull to refresh";
    };
    const onEnd = () => {
      const el = panel;
      const fire = engaged && pulled >= TRIGGER;
      panel = null; engaged = false; pulled = 0;
      if (!el) return;
      el.removeEventListener("touchmove", onMove);
      el.removeEventListener("touchend", onEnd);
      el.removeEventListener("touchcancel", onEnd);
      el.style.transition = "transform 180ms cubic-bezier(0.22, 1, 0.36, 1)";
      el.style.transform = "";
      setTimeout(() => { el.style.transition = ""; }, 220);
      if (fire) {
        indicator.classList.add("busy");
        label.textContent = "Refreshing…";
        click("key-refresh-all");
      }
    };
    document.addEventListener("touchstart", (ev) => {
      if (ev.touches.length !== 1 || panel) return;
      let el = ev.touches[0].target;
      for (; el && el !== document.body; el = el.parentElement) {
        if (el.closest("#inspect-modal")) return;
        const ov = getComputedStyle(el).overflowY;
        if (ov === "auto" || ov === "scroll") { panel = el; break; }
      }
      if (!panel || panel.scrollTop > 0) { panel = null; return; }
      startY = ev.touches[0].clientY;
      startX = ev.touches[0].clientX;
      panel.addEventListener("touchmove", onMove, { passive: false });
      panel.addEventListener("touchend", onEnd, { passive: true });
      panel.addEventListener("touchcancel", onEnd, { passive: true });
    }, { passive: true });
  })();

  // Provider submenu reordering: move provider groups in the Providers & boxes
  // submenu via drag handle (PointerEvents) or Up/Down switch buttons; the new order
  // persists in settings.json, cookies, and localStorage, and syncs dashboard views.
  (function providerSubmenuReorder() {
    let activeDrag = null;

    const orderFromDOM = () =>
      [...document.querySelectorAll(".providers-group")]
        .map((el) => el.dataset.provider)
        .filter(Boolean)
        .filter((id, i, arr) => arr.indexOf(id) === i);

    const updateButtonStates = () => {
      const groups = document.querySelectorAll(".providers-group");
      groups.forEach((g, idx) => {
        const upBtn = g.querySelector('.provider-move-btn[data-move="up"]');
        const downBtn = g.querySelector('.provider-move-btn[data-move="down"]');
        if (upBtn) {
          upBtn.disabled = (idx === 0);
          upBtn.setAttribute("aria-disabled", idx === 0 ? "true" : "false");
        }
        if (downBtn) {
          downBtn.disabled = (idx === groups.length - 1);
          downBtn.setAttribute("aria-disabled", idx === groups.length - 1 ? "true" : "false");
        }
      });
    };

    const flashGroup = (el) => {
      if (!el) return;
      el.classList.add("moved-highlight");
      setTimeout(() => el.classList.remove("moved-highlight"), 250);
    };

    let saveTimer = null;
    const scheduleSaveOrder = () => {
      if (saveTimer) clearTimeout(saveTimer);
      saveTimer = setTimeout(() => {
        saveTimer = null;
        saveOrder();
      }, 200);
    };

    const saveOrder = () => {
      const order = orderFromDOM();
      if (order.length < 2) return;

      const orderStr = order.join("|");
      try { localStorage.setItem("au_provider_order", orderStr); } catch {}

      const basePath = window.location.pathname.startsWith("/agentusage") ? "/agentusage" : "/";
      document.cookie = `au_provider_order=${encodeURIComponent(orderStr)}; path=${basePath}; max-age=31536000; SameSite=Lax`;

      fetch("actions/provider-order", {
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "X-Requested-With": "XMLHttpRequest",
        },
        body: new URLSearchParams({ order: orderStr, silent: "1" }),
      }).then(() => {
        if (window.htmx) {
          window.htmx.ajax("GET", "partial/app", { target: "#app", swap: "outerHTML" });
        }
      }).catch(() => {});
    };

    // Initial cookie sync from localStorage if needed
    try {
      const saved = localStorage.getItem("au_provider_order");
      if (saved && !document.cookie.includes("au_provider_order=")) {
        const basePath = window.location.pathname.startsWith("/agentusage") ? "/agentusage" : "/";
        document.cookie = `au_provider_order=${encodeURIComponent(saved)}; path=${basePath}; max-age=31536000; SameSite=Lax`;
      }
    } catch {}

    // Up / Down click handlers for quick and clean switching
    document.body.addEventListener("click", (ev) => {
      const btn = ev.target.closest?.(".provider-move-btn");
      if (!btn) return;
      ev.preventDefault();
      ev.stopPropagation();

      const group = btn.closest(".providers-group");
      if (!group) return;
      const dir = btn.dataset.move;
      const parent = group.parentElement;
      if (!parent) return;

      if (dir === "up") {
        const prev = group.previousElementSibling;
        if (prev && prev.classList.contains("providers-group")) {
          parent.insertBefore(group, prev);
          flashGroup(group);
          updateButtonStates();
          btn.focus();
          scheduleSaveOrder();
        }
      } else if (dir === "down") {
        const next = group.nextElementSibling;
        if (next && next.classList.contains("providers-group")) {
          parent.insertBefore(next, group);
          flashGroup(group);
          updateButtonStates();
          btn.focus();
          scheduleSaveOrder();
        }
      }
    });

    // Keyboard ArrowUp / ArrowDown support on drag handle
    document.body.addEventListener("keydown", (ev) => {
      const handle = ev.target.closest?.(".provider-drag-handle");
      if (!handle) return;
      const group = handle.closest(".providers-group");
      if (!group) return;
      const parent = group.parentElement;
      if (!parent) return;

      if (ev.key === "ArrowUp") {
        const prev = group.previousElementSibling;
        if (prev && prev.classList.contains("providers-group")) {
          ev.preventDefault();
          parent.insertBefore(group, prev);
          handle.focus();
          flashGroup(group);
          updateButtonStates();
          scheduleSaveOrder();
        }
      } else if (ev.key === "ArrowDown") {
        const next = group.nextElementSibling;
        if (next && next.classList.contains("providers-group")) {
          ev.preventDefault();
          parent.insertBefore(next, group);
          handle.focus();
          flashGroup(group);
          updateButtonStates();
          scheduleSaveOrder();
        }
      }
    });

    // PointerEvents-based dragging (works smoothly on mouse, trackpad, and touch)
    document.body.addEventListener("pointerdown", (ev) => {
      if (ev.button !== 0 && ev.pointerType === "mouse") return;
      const handle = ev.target.closest?.(".provider-drag-handle");
      if (!handle) return;

      const group = handle.closest(".providers-group");
      if (!group) return;
      const parent = group.parentElement;
      if (!parent) return;

      ev.preventDefault();
      try { handle.setPointerCapture(ev.pointerId); } catch {}

      const initialOrder = orderFromDOM().join("|");
      activeDrag = {
        pointerId: ev.pointerId,
        handle,
        group,
        parent,
        initialOrder,
        hasMoved: false,
      };

      group.classList.add("dragging");
      document.body.classList.add("is-dragging");
    });

    document.body.addEventListener("pointermove", (ev) => {
      if (!activeDrag || ev.pointerId !== activeDrag.pointerId) return;

      const { group, parent } = activeDrag;
      const groups = Array.from(parent.children).filter((el) => el.classList.contains("providers-group"));
      const idxThis = groups.indexOf(group);
      if (idxThis === -1) return;

      for (let i = 0; i < groups.length; i++) {
        const target = groups[i];
        if (target === group) continue;
        const rect = target.getBoundingClientRect();
        if (ev.clientY >= rect.top && ev.clientY <= rect.bottom) {
          const mid = rect.top + rect.height / 2;
          if (i < idxThis && ev.clientY < mid) {
            parent.insertBefore(group, target);
            activeDrag.hasMoved = true;
            updateButtonStates();
          } else if (i > idxThis && ev.clientY > mid) {
            parent.insertBefore(group, target.nextElementSibling);
            activeDrag.hasMoved = true;
            updateButtonStates();
          }
          break;
        }
      }
    });

    const finishPointerDrag = (ev) => {
      if (!activeDrag) return;
      if (ev && ev.pointerId !== activeDrag.pointerId) return;

      const { handle, group, initialOrder, hasMoved } = activeDrag;
      try { handle.releasePointerCapture(activeDrag.pointerId); } catch {}

      group.classList.remove("dragging");
      document.body.classList.remove("is-dragging");

      const newOrder = orderFromDOM().join("|");
      if (hasMoved && newOrder !== initialOrder) {
        flashGroup(group);
        scheduleSaveOrder();
      }
      updateButtonStates();
      activeDrag = null;
    };

    document.body.addEventListener("pointerup", finishPointerDrag);
    document.body.addEventListener("pointercancel", finishPointerDrag);

    // Sync button states whenever providers panel is swapped by htmx
    document.body.addEventListener("htmx:afterSwap", (ev) => {
      const target = ev.detail.target;
      if (target && (target.id === "providers-content" || target.querySelector?.(".providers-group"))) {
        updateButtonStates();
      }
    });

    // Run once on load
    updateButtonStates();
  })();

  ensureScrollContainer();
  syncTheme();
  syncFilter();
  scheduleToastDismiss();
  updateModalScrollLock();
})();
