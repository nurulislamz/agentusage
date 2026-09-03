(() => {
  "use strict";

  const LAYOUTS = [
    { id: "bars", label: "Bars", hint: "Linear gauges · OpenUsage-style cards" },
    { id: "dials", label: "Dials", hint: "Radial gauges · at-a-glance remaining" },
    { id: "strips", label: "Strips", hint: "Grafana bar-gauge wall" },
  ];

  function normalizeLayout(raw) {
    const id = String(raw || "").toLowerCase();
    return LAYOUTS.some((l) => l.id === id) ? id : "bars";
  }

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    filter: "",
    token: sessionStorage.getItem("au-serve-token") || "",
    themeOverride: localStorage.getItem("au-serve-theme-override") || "",
    layout: normalizeLayout(localStorage.getItem("au-serve-layout") || "bars"),
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
    document.querySelectorAll(".agent").forEach((el) => {
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
    const fetchVisible = state.refreshing ? "" : " hidden";
    const fetchText = esc(state.refreshText || "Fetching...");
    const spinChar = SPINNER[state.animFrame] || "⠋";
    const meta = layoutMeta();
    $("header").innerHTML = `
      <span class="bolt">⚡</span>
      <span class="brand">agentUsage</span>
      <span id="fetching-header" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
      <span class="spacer"></span>
      <span class="header-meta">⊞ ${n} agents${filteredNote} · ${esc(meta.label)} · ${esc(meta.hint)}</span>
    `;
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

  function clampPct(n) {
    const pct = Number(n);
    if (!Number.isFinite(pct)) return null;
    return Math.max(0, Math.min(100, pct));
  }

  function lineTone(line, v) {
    if (line.percent != null) {
      return line.tone || toneFromPercent(Number(line.percent), v);
    }
    return line.tone || "ok";
  }

  function isDepleted(pct) {
    if (pct == null) return false;
    return usageMode() === "used" ? pct >= 99.5 : pct <= 0.5;
  }

  function percentCaption(pct) {
    if (isDepleted(pct)) return "Limit reached";
    const n = Math.round(pct);
    return usageMode() === "used" ? n + "% used" : n + "% left";
  }

  function resetCaption(line) {
    const reset = stripResetPrefix(line.reset_in || "");
    if (!reset) return "";
    if (/^expired$/i.test(reset)) return "Expired";
    return "Resets in " + reset;
  }

  function formatCompact(n) {
    const v = Number(n);
    if (!Number.isFinite(v)) return "—";
    const abs = Math.abs(v);
    const trim = (x) => x.toFixed(1).replace(/\.0$/, "");
    if (abs >= 1e9) return trim(v / 1e9) + "B";
    if (abs >= 1e6) return trim(v / 1e6) + "M";
    if (abs >= 1e3) return trim(v / 1e3) + "K";
    if (abs >= 100 || Number.isInteger(v)) return String(Math.round(v));
    return v.toFixed(2).replace(/0+$/, "").replace(/\.$/, "");
  }

  function trendLooksMoney(points) {
    const vals = (points || []).map((p) => Number(p.value)).filter(Number.isFinite);
    if (!vals.length) return false;
    const max = Math.max(...vals.map(Math.abs));
    return max < 5000 && vals.some((v) => !Number.isInteger(v));
  }

  function formatTrendValue(n, points) {
    return trendLooksMoney(points) ? "$" + formatCompact(n) : formatCompact(n);
  }

  function trendStats(points) {
    const pts = (points || []).filter((p) => Number.isFinite(Number(p.value)));
    if (!pts.length) return [];
    const last = pts[pts.length - 1];
    const prev = pts.length > 1 ? pts[pts.length - 2] : null;
    const sum = pts.reduce((a, p) => a + Number(p.value), 0);
    const out = [{ label: "Today", value: formatTrendValue(last.value, pts) }];
    if (prev) out.push({ label: "Yesterday", value: formatTrendValue(prev.value, pts) });
    if (pts.length > 2) out.push({ label: pts.length + "d", value: formatTrendValue(sum, pts) });
    return out;
  }

  function sparkBars(points, w, h) {
    const vals = (points || []).map((p) => Number(p.value)).filter((v) => Number.isFinite(v) && v >= 0);
    if (vals.length < 2) return "";
    const width = w || 108;
    const height = h || 26;
    const max = Math.max(...vals, 1e-9);
    const gap = 1.2;
    const bw = Math.max(2, (width / vals.length) - gap);
    const bars = vals.map((v, i) => {
      const bh = Math.max(1.5, (v / max) * height);
      const x = i * (width / vals.length);
      return `<rect x="${x.toFixed(2)}" y="${(height - bh).toFixed(2)}" width="${bw.toFixed(2)}" height="${bh.toFixed(2)}" rx="0.6"/>`;
    }).join("");
    return `<svg class="spark spark-bars" viewBox="0 0 ${width} ${height}" width="${width}" height="${height}" aria-hidden="true">${bars}</svg>`;
  }

  function sparkLine(points, w, h) {
    const vals = (points || []).map((p) => Number(p.value)).filter(Number.isFinite);
    if (vals.length < 2) return "";
    const width = w || 88;
    const height = h || 28;
    const min = Math.min(...vals);
    const max = Math.max(...vals);
    const span = Math.max(max - min, 1e-9);
    const coords = vals.map((v, i) => {
      const x = vals.length === 1 ? width / 2 : (i / (vals.length - 1)) * width;
      const y = height - ((v - min) / span) * (height - 2) - 1;
      return x.toFixed(1) + "," + y.toFixed(1);
    });
    return `<svg class="spark spark-line" viewBox="0 0 ${width} ${height}" width="${width}" height="${height}" aria-hidden="true"><polyline fill="none" points="${coords.join(" ")}"/></svg>`;
  }

  function agentSubtitle(v) {
    return v.tag_label || v.detail || v.provider_name || v.provider_id || "";
  }

  function agentShell(v, i, inner) {
    const sel = i === state.selected;
    const refreshing = isViewRefreshing(v);
    return `<button type="button" class="agent${sel ? " selected" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}"${sel ? ` aria-current="true"` : ""} style="--p:${esc(v.accent_color || "var(--accent)")};--tone:${gaugeColor(lineTone(usageLines(v)[0] || {}, v))}">${inner}</button>`;
  }

  function agentHead(v, i) {
    const sub = agentSubtitle(v);
    return `<div class="agent-head">
      <span class="agent-name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
      <span class="agent-plan">${esc(sub)}</span>
      <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
    </div>`;
  }

  function renderLinearGauge(line, v) {
    const pct = clampPct(line.percent);
    const tone = lineTone(line, v);
    if (pct == null) {
      return `<div class="lin-stat">
        <span class="lin-name">${esc(line.label || line.short || "")}</span>
        <span class="lin-value">${esc(line.value || "—")}</span>
      </div>`;
    }
    const depleted = isDepleted(pct);
    return `<div class="lin-gauge ${toneClass(tone)}${depleted ? " depleted" : ""}">
      <div class="lin-name">${esc(line.label || line.short || "")}${depleted ? ` <span class="limit-flag">Limit reached</span>` : ""}</div>
      <div class="lin-track"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></div>
      <div class="lin-meta">
        <span>${esc(percentCaption(pct))}</span>
        <span class="${line.urgent ? "urgent" : ""}">${esc(resetCaption(line))}</span>
      </div>
    </div>`;
  }

  function renderArcGauge(line, v) {
    const pct = clampPct(line.percent);
    const tone = lineTone(line, v);
    if (pct == null) {
      return `<div class="dial dial-stat ${toneClass(tone)}">
        <span class="dial-num">${esc(line.value || "—")}</span>
        <span class="dial-lab">${esc(line.short || line.label || "")}</span>
      </div>`;
    }
    const color = gaugeColor(tone);
    return `<div class="dial ${toneClass(tone)}${isDepleted(pct) ? " depleted" : ""}">
      <svg viewBox="0 0 88 58" class="dial-svg" aria-hidden="true">
        <path class="dial-track" d="M10 48 A34 34 0 0 1 78 48" pathLength="100"/>
        <path class="dial-fill" d="M10 48 A34 34 0 0 1 78 48" pathLength="100" stroke="${color}" stroke-dasharray="${pct} 100"/>
        <text x="44" y="44" text-anchor="middle">${Math.round(pct)}%</text>
      </svg>
      <span class="dial-lab">${esc(line.short || line.label || "")}</span>
      <span class="dial-reset${line.urgent ? " urgent" : ""}">${esc(resetCaption(line) || percentCaption(pct))}</span>
    </div>`;
  }

  function renderStripGauge(line, v) {
    const pct = clampPct(line.percent);
    const tone = lineTone(line, v);
    if (pct == null) {
      return `<div class="strip-metric text ${toneClass(tone)}">
        <span class="strip-lab">${esc(line.short || line.label || "")}</span>
        <span class="strip-val">${esc(line.value || "—")}</span>
      </div>`;
    }
    return `<div class="strip-metric ${toneClass(tone)}${isDepleted(pct) ? " depleted" : ""}">
      <span class="strip-lab">${esc(line.short || line.label || "")}</span>
      <span class="strip-track">
        <i style="width:${pct}%;background:${gaugeColor(tone)}"></i>
        <em>${esc(isDepleted(pct) ? "Limit" : Math.round(pct) + "%")}</em>
      </span>
      <span class="strip-reset${line.urgent ? " urgent" : ""}">${esc(stripResetPrefix(line.reset_in || "") || "—")}</span>
    </div>`;
  }

  function renderBarCard(v, i) {
    const lines = usageLines(v);
    const gauges = lines.map((line) => renderLinearGauge(line, v)).join("");
    const stats = trendStats(v.daily_cost).map((s) =>
      `<div class="hist-row"><span>${esc(s.label)}</span><span>${esc(s.value)}</span></div>`
    ).join("");
    const spark = sparkBars(v.daily_cost);
    const inner = `
      ${agentHead(v, i)}
      <div class="lin-list">${gauges || `<p class="dim">${esc(v.message || "No usage yet")}</p>`}</div>
      ${stats || spark ? `<div class="agent-foot">${stats ? `<div class="hist">${stats}</div>` : ""}${spark ? `<div class="spark-wrap" title="Usage trend">${spark}</div>` : ""}</div>` : ""}
    `;
    return agentShell(v, i, inner);
  }

  function renderDialCard(v, i) {
    const lines = usageLines(v);
    const dials = lines.map((line) => renderArcGauge(line, v)).join("");
    const spark = sparkLine(v.daily_cost, 120, 32);
    const next = v.next_reset || stripResetPrefix(v.reset_hint || "");
    const inner = `
      ${agentHead(v, i)}
      <div class="dial-row">${dials || `<p class="dim">${esc(v.message || "No usage yet")}</p>`}</div>
      <div class="agent-foot">
        ${next ? `<span class="next-reset">Next reset ${esc(next)}</span>` : `<span class="dim">No reset</span>`}
        ${spark ? `<div class="spark-wrap">${spark}</div>` : ""}
      </div>
    `;
    return agentShell(v, i, inner);
  }

  function renderStripCard(v, i) {
    const lines = usageLines(v);
    const metrics = lines.map((line) => renderStripGauge(line, v)).join("");
    const spark = sparkBars(v.daily_cost, 72, 22);
    const sub = agentSubtitle(v);
    const inner = `
      <div class="strip-id">
        <span class="agent-name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
        <span class="agent-plan">${esc(sub)}</span>
        <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
      </div>
      <div class="strip-gauges">${metrics || `<span class="dim">${esc(v.summary || v.message || "")}</span>`}</div>
      <div class="strip-trend">${spark || ""}</div>
    `;
    return agentShell(v, i, inner);
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

  function renderBoard() {
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
    const id = state.layout;
    const renderer = id === "dials" ? renderDialCard : id === "strips" ? renderStripCard : renderBarCard;
    const cards = views.map((v, i) => renderer(v, i)).join("");
    panel.innerHTML = `
      <span id="fetching-detail" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>
      <div class="board board-${esc(id)}" role="list" aria-label="${esc(layoutMeta().hint)}">${cards}</div>
    `;
    panel.querySelectorAll(".agent").forEach((el) => {
      el.addEventListener("click", () => {
        state.selected = Number(el.dataset.idx);
        render();
      });
    });
    const selectedEl = panel.querySelector(".agent.selected");
    if (selectedEl) selectedEl.scrollIntoView({ block: "nearest" });
  }

  function applyLayout() {
    const shell = $("app");
    if (!shell) return;
    const id = normalizeLayout(state.layout);
    state.layout = id;
    shell.dataset.layout = id;
    LAYOUTS.forEach((l) => shell.classList.toggle("layout-" + l.id, l.id === id));
    const panel = document.querySelector("main.panel");
    if (panel) panel.setAttribute("aria-label", layoutMeta().hint);
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
    renderBoard();
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
    const views = filteredViews();
    document.querySelectorAll("[data-refreshed-idx]").forEach((el) => {
      const text = lastRefreshedText(views[Number(el.dataset.refreshedIdx)]);
      if (text && el.textContent !== text) el.textContent = text;
    });
  }, 1000);
})();
