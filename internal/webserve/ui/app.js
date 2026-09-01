(() => {
  "use strict";

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    filter: "",
    token: sessionStorage.getItem("au-serve-token") || "",
    themeOverride: localStorage.getItem("au-serve-theme-override") || "",
    loading: true,
    refreshing: false,
    refreshOpts: null,
    refreshText: "Fetching...",
    animFrame: 0,
    error: null,
    filterOpen: false,
  };

  const $ = (id) => document.getElementById(id);

  function applyThemeTokens(tokens, override) {
    const root = document.documentElement;
    if (override === "light") {
      root.style.cssText = [
        "--bg:#f4f1ea", "--base:#f4f1ea", "--mantle:#fffdf8", "--surface-warm:#fffdf8",
        "--surface:#ebe6dc", "--surface0:#ebe6dc", "--surface1:#ddd6c8", "--surface2:#cfc7b8",
        "--fg:#2c2823", "--text:#2c2823", "--fg-2:#6d6760", "--subtext:#6d6760",
        "--muted:#9c958d", "--dim:#9c958d", "--accent:#7c5cbf", "--accent-on:#fffdf8",
        "--lavender:#6b5cae", "--teal:#2f8f86", "--sapphire:#2f8f86",
        "--success:#2f8f86", "--warn:#b8860b", "--danger:#c65746", "--crit:#c65746",
        "--peach:#c65746", "--border:#ddd6c8", "--border-soft:#cfc7b8",
      ].join(";");
      return;
    }
    if (!tokens || !tokens.base) {
      root.style.cssText = "";
      return;
    }
    const map = {
      "--bg": tokens.base,
      "--base": tokens.base,
      "--mantle": tokens.mantle,
      "--surface-warm": tokens.mantle,
      "--surface": tokens.surface0,
      "--surface0": tokens.surface0,
      "--surface1": tokens.surface1,
      "--surface2": tokens.surface2,
      "--fg": tokens.text,
      "--text": tokens.text,
      "--fg-2": tokens.subtext,
      "--subtext": tokens.subtext,
      "--muted": tokens.dim,
      "--dim": tokens.dim,
      "--accent": tokens.accent,
      "--accent-on": tokens.mantle || "#080a11",
      "--lavender": tokens.lavender,
      "--teal": tokens.teal,
      "--sapphire": tokens.sapphire,
      "--success": tokens.green,
      "--warn": tokens.yellow,
      "--danger": tokens.red,
      "--crit": tokens.red,
      "--peach": tokens.peach,
      "--border": tokens.surface1,
      "--border-soft": tokens.surface2,
    };
    root.style.cssText = Object.entries(map)
      .filter(([, v]) => v)
      .map(([k, v]) => `${k}:${v}`)
      .join(";");
  }

  function cycleThemeOverride() {
    const order = ["", "light"];
    const idx = order.indexOf(state.themeOverride);
    state.themeOverride = order[(idx + 1) % order.length];
    localStorage.setItem("au-serve-theme-override", state.themeOverride);
    applyThemeTokens(state.envelope?.theme_tokens, state.themeOverride);
  }

  function filteredViews() {
    const q = state.filter.trim().toLowerCase();
    if (!q) return state.views;
    return state.views.filter((v) =>
      v.provider_id.toLowerCase().includes(q)
      || v.provider_name.toLowerCase().includes(q)
      || v.account_id.toLowerCase().includes(q)
      || (v.summary || "").toLowerCase().includes(q),
    );
  }

  function headers() {
    const h = {};
    if (state.token) h.Authorization = `Bearer ${state.token}`;
    return h;
  }

  function esc(s) {
    return String(s ?? "").replace(/[&<>"']/g, (c) => (
      { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]
    ));
  }

  function pillClass(badge) {
    const b = (badge || "").toUpperCase();
    if (b.includes("LIMIT") || b.includes("ERROR") || b.includes("CRIT") || b.includes("DANGER")) return "crit";
    if (b.includes("LOW") || b.includes("NEAR") || b.includes("WARN")) return "warn";
    if (b.includes("PEACH")) return "peach";
    if (b.includes("AUTH")) return "auth";
    if (b.includes("DIM")) return "dim";
    if (b === "OK") return "ok";
    return "ok";
  }

  function toneClass(tone) {
    if (tone === "crit" || tone === "warn" || tone === "peach" || tone === "ok" || tone === "dim" || tone === "auth") {
      return "tone-" + tone;
    }
    return "tone-ok";
  }

  function gaugeColor(tone) {
    if (tone === "crit") return "var(--danger)";
    if (tone === "warn") return "var(--warn)";
    if (tone === "peach") return "var(--peach)";
    if (tone === "dim") return "var(--muted)";
    return "var(--success)";
  }

  const SPINNER = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
  let spinnerTimer = 0;
  let loadInFlight = false;

  function isViewRefreshing(v) {
    if (!state.refreshing) return false;
    if (!state.refreshOpts || state.refreshOpts.all) return true;
    if (state.refreshOpts.accountID) return state.refreshOpts.accountID === v.account_id;
    return true;
  }

  function setRefreshing(on, opts) {
    state.refreshing = !!on;
    state.refreshOpts = on ? (opts || null) : null;
    const accountID = opts?.accountID;
    const all = opts?.all;
    let text = "Fetching...";
    if (accountID) {
      text = "Fetching (" + accountID + ")...";
    } else if (all) {
      text = "Fetching all...";
    }
    state.refreshText = text;

    const appShell = $("app");
    if (appShell) appShell.classList.toggle("refreshing", state.refreshing);
    ["fetching-header", "fetching-detail", "fetching-footer"].forEach((id) => {
      const el = $(id);
      if (el) el.hidden = !state.refreshing;
    });
    document.querySelectorAll(".fetching-text").forEach((el) => {
      el.textContent = text;
    });

    const views = filteredViews();
    $("nav")?.querySelectorAll(".item").forEach((el) => {
      const idx = Number(el.dataset.idx);
      const v = views[idx];
      const refreshing = v && isViewRefreshing(v);
      el.classList.toggle("refreshing", !!refreshing);
    });

    if (state.refreshing) {
      state.animFrame = 0;
      const frame = SPINNER[0];
      document.querySelectorAll(".spin").forEach((el) => {
        el.textContent = frame;
      });
      startSpinner();
    } else {
      stopSpinner();
    }
  }

  function applyEnvelope(env) {
    state.envelope = env;
    state.views = env.views || [];
    state.error = null;
    applyThemeTokens(state.envelope.theme_tokens, state.themeOverride);
    const visible = filteredViews();
    if (state.selected >= visible.length) {
      state.selected = Math.max(0, visible.length - 1);
    }
    showDashboard(state.views.length > 0);
    if ($("status-bar")) $("status-bar").hidden = true;
  }

  function usageMode() {
    return (state.envelope?.usage_mode || "remaining").toLowerCase() === "used" ? "used" : "remaining";
  }

  function usageModeLabel() {
    return usageMode() === "used" ? "Used" : "Remaining";
  }

  function startSpinner() {
    stopSpinner();
    spinnerTimer = setInterval(() => {
      if (!state.refreshing) return;
      state.animFrame = (state.animFrame + 1) % SPINNER.length;
      document.querySelectorAll(".spin").forEach((el) => {
        el.textContent = SPINNER[state.animFrame];
      });
    }, 150);
  }

  function stopSpinner() {
    clearInterval(spinnerTimer);
    spinnerTimer = 0;
  }

  async function load(opts) {
    const manual = !!(opts && opts.manual);
    if (loadInFlight) return;
    loadInFlight = true;
    const showFetching = manual && state.views.length > 0;
    if (showFetching) {
      setRefreshing(true, { accountID: opts?.accountID, all: opts ? (opts.all || !opts.accountID) : true });
    }
    try {
      let qs = manual ? "?refresh=1" : "";
      if (opts && opts.accountID) {
        qs += (qs ? "&" : "?") + `account_id=${encodeURIComponent(opts.accountID)}`;
      }
      const fetchPromise = fetch("api/v1/snapshots" + qs, { headers: headers() });
      const minDurationPromise = manual ? new Promise((r) => setTimeout(r, 350)) : Promise.resolve();
      const [res] = await Promise.all([fetchPromise, minDurationPromise]);
      if (res.status === 401) {
        $("token-modal").hidden = false;
        $("token-error").hidden = false;
        $("token-error").textContent = state.token ? "Invalid token" : "Token required";
        return;
      }
      if (!res.ok) throw new Error(`snapshots ${res.status}`);
      $("token-modal").hidden = true;
      applyEnvelope(await res.json());
    } catch (err) {
      state.error = String(err);
      if ($("status-bar")) {
        $("status-bar").hidden = false;
        $("status-bar").textContent = "offline - reconnecting…";
      }
      if (state.views.length === 0) {
        showDashboard(false);
        $("empty-state").hidden = false;
        $("splash").hidden = true;
        $("app").hidden = true;
      }
    } finally {
      loadInFlight = false;
      state.loading = false;
      if (state.views.length > 0 && $("token-modal").hidden) {
        render();
      }
      setRefreshing(false);
    }
  }

  let toastTimer = 0;
  function showToast(message, isInfo = true) {
    const sb = $("status-bar");
    if (!sb) return;
    clearTimeout(toastTimer);
    sb.textContent = message;
    sb.classList.toggle("info", !!isInfo);
    sb.hidden = false;
    sb.style.opacity = "1";
    toastTimer = setTimeout(() => {
      sb.style.opacity = "0";
      setTimeout(() => {
        if (sb.textContent === message) {
          sb.hidden = true;
          sb.style.opacity = "1";
        }
      }, 250);
    }, 2500);
  }

  async function cycleUsageMode() {
    const next = usageMode() === "used" ? "remaining" : "used";
    try {
      const res = await fetch("api/v1/usage-mode", {
        method: "POST",
        headers: { ...headers(), "Content-Type": "application/json" },
        body: JSON.stringify({ usage_mode: next }),
      });
      if (res.status === 401) {
        $("token-modal").hidden = false;
        return;
      }
      if (!res.ok) throw new Error(`usage-mode ${res.status}`);
      applyEnvelope(await res.json());
      render();
      showToast(`Usage mode: ${usageModeLabel()}`);
    } catch (err) {
      console.error(err);
      showToast(`Failed to update mode: ${err.message || err}`, false);
    }
  }

  function showDashboard(hasData) {
    $("splash").hidden = true;
    $("empty-state").hidden = hasData;
    $("app").hidden = !hasData;
  }

  function renderHeader() {
    const views = filteredViews();
    const n = views.length;
    const filteredNote = state.filter ? " (filtered)" : "";
    const switcher = views.length
      ? `<select id="switcher" class="switcher" aria-label="Account">${views.map((v, i) =>
          `<option value="${i}"${i === state.selected ? " selected" : ""}>${esc(v.account_id)}</option>`
        ).join("")}</select>`
      : "";
    const fetchVisible = state.refreshing ? "" : " hidden";
    const fetchText = esc(state.refreshText || "Fetching...");
    const spinChar = SPINNER[state.animFrame] || "⠋";
    $("header").innerHTML = `
      <span class="bolt">⚡</span>
      <span class="brand">agentUsage</span>
      <span id="fetching-header" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
      ${switcher}
      <span class="spacer"></span>
      <span class="header-meta">⊞ ${n} providers${filteredNote}</span>
    `;
    const sel = $("switcher");
    if (sel) {
      sel.addEventListener("change", () => {
        state.selected = Number(sel.value);
        render();
      });
    }
  }

  function renderNav() {
    const views = filteredViews();
    if (!views.length) {
      $("nav").innerHTML = `<div class="group">${state.filter ? "No matches." : "Loading providers…"}</div>`;
      return;
    }
    const selected = views[state.selected] || {};
    const counts = {};
    views.forEach((v) => { counts[v.provider_id] = (counts[v.provider_id] || 0) + 1; });
    let html = "";
    let last = "";
    views.forEach((v, i) => {
      if (v.provider_id !== last) {
        const active = v.provider_id === selected.provider_id;
        html += `<div class="group${active ? " active" : ""}" style="--p:${esc(v.accent_color || "var(--accent)")}">${esc((v.provider_id || "").toUpperCase())} (${counts[v.provider_id]})</div>`;
        last = v.provider_id;
      }
      const sel = i === state.selected;
      const pct = v.has_gauge ? Math.max(0, Math.min(100, v.gauge_percent || 0)) : null;
      const reset = v.reset_hint || "";
      const summary = v.has_gauge ? (v.summary || `${pct.toFixed(2)}%`) : (v.summary || v.message || "");
      const inGroup = v.provider_id === selected.provider_id && counts[v.provider_id] > 1;
      const refreshing = isViewRefreshing(v);
      html += `<button type="button" class="item${sel ? " selected" : ""}${inGroup ? " in-group" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}"${sel ? ` aria-current="true"` : ""} style="--p:${esc(v.accent_color || "var(--accent)")}">
        <span class="rail"></span>
        <span class="name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
        <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
        <span class="sum">${pct !== null ? `<span class="mini"><i style="width:${pct}%;background:${gaugeColor(toneFromPercent(pct, v))}"></i></span>` : ""}<span class="sum-text">${esc(summary)}</span>${reset ? `<span class="reset">${esc(reset)}</span>` : ""}</span>
      </button>`;
    });
    $("nav").innerHTML = html;
    $("nav").querySelectorAll(".item").forEach((el) => {
      el.addEventListener("click", () => {
        state.selected = Number(el.dataset.idx);
        render();
      });
    });
    const selectedEl = $("nav").querySelector(".item.selected");
    if (selectedEl) selectedEl.scrollIntoView({ block: "nearest" });
  }

  function toneFromPercent(pct, view) {
    const used = (state.envelope?.usage_mode || "").toLowerCase() === "used";
    if (used) {
      if (pct >= 90) return "crit";
      if (pct >= 75) return "warn";
      if (pct >= 50) return "peach";
      return "ok";
    }
    if (pct <= 10) return "crit";
    if (pct <= 25) return "warn";
    if (pct <= 50) return "peach";
    return "ok";
  }

  function renderCards(cards) {
    if (!cards || !cards.length) return "";
    const isUsed = (state.envelope?.usage_mode || "remaining").toLowerCase() === "used";
    const modeWord = isUsed ? "used" : "remaining";
    const threshPct = isUsed ? 80 : 20;

    const cardsHtml = cards.map((card, idx) => {
      const color = card.color || "var(--fg-2)";
      const cardId = (card.id || "").toLowerCase();
      const cardTitle = (card.title || "").toLowerCase();
      const isFeatured = idx === 0 || ["usage", "hero", "overview", "quota"].includes(cardId) || cardTitle === "usage";
      const heroClass = isFeatured ? " card-hero" : "";
      const rows = (card.rows || []).map((row) => {
        if (row.kind === "heading") {
          return `<div class="heading">${esc(row.value || row.label || "")}</div>`;
        }
        if (row.kind === "gauge") {
          const pct = row.percent == null ? 0 : Number(row.percent);
          const tone = row.tone || "ok";
          return `<div class="gauge-block ${toneClass(tone)}">
            <div class="gauge-header">
              <span class="gauge-label">${esc(row.label || "")}</span>
              <span class="gauge-stat">
                <b class="stat-val">${pct.toFixed(2)}%</b>
                <span class="pill ${pillClass(tone)}">${esc(modeWord)}</span>
              </span>
            </div>
            <div class="gauge" style="color:${gaugeColor(tone)}">
              <i style="width:${Math.max(0, Math.min(100, pct))}%;background:currentColor"></i>
              <span class="gauge-threshold" style="left:${threshPct}%" aria-hidden="true"></span>
            </div>
            <div class="caption">${row.hint ? esc(row.hint) : `${pct.toFixed(2)}% ${esc(modeWord)}`}</div>
          </div>`;
        }
        if (row.kind === "timer") {
          return `<div class="timer ${toneClass(row.tone || "ok")}">
            <span class="dot"></span>
            <span>${esc(row.label || "")}</span>
            <span class="when">${esc(row.value || "")}</span>
            <span class="hint">${esc(row.hint || "")}</span>
          </div>`;
        }
        if (row.kind === "kv") {
          return `<div class="kv"><span class="dim">${esc(row.label || "")}</span><span class="kv-val">${esc(row.value || "")}</span></div>`;
        }
        return `<div class="text-row">${esc(row.value || row.label || "")}</div>`;
      }).join("");
      return `<section class="card${heroClass}" style="--card:${esc(color)}"><div class="card-header"><h2>${esc(card.icon || "")} ${esc((card.title || "").toUpperCase())}</h2></div>${rows}</section>`;
    }).join("");

    return `<div class="panel-cards-grid">${cardsHtml}</div>`;
  }

  function formatAge(ms) {
    if (ms < 0) ms = 0;
    const s = Math.floor(ms / 1000);
    if (s < 60) return s + "s";
    const m = Math.floor(s / 60);
    if (s < 3600) return m + "m" + (s % 60) + "s";
    const h = Math.floor(s / 3600);
    if (s < 86400) return h + "h" + (m % 60) + "m";
    return Math.floor(h / 24) + "d" + (h % 24) + "h";
  }

  function parseTimestamp(iso) {
    if (!iso) return NaN;
    const ts = Date.parse(iso);
    if (!Number.isFinite(ts) || ts < Date.parse("2000-01-01T00:00:00Z")) return NaN;
    return ts;
  }

  function formatLastRefreshed(iso, fallback) {
    const ts = parseTimestamp(iso);
    if (!Number.isFinite(ts)) return fallback || "";
    const age = Date.now() - ts;
    if (age < 5000) return "Last refreshed just now";
    return "Last refreshed " + formatAge(age) + " ago";
  }

  function lastRefreshedText(view) {
    if (!view) return "";
    return formatLastRefreshed(view.timestamp, view.last_refreshed || "");
  }

  function renderDetail() {
    const views = filteredViews();
    const panel = $("panel");
    const fetchVisible = state.refreshing ? "" : " hidden";
    const fetchText = esc(state.refreshText || "Fetching...");
    const spinChar = SPINNER[state.animFrame] || "⠋";
    if (!views.length) {
      panel.innerHTML = `
        <span id="fetching-detail" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
        <p class="dim">${state.filter ? "No matches." : "No providers."}</p>
      `;
      return;
    }
    const v = views[state.selected];
    const meta = [v.provider_id, v.detail].filter(Boolean).join(" · ");
    const schedule = v.cycle_schedule || "";
    const summary = v.summary || "";
    const refreshed = lastRefreshedText(v);
    const cards = v.detail_cards && v.detail_cards.length
      ? renderCards(v.detail_cards)
      : (v.detail_html ? `<div class="panel-cards-grid"><section class="card">${v.detail_html}</section></div>` : "");
    panel.innerHTML = `
      <div class="hero">
        <h1>
          ${esc(v.status_icon || "●")} ${esc(v.account_id)}
          <span id="fetching-detail" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
        </h1>
        <div class="hero-right">${esc(meta)} <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || "")}</span></div>
      </div>
      <div class="subhero">
        <div>${summary ? `<strong>${esc(summary)}</strong>` : ""}${schedule ? ` · ${esc(schedule)}` : ""}</div>
        ${refreshed ? `<div id="last-refreshed" class="last-refreshed">${esc(refreshed)}</div>` : ""}
      </div>
      <div class="accent-line ${esc(v.header_tone || "ok")}"></div>
      ${cards}
    `;
  }

  function renderFooter() {
    const sec = Math.max(5, state.envelope?.refresh_interval_seconds || 30);
    const theme = state.envelope?.theme_tokens?.name || state.envelope?.theme || "";
    const fetchVisible = state.refreshing ? "" : " hidden";
    const fetchText = esc(state.refreshText || "Fetching...");
    const spinChar = SPINNER[state.animFrame] || "⠋";
    $("footer").innerHTML = `
      <span id="fetching-footer" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
      <span>auto-refresh ⟳ ${sec}s</span>
      <span><kbd>j</kbd>/<kbd>k</kbd> move</span>
      <button type="button" class="footer-btn" id="footer-btn-filter" title="Filter providers (/)"><kbd>/</kbd> filter</button>
      <button type="button" class="footer-btn" id="footer-btn-mode" title="Toggle usage mode (u)"><kbd>u</kbd> <span>${esc(usageModeLabel())}</span></button>
      <button type="button" class="footer-btn" id="footer-btn-refresh" title="Refresh focused account (r) / all (R)"><kbd>r</kbd> refresh</button>
      <button type="button" class="footer-btn" id="footer-btn-theme" title="Cycle theme (t)"><kbd>t</kbd> theme</button>
      <span class="grow"></span>
      <span>${esc(theme)}</span>
    `;

    $("footer-btn-filter")?.addEventListener("click", () => {
      openFilter();
    });
    $("footer-btn-mode")?.addEventListener("click", () => {
      cycleUsageMode().catch(console.error);
    });
    $("footer-btn-refresh")?.addEventListener("click", () => {
      load({ manual: true, accountID: filteredViews()[state.selected]?.account_id }).catch(console.error);
    });
    $("footer-btn-theme")?.addEventListener("click", () => {
      cycleThemeOverride();
    });
  }

  function render() {
    renderHeader();
    renderNav();
    renderDetail();
    renderFooter();
  }

  function moveSelection(delta) {
    const views = filteredViews();
    if (!views.length) return;
    state.selected = Math.max(0, Math.min(views.length - 1, state.selected + delta));
    render();
  }

  function openFilter() {
    state.filterOpen = true;
    const bar = $("filter-bar");
    const input = $("filter-input");
    bar.hidden = false;
    input.value = state.filter;
    input.focus();
    input.select();
  }

  function closeFilter(apply) {
    const bar = $("filter-bar");
    const input = $("filter-input");
    if (apply) {
      state.filter = input.value;
      state.selected = 0;
    }
    state.filterOpen = false;
    bar.hidden = true;
    render();
  }

  $("token-form").addEventListener("submit", (ev) => {
    ev.preventDefault();
    state.token = $("token-input").value.trim();
    sessionStorage.setItem("au-serve-token", state.token);
    load().catch(console.error);
  });

  $("filter-input").addEventListener("keydown", (ev) => {
    if (ev.key === "Enter") {
      ev.preventDefault();
      closeFilter(true);
    } else if (ev.key === "Escape") {
      ev.preventDefault();
      closeFilter(false);
    }
  });

  document.addEventListener("keydown", (ev) => {
    if ($("token-modal").hidden === false) return;
    if (state.filterOpen) return;
    if (ev.ctrlKey || ev.metaKey || ev.altKey) return;
    // Don't intercept typing in text input, textarea, or select
    if (ev.target && typeof ev.target.matches === "function" && (ev.target.matches("input, textarea, select") || ev.target.isContentEditable)) return;

    const key = ev.key;
    if (["ArrowUp", "ArrowDown", "j", "J", "k", "K", "/", "r", "R", "u", "U", "t", "T"].includes(key)) {
      if (ev.target && typeof ev.target.blur === "function" && ev.target !== document.body) {
        ev.target.blur();
      }
    }

    switch (key) {
      case "ArrowUp":
      case "k":
      case "K":
        ev.preventDefault();
        moveSelection(-1);
        break;
      case "ArrowDown":
      case "j":
      case "J":
        ev.preventDefault();
        moveSelection(1);
        break;
      case "/":
        ev.preventDefault();
        openFilter();
        break;
      case "r":
        ev.preventDefault();
        load({ manual: true, accountID: filteredViews()[state.selected]?.account_id }).catch(console.error);
        break;
      case "R":
        ev.preventDefault();
        load({ manual: true, all: true }).catch(console.error);
        break;
      case "u":
      case "U":
        ev.preventDefault();
        cycleUsageMode().catch(console.error);
        break;
      case "t":
      case "T":
        ev.preventDefault();
        cycleThemeOverride();
        break;
      default:
        break;
    }
  });

  let refreshTimer = 0;
  const scheduleRefresh = () => {
    clearTimeout(refreshTimer);
    const seconds = Math.max(5, state.envelope?.refresh_interval_seconds || 30);
    refreshTimer = setTimeout(() => {
      if (!document.hidden) load().catch(() => {});
      scheduleRefresh();
    }, seconds * 1000);
  };

  load().then(scheduleRefresh).catch((err) => {
    console.error(err);
    $("splash").hidden = true;
    $("empty-state").hidden = false;
  });

  setInterval(() => {
    const el = $("last-refreshed");
    if (!el) return;
    const views = filteredViews();
    const text = lastRefreshedText(views[state.selected]);
    if (text && el.textContent !== text) el.textContent = text;
  }, 1000);
})();
