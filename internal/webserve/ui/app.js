(() => {
  "use strict";

  const LAYOUTS = [
    { id: "split", label: "Split", hint: "Glanceable Submenu + Deep Inspector" },
    { id: "matrix", label: "Matrix", hint: "Dense Roster Matrix HUD" },
    { id: "bento", label: "Bento", hint: "Viewport Bento Glance Tiles" },
    { id: "bars", label: "Bars", hint: "Linear gauges · OpenUsage-style cards" },
    { id: "dials", label: "Dials", hint: "Radial gauges · at-a-glance remaining" },
    { id: "strips", label: "Strips", hint: "Grafana bar-gauge wall" },
  ];

  function normalizeLayout(raw) {
    const id = String(raw || "").toLowerCase();
    return LAYOUTS.some((l) => l.id === id) ? id : "split";
  }

  function detectDefaultLayout() {
    const params = new URLSearchParams(window.location.search);
    const q = params.get("layout");
    if (q) {
      const norm = normalizeLayout(q);
      if (norm) return norm;
    }
    if (window.__DEFAULT_LAYOUT__) {
      return normalizeLayout(window.__DEFAULT_LAYOUT__);
    }
    const port = String(window.location.port || "");
    if (port === "8080") return "split";
    if (port === "8081") return "matrix";
    if (port === "8082") return "bento";
    const stored = localStorage.getItem("au-serve-layout");
    if (stored) return normalizeLayout(stored);
    return "split";
  }

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    matrixExpanded: -1,
    inspectOpen: false,
    inspectView: null,
    filter: "",
    token: sessionStorage.getItem("au-serve-token") || "",
    themeOverride: localStorage.getItem("au-serve-theme-override") || "",
    layout: detectDefaultLayout(),
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
    const ctrl = new AbortController();
    const abortTimer = setTimeout(() => ctrl.abort(), 15000);
    try {
      let qs = manual ? "?refresh=1" : "";
      if (opts && opts.accountID) {
        qs += (qs ? "&" : "?") + `account_id=${encodeURIComponent(opts.accountID)}`;
      }
      const fetchPromise = fetch("api/v1/snapshots" + qs, { headers: headers(), signal: ctrl.signal });
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
        const hint = $("empty-hint");
        if (hint && (err && (err.name === "AbortError" || String(err).includes("abort")))) {
          hint.textContent = "Timed out loading usage data. Refresh the page, or start the telemetry daemon.";
        }
      }
    } finally {
      clearTimeout(abortTimer);
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
      <nav class="layout-nav" aria-label="Layout view">
        ${LAYOUTS.map((l) => `<button type="button" class="layout-btn${state.layout === l.id ? " active" : ""}" data-layout="${esc(l.id)}" title="${esc(l.hint)}">${esc(l.label)}</button>`).join("")}
      </nav>
      <span class="header-meta">⊞ ${n} agents${filteredNote} · ${esc(meta.label)} · ${esc(meta.hint)}</span>
    `;
    $("header").querySelectorAll(".layout-btn").forEach((btn) => {
      btn.addEventListener("click", () => {
        state.layout = btn.dataset.layout;
        localStorage.setItem("au-serve-layout", state.layout);
        render();
        showToast("Layout: " + layoutMeta().label);
      });
    });
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

  function quotaWindowPriority(line) {
    const s = String((line.label || "") + " " + (line.short || "") + " " + (line.hint || "") + " " + (line.value || "")).toLowerCase().replace(/[-_]/g, " ");
    if (s.includes("five hour") || s.includes("5 hour") || s.includes("5h") || s.includes("rolling")) return 1;
    if (s.includes("week") || s.includes("7d") || s.includes("7 day") || s.includes("seven day")) return 2;
    if (s.includes("month") || s.includes("30d") || s.includes("monthly")) return 3;
    if (s.includes("day") || s.includes("daily") || s.includes("today")) return 4;
    if (s.includes("session")) return 5;
    return 10;
  }

  function sortUsageLines(lines) {
    if (!lines || lines.length <= 1) return lines;
    const groupOrder = new Map();
    lines.forEach((l) => {
      const g = l.group || "";
      if (!groupOrder.has(g)) groupOrder.set(g, groupOrder.size);
    });
    return lines.slice().sort((a, b) => {
      const ga = groupOrder.get(a.group || "");
      const gb = groupOrder.get(b.group || "");
      if (ga !== gb) return ga - gb;
      return quotaWindowPriority(a) - quotaWindowPriority(b);
    });
  }

  function usageLines(v) {
    if (!v) return [];
    if (v.usage_lines && v.usage_lines.length) return sortUsageLines(v.usage_lines);
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
    return sortUsageLines(lines);
  }

  function parseRatio(s) {
    const m = String(s || "").match(/\$?\s*([\d,.]+)\s*\/\s*\$?\s*([\d,.]+)/);
    if (!m) return null;
    const used = Number(m[1].replace(/,/g, ""));
    const limit = Number(m[2].replace(/,/g, ""));
    if (!Number.isFinite(used) || !Number.isFinite(limit) || limit <= 0) return null;
    const usedPct = Math.max(0, Math.min(150, (used / limit) * 100));
    const remPct = Math.max(0, Math.min(100, 100 - Math.min(100, usedPct)));
    const money = /\$/.test(s);
    const fmt = (n) => (money ? "$" : "") + formatCompact(n);
    return {
      used,
      limit,
      percent: usageMode() === "used" ? Math.min(100, usedPct) : remPct,
      display: fmt(used) + " / " + fmt(limit),
    };
  }

  function parseMagnitude(s) {
    if (parseRatio(s)) return null;
    const str = String(s || "");
    const money = str.match(/\$([\d,.]+)/);
    if (money) {
      const n = Number(money[1].replace(/,/g, ""));
      return Number.isFinite(n) ? n : null;
    }
    const m = str.replace(/,/g, "").match(/(\d+(?:\.\d+)?)\s*([KMB])?/i);
    if (!m) return null;
    const n = Number(m[1]);
    if (!Number.isFinite(n)) return null;
    const mul = { K: 1e3, M: 1e6, B: 1e9 }[(m[2] || "").toUpperCase()] || 1;
    return n * mul;
  }

  function partShort(part, line) {
    const lab = String(line.label || "") + " " + part;
    if (/input/i.test(lab)) return "In";
    if (/output/i.test(lab)) return "Out";
    const head = String(part).replace(/[$\d].*$/, "").replace(/[:]/g, "").trim();
    if (head) return head.charAt(0).toUpperCase() + head.slice(1);
    return line.short || line.label || "Value";
  }

  function usageItems(v) {
    const items = [];
    const hasTrend = (v.daily_cost || []).length > 1;
    const isCursor = (v.provider_id || "") === "cursor";
    (usageLines(v) || []).forEach((line) => {
      const pct = clampPct(line.percent);
      if (pct != null) {
        items.push({
          kind: "quota",
          label: line.label,
          short: line.short || line.label,
          percent: pct,
          value: line.value || "",
          reset_in: line.reset_in,
          urgent: line.urgent,
          tone: line.tone || toneFromPercent(pct, v),
          group: line.group || "",
        });
        return;
      }
      if (isCursor) return;
      const parts = String(line.value || "").split(/\s*·\s*/).map((p) => p.trim()).filter(Boolean);
      if (!parts.length) {
        return;
      }
      parts.forEach((part) => {
        if (hasTrend && /^today\b/i.test(part)) return;
        const ratio = parseRatio(part);
        if (ratio) {
          items.push({
            kind: "quota",
            label: partShort(part, line),
            short: partShort(part, line),
            percent: ratio.percent,
            value: ratio.display,
            display: ratio.display,
            reset_in: line.reset_in,
            urgent: line.urgent,
            tone: line.tone || toneFromPercent(ratio.percent, v),
            group: line.group || "",
          });
          return;
        }
        const amount = parseMagnitude(part);
        if (amount != null) {
          items.push({
            kind: "amount",
            label: line.label || partShort(part, line),
            short: partShort(part, line),
            value: part,
            amount,
            tone: line.tone || "ok",
            group: line.group || "",
          });
          return;
        }
        items.push({ kind: "table", label: partShort(part, line), value: part });
      });
    });
    const amounts = items.filter((i) => i.kind === "amount");
    if (amounts.length >= 2) {
      const max = Math.max(...amounts.map((a) => a.amount), 1e-9);
      amounts.forEach((a) => {
        a.kind = "rel";
        a.percent = Math.max(0, Math.min(100, (a.amount / max) * 100));
      });
    } else {
      amounts.forEach((a) => {
        a.kind = "table";
      });
    }
    return items;
  }

  function renderMetricTable(rows) {
    if (!rows || !rows.length) return "";
    const body = rows.map((r) =>
      `<tr><th scope="row">${esc(r.label || "")}</th><td>${esc(r.value || "—")}</td></tr>`
    ).join("");
    return `<table class="metric-table"><tbody>${body}</tbody></table>`;
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

  function graphGroups(items, v) {
    const graphs = (items || []).filter((it) => it.kind !== "table");
    if (!graphs.length) return [];
    if ((v.provider_id || "") !== "antigravity") {
      return [{ title: "", items: graphs }];
    }
    const gemini = [];
    const claude = [];
    graphs.forEach((it) => {
      const text = [it.group, it.label, it.short].join(" ").toLowerCase();
      if (/gemini/.test(text)) {
        gemini.push(it);
        return;
      }
      if (/claude|opus|sonnet|gpt|\b3p\b/.test(text)) {
        claude.push(it);
        return;
      }
    });
    const groups = [];
    if (gemini.length) groups.push({ title: "Gemini", items: gemini });
    if (claude.length) groups.push({ title: "Claude / GPT", items: claude });
    if (!groups.length) groups.push({ title: "", items: graphs });
    return groups;
  }

  function renderGaugeGroups(groups, renderer) {
    if (groups.length === 1 && !groups[0].title) {
      return groups[0].items.map(renderer).join("");
    }
    return groups.map((g) =>
      `<div class="gauge-group">
        ${g.title ? `<div class="gauge-group-title">${esc(g.title)}</div>` : ""}
        <div class="gauge-group-body">${g.items.map(renderer).join("")}</div>
      </div>`
    ).join("");
  }

  function tableItems(items, v) {
    return (items || []).filter((it) => it.kind === "table");
  }

  function renderLinearGauge(item, v) {
    const pct = clampPct(item.percent);
    const tone = item.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
    if (pct == null) return "";
    const depleted = item.kind === "quota" && isDepleted(pct);
    const left = item.kind === "rel"
      ? (item.value || formatCompact(item.amount))
      : (item.display || percentCaption(pct));
    const right = item.kind === "quota" ? resetCaption(item) : "";
    return `<div class="lin-gauge ${toneClass(tone)}${depleted ? " depleted" : ""}">
      <div class="lin-name">${esc(item.short || item.label || "")}${depleted ? ` <span class="limit-flag">Limit reached</span>` : ""}</div>
      <div class="lin-track"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></div>
      <div class="lin-meta">
        <span>${esc(left)}</span>
        <span class="${item.urgent ? "urgent" : ""}">${esc(right)}</span>
      </div>
    </div>`;
  }

  function renderArcGauge(item, v) {
    const pct = clampPct(item.percent);
    const tone = item.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
    if (pct == null) return "";
    const color = gaugeColor(tone);
    const hub = item.kind === "rel"
      ? (String(item.value || "").includes("%") ? Math.round(item.amount) + "%" : formatCompact(item.amount))
      : Math.round(pct) + "%";
    return `<div class="dial ${toneClass(tone)}${item.kind === "quota" && isDepleted(pct) ? " depleted" : ""}">
      <svg viewBox="0 0 88 58" class="dial-svg" aria-hidden="true">
        <path class="dial-track" d="M10 48 A34 34 0 0 1 78 48" pathLength="100"/>
        <path class="dial-fill" d="M10 48 A34 34 0 0 1 78 48" pathLength="100" stroke="${color}" stroke-dasharray="${pct} 100"/>
        <text x="44" y="44" text-anchor="middle">${esc(hub)}</text>
      </svg>
      <span class="dial-lab">${esc(item.short || item.label || "")}</span>
      <span class="dial-reset${item.urgent ? " urgent" : ""}">${esc(item.kind === "rel" ? (item.value || "") : (resetCaption(item) || item.display || percentCaption(pct)))}</span>
    </div>`;
  }

  function renderStripGauge(item, v) {
    const pct = clampPct(item.percent);
    const tone = item.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
    if (pct == null) return "";
    const overlay = item.kind === "rel"
      ? formatCompact(item.amount)
      : (isDepleted(pct) ? "Limit" : Math.round(pct) + "%");
    return `<div class="strip-metric ${toneClass(tone)}${item.kind === "quota" && isDepleted(pct) ? " depleted" : ""}">
      <span class="strip-lab">${esc(item.short || item.label || "")}</span>
      <span class="strip-track">
        <i style="width:${pct}%;background:${gaugeColor(tone)}"></i>
        <em>${esc(overlay)}</em>
      </span>
      <span class="strip-reset${item.urgent ? " urgent" : ""}">${esc(item.kind === "rel" ? (item.value || "") : (stripResetPrefix(item.reset_in || "") || item.display || "—"))}</span>
    </div>`;
  }

  function renderBarCard(v, i) {
    const items = usageItems(v);
    const groups = graphGroups(items, v);
    const graphs = renderGaugeGroups(groups, (it) => renderLinearGauge(it, v));
    const table = renderMetricTable(tableItems(items, v));
    const hist = renderMetricTable(trendStats(v.daily_cost));
    const spark = sparkBars(v.daily_cost);
    const body = graphs || table || hist || spark
      ? `${graphs ? `<div class="lin-list">${graphs}</div>` : ""}${table}${hist || spark ? `<div class="agent-foot">${hist}${spark ? `<div class="spark-wrap" title="Usage trend">${spark}</div>` : ""}</div>` : ""}`
      : renderMetricTable([{ label: "Usage", value: v.message || "No data" }]);
    const inner = `${agentHead(v, i)}${body}`;
    return agentShell(v, i, inner);
  }

  function renderDialCard(v, i) {
    const items = usageItems(v);
    const groups = graphGroups(items, v);
    const dials = renderGaugeGroups(groups, (it) => renderArcGauge(it, v));
    const table = renderMetricTable(tableItems(items, v));
    const hist = renderMetricTable(trendStats(v.daily_cost));
    const spark = sparkLine(v.daily_cost, 120, 32);
    const body = dials || table || hist || spark
      ? `${dials ? `<div class="dial-row">${dials}</div>` : ""}${table}${hist || spark ? `<div class="agent-foot">${hist}${spark ? `<div class="spark-wrap">${spark}</div>` : ""}</div>` : ""}`
      : renderMetricTable([{ label: "Usage", value: v.message || "No data" }]);
    const inner = `${agentHead(v, i)}${body}`;
    return agentShell(v, i, inner);
  }

  function renderStripCard(v, i) {
    const items = usageItems(v);
    const groups = graphGroups(items, v);
    const metrics = renderGaugeGroups(groups, (it) => renderStripGauge(it, v));
    const table = renderMetricTable(tableItems(items, v).concat(trendStats(v.daily_cost)));
    const spark = sparkBars(v.daily_cost, 72, 22);
    const sub = agentSubtitle(v);
    const body = metrics || table || spark
      ? `<div class="strip-gauges">${metrics}${table}</div>`
      : `<div class="strip-gauges">${renderMetricTable([{ label: "Usage", value: v.summary || v.message || "No data" }])}</div>`;
    const inner = `
      <div class="strip-id">
        <span class="agent-name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
        <span class="agent-plan">${esc(sub)}</span>
        <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
      </div>
      ${body}
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

  function renderCards(cards) {
    if (!cards || !cards.length) return "";
    const isUsed = (state.envelope?.usage_mode || "remaining").toLowerCase() === "used";
    const modeWord = isUsed ? "used" : "remaining";
    const threshPct = isUsed ? 80 : 20;

    return cards.map((card) => {
      const color = card.color || "var(--fg-2)";
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
                <b class="stat-val">${pct.toFixed(1)}%</b>
                <span class="pill ${pillClass(tone)}">${esc(modeWord)}</span>
              </span>
            </div>
            <div class="gauge" style="color:${gaugeColor(tone)}">
              <i style="width:${Math.max(0, Math.min(100, pct))}%;background:currentColor"></i>
              <span class="gauge-threshold" style="left:${threshPct}%" aria-hidden="true"></span>
            </div>
            <div class="caption">${row.hint ? esc(row.hint) : `${pct.toFixed(1)}% ${esc(modeWord)}`}</div>
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
      return `<section class="card" style="--card:${esc(color)}"><div class="card-header"><h2>${esc(card.icon || "")} ${esc((card.title || "").toUpperCase())}</h2></div>${rows}</section>`;
    }).join("");
  }

  function renderCockpit(v) {
    if (!v) return "";
    const meta = [v.provider_name || v.provider_id, v.detail].filter(Boolean).join(" · ");
    const schedule = v.cycle_schedule || "";
    const summary = v.summary || "";
    const refreshed = lastRefreshedText(v);
    const fetchVisible = state.refreshing && isViewRefreshing(v) ? "" : " hidden";
    const spinChar = SPINNER[state.animFrame] || "⠋";

    const items = usageItems(v);
    const groups = graphGroups(items, v);
    const gaugeCards = renderGaugeGroups(groups, (it) => renderLinearGauge(it, v));
    const quotaSection = gaugeCards ? `<section class="card"><div class="card-header"><h2>⚡ USAGE &amp; QUOTAS</h2></div><div class="lin-list">${gaugeCards}</div></section>` : "";

    const resets = (v.resets || []).filter((r) => r && r.duration);
    let timerRows = resets.map((r) => {
      const hasDistinctTarget = r.target && r.target !== r.duration;
      const durText = r.duration ? (r.duration.startsWith("in ") ? r.duration : "in " + r.duration) : "";
      return `
      <div class="timer ${toneClass(r.urgent ? "warn" : "ok")}">
        <span class="dot"></span>
        <span>${esc(r.label || "Reset")}</span>
        <span class="when">${esc(hasDistinctTarget ? r.target : durText)}</span>
        ${hasDistinctTarget ? `<span class="hint">${esc(durText)}</span>` : ""}
      </div>
    `;
    }).join("");
    if (!timerRows && v.next_reset) {
      const nr = v.next_reset.startsWith("in ") ? v.next_reset : "in " + v.next_reset;
      timerRows = `<div class="timer tone-ok"><span class="dot"></span><span>Next Reset</span><span class="when">${esc(nr)}</span></div>`;
    }
    const timerSection = timerRows ? `<section class="card"><div class="card-header"><h2>⏱ TIMERS &amp; SCHEDULE</h2></div>${timerRows}</section>` : "";

    const dailyPoints = v.daily_cost || [];
    const stats = trendStats(dailyPoints);
    const spark = sparkLine(dailyPoints, 180, 42) || sparkBars(dailyPoints, 180, 42);
    let activitySection = "";
    if (dailyPoints.length > 0 || stats.length > 0) {
      const statsHtml = renderMetricTable(stats);
      activitySection = `<section class="card"><div class="card-header"><h2>📈 ACTIVITY &amp; TREND</h2></div>${spark ? `<div style="padding:4px 0 8px">${spark}</div>` : ""}${statsHtml}</section>`;
    }

    let extraCards = "";
    if (v.detail_cards && v.detail_cards.length > 0) {
      const filtered = v.detail_cards.filter((c) => {
        const title = String(c.title || "").toLowerCase();
        return title !== "usage" && title !== "timers";
      });
      if (filtered.length > 0) {
        extraCards = renderCards(filtered);
      }
    }

    let summaryDisplay = summary;
    if (summaryDisplay && /^\d+(\.\d+)?%$/.test(summaryDisplay.trim())) {
      summaryDisplay = `${summaryDisplay.trim()} remaining`;
    }

    return `
      <div class="hero">
        <h1>
          ${esc(v.status_icon || "●")} ${esc(v.account_id)}
          <span class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span></span>
        </h1>
        <div class="hero-right">
          <span>${esc(meta)}</span>
          <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
          <button type="button" class="footer-btn btn-cockpit-refresh" data-acc="${esc(v.account_id)}" title="Refresh account">⟳</button>
        </div>
      </div>
      <div class="subhero">
        <div>${summaryDisplay ? `<strong>${esc(summaryDisplay)}</strong>` : ""}${schedule ? ` · ${esc(schedule)}` : ""}</div>
        ${refreshed ? `<div class="last-refreshed">${esc(refreshed)}</div>` : ""}
      </div>
      <div class="accent-line ${esc(v.header_tone || "ok")}"></div>
      <div class="panel-cards-grid">
        ${quotaSection}
        ${timerSection}
        ${activitySection}
        ${extraCards}
      </div>
    `;
  }

  function openInspectModal(v) {
    if (!v) return;
    state.inspectOpen = true;
    state.inspectView = v;
    const modal = $("inspect-modal");
    const title = $("inspect-title");
    const content = $("inspect-content");
    if (!modal || !content) return;
    if (title) {
      title.innerHTML = `${esc(v.status_icon || "●")} ${esc(v.account_id)} <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || "")}</span>`;
    }
    content.innerHTML = renderCockpit(v);
    modal.hidden = false;

    content.querySelectorAll(".btn-cockpit-refresh").forEach((btn) => {
      btn.addEventListener("click", (e) => {
        e.stopPropagation();
        load({ manual: true, accountID: btn.dataset.acc }).catch(console.error);
      });
    });
  }

  function closeInspectModal() {
    state.inspectOpen = false;
    state.inspectView = null;
    const modal = $("inspect-modal");
    if (modal) modal.hidden = true;
  }

  function renderSplitLayout(views, fetchHtml) {
    const nav = $("nav");
    const panel = $("panel");
    if (!views.length) return;

    const counts = {};
    views.forEach((v) => { counts[v.provider_id] = (counts[v.provider_id] || 0) + 1; });

    let navHtml = "";
    let lastProvider = "";
    views.forEach((v, i) => {
      if (v.provider_id !== lastProvider) {
        const active = views[state.selected]?.provider_id === v.provider_id;
        navHtml += `<div class="group${active ? " active" : ""}" style="--p:${esc(v.accent_color || "var(--accent)")}">${esc((v.provider_id || "").toUpperCase())} (${counts[v.provider_id]})</div>`;
        lastProvider = v.provider_id;
      }
      const sel = i === state.selected;
      const refreshing = isViewRefreshing(v);
      const next = v.next_reset || stripResetPrefix(v.reset_hint || "");
      const lines = usageLines(v);
      const isUrgent = (v.resets || []).some((r) => r.urgent) || lines.some((l) => l.urgent);

      const microMeters = lines.slice(0, 2).map((l) => {
        const pct = clampPct(l.percent);
        const tone = l.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
        if (pct == null) {
          return `<span class="meter text"><span class="meter-lab">${esc(l.short || l.label || "")}</span><span class="meter-val">${esc(l.value || "")}</span></span>`;
        }
        return `<span class="meter ${toneClass(tone)}">
          <span class="meter-lab">${esc(l.short || l.label || "")}</span>
          <span class="mini"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></span>
          <span class="meter-val">${pct.toFixed(0)}%</span>
        </span>`;
      }).join("");

      navHtml += `
        <button type="button" class="item nav-item${sel ? " selected" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}"${sel ? ` aria-current="true"` : ""} style="--p:${esc(v.accent_color || "var(--accent)")}">
          <span class="rail"></span>
          <span class="name">${esc(v.status_icon || "●")} ${esc(v.account_id)}</span>
          <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
          <div class="meters">
            ${microMeters || `<span class="dim">${esc(v.summary || "")}</span>`}
            <span class="next${isUrgent ? " urgent" : ""}" title="Next reset">${esc(next ? "⏱ " + next : "—")}</span>
          </div>
        </button>
      `;
    });

    if (nav) {
      nav.innerHTML = navHtml;
      nav.querySelectorAll(".nav-item").forEach((el) => {
        el.addEventListener("click", () => {
          state.selected = Number(el.dataset.idx);
          render();
        });
      });
      const selectedEl = nav.querySelector(".nav-item.selected");
      if (selectedEl) selectedEl.scrollIntoView({ block: "nearest" });
    }

    const selectedView = views[state.selected] || views[0];
    panel.innerHTML = `
      ${fetchHtml}
      ${renderCockpit(selectedView)}
    `;

    panel.querySelectorAll(".btn-cockpit-refresh").forEach((btn) => {
      btn.addEventListener("click", (e) => {
        e.stopPropagation();
        load({ manual: true, accountID: btn.dataset.acc }).catch(console.error);
      });
    });
  }

  function renderMatrixLayout(views, fetchHtml) {
    const panel = $("panel");
    const hasAnyTrends = views.some((v) => v.daily_cost && v.daily_cost.length > 1);

    const rows = views.map((v, i) => {
      const sel = i === state.selected;
      const isExp = state.matrixExpanded === i;
      const lines = usageLines(v);
      const refreshing = isViewRefreshing(v);

      const renderQuotaCell = (l) => {
        if (!l) return `<span class="dim">—</span>`;
        const pct = clampPct(l.percent);
        const tone = l.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
        if (pct == null) {
          return `<span class="matrix-metric-wrap"><span class="matrix-metric-lab">${esc(l.short || l.label || "")}</span><span class="matrix-metric-val">${esc(l.value || "—")}</span></span>`;
        }
        return `
          <div class="matrix-metric-wrap ${toneClass(tone)}">
            <span class="matrix-metric-lab">${esc(l.short || l.label || "")}</span>
            <span class="matrix-metric-bar"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></span>
            <span class="matrix-metric-val">${pct.toFixed(0)}%</span>
          </div>
        `;
      };

      // In the matrix view, Antigravity only has 2 quotas (5h and Week); it does not have a Quota 3.
      // Secondary model pools should not spill into Quota 3, and quota windows should not duplicate across columns.
      let matrixLines = lines;
      if ((v.provider_id || "") === "antigravity") {
        const firstGroup = lines[0]?.group || "";
        matrixLines = lines.filter((l) => (l.group || "") === firstGroup).slice(0, 2);
      } else {
        const seen = new Set();
        matrixLines = lines.filter((l) => {
          const k = (l.short || l.label || "").toLowerCase().trim();
          if (!k || seen.has(k)) return false;
          seen.add(k);
          return true;
        });
      }

      const q1 = renderQuotaCell(matrixLines[0]);
      const q2 = renderQuotaCell(matrixLines[1]);
      const q3 = renderQuotaCell(matrixLines[2]);
      const next = v.next_reset || stripResetPrefix(v.reset_hint || "") || "—";
      const isUrgent = (v.resets || []).some((r) => r.urgent) || lines.some((l) => l.urgent);
      const spark = (v.daily_cost && v.daily_cost.length > 1)
        ? sparkBars(v.daily_cost, 72, 18) || sparkLine(v.daily_cost, 72, 18)
        : `<span class="dim">—</span>`;

      const trendTd = hasAnyTrends ? `<td><div class="matrix-trend">${spark}</div></td>` : "";
      const colSpan = hasAnyTrends ? 8 : 7;

      const mainRow = `
        <tr class="matrix-row${sel ? " selected" : ""}${isExp ? " expanded" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}" style="--p:${esc(v.accent_color || "var(--accent)")}">
          <td>
            <div class="matrix-app">
              <span class="matrix-app-icon">${esc(v.status_icon || "●")}</span>
              <div>
                <div class="matrix-app-title">${esc(v.account_id)}</div>
                <div class="matrix-app-sub">${esc(v.provider_id)}</div>
              </div>
            </div>
          </td>
          <td><span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span></td>
          <td>${q1}</td>
          <td>${q2}</td>
          <td>${q3}</td>
          <td><span class="matrix-reset${isUrgent ? " urgent" : ""}">${esc(next ? "⏱ " + next : "—")}</span></td>
          ${trendTd}
          <td><span class="matrix-toggle" title="Toggle details">${isExp ? "▾" : "▸"}</span></td>
        </tr>
      `;

      const drawerRow = isExp ? `
        <tr class="matrix-drawer-row">
          <td colspan="${colSpan}">
            <div class="matrix-drawer">
              ${renderCockpit(v)}
            </div>
          </td>
        </tr>
      ` : "";

      return mainRow + drawerRow;
    }).join("");

    const trendTh = hasAnyTrends ? `<th scope="col">Trend</th>` : "";

    panel.innerHTML = `
      ${fetchHtml}
      <div class="matrix-container">
        <table class="matrix-table" role="table" aria-label="Dense Roster Matrix">
          <thead>
            <tr>
              <th scope="col">Account</th>
              <th scope="col">Status</th>
              <th scope="col">Quota 1</th>
              <th scope="col">Quota 2</th>
              <th scope="col">Quota 3</th>
              <th scope="col">Next Reset</th>
              ${trendTh}
              <th scope="col" aria-label="Toggle"></th>
            </tr>
          </thead>
          <tbody>
            ${rows}
          </tbody>
        </table>
      </div>
    `;

    panel.querySelectorAll(".matrix-row").forEach((el) => {
      el.addEventListener("click", () => {
        const idx = Number(el.dataset.idx);
        state.selected = idx;
        state.matrixExpanded = (state.matrixExpanded === idx ? -1 : idx);
        render();
      });
    });

    panel.querySelectorAll(".btn-cockpit-refresh").forEach((btn) => {
      btn.addEventListener("click", (e) => {
        e.stopPropagation();
        load({ manual: true, accountID: btn.dataset.acc }).catch(console.error);
      });
    });
  }

  function groupViewsByProvider(views) {
    const groups = [];
    const groupMap = new Map();
    views.forEach((v, i) => {
      const pid = v.provider_id || "other";
      if (!groupMap.has(pid)) {
        const grp = {
          provider_id: pid,
          provider_name: v.provider_name || pid,
          accent_color: v.accent_color || "var(--accent)",
          items: [],
        };
        groupMap.set(pid, grp);
        groups.push(grp);
      }
      groupMap.get(pid).items.push({ view: v, index: i });
    });
    return groups;
  }

  function renderBentoLayout(views, fetchHtml) {
    const panel = $("panel");
    const groups = groupViewsByProvider(views);

    const groupHtml = groups.map((grp) => {
      const tiles = grp.items.map(({ view: v, index: i }) => {
        const sel = i === state.selected;
        const lines = usageLines(v);
        const refreshing = isViewRefreshing(v);
        const next = v.next_reset || stripResetPrefix(v.reset_hint || "") || "—";
        const isUrgent = (v.resets || []).some((r) => r.urgent) || lines.some((l) => l.urgent);
        const hasSpark = v.daily_cost && v.daily_cost.length > 1;
        const spark = hasSpark
          ? sparkBars(v.daily_cost, 60, 16) || sparkLine(v.daily_cost, 60, 16)
          : "";

        const labelCounts = {};
        lines.forEach((l) => {
          const k = (l.short || l.label || "").toLowerCase();
          labelCounts[k] = (labelCounts[k] || 0) + 1;
        });

        const quotaRows = lines.slice(0, 3).map((l) => {
          const pct = clampPct(l.percent);
          const tone = l.tone || (pct != null ? toneFromPercent(pct, v) : "ok");
          const rawLab = l.short || l.label || "";
          let displayLab = rawLab;
          if (labelCounts[rawLab.toLowerCase()] > 1 && l.group) {
            if (/gemini/i.test(l.group)) {
              displayLab = "G-" + rawLab;
            } else if (/claude|gpt|opus|sonnet/i.test(l.group)) {
              displayLab = "C-" + rawLab;
            } else {
              displayLab = l.group.slice(0, 3).toUpperCase() + "-" + rawLab;
            }
          }
          if (pct == null) {
            return `<div class="bento-quota-row"><span class="bento-quota-lab">${esc(displayLab)}</span><span class="dim" style="grid-column:span 2">${esc(l.value || "—")}</span></div>`;
          }
          return `
            <div class="bento-quota-row ${toneClass(tone)}">
              <span class="bento-quota-lab">${esc(displayLab)}</span>
              <span class="bento-quota-bar"><i style="width:${pct}%;background:${gaugeColor(tone)}"></i></span>
              <span class="bento-quota-pct">${pct.toFixed(0)}%</span>
            </div>
          `;
        }).join("");

        return `
          <div class="bento-tile${sel ? " selected" : ""}${refreshing ? " refreshing" : ""}" data-idx="${i}" style="--p:${esc(v.accent_color || "var(--accent)")}" title="Click to inspect ${esc(v.account_id)}">
            <div class="bento-tile-head">
              <div class="bento-head-left">
                <span>${esc(v.status_icon || "●")}</span>
                <span class="bento-acc-name">${esc(v.account_id)}</span>
              </div>
              <span class="pill ${pillClass(v.status_badge)}">${esc(v.status_badge || v.status || "")}</span>
            </div>
            <div class="bento-tile-body">
              ${quotaRows || `<div class="dim" style="font-size:11px">${esc(v.summary || "No active quotas")}</div>`}
            </div>
            <div class="bento-tile-foot">
              <span class="bento-reset-hint${isUrgent ? " urgent" : ""}">${esc(next ? "⏱ " + next : "—")}</span>
              ${hasSpark ? `<div class="bento-spark-wrap">${spark}</div>` : ""}
            </div>
          </div>
        `;
      }).join("");

      const anyCrit = grp.items.some(({ view }) => {
        const s = (view.status_badge || view.status || "").toLowerCase();
        return s.includes("limit") || s.includes("err") || s.includes("crit");
      });
      const anyWarn = !anyCrit && grp.items.some(({ view }) => {
        const s = (view.status_badge || view.status || "").toLowerCase();
        return s.includes("warn") || s.includes("auth");
      });
      const statusBadge = anyCrit
        ? `<span class="pill pill-crit">ATTENTION</span>`
        : anyWarn
        ? `<span class="pill pill-warn">WARNING</span>`
        : `<span class="pill pill-ok">ALL OK</span>`;

      return `
        <section class="provider-group-box" style="--p:${esc(grp.accent_color)}">
          <header class="provider-group-header">
            <div class="provider-group-title">
              <span class="provider-indicator" aria-hidden="true"></span>
              <span class="provider-name">${esc(grp.provider_name.toUpperCase())}</span>
              <span class="provider-count">${grp.items.length} ${grp.items.length === 1 ? 'AGENT' : 'AGENTS'}</span>
            </div>
            <div class="provider-group-meta">
              ${statusBadge}
            </div>
          </header>
          <div class="bento-tiles-grid">
            ${tiles}
          </div>
        </section>
      `;
    }).join("");

    panel.innerHTML = `
      ${fetchHtml}
      <div class="bento-container" role="grid" aria-label="Viewport Bento Glance Tiles">
        ${groupHtml}
      </div>
    `;

    panel.querySelectorAll(".bento-tile").forEach((el) => {
      el.addEventListener("click", () => {
        const idx = Number(el.dataset.idx);
        state.selected = idx;
        openInspectModal(views[idx]);
      });
    });
  }

  function renderLegacyBoard(views, fetchHtml) {
    const panel = $("panel");
    const id = state.layout;
    const renderer = id === "dials" ? renderDialCard : id === "strips" ? renderStripCard : renderBarCard;
    const groups = groupViewsByProvider(views);

    const groupHtml = groups.map((grp) => {
      const cards = grp.items.map(({ view: v, index: i }) => renderer(v, i)).join("");
      const anyCrit = grp.items.some(({ view }) => {
        const s = (view.status_badge || view.status || "").toLowerCase();
        return s.includes("limit") || s.includes("err") || s.includes("crit");
      });
      const anyWarn = !anyCrit && grp.items.some(({ view }) => {
        const s = (view.status_badge || view.status || "").toLowerCase();
        return s.includes("warn") || s.includes("auth");
      });
      const statusBadge = anyCrit
        ? `<span class="pill pill-crit">ATTENTION</span>`
        : anyWarn
        ? `<span class="pill pill-warn">WARNING</span>`
        : `<span class="pill pill-ok">ALL OK</span>`;

      return `
        <section class="provider-group-box" style="--p:${esc(grp.accent_color)}">
          <header class="provider-group-header">
            <div class="provider-group-title">
              <span class="provider-indicator" aria-hidden="true"></span>
              <span class="provider-name">${esc(grp.provider_name.toUpperCase())}</span>
              <span class="provider-count">${grp.items.length} ${grp.items.length === 1 ? 'AGENT' : 'AGENTS'}</span>
            </div>
            <div class="provider-group-meta">
              ${statusBadge}
            </div>
          </header>
          <div class="board-grid board-${esc(id)}">
            ${cards}
          </div>
        </section>
      `;
    }).join("");

    panel.innerHTML = `
      ${fetchHtml}
      <div class="board-container" role="list" aria-label="${esc(layoutMeta().hint)}">
        ${groupHtml}
      </div>
    `;

    panel.querySelectorAll(".agent").forEach((el) => {
      el.addEventListener("click", () => {
        state.selected = Number(el.dataset.idx);
        render();
      });
    });
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
      if ($("nav")) $("nav").innerHTML = "";
      return;
    }

    const fetchHtml = `<span id="fetching-detail" class="fetching"${fetchVisible}><span class="spin" aria-hidden="true">${spinChar}</span> <span class="fetching-text">${fetchText}</span></span>`;

    if (state.layout === "split") {
      renderSplitLayout(views, fetchHtml);
    } else if (state.layout === "matrix") {
      renderMatrixLayout(views, fetchHtml);
    } else if (state.layout === "bento") {
      renderBentoLayout(views, fetchHtml);
    } else {
      renderLegacyBoard(views, fetchHtml);
    }
  }

  function applyLayout() {
    const shell = $("app");
    if (!shell) return;
    const id = normalizeLayout(state.layout);
    state.layout = id;
    shell.dataset.layout = id;
    LAYOUTS.forEach((l) => shell.classList.toggle("layout-" + l.id, l.id === id));
    const nav = $("nav");
    const vsep = document.querySelector(".vsep");
    if (id === "split") {
      if (nav) {
        nav.hidden = false;
        nav.removeAttribute("aria-hidden");
        nav.setAttribute("aria-label", "Account roster");
      }
      if (vsep) vsep.hidden = false;
    } else {
      if (nav) {
        nav.hidden = true;
        nav.setAttribute("aria-hidden", "true");
      }
      if (vsep) vsep.hidden = true;
    }
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

  function cycleMainLayout() {
    const main = ["split", "matrix", "bento"];
    const idx = main.indexOf(state.layout);
    state.layout = main[(idx + 1) % main.length];
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
      <button type="button" class="footer-btn" id="footer-btn-layout" title="Cycle dashboard layout (Tab / l / v)"><kbd>v</kbd> <span>${esc(layoutMeta().label)}</span></button>
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

  let filterBackup = "";

  function openFilter() {
    state.filterOpen = true;
    filterBackup = state.filter;
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
    } else {
      state.filter = filterBackup;
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

  $("filter-input").addEventListener("input", () => {
    state.filter = $("filter-input").value;
    state.selected = 0;
    render();
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
    if (["ArrowUp", "ArrowDown", "j", "J", "k", "K", "/", "r", "R", "u", "U", "t", "T", "v", "V", "Tab", "l", "L", "Enter", " "].includes(key)) {
      if (ev.target && typeof ev.target.blur === "function" && ev.target !== document.body) {
        ev.target.blur();
      }
    }

    switch (key) {
      case "Tab":
      case "l":
      case "L":
        ev.preventDefault();
        cycleMainLayout();
        break;
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
      case "Enter":
      case " ":
        if (state.layout === "matrix") {
          ev.preventDefault();
          state.matrixExpanded = (state.matrixExpanded === state.selected ? -1 : state.selected);
          render();
        } else if (state.layout === "bento") {
          ev.preventDefault();
          openInspectModal(filteredViews()[state.selected]);
        }
        break;
      case "Escape":
        if (state.inspectOpen) {
          ev.preventDefault();
          closeInspectModal();
        }
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

  $("inspect-close")?.addEventListener("click", closeInspectModal);
  $("inspect-modal")?.addEventListener("click", (ev) => {
    if (ev.target === $("inspect-modal")) closeInspectModal();
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
