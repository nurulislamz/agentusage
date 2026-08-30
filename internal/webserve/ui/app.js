(() => {
  "use strict";

  const SPINNER = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
  const SEP = "━".repeat(160);
  const VSEP = Array.from({ length: 80 }, () => "┃").join("\n");

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    filter: "",
    token: sessionStorage.getItem("au-serve-token") || "",
    themeOverride: localStorage.getItem("au-serve-theme-override") || "",
    loading: true,
    refreshing: false,
    spinnerFrame: 0,
    error: null,
    filterOpen: false,
  };

  const $ = (id) => document.getElementById(id);

  const el = (tag, attrs = {}, children = []) => {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      if (k === "class") node.className = v;
      else if (k === "text") node.textContent = v;
      else if (k === "html") node.innerHTML = v;
      else if (k.startsWith("on") && typeof v === "function") {
        node.addEventListener(k.slice(2).toLowerCase(), v);
      } else if (v != null) node.setAttribute(k, v);
    }
    for (const child of children) {
      if (child != null) node.append(child);
    }
    return node;
  };

  function statusClass(status) {
    switch (status) {
      case "OK": return "ok";
      case "NEAR_LIMIT": return "warn";
      case "LIMITED":
      case "ERROR": return "crit";
      case "AUTH_REQUIRED": return "auth";
      default: return "dim";
    }
  }

  function applyThemeTokens(tokens, override) {
    const root = document.documentElement;
    if (override === "light") {
      root.dataset.theme = "light";
      root.style.cssText = [
        "--base:#f4f1ea", "--mantle:#fffdf8", "--surface0:#ebe6dc",
        "--surface1:#ddd6c8", "--surface2:#cfc7b8", "--text:#2c2823",
        "--subtext:#6d6760", "--dim:#9c958d", "--accent:#7c5cbf",
        "--blue:#3d6fb8", "--sapphire:#2f8f86", "--green:#2f9e6d",
        "--yellow:#b5811a", "--red:#c65746", "--peach:#c67a3a",
        "--teal:#2f8f86", "--lavender:#6b5cae", "--mauve:#6b5cae",
        "--ok:#2f9e6d", "--warn:#b5811a", "--crit:#c65746", "--auth:#c67a3a",
      ].join(";");
      return;
    }
    if (!tokens || !tokens.base) {
      root.dataset.theme = "deep-space";
      root.style.cssText = "";
      return;
    }
    root.dataset.theme = "dynamic";
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
      "--blue": tokens.blue,
      "--sapphire": tokens.sapphire,
      "--green": tokens.green,
      "--yellow": tokens.yellow,
      "--red": tokens.red,
      "--peach": tokens.peach,
      "--teal": tokens.teal,
      "--lavender": tokens.lavender,
      "--mauve": tokens.mauve,
      "--ok": tokens.green,
      "--warn": tokens.yellow,
      "--crit": tokens.red,
      "--auth": tokens.peach,
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
    state.refreshing = true;
    renderHeader();
    try {
      const res = await fetch("/api/v1/snapshots", { headers: headers() });
      if (res.status === 401) {
        $("token-modal").hidden = false;
        $("token-error").hidden = false;
        $("token-error").textContent = state.token ? "Invalid token" : "Token required";
        state.refreshing = false;
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
      state.refreshing = false;
      renderHeader();
      renderFooter();
    }
  }

  function showDashboard(hasData) {
    $("splash").hidden = true;
    $("empty-state").hidden = hasData;
    $("app").hidden = !hasData;
  }

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

  function renderNav() {
    const views = filteredViews();
    const root = $("nav");
    root.replaceChildren();
    if (!views.length) {
      root.append(el("p", { class: "dim", style: "padding:12px", text: state.filter ? "No matches." : "No providers." }));
      return;
    }

    const selectedView = views[state.selected];
    const selectedProvider = selectedView?.provider_id;
    let lastProvider = "";

    views.forEach((view, index) => {
      const pID = view.provider_id;
      const isFirst = pID !== lastProvider;
      lastProvider = pID;

      if (isFirst) {
        const groupCount = views.filter((v) => v.provider_id === pID).length;
        const active = pID === selectedProvider;
        const prefix = active ? "✦ " : "◈ ";
        root.append(el("div", {
          class: `nav-group${active ? " active" : ""}`,
          style: active ? `--item-accent:${view.accent_color};color:${view.accent_color}` : "",
          text: `${prefix}${pID.toUpperCase()} (${groupCount})`,
        }));
      }

      const inGroup = pID === selectedProvider;
      const selected = index === state.selected;
      const railChar = selected ? "┃" : (inGroup ? "│" : " ");

      const summaryNode = view.summary_html
        ? el("span", { class: "nav-summary", html: view.summary_html })
        : el("span", { class: "nav-summary", text: view.summary || "—" });

      const stripNode = view.strip_html
        ? el("span", { class: "nav-strip", html: view.strip_html })
        : null;

      const badgeNode = view.badge_html
        ? el("span", { class: "nav-badge", html: view.badge_html })
        : el("span", { class: `nav-badge ${statusClass(view.status)}`, text: view.status_badge || "" });

      const iconNode = view.icon_html
        ? el("span", { html: view.icon_html })
        : el("span", { class: statusClass(view.status), text: view.status_icon || "·" });

      const resetNode = view.reset_hint
        ? el("span", {
          class: `nav-reset${view.resets?.[0]?.urgent ? " urgent" : ""}`,
          text: ` · ${view.reset_hint}`,
        })
        : null;

      root.append(el("button", {
        class: `nav-item${selected ? " selected" : ""}${inGroup && !selected ? " in-group" : ""}`,
        type: "button",
        style: `--item-accent:${view.accent_color}`,
        onclick: () => { state.selected = index; render(); },
      }, [
        el("span", { class: "rail", text: railChar }),
        el("div", { class: "nav-rows" }, [
          el("div", { class: "nav-row1" }, [
            iconNode,
            el("span", { class: "nav-name", text: view.account_id }),
            badgeNode,
          ]),
          el("div", { class: "nav-row2" }, [
            stripNode,
            summaryNode,
            resetNode,
          ]),
          el("hr", { class: "nav-sep" }),
        ]),
      ]));
    });

    root.querySelector(".nav-item.selected")?.scrollIntoView({ block: "nearest" });
  }

  function renderPanel() {
    const views = filteredViews();
    const root = $("panel");
    root.replaceChildren();
    if (!views.length) {
      root.append(el("div", { class: "detail-empty", text: "No provider selected." }));
      return;
    }

    const view = views[state.selected];
    if (view.detail_html) {
      root.append(el("pre", { class: "detail-term", html: view.detail_html }));
      return;
    }

    // Fallback if detail_html missing (should not happen in current API).
    const lines = [];
    for (const sec of view.detail_sections || []) {
      const title = [sec.icon, sec.title].filter(Boolean).join(" ");
      if (title) lines.push(title);
      lines.push(...(sec.lines || []));
      lines.push("");
    }
    root.append(el("pre", { class: "detail-term", text: lines.join("\n") || view.summary || "—" }));
  }

  function renderVSep() {
    const sep = document.querySelector(".term-vsep");
    if (sep) sep.textContent = VSEP;
  }

  function render() {
    if (!state.envelope) return;
    renderHeader();
    renderNav();
    renderPanel();
    renderFooter();
    renderVSep();
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

  setInterval(() => {
    if (!state.refreshing) return;
    state.spinnerFrame += 1;
    const spin = document.querySelector(".spinner");
    if (spin) spin.textContent = SPINNER[state.spinnerFrame % SPINNER.length];
  }, 80);

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
