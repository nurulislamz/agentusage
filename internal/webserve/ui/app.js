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
    error: null,
    filterOpen: false,
  };

  const $ = (id) => document.getElementById(id);

  function applyThemeTokens(tokens, override) {
    const root = document.documentElement;
    if (override === "light") {
      root.style.cssText = [
        "--base:#f4f1ea", "--mantle:#fffdf8", "--surface0:#ebe6dc",
        "--surface1:#ddd6c8", "--surface2:#cfc7b8", "--text:#2c2823",
        "--subtext:#6d6760", "--dim:#9c958d", "--accent:#7c5cbf",
        "--lavender:#6b5cae", "--teal:#2f8f86", "--crit:#c65746",
      ].join(";");
      return;
    }
    if (!tokens || !tokens.base) {
      root.style.cssText = "";
      return;
    }
    const map = {
      "--base": tokens.base,
      "--mantle": tokens.mantle,
      "--surface0": tokens.surface0,
      "--surface1": tokens.surface1,
      "--surface2": tokens.surface2,
      "--text": tokens.text,
      "--subtext": tokens.subtext,
      "--dim": tokens.dim,
      "--accent": tokens.accent,
      "--lavender": tokens.lavender,
      "--teal": tokens.teal,
      "--crit": tokens.red,
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

  async function load() {
    try {
      const res = await fetch("/api/v1/snapshots", { headers: headers() });
      if (res.status === 401) {
        $("token-modal").hidden = false;
        $("token-error").hidden = false;
        $("token-error").textContent = state.token ? "Invalid token" : "Token required";
        return;
      }
      if (!res.ok) throw new Error(`snapshots ${res.status}`);
      $("token-modal").hidden = true;
      state.envelope = await res.json();
      state.views = state.envelope.views || [];
      state.error = null;
      applyThemeTokens(state.envelope.theme_tokens, state.themeOverride);
      const visible = filteredViews();
      if (state.selected >= visible.length) {
        state.selected = Math.max(0, visible.length - 1);
      }
      showDashboard(state.views.length > 0);
      render();
    } catch (err) {
      state.error = String(err);
      showDashboard(false);
      $("empty-state").hidden = false;
      $("splash").hidden = true;
      $("app").hidden = true;
    } finally {
      state.loading = false;
    }
  }

  function showDashboard(hasData) {
    $("splash").hidden = true;
    $("empty-state").hidden = hasData;
    $("app").hidden = !hasData;
  }

<<<<<<< HEAD
  function countStatuses(views) {
    let ok = 0;
    let warn = 0;
    let err = 0;
    for (const v of views) {
      switch (v.status) {
        case "OK": ok++; break;
        case "NEAR_LIMIT": warn++; break;
        case "LIMITED":
        case "ERROR": err++; break;
        default: break;
      }
    }
    return { ok, warn, err };
  }

  function renderHeader() {
    const env = state.envelope;
    const views = state.views;
    const counts = countStatuses(views);
    const root = $("header");
    root.replaceChildren();

    const spinner = state.refreshing
      ? el("span", { class: "spinner", text: SPINNER[state.spinnerFrame % SPINNER.length] })
      : null;

    const countParts = [];
    if (counts.ok) countParts.push(el("span", { class: "ok", text: `${counts.ok}●` }));
    if (counts.warn) countParts.push(el("span", { class: "warn", text: `${counts.warn}◐` }));
    if (counts.err) countParts.push(el("span", { class: "crit", text: `${counts.err}✗` }));

    const filterLabel = state.filter ? " (filtered)" : "";
    const usageMode = env?.usage_mode || "remaining";
    const modeLabel = usageMode === "used" ? "Used" : "Remaining";

    const line = el("div", { class: "header-line" }, [
      el("span", { class: "brand-bolt", text: "⚡" }),
      el("span", { class: "brand-name", text: "agentUsage" }),
      el("span", { class: "screen-tab active", text: "1:Dashboard" }),
      el("span", { class: "header-counts" }, countParts),
      spinner,
      el("span", {
        class: "header-meta",
        text: env
          ? `⊞ ${views.length} provider${views.length === 1 ? "" : "s"}${filterLabel} · ${modeLabel} · ${env.time_window || ""}`
          : "connecting…",
      }),
    ]);

    root.append(line, el("div", { class: "header-sep", text: SEP }));
  }

  function renderFooter() {
    const env = state.envelope;
    const root = $("footer");
    root.replaceChildren();

    const line = el("div", { class: "footer-line" }, [
      el("span", {}, [
        el("kbd", { text: "↑" }), document.createTextNode(" "),
        el("kbd", { text: "↓" }), document.createTextNode(" navigate"),
      ]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "/" }), document.createTextNode(" filter")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "r" }), document.createTextNode(" refresh")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "t" }), document.createTextNode(" theme")]),
      el("span", { style: "flex:1" }),
      el("span", {
        class: "dim",
        text: env
          ? `${env.theme || ""} · ${env.source || ""} · refresh ${env.refresh_interval_seconds || 30}s · ${new Date(env.generated_at).toLocaleTimeString()}`
          : state.error || "offline",
      }),
    ]);

    root.append(el("div", { class: "footer-sep", text: SEP }), line);
  }
  function render() {
    const views = filteredViews();
    const frame = $("frame");
    if (!views.length) {
      frame.textContent = state.filter ? "No matches." : "No providers.";
      return;
    }
    const view = views[state.selected];
    if (view.frame_html) {
      frame.innerHTML = view.frame_html;
    } else if (view.detail_html) {
      // Fallback if frame projection unavailable.
      frame.innerHTML = view.detail_html;
    } else {
      frame.textContent = view.summary || view.account_id;
    }
    frame.tabIndex = 0;
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

  // Approximate TUI mouse hit-test: clicks in the left third pick a provider.
  $("frame").addEventListener("click", (ev) => {
    const views = filteredViews();
    if (!views.length) return;
    const rect = ev.currentTarget.getBoundingClientRect();
    const xRatio = (ev.clientX - rect.left) / rect.width;
    if (xRatio > 0.36) return;
    const yRatio = (ev.clientY - rect.top) / Math.max(1, rect.height);
    // Skip header (~2 lines) and footer (~2 lines) roughly.
    const usable = Math.max(0, Math.min(1, (yRatio - 0.08) / 0.84));
    state.selected = Math.max(0, Math.min(views.length - 1, Math.floor(usable * views.length)));
    render();
  });

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
    if (ev.target.matches("input, textarea")) return;
    switch (ev.key) {
      case "ArrowUp":
      case "k":
        ev.preventDefault();
        moveSelection(-1);
        break;
      case "ArrowDown":
      case "j":
        ev.preventDefault();
        moveSelection(1);
        break;
      case "/":
        ev.preventDefault();
        openFilter();
        break;
      case "r":
      case "R":
        ev.preventDefault();
        load().catch(console.error);
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
})();
