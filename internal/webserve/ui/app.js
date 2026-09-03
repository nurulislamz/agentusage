(() => {
  "use strict";

  const LAYOUTS = [
    { id: "split", label: "Split", hint: "Navigator + usage pane" },
    { id: "roster", label: "Roster", hint: "All apps with usage and resets" },
    { id: "matrix", label: "Matrix", hint: "Dense usage table" },
  ];

  function normalizeLayout(raw) {
    const id = String(raw || "").toLowerCase();
    return LAYOUTS.some((l) => l.id === id) ? id : "split";
  }

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    filter: "",
    token: sessionStorage.getItem("au-serve-token") || "",
    themeOverride: localStorage.getItem("au-serve-theme-override") || "",
    layout: normalizeLayout(localStorage.getItem("au-serve-layout") || "split"),
    loading: true,
    refreshing: false,
    refreshOpts: null,
    refreshText: "Fetching...",
    animFrame: 0,
    error: null,
    filterOpen: false,
  };

  function layoutMeta(id) {
    return LAYOUTS.find((l) => l.id === (id || state.layout)) || LAYOUTS[0];
  }

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
      <span class="header-meta">⊞ ${n} providers${filteredNote} · ${esc(layoutMeta().label)}</span>
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
      const inGroup = v.provider_id === selected.provider_id && counts[v.provider_id] > 1;
      const refreshing = isViewRefreshing(v);
      const next = v.next_reset || stripResetPrefix(v.reset_hint || "");
      html += `<button type="button" class="item${sel ? " selected" : ""}${inGroup ? " in-group" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}"${sel ? ` aria-current="true"` : ""} style="--p:${esc(v.accent_color || "var(--accent)")}">
        <span class="rail"></span>
        <span class="name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
        <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
        ${renderUsageMeters(v)}
        ${next ? `<span class="next${(v.resets || []).some((r) => r.urgent) ? " urgent" : ""}" title="Next reset">${esc(next)}</span>` : `<span class="next dim">—</span>`}
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

  function stripResetPrefix(s) {
    return String(s || "").replace(/^(resets?\s+in\s+|in\s+)/i, "").trim();
  }

  function usageLines(v) {
    if (!v) return [];
    if (v.usage_lines && v.usage_lines.length) return v.usage_lines;
    const lines = [];
    if (v.has_gauge) {
      lines.push({
        label: v.summary || "Usage",
        short: "Usage",
        percent: v.gauge_percent,
        reset_in: stripResetPrefix(v.reset_hint || v.next_reset || ""),
        tone: toneFromPercent(v.gauge_percent || 0, v),
      });
    } else if (v.summary) {
      lines.push({ label: "Status", value: v.summary, tone: "dim" });
    }
    (v.resets || []).forEach((r) => {
      if (!r || !r.duration) return;
      if (lines.some((l) => stripResetPrefix(l.reset_in) === r.duration)) return;
      lines.push({
        label: r.label,
        short: r.label,
        reset_in: r.duration,
        urgent: r.urgent,
        tone: r.urgent ? "crit" : "ok",
      });
    });
    return lines;
  }

  function isUsageCard(card) {
    const id = (card.id || "").toLowerCase();
    const title = (card.title || "").toLowerCase();
    return id === "usage" || title === "usage" || ["hero", "overview", "quota"].includes(id);
  }

  function usageOnlyCards(v) {
    const cards = v.detail_cards || [];
    let usage = cards.filter(isUsageCard);
    if (!usage.length) {
      const gauged = cards.find((c) => (c.rows || []).some((r) => r.kind === "gauge"));
      if (gauged) usage = [gauged];
    }
    if (!usage.length) {
      const lines = usageLines(v);
      if (!lines.length) return [];
      usage = [{
        id: "usage",
        title: "Usage",
        icon: "⚡",
        rows: lines.map((line) => usageLineToRow(line)),
      }];
    }
    return usage.map((card) => attachResetsToUsageCard(card, v));
  }

  function usageLineToRow(line) {
    if (line.percent != null) {
      return {
        kind: "gauge",
        label: line.label,
        percent: line.percent,
        hint: line.hint || (line.reset_in ? "Resets in " + line.reset_in : ""),
        tone: line.tone || "ok",
      };
    }
    if (line.reset_in) {
      return {
        kind: "timer",
        label: line.label,
        value: line.value || "",
        hint: "in " + line.reset_in,
        tone: line.tone || "ok",
      };
    }
    return { kind: "kv", label: line.label, value: line.value || "" };
  }

  function attachResetsToUsageCard(card, v) {
    const lines = usageLines(v);
    const rows = (card.rows || []).map((row) => {
      if (row.kind !== "gauge") return row;
      if (/reset/i.test(row.hint || "")) return row;
      const match = lines.find((l) => l.label === row.label || l.short === row.label);
      if (match && match.reset_in) {
        const extra = "Resets in " + match.reset_in;
        return { ...row, hint: row.hint ? row.hint + " · " + extra : extra };
      }
      return row;
    });
    return { ...card, rows };
  }

  function renderUsageMeters(v) {
    const lines = usageLines(v);
    if (!lines.length) {
      const summary = v.summary || v.message || "";
      const reset = v.reset_hint || v.next_reset || "";
      return `<span class="meters"><span class="meter dim"><span class="meter-val">${esc(summary)}</span>${reset ? `<span class="meter-reset">${esc(stripResetPrefix(reset))}</span>` : ""}</span></span>`;
    }
    const meters = lines.map((line) => {
      const pct = line.percent == null ? null : Math.max(0, Math.min(100, Number(line.percent)));
      const tone = line.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
      const val = pct != null ? `${pct.toFixed(0)}%` : esc(line.value || "");
      const reset = line.reset_in
        ? `<span class="meter-reset${line.urgent ? " urgent" : ""}">${esc(line.reset_in)}</span>`
        : "";
      const bar = pct != null
        ? `<span class="mini"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></span>`
        : "";
      const layout = pct != null ? "" : " text";
      return `<span class="meter${layout} ${toneClass(tone)}">
        <span class="meter-lab">${esc(line.short || line.label || "")}</span>
        ${bar}
        <span class="meter-val">${val}</span>
        ${reset}
      </span>`;
    }).join("");
    return `<span class="meters">${meters}</span>`;
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
    const cards = renderCards(usageOnlyCards(v));
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

  function renderMatrix() {
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
    const rows = views.map((v, i) => {
      const sel = i === state.selected;
      const lines = usageLines(v);
      const usage = lines.map((line) => {
        const pct = line.percent == null ? null : Math.max(0, Math.min(100, Number(line.percent)));
        const tone = line.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
        const val = pct != null ? `${pct.toFixed(0)}%` : esc(line.value || "—");
        const bar = pct != null
          ? `<span class="mini"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></span>`
          : "";
        return `<span class="matrix-metric ${toneClass(tone)}">
          <span class="meter-lab">${esc(line.short || line.label || "")}</span>
          ${bar}
          <span class="meter-val">${val}</span>
          ${line.reset_in ? `<span class="meter-reset${line.urgent ? " urgent" : ""}">${esc(line.reset_in)}</span>` : ""}
        </span>`;
      }).join("");
      const next = v.next_reset || stripResetPrefix(v.reset_hint || "") || "—";
      return `<button type="button" class="matrix-row${sel ? " selected" : ""}" data-idx="${i}"${sel ? ` aria-current="true"` : ""} style="--p:${esc(v.accent_color || "var(--accent)")}">
        <span class="matrix-app">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
        <span class="matrix-provider">${esc(v.provider_id || "")}</span>
        <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
        <span class="matrix-usage">${usage || `<span class="dim">${esc(v.summary || "")}</span>`}</span>
        <span class="matrix-reset">${esc(next)}</span>
      </button>`;
    }).join("");
    panel.innerHTML = `
      <span id="fetching-detail" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
      <div class="matrix" role="table" aria-label="Usage matrix">
        <div class="matrix-head" role="row">
          <span>App</span>
          <span>Provider</span>
          <span>Status</span>
          <span>Usage</span>
          <span>Next reset</span>
        </div>
        ${rows}
      </div>
    `;
    panel.querySelectorAll(".matrix-row").forEach((el) => {
      el.addEventListener("click", () => {
        state.selected = Number(el.dataset.idx);
        render();
      });
    });
    const selectedEl = panel.querySelector(".matrix-row.selected");
    if (selectedEl) selectedEl.scrollIntoView({ block: "nearest" });
  }

  function applyLayout() {
    const shell = $("app");
    if (!shell) return;
    const id = normalizeLayout(state.layout);
    state.layout = id;
    shell.dataset.layout = id;
    LAYOUTS.forEach((l) => shell.classList.toggle("layout-" + l.id, l.id === id));
    const nav = $("nav");
    const panel = document.querySelector("main.panel");
    if (nav) {
      nav.setAttribute("aria-label", id === "roster" ? "Usage board" : "Providers");
    }
    if (panel) {
      panel.setAttribute("aria-label", id === "matrix" ? "Usage matrix" : "Detail");
    }
  }

  function cycleLayout() {
    const ids = LAYOUTS.map((l) => l.id);
    const idx = Math.max(0, ids.indexOf(state.layout));
    state.layout = ids[(idx + 1) % ids.length];
    localStorage.setItem("au-serve-layout", state.layout);
    render();
    showToast("Layout: " + layoutMeta().label);
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
      <button type="button" class="footer-btn" id="footer-btn-layout" title="Cycle dashboard layout (v)"><kbd>v</kbd> <span>${esc(layoutMeta().label)}</span></button>
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
    $("footer-btn-layout")?.addEventListener("click", () => {
      cycleLayout();
    });
  }

  function render() {
    applyLayout();
    renderHeader();
    renderNav();
    if (state.layout === "matrix") {
      renderMatrix();
    } else {
      renderDetail();
    }
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
    if (["ArrowUp", "ArrowDown", "j", "J", "k", "K", "/", "r", "R", "u", "U", "t", "T", "v", "V"].includes(key)) {
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
      case "v":
      case "V":
        ev.preventDefault();
        cycleLayout();
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
