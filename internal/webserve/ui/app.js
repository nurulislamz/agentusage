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

  document.addEventListener("htmx:responseError", (ev) => {
    if (!ev.detail.xhr || ev.detail.xhr.status !== 401) return;
    const error = $("token-error");
    if (error) {
      error.textContent = sessionStorage.getItem(TOKEN_KEY) ? "Invalid token" : "Token required";
      error.hidden = false;
    }
    if ($("token-modal")) $("token-modal").hidden = false;
    if ($("token-input")) $("token-input").focus();
  });

  if ($("token-form")) {
    $("token-form").addEventListener("submit", (ev) => {
      ev.preventDefault();
      sessionStorage.setItem(TOKEN_KEY, $("token-input").value.trim());
      $("token-modal").hidden = true;
      htmx.ajax("GET", "partial/app", { target: "#app", swap: "outerHTML" });
    });
  }

  const closeInspect = () => { if ($("inspect-modal")) $("inspect-modal").hidden = true; };
  document.addEventListener("htmx:afterSwap", (ev) => {
    if (ev.detail.target && ev.detail.target.id === "inspect-content") {
      if ($("inspect-modal")) $("inspect-modal").hidden = false;
    }
    const selected = document.querySelector(".nav-item.selected");
    if (selected && selected.scrollIntoView) selected.scrollIntoView({ block: "nearest" });
  });
  if ($("inspect-close")) $("inspect-close").addEventListener("click", closeInspect);
  if ($("inspect-modal")) {
    $("inspect-modal").addEventListener("click", (ev) => { if (ev.target === $("inspect-modal")) closeInspect(); });
  }

  let filterBackup = "";
  const closeFilter = (apply) => {
    const bar = $("filter-bar"), input = $("filter-input");
    if (!bar || !input) return;
    if (!apply) { input.value = filterBackup; htmx.trigger(input, "search"); }
    bar.hidden = true;
    input.blur();
  };
  const openFilter = () => {
    const bar = $("filter-bar"), input = $("filter-input");
    if (!bar || !input) return;
    filterBackup = input.value;
    bar.hidden = false;
    input.focus();
    input.select();
  };
  // Footer buttons live inside the swapped #app fragment: delegate.
  document.addEventListener("click", (ev) => {
    const target = ev.target;
    if (!target || !target.closest) return;
    if (target.closest("#footer-btn-filter")) {
      openFilter();
      return;
    }
    if (ev.shiftKey && target.closest("#footer-btn-theme")) {
      ev.preventDefault();
      ev.stopPropagation();
      click("key-theme-back");
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
      if (ev.key === "Enter") { ev.preventDefault(); closeFilter(true); }
      else if (ev.key === "Escape") { ev.preventDefault(); closeFilter(false); }
    });
  }

  const selectedRow = (sel) => document.querySelector(sel);
  const rowAction = () => {
    const layout = ($("app") || {}).dataset ? $("app").dataset.layout : "";
    if (layout === "matrix") selectedRow(".matrix-row.selected")?.click();
    else if (layout === "bento") selectedRow(".bento-tile.selected")?.click();
  };

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
      case "ArrowDown": case "j": case "J": click("key-next"); break;
      case "/": ev.preventDefault(); openFilter(); break;
      case "r": click("key-refresh"); break;
      case "R": click("key-refresh-all"); break;
      case "u": case "U": click("key-mode"); break;
      case "t": click("key-theme-fwd"); break;
      case "T": click("key-theme-back"); break;
      case "v": case "V": click("key-layout-cycle"); break;
      case "Enter": case " ": rowAction(); break;
      case "Escape": closeInspect(); break;
    }
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
})();
