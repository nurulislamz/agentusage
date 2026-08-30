(() => {
  "use strict";

  const SPINNER = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];

  const state = {
    envelope: null,
    views: [],
    selected: 0,
    filter: "",
    token: sessionStorage.getItem("ou-serve-token") || "",
    themeOverride: localStorage.getItem("ou-serve-theme-override") || "",
    viewMode: localStorage.getItem("ou-serve-view-mode") || "structured",
    loading: true,
    refreshing: false,
    spinnerFrame: 0,
    error: null,
  };

  const $ = (id) => document.getElementById(id);

  const el = (tag, attrs = {}, children = []) => {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs)) {
      if (k === "class") node.className = v;
      else if (k === "text") node.textContent = v;
      else if (k === "html") node.innerHTML = v;
      else if (k.startsWith("on") && typeof v === "function") node.addEventListener(k.slice(2).toLowerCase(), v);
      else if (v != null) node.setAttribute(k, v);
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
      default: return "";
    }
  }

  function gaugeColor(pct) {
    if (pct < 50) return "var(--ok, #a6e3a1)";
    if (pct < 80) return "var(--warn, #f9e2af)";
    return "var(--crit, #f38ba8)";
  }

  function renderMiniGauge(percent) {
    if (percent < 0 || !Number.isFinite(percent)) return "";
    const pct = Math.min(100, Math.max(0, percent));
    return el("div", { class: "mini-gauge-track" }, [
      el("div", { class: "mini-gauge-fill", style: `width: ${pct}%; background: ${gaugeColor(pct)}` })
    ]).outerHTML;
  }

  function applyThemeTokens(tokens, override) {
    const root = document.documentElement;
    if (override) {
      root.dataset.theme = override;
    } else {
      if (!tokens || !tokens.base) {
        root.dataset.theme = "deep-space";
        return;
      }
      root.dataset.theme = "dynamic";
    }
    
    if (override === "light" || override === "deep-space" || override === "tokyo-night") return;

    if (!tokens) return;
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
    for (const [key, val] of Object.entries(map)) {
      if (val) root.style.setProperty(key, val);
    }
  }

  function cycleThemeOverride() {
    const order = ["", "deep-space", "tokyo-night", "light"];
    const idx = order.indexOf(state.themeOverride);
    state.themeOverride = order[(idx + 1) % order.length];
    localStorage.setItem("ou-serve-theme-override", state.themeOverride);
    applyThemeTokens(state.envelope?.theme_tokens, state.themeOverride);
    renderHeader(); // Re-render to update the theme button state if needed
  }

  function toggleViewMode() {
    state.viewMode = state.viewMode === "structured" ? "raw" : "structured";
    localStorage.setItem("ou-serve-view-mode", state.viewMode);
    renderPanel();
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
      if (state.selected >= visible.length) state.selected = Math.max(0, visible.length - 1);
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
    const filtered = filteredViews();
    const counts = countStatuses(views);
    const root = $("header");

    const spinner = state.refreshing
      ? el("span", { class: "spinner", text: SPINNER[state.spinnerFrame % SPINNER.length] })
      : null;

    const countParts = [];
    if (counts.ok) countParts.push(el("span", { class: "ok", text: `${counts.ok}●` }));
    if (counts.warn) countParts.push(el("span", { class: "warn", text: `${counts.warn}◐` }));
    if (counts.err) countParts.push(el("span", { class: "crit", text: `${counts.err}✗` }));

    const searchInput = el("input", {
      type: "text",
      id: "filter-input",
      placeholder: "Filter providers (/ or Ctrl+K)",
      value: state.filter,
      oninput: (e) => {
        state.filter = e.target.value;
        state.selected = 0;
        renderNav();
        renderPanel();
        const meta = document.getElementById("header-meta");
        if (meta) {
          const visible = filteredViews();
          meta.textContent = env ? `⊞ ${visible.length} of ${views.length} provider${views.length === 1 ? "" : "s"} · ${env.time_window || ""} · ${env.source}` : "connecting…";
        }
      }
    });

    const themeBtn = el("button", {
      class: "theme-toggle-btn",
      text: "Theme: " + (state.themeOverride || "dynamic"),
      onclick: cycleThemeOverride
    });

    const line = el("div", { class: "dash-header-line" }, [
      el("span", { class: "brand-bolt", text: "⚡" }),
      el("span", { class: "brand-name", text: "OpenUsage" }),
      el("span", { class: "header-counts" }, countParts),
      spinner,
      el("div", { class: "header-search" }, [searchInput]),
      themeBtn,
      el("span", {
        id: "header-meta",
        class: "header-meta",
        text: env
          ? `⊞ ${filtered.length} of ${views.length} provider${views.length === 1 ? "" : "s"} · ${env.time_window || ""} · ${env.source}`
          : "connecting…",
      }),
    ]);
    
    // Check if input had focus
    const inputFocused = document.activeElement && document.activeElement.id === "filter-input";
    
    root.replaceChildren();
    root.append(line);
    
    if (inputFocused) {
      document.getElementById("filter-input").focus();
    }
  }

  function renderFooter() {
    const env = state.envelope;
    const root = $("footer");
    root.replaceChildren();

    const pulseDot = el("span", { class: "pulse-dot" });

    const line = el("div", { class: "footer-line" }, [
      el("span", {}, [el("kbd", { text: "↑" }), document.createTextNode(" "), el("kbd", { text: "↓" }), document.createTextNode(" navigate")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "/" }), document.createTextNode(" filter")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "r" }), document.createTextNode(" refresh")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "t" }), document.createTextNode(" theme")]),
      el("span", { text: "·" }),
      el("span", {}, [el("kbd", { text: "v" }), document.createTextNode(" view mode")]),
      el("span", { style: "flex:1" }),
      el("span", {
        class: "dim flex-row align-center",
        text: env
          ? `${env.theme || ""} · ${env.usage_mode || "remaining"} · refresh ${env.refresh_interval_seconds || 30}s · ${new Date(env.generated_at).toLocaleTimeString()}`
          : state.error || "offline",
      }),
      env && !state.error ? pulseDot : null
    ]);

    root.append(line);
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
        root.append(el("div", {
          class: `nav-group-header ${active ? "active" : "inactive"}`,
          style: active ? `color:${view.accent_color}` : "",
          text: `${pID.toUpperCase()} (${groupCount})`,
        }));
      }

      const inGroup = pID === selectedProvider;
      const selected = index === state.selected;
      const miniHtml = view.gauge_percent >= 0 ? renderMiniGauge(view.gauge_percent) : "";

      root.append(el("button", {
        class: `nav-item${selected ? " selected" : ""}${inGroup && !selected ? " in-group" : ""}`,
        type: "button",
        style: `--item-accent:${view.accent_color}`,
        onclick: () => { state.selected = index; render(); },
      }, [
        el("div", { class: "nav-row1" }, [
          el("span", { class: `nav-status ${statusClass(view.status)}`, text: view.status_icon }),
          el("span", { class: "nav-name", text: view.account_id }),
          el("span", { class: `nav-badge ${statusClass(view.status)}`, text: view.status_badge }),
        ]),
        el("div", { class: "nav-row2" }, [
          el("span", { class: "nav-summary", text: view.summary || "—" }),
          miniHtml ? el("span", { class: "nav-mini-gauge", html: miniHtml }) : null,
        ]),
        el("hr", { class: "nav-sep" }),
      ]));
    });

    root.querySelector(".nav-item.selected")?.scrollIntoView({ block: "nearest" });
  }

  function renderSparkline(points) {
    if (!points?.length) return null;
    const series = points.slice(-14);
    const max = Math.max(...series.map((p) => p.value || 0), 0.01);
    const row = el("div", { class: "spark-row" });
    for (const pt of series) {
      const bar = el("div", { class: "spark-bar", title: `${pt.date}: $${(pt.value || 0).toFixed(2)}` }, [
        el("div", { class: "spark-bar-fill", style: `height: ${Math.max(4, ((pt.value || 0) / max) * 100)}%` })
      ]);
      row.append(bar);
    }
    const dates = el("div", { class: "spark-dates" });
    for (const pt of series) {
      dates.append(el("span", { class: "spark-date-label", text: pt.date.slice(5) }));
    }
    return el("div", { class: "spark-wrap card" }, [
      el("div", { class: "tile-section-title", text: "Daily spend" }),
      row,
      dates
    ]);
  }

  function renderParsedDetailSections(sections) {
    const els = [];
    for (const sec of sections || []) {
      if (!sec.lines?.length) continue;
      const title = [sec.icon, sec.title].filter(Boolean).join(" ");
      
      const card = el("div", { class: "card section-card" });
      if (title.trim()) card.append(el("div", { class: "tile-section-title", text: title }));
      
      let isBraille = false;
      const parsedLines = [];
      
      for (const line of sec.lines) {
        if (line.match(/[⠒⠤⠇⠏⠋⠙⠹⠸⠼⠴⠦⠧⣿⣶⣤⣀]/)) {
          isBraille = true;
          break;
        }
      }
      
      if (isBraille) {
        card.append(el("pre", { class: "chart-pre", text: sec.lines.join("\n") }));
      } else {
        const rows = el("div", { class: "structured-rows" });
        for (const line of sec.lines) {
          const dotMatch = line.match(/^([a-zA-Z0-9\s]+)\s+([·\.]+)\s+(.+)$/);
          if (dotMatch) {
            rows.append(el("div", { class: "dot-row" }, [
              el("span", { class: "dot-label", text: dotMatch[1].trim() }),
              el("span", { class: "dot-leader" }),
              el("span", { class: "dot-value", text: dotMatch[3].trim() })
            ]));
          } else {
            const attrMatch = line.match(/^\s+([a-zA-Z0-9\s]+)\s{2,}(.+)$/);
            if (attrMatch) {
              rows.append(el("div", { class: "attr-row" }, [
                el("span", { class: "attr-key", text: attrMatch[1].trim() }),
                el("span", { class: "attr-value", text: attrMatch[2].trim() })
              ]));
            } else {
              rows.append(el("div", { class: "raw-row", text: line }));
            }
          }
        }
        card.append(rows);
      }
      els.push(card);
    }
    return els;
  }

  function renderStructuredPanel(view, snapshot) {
    const tile = el("article", { class: "tile structured-mode", style: `--tile-accent:${view.accent_color}` });
    
    // Header Hero
    const header = el("div", { class: "tile-header-hero" }, [
      el("div", { class: "hero-top" }, [
        el("span", { class: `hero-status ${statusClass(view.status)}`, text: view.status_icon }),
        el("span", { class: "hero-title", text: view.account_id }),
        el("span", { class: `hero-badge ${statusClass(view.status)}`, text: view.status_badge }),
        el("span", { class: "hero-provider", text: view.provider_name }),
        el("div", { class: "flex-spacer" }),
        el("button", { 
          class: "view-mode-toggle", 
          text: "[ ▤ Raw TUI ]",
          onclick: toggleViewMode
        })
      ])
    ]);
    
    if (view.timestamp) {
      const ts = new Date(view.timestamp);
      const diff = Date.now() - ts.getTime();
      const age = diff > 60000 ? `${Math.floor(diff / 60000)}m ago` : ts.toLocaleTimeString();
      header.append(el("div", { class: "hero-meta" }, [
        el("span", { class: "hero-time", text: `⏱ ${age}` })
      ]));
    }
    tile.append(header);
    tile.append(el("hr", { class: "tile-accent-sep" }));

    // Primary Gauge
    if (view.gauge_percent >= 0 || view.summary) {
      const gaugeCard = el("div", { class: "card gauge-card" });
      if (view.summary || view.message) {
        gaugeCard.append(el("div", { class: "hero-message", text: view.message || view.summary }));
      }
      if (view.gauge_percent >= 0) {
        const pct = Math.min(100, Math.max(0, view.gauge_percent));
        gaugeCard.append(el("div", { class: "gauge-bar-wrap" }, [
          el("div", { class: "gauge-bar-fill", style: `width: ${pct}%; background: ${gaugeColor(pct)}` })
        ]));
      }
      if (view.resets?.length) {
        gaugeCard.append(el("div", { class: "reset-pills" }, view.resets.map((p) =>
          el("span", { class: `reset-pill${p.urgent ? " urgent" : ""}`, html: `◷ ${p.label} <strong>${p.duration}</strong>` }),
        )));
      }
      tile.append(gaugeCard);
    }
    
    // Model Breakdown
    if (snapshot?.model_usage?.length) {
      const modelCard = el("div", { class: "card model-card" }, [
        el("div", { class: "tile-section-title", text: "Model Breakdown" })
      ]);
      const getVal = (m) => (m.cost_usd != null && m.cost_usd > 0 ? m.cost_usd : (m.total_tokens || ((m.input_tokens || 0) + (m.output_tokens || 0))));
      const formatVal = (m) => {
        if (m.cost_usd != null && m.cost_usd > 0) return `$${m.cost_usd.toFixed(2)}`;
        const tok = m.total_tokens || ((m.input_tokens || 0) + (m.output_tokens || 0));
        if (tok >= 1e6) return `${(tok / 1e6).toFixed(1)}M tok`;
        if (tok >= 1e3) return `${(tok / 1e3).toFixed(1)}k tok`;
        return `${tok} tok`;
      };
      const getModelName = (m) => m.canonical || m.raw_model_id || m.canonical_lineage_id || "unknown";

      const total = snapshot.model_usage.reduce((sum, m) => sum + getVal(m), 0);
      const bar = el("div", { class: "stacked-bar" });
      const legend = el("div", { class: "model-legend" });

      const colors = ["#89b4fa", "#f38ba8", "#a6e3a1", "#f9e2af", "#cba6f7", "#94e2d5", "#fab387", "#89dceb"];

      snapshot.model_usage.forEach((m, i) => {
        const val = getVal(m);
        const pct = total > 0 ? (val / total) * 100 : 0;
        const color = colors[i % colors.length];

        bar.append(el("div", { class: "stacked-segment", style: `width: ${pct}%; background-color: ${color}`, title: `${getModelName(m)}: ${formatVal(m)} (${pct.toFixed(1)}%)` }));

        legend.append(el("div", { class: "legend-item" }, [
          el("span", { class: "legend-dot", style: `background-color: ${color}` }),
          el("span", { class: "legend-name", text: getModelName(m) }),
          el("span", { class: "legend-val dim", text: formatVal(m) }),
          el("span", { class: "legend-pct", text: `${pct.toFixed(1)}%` })
        ]));
      });
      modelCard.append(bar, legend);
      tile.append(modelCard);
    }

    const spark = renderSparkline(view.daily_cost);
    if (spark) tile.append(spark);

    const parsedSections = renderParsedDetailSections(view.detail_sections);
    parsedSections.forEach(s => tile.append(s));

    return tile;
  }

  function renderRawPanel(view) {
    const tile = el("article", { class: "tile raw-mode", style: `--tile-accent:${view.accent_color}` });
    
    tile.append(el("div", { class: "tile-header" }, [
      el("span", { class: statusClass(view.status), text: view.status_icon }),
      el("span", { class: "tile-title", text: view.account_id }),
      el("span", { class: `nav-badge ${statusClass(view.status)}`, text: view.status_badge }),
      el("span", { class: "dim", text: view.provider_name }),
      el("div", { class: "flex-spacer" }),
      el("button", { 
        class: "view-mode-toggle", 
        text: "[ ✦ Dashboard ]",
        onclick: toggleViewMode
      })
    ]));
    tile.append(el("hr", { class: "tile-accent-sep" }));

    if (view.tag_label || view.tag_emoji) {
      tile.append(el("div", { class: "tile-tagline" }, [
        el("span", { text: `${view.tag_emoji || ""} ${view.tag_label || ""}`.trim() }),
        view.detail ? el("span", { class: "dim", text: view.detail }) : null,
      ]));
    }

    if (view.message) {
      tile.append(el("div", { class: "tile-hero", text: view.message }));
    } else if (view.summary) {
      tile.append(el("div", { class: "tile-hero", text: view.summary }));
    }

    if (view.resets?.length) {
      tile.append(el("div", { class: "reset-pills" }, view.resets.map((p) =>
        el("span", { class: `reset-pill${p.urgent ? " urgent" : ""}`, html: `◷ ${p.label} <strong>${p.duration}</strong>` }),
      )));
    }

    const spark = renderSparkline(view.daily_cost);
    if (spark) tile.append(spark);

    if (view.tile_lines?.length) {
      tile.append(el("pre", { class: "tile-pre", text: view.tile_lines.join("\n") }));
    }

    for (const sec of view.detail_sections || []) {
      if (!sec.lines?.length) continue;
      const title = [sec.icon, sec.title].filter(Boolean).join(" ");
      if (title.trim()) tile.append(el("div", { class: "tile-section-title", text: title }));
      const block = el("pre", { class: "detail-pre" });
      block.textContent = sec.lines.join("\n");
      tile.append(block);
    }

    if (view.timestamp) {
      const ts = new Date(view.timestamp);
      const diff = Date.now() - ts.getTime();
      const age = diff > 60000
        ? `${Math.floor(diff / 60000)}m ago`
        : ts.toLocaleTimeString();
      tile.append(el("div", { class: "tile-footer", text: `⏱ ${age}` }));
    }

    return tile;
  }

  function renderPanel() {
    const views = filteredViews();
    const root = $("panel");
    root.replaceChildren();
    if (!views.length) return;

    const view = views[state.selected];
    const snap = state.envelope.snapshots?.find(s => s.provider_id === view.provider_id && s.account_id === view.account_id);
    
    if (state.viewMode === "raw") {
      root.append(renderRawPanel(view));
    } else {
      root.append(renderStructuredPanel(view, snap));
    }
  }

  function render() {
    if (!state.envelope) return;
    renderHeader();
    renderNav();
    renderPanel();
    renderFooter();
  }

  function moveSelection(delta) {
    const views = filteredViews();
    if (!views.length) return;
    state.selected = Math.max(0, Math.min(views.length - 1, state.selected + delta));
    render();
  }

  $("token-form").addEventListener("submit", (ev) => {
    ev.preventDefault();
    state.token = $("token-input").value.trim();
    sessionStorage.setItem("ou-serve-token", state.token);
    load().catch(console.error);
  });

  document.addEventListener("keydown", (ev) => {
    if ($("token-modal").hidden === false) return;
    
    // Focus search input on / or Ctrl+K
    if (ev.key === "/" || (ev.key === "k" && (ev.ctrlKey || ev.metaKey))) {
      ev.preventDefault();
      const input = $("filter-input");
      if (input) input.focus();
      return;
    }

    if (ev.target.matches("input, textarea")) {
      if (ev.key === "Escape") {
        ev.target.value = "";
        state.filter = "";
        ev.target.blur();
        state.selected = 0;
        render();
      }
      return;
    }

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
      case "v":
      case "V":
        ev.preventDefault();
        toggleViewMode();
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
