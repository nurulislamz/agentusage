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
    const appShell = $("app");
    if (appShell) appShell.classList.add("refreshing");
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
      if ($("status-bar")) $("status-bar").hidden = true;
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
      state.loading = false;
      if (appShell) {
        setTimeout(() => appShell.classList.remove("refreshing"), 300);
      }
    }
  }

  function showDashboard(hasData) {
    $("splash").hidden = true;
    $("empty-state").hidden = hasData;
    $("app").hidden = !hasData;
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
