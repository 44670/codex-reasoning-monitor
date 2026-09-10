"use strict";
const $ = (s) => document.querySelector(s);
const $$ = (s) => [...document.querySelectorAll(s)];
const words = {
  en: {
    brand: "Codex Local Intelligence Radar",
    eyebrow: "LOCAL REASONING OBSERVATORY",
    headline: "Reasoning.",
    headlineAccent: "Revealed.",
    method:
      "Measured in model-reported reasoning tokens. An observable trace of reasoning effort, not an intelligence score.",
    reasoningShare: "Reasoning / output tokens",
    depthLabel: "MAX REASONING / LAST 30 MINUTES",
    peakResponse: "Response with the 30-minute maximum",
    count516: "516 EVENTS",
    inspect516: "Click to filter events and sessions ↗",
    ratio516: "{n}% of token events",
    medianLabel: "Median reasoning",
    medianNote: "Per token event in the selected period",
    p95Label: "P95 reasoning",
    p95Note: "95th percentile · raw event distribution",
    coverage: "Observed {h}h {m}m / {total}h",
    coverageNote: "Shaded chart regions indicate missing observation coverage.",
    storageError:
      "LOG RECORDING FAILED — statistics may be incomplete. Check the terminal, fix storage, then restart.",
    loadMore: "Load older",
    filtered516: "516 events only · click the 516 card to clear",
    no516: "No 516 events in this period",
    no516Note: "Choose another time range or clear the 516 filter.",
    eventLoadError: "Could not load events; retrying.",
    range1440: "Last 24 hours",
    range10080: "Last 7 days",
    rangeScope: "Selected period · {range}",
    peakRule: "Maximum = 1000: green · 516: red · otherwise: yellow",
    observationLabel: "{range} STATISTICS",
    awaitingSignal: "No token events in the last 30 minutes",
    subtitle:
      "How much reasoning went into that response? See the model’s reported reasoning effort, as it arrives.",
    observing: "Observation time",
    reasoning: "Reasoning tokens",
    reasoningTag: "REASONING",
    sinceLaunch: "selected period",
    output: "Output tokens",
    outputNote: "Reported output, including reasoning",
    sessionsActive: "Active sessions",
    activeNote: "Sessions with a writer lock",
    responses: "Token events",
    lastEvent: "Last event",
    noEventsYet: "Listening for the first event",
    signalLabel: "TOKEN TELEMETRY",
    signalTitle: "Reasoning waveform",
    reasoningShort: "Reasoning",
    outputShort: "Output",
    chartNote: "All sessions · grouped by arrival time",
    radarLabel: "SESSION RADAR",
    radarTitle: "In your orbit",
    radarCenter: "active sessions",
    radarWaiting: "Scanning for session activity",
    radarTracking: "{n} active · {w} waiting",
    sessionsTitle: "Sessions",
    eventsTitle: "Live feed",
    all: "All sessions",
    active: "Active",
    waiting: "Waiting",
    stopped: "Inactive",
    eventHead: "EVENT / SESSION",
    eventValues: "TOKENS",
    eventsLimit: "Newest first · more history available below",
    localOnly: "Local data. Local connection.",
    disclaimer: "Reasoning volume is not a measure of intelligence or quality.",
    detailTitle: "SESSION INSPECTOR",
    demo: "Explore demo",
    exitDemo: "Exit demo",
    demoNotice:
      "DEMO MODE — Illustrative data only. Live monitoring continues separately.",
    demoSource: "Illustrative data",
    liveSource: "Since monitor launch",
    connected: "Live connection",
    offline: "Reconnecting",
    offlineNotice:
      "Connection lost. Showing the last received data; reconnecting automatically.",
    connecting: "Connecting",
    search: "Find a session…",
    pause: "Pause",
    resume: "Resume",
    untitled: "Untitled session",
    waitingCount: "{n} waiting",
    ago: "{n}s ago",
    minutesAgo: "{n}m ago",
    hoursAgo: "{n}h ago",
    justNow: "just now",
    perBucket: "tokens / {n}s",
    windowTotal: "{n} reasoning",
    chartEmpty: "Your next response starts here.",
    chartEmptyNote: "Token events will appear as Codex reports them.",
    sessionsEmpty: "Your radar is ready.",
    sessionsEmptyNote:
      "Start or resume a Codex session on this machine. Its activity will appear here automatically.",
    noMatches: "No matching sessions",
    noMatchesNote: "Try another title, ID, or status filter.",
    feedEmpty: "Listening for signals",
    feedEmptyNote: "Session lifecycle and token events will appear here.",
    latest: "Latest response",
    sessionTotal: "Observed reasoning",
    eventCount: "{n} events",
    viewSession: "View session",
    metadataWaiting: "Waiting for session metadata",
    copy: "Copy ID",
    copied: "Copied",
    copyFailed: "Select the ID to copy",
    observedTotal: "Observation totals",
    reasoningHistory: "Reasoning · last 24 token events",
    recentEvents: "Recent session events",
    startedAt: "First observed",
    detailNote:
      "Totals cover the selected period, including catch-up reads. Inactive sessions may be historical records.",
    waitingNote:
      "The writer lock exists; session metadata is not available yet. Retrying automatically.",
    paused: "Feed paused",
    moreSessions: "{n} sessions on radar",
    START: "Started",
    ATTACH: "Attached",
    STOP: "Ended",
    WAITING: "Waiting",
    TOKENS: "Tokens",
    range15: "Last 15 minutes",
    range60: "Last hour",
    chartAria:
      "Token activity chart. {r} reasoning tokens and {o} output tokens in the selected period.",
  },
  zh: {
    brand: "Codex 本地智力雷达",
    eyebrow: "CODEX 本地推理观测站",
    headline: "推理深度。",
    headlineAccent: "此刻显形。",
    method: "以模型报告的推理 token 量化推理投入，不将其视为智力或正确率评分。",
    reasoningShare: "推理 token / 输出 token",
    depthLabel: "最近 30 分钟 · 最大思考量",
    peakResponse: "30 分钟内推理量最大的响应",
    count516: "516 事件数",
    inspect516: "点击筛选事件与会话 ↗",
    ratio516: "占全部 token 事件的 {n}%",
    medianLabel: "推理量中位数",
    medianNote: "所选时段 · 单个 token 事件",
    p95Label: "P95 推理量",
    p95Note: "原始事件推理量的第 95 百分位",
    coverage: "已观测 {h} 小时 {m} 分钟 / {total} 小时",
    coverageNote: "图中阴影表示缺少观测覆盖，不代表零活动。",
    storageError:
      "日志记录失败，统计可能不完整。请查看终端，解决存储问题后重启。",
    loadMore: "加载更早",
    filtered516: "仅显示 516 事件 · 再次点击 516 卡片可取消",
    no516: "此时段没有 516 事件",
    no516Note: "可切换时间范围，或取消 516 筛选。",
    eventLoadError: "事件加载失败，正在重试。",
    range1440: "最近 24 小时",
    range10080: "最近 7 天",
    rangeScope: "所选时段 · {range}",
    peakRule: "最大值 = 1000：绿 · 516：红 · 其它：黄",
    observationLabel: "{range} 统计",
    awaitingSignal: "最近 30 分钟暂无数据",
    subtitle:
      "这一次回答，模型究竟投入了多少推理？捕捉实际报告的推理量，让每次响应的投入清晰可见。",
    observing: "持续观测",
    reasoning: "推理 TOKEN",
    reasoningTag: "推理量",
    sinceLaunch: "所选时段累计",
    output: "输出 TOKEN",
    outputNote: "报告的 output，含推理 token",
    sessionsActive: "活跃会话",
    activeNote: "持有会话写锁的会话",
    responses: "TOKEN 事件",
    lastEvent: "最近事件",
    noEventsYet: "正在等待第一个事件",
    signalLabel: "TOKEN 遥测",
    signalTitle: "推理波形",
    reasoningShort: "推理",
    outputShort: "输出",
    chartNote: "全部会话 · 按接收时间分组",
    radarLabel: "SESSION RADAR",
    radarTitle: "你的会话轨道",
    radarCenter: "活跃会话",
    radarWaiting: "正在扫描会话活动",
    radarTracking: "{n} 个活跃 · {w} 个等待",
    sessionsTitle: "会话空间",
    eventsTitle: "实时信号",
    all: "全部会话",
    active: "活跃",
    waiting: "等待中",
    stopped: "未活跃",
    eventHead: "事件 / 会话",
    eventValues: "TOKEN 数",
    eventsLimit: "新事件在前 · 可加载更早记录",
    localOnly: "本地数据，本地连接。",
    disclaimer: "推理量不代表智力水平或回答质量。",
    detailTitle: "会话观测详情",
    demo: "体验演示",
    exitDemo: "退出演示",
    demoNotice: "演示模式 — 当前展示的是示例数据。真实会话仍在后台独立监控。",
    demoSource: "演示数据",
    liveSource: "自监控程序启动起",
    connected: "实时连接",
    offline: "正在重连",
    offlineNotice: "连接已中断，当前保留最后收到的数据，正在自动重连。",
    connecting: "正在连接",
    search: "搜索会话…",
    pause: "暂停",
    resume: "继续",
    untitled: "未命名会话",
    waitingCount: "{n} 个等待",
    ago: "{n} 秒前",
    minutesAgo: "{n} 分钟前",
    hoursAgo: "{n} 小时前",
    justNow: "刚刚",
    perBucket: "token / {n} 秒",
    windowTotal: "{n} 推理 token",
    chartEmpty: "下一次响应，从这里亮起。",
    chartEmptyNote: "Codex 报告 token 用量后，活动将在此呈现。",
    sessionsEmpty: "雷达已就绪。",
    sessionsEmptyNote:
      "在这台机器上新建或恢复 Codex 会话，活动将在这里自动出现。",
    noMatches: "没有匹配的会话",
    noMatchesNote: "试试其他标题、ID 或状态筛选。",
    feedEmpty: "正在聆听信号",
    feedEmptyNote: "会话开始、结束与 token 事件将在这里出现。",
    latest: "最近一次响应",
    sessionTotal: "已观测推理量",
    eventCount: "{n} 个事件",
    viewSession: "查看会话",
    metadataWaiting: "等待会话元数据",
    copy: "复制 ID",
    copied: "已复制",
    copyFailed: "请选中 ID 复制",
    observedTotal: "本次观测累计",
    reasoningHistory: "推理量 · 最近 24 个 token 事件",
    recentEvents: "此会话最近事件",
    startedAt: "首次观测",
    detailNote:
      "累计数据来自所选时段的观测事件，包含可能的补读。未活跃会话可能是恢复的历史记录。",
    waitingNote: "已发现写锁，但暂时无法获取会话元数据，正在自动重试。",
    paused: "事件流已暂停",
    moreSessions: "雷达上有 {n} 个会话",
    START: "开始",
    ATTACH: "接入",
    STOP: "结束",
    WAITING: "等待",
    TOKENS: "用量",
    range15: "最近 15 分钟",
    range60: "最近一小时",
    chartAria:
      "Token 活动图。所选时段内有 {r} 个推理 token，{o} 个输出 token。",
  },
};
let lang;
try {
  lang = localStorage.getItem("codex-radar-language");
} catch {}
if (!["zh", "en"].includes(lang))
  lang = (navigator.languages?.[0] || navigator.language || "en")
    .toLowerCase()
    .startsWith("zh")
    ? "zh"
    : "en";
let data = null,
  liveData = null,
  connected = false,
  demo = false,
  demoData = null;
let range = 1440,
  filter = "all",
  search = "",
  paused = false,
  frozenEvents = [],
  selectedID = null;
let oldSequence = 0,
  connectionAttempted = false,
  chartPoints = [];
let only516 = false,
  extraEvents = [],
  historyExhausted = false,
  filteredEvents = [],
  nextEventCursor = 0,
  feedError = false,
  feedLoading = false;
let pollTimer,
  requestSerial = 0,
  feedSerial = 0,
  signalColor = "#ffd66b";
let chartGeometry = { w: 800, left: 40, right: 12 };
const t = (key, values = {}) =>
  Object.entries(values).reduce(
    (s, [k, v]) => s.replaceAll(`{${k}}`, String(v)),
    words[lang][key] || key,
  );
const escapeHTML = (s) =>
  String(s ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
const number = (n) =>
  new Intl.NumberFormat(lang === "zh" ? "zh-CN" : "en-US").format(n || 0);
const compact = (n) =>
  n >= 1e6
    ? (n / 1e6).toFixed(1) + "M"
    : n >= 1e3
      ? (n / 1e3).toFixed(n >= 1e4 ? 0 : 1) + "k"
      : String(n);
const title = (s) =>
  s.title ||
  (s.status === "waiting" || s.kind === "WAITING"
    ? t("metadataWaiting")
    : t("untitled"));
const clock = (at, seconds = true) =>
  new Date(at).toLocaleTimeString(lang === "zh" ? "zh-CN" : "en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    ...(seconds ? { second: "2-digit" } : {}),
    hour12: false,
  });
const ago = (at) => {
  const n = Math.max(
    0,
    Math.floor((new Date(data?.now || Date.now()) - new Date(at)) / 1000),
  );
  return n < 5
    ? t("justNow")
    : n < 60
      ? t("ago", { n })
      : n < 3600
        ? t("minutesAgo", { n: Math.floor(n / 60) })
        : t("hoursAgo", { n: Math.floor(n / 3600) });
};
function setLanguage(next) {
  lang = next;
  try {
    localStorage.setItem("codex-radar-language", lang);
  } catch {}
  localize();
  render(true);
}
function localize() {
  document.documentElement.lang = lang === "zh" ? "zh-CN" : "en";
  document.title = t("brand");
  $$("[data-i18n]").forEach(
    (el) =>
      (el.textContent = t(el.dataset.i18n, {
        range: { 15: "15M", 60: "1H", 1440: "24H", 10080: "7D" }[range],
      })),
  );
  $("#lang-zh").classList.toggle("selected", lang === "zh");
  $("#lang-en").classList.toggle("selected", lang === "en");
  $("#lang-zh").setAttribute("aria-pressed", String(lang === "zh"));
  $("#lang-en").setAttribute("aria-pressed", String(lang === "en"));
  $("#search").placeholder = t("search");
  $("#search").setAttribute("aria-label", t("search"));
  $$("[data-range]").forEach((b) => {
    b.title = t("range" + b.dataset.range);
    b.setAttribute("aria-pressed", String(Number(b.dataset.range) === range));
  });
  $$("[data-filter]").forEach((b) =>
    b.setAttribute("aria-pressed", String(b.dataset.filter === filter)),
  );
  $("#pause-events").textContent = t(paused ? "resume" : "pause");
  updateConnection();
}
function updateConnection() {
  const badge = $("#connection");
  badge.classList.toggle("offline", !connected && !demo);
  badge.querySelector("span").textContent = t(
    demo
      ? "demoSource"
      : connected
        ? "connected"
        : connectionAttempted
          ? "offline"
          : "connecting",
  );
  $("#notice").hidden = !demo && (connected || !connectionAttempted);
  $("#notice").textContent = t(demo ? "demoNotice" : "offlineNotice");
  $("#demo-toggle").textContent = t(demo ? "exitDemo" : "demo");
  $("#demo-toggle").setAttribute("aria-pressed", String(demo));
  $("#source-label").textContent = demo
    ? t("demoSource")
    : t("rangeScope", {
        range: { 15: "15m", 60: "1h", 1440: "24h", 10080: "7d" }[range],
      });
  $("#storage-notice").hidden = demo || data?.storage !== "error";
  $("#storage-notice").textContent = t("storageError");
  $(".radar-indicator").textContent = demo ? "DEMO" : connected ? "LIVE" : "—";
}
function emptyState(heading, note, icon = "◎") {
  return `<div class="empty-state"><span class="empty-icon" aria-hidden="true">${icon}</span><strong>${t(heading)}</strong><p>${t(note)}</p></div>`;
}
function render(force = false) {
  updateConnection();
  renderDepth(force);
  renderInsights();
  if (!data) {
    for (const id of [
      "reasoning-total",
      "output-total",
      "responses-total",
      "active-total",
      "sessions-count",
    ])
      $("#" + id).textContent = "0";
    $("#waiting-total").textContent = "";
    $("#radar-caption").textContent = t("radarWaiting");
    $("#last-event").textContent = t("noEventsYet");
    $("#uptime").textContent = "00:00:00";
    $("#sessions").innerHTML = emptyState("sessionsEmpty", "sessionsEmptyNote");
    $("#events").innerHTML = emptyState("feedEmpty", "feedEmptyNote", "≋");
    renderChart();
    return;
  }
  $("#reasoning-total").textContent = number(data.reasoning);
  $("#output-total").textContent = number(data.output);
  $("#responses-total").textContent = number(data.responses);
  const active = data.sessions.filter((s) => s.status === "active").length,
    waiting = data.sessions.filter((s) => s.status === "waiting").length;
  $("#active-total").textContent = number(active);
  $("#waiting-total").textContent = waiting
    ? t("waitingCount", { n: waiting })
    : "";
  $("#radar-caption").textContent =
    active || waiting
      ? t("radarTracking", { n: active, w: waiting })
      : t("radarWaiting");
  const last = data.events.at(-1);
  $("#last-event").textContent = last
    ? `${t("lastEvent")} · ${ago(last.observed)}`
    : t("noEventsYet");
  const elapsed = Math.max(
    0,
    Math.floor((new Date(data.now) - new Date(data.started)) / 1000),
  );
  $("#uptime").textContent = [
    Math.floor(elapsed / 3600),
    Math.floor(elapsed / 60) % 60,
    elapsed % 60,
  ]
    .map((n) => String(n).padStart(2, "0"))
    .join(":");
  renderChart();
  renderSessions(force);
  renderEvents();
  if (selectedID && (force || (last?.sequence || 0) !== oldSequence))
    renderDetail();
  oldSequence = last?.sequence || 0;
}
// The thirty-minute maximum is calculated by the backend independently of range and feed pagination.
let lastDepthSequence = null;
function renderDepth(force) {
  const latest = data?.peak30 || null;
  const value = $("#depth-value");
  value.textContent = latest ? number(latest.reasoning) : "—";
  value.title = latest ? number(latest.reasoning) : t("awaitingSignal");
  const tone = !latest
    ? "none"
    : latest.reasoning === 1000
      ? "green"
      : latest.reasoning === 516
        ? "red"
        : "yellow";
  signalColor = {
    none: "#8c94a6",
    green: "#65f5a6",
    red: "#ff6478",
    yellow: "#ffd66b",
  }[tone];
  $(".depth-readout").dataset.tone = tone;
  $(".instrument").style.setProperty("--signal", signalColor);
  $(".readout-unit").title = t("peakRule");
  value.style.fontSize = "";
  const maximum = parseFloat(getComputedStyle(value).fontSize);
  const available = $(".radar-stage").clientWidth * 0.76;
  value.style.fontSize =
    Math.min(
      maximum,
      available / Math.max(1, value.textContent.length * 0.61),
    ) + "px";
  $("#depth-time").textContent = latest
    ? clock(latest.at) + " · " + ago(latest.observed)
    : t("awaitingSignal");
  $("#latest-session").textContent = latest ? title(latest) : t("radarWaiting");
  $("#latest-output").textContent = latest
    ? t("outputShort") + " " + number(latest.output)
    : "—";
  const ratio =
    latest && latest.output > 0 && latest.reasoning <= latest.output
      ? (latest.reasoning / latest.output) * 100
      : null;
  $("#depth-share").textContent = ratio === null ? "—" : ratio.toFixed(1) + "%";
  const meter = $("#share-meter");
  meter.setAttribute("aria-label", t("reasoningShare"));
  if (ratio === null) meter.removeAttribute("aria-valuenow");
  else meter.setAttribute("aria-valuenow", ratio.toFixed(1));
  meter.querySelector("span").style.width = (ratio ?? 0) + "%";
  const key = latest ? (demo ? "demo-" : "live-") + latest.sequence : null;
  if (key !== lastDepthSequence && !force) {
    const readout = $(".depth-readout");
    readout.classList.remove("received");
    void readout.offsetWidth;
    readout.classList.add("received");
  }
  lastDepthSequence = key;
}
function renderChart() {
  chartPoints = data?.series || [];
  const group =
    (data?.step || { 15: 10, 60: 60, 1440: 300, 10080: 3600 }[range]) / 10;
  const points = chartPoints,
    w = Math.max(260, $("#chart").clientWidth),
    h = $("#chart").clientHeight,
    left = 40,
    right = 12,
    top = 16,
    bottom = 30,
    pw = w - left - right,
    ph = h - top - bottom;
  chartGeometry = { w, left, right };
  const maximum = Math.max(
      1,
      ...points.map((p) => Math.max(p.output, p.reasoning)),
    ),
    ceiling =
      maximum === 1
        ? 100
        : Math.ceil(maximum / Math.pow(10, Math.floor(Math.log10(maximum)))) *
          Math.pow(10, Math.floor(Math.log10(maximum)));
  const x = (i) => left + (i / Math.max(1, points.length - 1)) * pw,
    y = (v) => top + ph - (v / ceiling) * ph;
  const valid = (p) => p.covered > 0 || p.responses > 0;
  const path = (key) => {
    let previous = false;
    return points
      .map((p, i) => {
        if (!valid(p)) {
          previous = false;
          return "";
        }
        const command = previous ? "L" : "M";
        previous = true;
        return `${command}${x(i).toFixed(1)},${y(p[key]).toFixed(1)}`;
      })
      .join(" ");
  };
  const area = () => {
    let result = "",
      first = -1,
      last = -1;
    for (let i = 0; i <= points.length; i++) {
      if (i < points.length && valid(points[i])) {
        if (first < 0) {
          first = i;
          result += `M${x(i)},${y(points[i].reasoning)}`;
        } else result += ` L${x(i)},${y(points[i].reasoning)}`;
        last = i;
      } else if (first >= 0) {
        result += ` L${x(last)},${y(0)} L${x(first)},${y(0)} Z `;
        first = -1;
      }
    }
    return result;
  };
  let svg = `<svg viewBox="0 0 ${w} ${h}" preserveAspectRatio="none" aria-hidden="true"><defs><pattern id="coverage-gap" width="7" height="7" patternUnits="userSpaceOnUse"><path d="M-1 1L1-1M0 7L7 0M6 8L8 6" stroke="#80849b" stroke-opacity=".17"/></pattern><linearGradient id="mint-area" x1="0" y1="0" x2="0" y2="1"><stop stop-color="#dcff63" stop-opacity=".22"/><stop offset="1" stop-color="#dcff63" stop-opacity="0"/></linearGradient></defs>`;
  for (let i = 0; i <= 4; i++) {
    const val = (ceiling * i) / 4,
      yy = y(val);
    svg += `<line class="grid" x1="${left}" x2="${w - right}" y1="${yy}" y2="${yy}"/><text class="axis-label" x="${left - 10}" y="${yy + 3}" text-anchor="end">${compact(Math.round(val))}</text>`;
  }
  if (points.length) {
    const bw = pw / Math.max(1, points.length - 1);
    points.forEach((p, i) => {
      if (p.covered <= 0)
        svg += `<rect x="${Math.max(left, x(i) - bw / 2)}" y="${top}" width="${Math.min(bw, w - right - Math.max(left, x(i) - bw / 2))}" height="${ph}" fill="url(#coverage-gap)"/>`;
    });
    for (let i = 0; i <= 4; i++) {
      const j = Math.round(((points.length - 1) * i) / 4);
      svg += `<text class="axis-label" x="${x(j)}" y="${h - 5}" text-anchor="${i === 0 ? "start" : i === 4 ? "end" : "middle"}">${chartTime(points[j].at * 1000)}</text>`;
    }
    svg += `<path d="${area()}" fill="url(#mint-area)"/><path d="${path("output")}" fill="none" stroke="#a89aff" stroke-width="2" vector-effect="non-scaling-stroke" opacity=".75"/><path d="${path("reasoning")}" fill="none" stroke="#dcff63" stroke-width="2" vector-effect="non-scaling-stroke"/><line id="chart-crosshair" x1="0" x2="0" y1="${top}" y2="${y(0)}" stroke="#aec4b4" stroke-dasharray="3 4" visibility="hidden"/>`;
  }
  svg += "</svg>";
  const sumR = points.reduce((n, p) => n + p.reasoning, 0),
    sumO = points.reduce((n, p) => n + p.output, 0);
  if (!sumR && !sumO)
    svg += `<div class="chart-empty"><strong>${t("chartEmpty")}</strong><span>${t("chartEmptyNote")}</span></div>`;
  $("#chart").innerHTML = svg;
  $("#chart").setAttribute(
    "aria-label",
    t("chartAria", { r: number(sumR), o: number(sumO) }),
  );
  $("#bucket-label").textContent = t("perBucket", { n: group * 10 });
  $("#chart-window-total").textContent = t("windowTotal", { n: number(sumR) });
}
function sparkline(values, large = false) {
  const vals = values.length ? values : [0, 0],
    max = Math.max(1, ...vals),
    w = large ? 420 : 90,
    h = large ? 70 : 37;
  const path = vals
    .map(
      (n, i) =>
        `${i ? "L" : "M"}${(i / Math.max(1, vals.length - 1)) * w},${h - 3 - (n / max) * (h - 7)}`,
    )
    .join(" ");
  return `<svg class="sparkline" viewBox="0 0 ${w} ${h}" preserveAspectRatio="none" aria-hidden="true"><path d="${path}" fill="none" stroke="#dcff63" stroke-width="1.5" vector-effect="non-scaling-stroke"/></svg>`;
}
function renderSessions(force) {
  const list = data.sessions.filter(
    (s) =>
      (filter === "all" || s.status === filter) &&
      (!only516 || s.count516 > 0) &&
      `${title(s)} ${s.id}`.toLowerCase().includes(search.toLowerCase()),
  );
  $("#sessions-count").textContent = data.sessions.length;
  const focusID = document.activeElement?.dataset.session;
  $("#sessions").innerHTML = list.length
    ? list
        .map((s) => {
          const fresh =
            !force &&
            data.events.some(
              (e) =>
                e.sequence > oldSequence &&
                e.id === s.id &&
                e.kind === "TOKENS",
            );
          return `<button class="session-card${fresh ? " flash" : ""}" data-session="${escapeHTML(s.id)}"><div class="session-top"><span class="status-label ${s.status}"><i class="tiny-dot"></i>${t(s.status)}</span><span>${ago(s.updated)}</span></div><h3 title="${escapeHTML(title(s))}">${escapeHTML(title(s))}</h3><div class="session-values"><div><small>${t("reasoningShort")} · ${t("latest")}</small><strong class="mint">${number(s.lastReasoning)}</strong></div><div><small>${t("outputShort")} · ${t("latest")}</small><strong>${number(s.lastOutput)}</strong></div>${sparkline(s.history)}</div><div class="session-bottom"><span>${t("sessionTotal")} <span class="mint">${number(s.reasoning)}</span></span><span>${t("eventCount", { n: number(s.responses) })} <span class="arrow">↗</span></span></div></button>`;
        })
        .join("")
    : emptyState(
        data.sessions.length ? "noMatches" : "sessionsEmpty",
        data.sessions.length ? "noMatchesNote" : "sessionsEmptyNote",
      );
  if (focusID)
    $$("[data-session]")
      .find((b) => b.dataset.session === focusID)
      ?.focus({ preventScroll: true });
}
function eventHTML(e) {
  return `<div class="event"><div class="event-main"><div class="event-title" title="${escapeHTML(title(e))}">${escapeHTML(title(e))}</div><div class="event-info"><time>${clock(e.at)}</time><span class="event-type ${e.kind}">${t(e.kind)}</span></div></div><div class="event-numbers">${e.kind === "TOKENS" ? `<div class="mint">${number(e.reasoning)}<small>${t("reasoningShort")}</small></div><div class="purple">${number(e.output)}<small>${t("outputShort")}</small></div>` : '<span class="mono">—</span>'}</div></div>`;
}
function currentFeed() {
  if (only516) return filteredEvents;
  const unique = new Map(
    [...extraEvents, ...(data?.events || [])].map((e) => [e.sequence, e]),
  );
  const cutoff = new Date(data?.now || Date.now()).getTime() - range * 60000;
  return [...unique.values()]
    .filter((e) => new Date(e.observed).getTime() >= cutoff)
    .sort((a, b) => a.sequence - b.sequence);
}
function renderEvents() {
  const source = paused ? frozenEvents : currentFeed(),
    list = $("#events"),
    scroll = list.scrollTop;
  list.innerHTML = source.length
    ? [...source].reverse().map(eventHTML).join("")
    : emptyState(
        only516 ? "no516" : "feedEmpty",
        only516 ? "no516Note" : "feedEmptyNote",
        "≋",
      );
  list.scrollTop = scroll;
  $("#pause-events").textContent = t(paused ? "resume" : "pause");
  $("#pause-events").setAttribute("aria-pressed", String(paused));
  $("#feed-caption").textContent = t(
    feedError
      ? "eventLoadError"
      : paused
        ? "paused"
        : only516
          ? "filtered516"
          : "eventsLimit",
  );
  $("#load-events").disabled =
    paused ||
    feedLoading ||
    demo ||
    (only516 ? !nextEventCursor : historyExhausted || !source.length);
}
async function fetchEvents(older = false) {
  if (demo) {
    filteredEvents = demoData.allEvents.filter(
      (e) => e.kind === "TOKENS" && e.reasoning === 516,
    );
    nextEventCursor = 0;
    renderEvents();
    return;
  }
  const requestedRange = range,
    requestedFilter = only516,
    serial = ++feedSerial;
  const current = currentFeed(),
    before = older ? (only516 ? nextEventCursor : current[0]?.sequence) : 0;
  feedLoading = true;
  renderEvents();
  try {
    const response = await fetch(
      `/api/events?range=${range}${only516 ? "&reasoning=516" : ""}${before ? "&before=" + before : ""}`,
      { signal: AbortSignal.timeout(4000), cache: "no-store" },
    );
    if (!response.ok) throw Error(response.status);
    const page = await response.json();
    if (
      serial !== feedSerial ||
      requestedRange !== range ||
      requestedFilter !== only516
    )
      return;
    if (only516) {
      const records = [...filteredEvents, ...page.events];
      const cutoff =
        new Date(data?.now || Date.now()).getTime() - range * 60000;
      filteredEvents = [
        ...new Map(records.map((e) => [e.sequence, e])).values(),
      ]
        .filter((e) => new Date(e.observed).getTime() >= cutoff)
        .sort((a, b) => a.sequence - b.sequence);
      if (older || filteredEvents.length <= page.events.length)
        nextEventCursor = page.next;
    } else {
      extraEvents = [...extraEvents, ...page.events];
      historyExhausted = !page.next;
    }
    feedError = false;
  } catch {
    if (serial === feedSerial) feedError = true;
  } finally {
    if (serial === feedSerial) {
      feedLoading = false;
      renderEvents();
    }
  }
}
function renderInsights() {
  const count = data?.count516 || 0,
    n = data?.responses || 0;
  $("#count-516").textContent = number(count);
  $("#ratio-516").textContent = n
    ? t("ratio516", { n: ((count / n) * 100).toFixed(1) })
    : "—";
  $("#median-value").textContent = n ? number(data.median) : "—";
  $("#p95-value").textContent = n ? number(data.p95) : "—";
  const seconds = data?.covered || 0;
  $("#coverage-label").textContent = t("coverage", {
    h: Math.floor(seconds / 3600),
    m: Math.floor(seconds / 60) % 60,
    total: number(range / 60),
  });
  $("#filter-516").classList.toggle("selected", only516);
  $("#filter-516").setAttribute("aria-pressed", String(only516));
}
function chartTime(at, detail = false) {
  if (range < 1440) return clock(at, detail);
  return new Date(at).toLocaleString(lang === "zh" ? "zh-CN" : "en-GB", {
    month: "2-digit",
    day: "2-digit",
    ...(range === 1440 || detail
      ? { hour: "2-digit", minute: "2-digit", hour12: false }
      : {}),
  });
}
function renderDetail() {
  const s = data.sessions.find((s) => s.id === selectedID);
  if (!s) {
    $("#detail").close();
    selectedID = null;
    return;
  }
  const content = $("#detail-content"),
    scroll = $("#detail").scrollTop,
    copyFocused = document.activeElement?.id === "copy-id";
  content.innerHTML = `<h2 class="detail-title">${escapeHTML(title(s))}</h2><span class="status-label ${s.status}"><i class="tiny-dot"></i>${t(s.status)}</span><div class="detail-id"><code>${escapeHTML(s.id)}</code><button class="quiet-button" id="copy-id">${t("copy")}</button></div>${s.status === "waiting" ? `<p class="detail-note">${t("waitingNote")}</p>` : ""}<p class="detail-note">${t("startedAt")} · ${new Date(s.since).toLocaleString(lang === "zh" ? "zh-CN" : "en-GB")}</p><div class="detail-metrics"><div><small>${t("reasoningShort")} · ${t("observedTotal")}</small><strong class="mint">${number(s.reasoning)}</strong></div><div><small>${t("outputShort")} · ${t("observedTotal")}</small><strong class="purple">${number(s.output)}</strong></div></div><h3 class="detail-subtitle">${t("reasoningHistory")}</h3><div class="detail-spark">${sparkline(s.history, true)}</div><h3 class="detail-subtitle">${t("recentEvents")}</h3><div class="detail-events">${[
    ...new Map(
      [...data.events, ...currentFeed()].map((e) => [e.sequence, e]),
    ).values(),
  ]
    .sort((a, b) => b.sequence - a.sequence)
    .filter((e) => e.id === s.id)
    .slice(0, 20)
    .map(eventHTML)
    .join("")}</div><p class="detail-note">${t("detailNote")}</p>`;
  $("#detail").scrollTop = scroll;
  if (copyFocused) $("#copy-id").focus({ preventScroll: true });
}
function demoSnapshot() {
  const now = Date.now(),
    iso = (n) => new Date(n).toISOString(),
    step = { 15: 10, 60: 60, 1440: 300, 10080: 3600 }[range];
  const names =
    lang === "zh"
      ? [
          "构建 Codex 本地智力雷达",
          "重构会话存储与并发控制",
          "分析 ARM64 指令执行性能",
          "完善组件库的交互细节",
          "等待新会话元数据",
        ]
      : [
          "Building the local intelligence radar",
          "Refactoring session storage & concurrency",
          "Profiling ARM64 execution performance",
          "Polishing the component library",
          "Waiting for session metadata",
        ];
  const sessions = names.map((name, i) => ({
    id: `019f-radar-demo-session-${i + 1}`,
    title: name,
    status: i === 4 ? "waiting" : i === 3 ? "stopped" : "active",
    since: iso(now - 7 * 86400000),
    updated: iso(now - i * 17000),
    output: 0,
    reasoning: 0,
    responses: 0,
    count516: 0,
    lastOutput: 0,
    lastReasoning: 0,
    history: [],
  }));
  const all = [];
  let seq = 0;
  for (let i = 0; i < 2376; i++) {
    const at =
      i < 2016
        ? now - 7 * 86400000 + i * 300000
        : now - 3600000 + (i - 2016) * 10000;
    if (i < 2016 && at >= now - 3600000) continue;
    const reasoning =
        i === 2375
          ? 8256
          : i % 7 === 0
            ? 516
            : i % 11 === 0
              ? 1000
              : Math.round(
                  80 +
                    Math.pow(
                      Math.max(
                        0,
                        Math.sin(i * 0.26) + Math.cos(i * 0.61) * 0.45,
                      ),
                      2,
                    ) *
                      1300,
                ),
      output = Math.round(reasoning * 1.27 + 22);
    const session = sessions[i % 3];
    all.push({
      sequence: ++seq,
      id: session.id,
      title: session.title,
      kind: "TOKENS",
      at: iso(at),
      observed: iso(at),
      reasoning,
      output,
    });
  }
  all.sort((a, b) => new Date(a.observed) - new Date(b.observed));
  all.forEach((e, i) => (e.sequence = i + 1));
  const first = Math.floor((now - range * 60000) / 1000 / step) * step,
    last = Math.floor(now / 1000 / step) * step;
  const series = Array.from({ length: (last - first) / step + 1 }, (_, i) => ({
    at: first + i * step,
    output: 0,
    reasoning: 0,
    responses: 0,
    count516: 0,
    covered: step,
  }));
  const events = all.filter((e) => new Date(e.observed) >= now - range * 60000),
    values = [];
  let output = 0,
    reasoning = 0,
    count516 = 0;
  for (const e of events) {
    const bucket =
      series[
        Math.floor(
          (Math.floor(new Date(e.observed).getTime() / 1000 / step) * step -
            first) /
            step,
        )
      ];
    bucket.output += e.output;
    bucket.reasoning += e.reasoning;
    bucket.responses++;
    if (e.reasoning === 516) {
      bucket.count516++;
      count516++;
    }
    const session = sessions.find((s) => s.id === e.id);
    session.output += e.output;
    session.reasoning += e.reasoning;
    session.responses++;
    session.count516 += e.reasoning === 516 ? 1 : 0;
    session.lastOutput = e.output;
    session.lastReasoning = e.reasoning;
    session.history.push(e.reasoning);
    session.history = session.history.slice(-24);
    values.push(e.reasoning);
    output += e.output;
    reasoning += e.reasoning;
  }
  values.sort((a, b) => a - b);
  const n = values.length;
  const peak30 = all
    .filter((e) => new Date(e.observed) >= now - 1800000)
    .reduce(
      (peak, e) => (!peak || e.reasoning >= peak.reasoning ? e : peak),
      null,
    );
  return {
    started: iso(now - 3671000),
    now: iso(now),
    range,
    step,
    storage: "demo",
    output,
    reasoning,
    responses: n,
    count516,
    median: n
      ? (values[Math.floor((n - 1) / 2)] + values[Math.floor(n / 2)]) / 2
      : 0,
    p95: n ? values[Math.ceil(n * 0.95) - 1] : 0,
    covered: range * 60,
    sessions,
    events: events.slice(-200),
    allEvents: events,
    series,
    peak30,
  };
}
async function poll() {
  clearTimeout(pollTimer);
  const serial = ++requestSerial,
    requested = range;
  try {
    const response = await fetch(`/api/snapshot?range=${requested}`, {
      cache: "no-store",
      signal: AbortSignal.timeout(4000),
    });
    if (!response.ok) throw Error(response.status);
    const incoming = await response.json();
    if (serial !== requestSerial || requested !== range) return;
    if (liveData && liveData.started !== incoming.started) {
      oldSequence = 0;
      frozenEvents = [];
      paused = false;
      extraEvents = [];
      historyExhausted = false;
      filteredEvents = [];
      nextEventCursor = 0;
      feedLoading = false;
      feedError = false;
      feedSerial++;
    }
    liveData = incoming;
    connected = true;
  } catch {
    if (serial !== requestSerial) return;
    connected = false;
  }
  connectionAttempted = true;
  if (!demo) {
    data = liveData;
    render();
    if (only516 && !paused && !feedLoading) fetchEvents();
  } else updateConnection();
  if (serial === requestSerial) pollTimer = setTimeout(poll, 1000);
}
$("#lang-zh").addEventListener("click", () => {
  if (demo) {
    lang = "zh";
    demoData = demoSnapshot();
    data = demoData;
  }
  if (demo && only516) fetchEvents();
  setLanguage("zh");
});
$("#lang-en").addEventListener("click", () => {
  if (demo) {
    lang = "en";
    demoData = demoSnapshot();
    data = demoData;
  }
  if (demo && only516) fetchEvents();
  setLanguage("en");
});
$("#demo-toggle").addEventListener("click", () => {
  demo = !demo;
  paused = false;
  extraEvents = [];
  historyExhausted = false;
  filteredEvents = [];
  nextEventCursor = 0;
  feedSerial++;
  feedLoading = false;
  feedError = false;
  oldSequence = 0;
  selectedID = null;
  $("#detail").close();
  demoData = demo ? demoSnapshot() : null;
  data = demo ? demoData : liveData;
  if (only516) fetchEvents();
  localize();
  render(true);
});
$$("[data-range]").forEach((b) =>
  b.addEventListener("click", () => {
    range = Number(b.dataset.range);
    $$("[data-range]").forEach((el) =>
      el.classList.toggle("selected", el === b),
    );
    paused = false;
    extraEvents = [];
    historyExhausted = false;
    filteredEvents = [];
    nextEventCursor = 0;
    feedSerial++;
    feedLoading = false;
    feedError = false;
    localize();
    if (demo) {
      demoData = demoSnapshot();
      data = demoData;
      if (only516) fetchEvents();
      render(true);
    } else {
      liveData = null;
      data = null;
      render(true);
      poll();
    }
  }),
);
$$("[data-filter]").forEach((b) =>
  b.addEventListener("click", () => {
    filter = b.dataset.filter;
    $$("[data-filter]").forEach((el) =>
      el.classList.toggle("selected", el === b),
    );
    localize();
    if (data) renderSessions(true);
  }),
);
$("#search").addEventListener("input", (e) => {
  search = e.target.value;
  if (data) renderSessions(true);
});
$("#pause-events").addEventListener("click", () => {
  paused = !paused;
  if (paused) frozenEvents = [...currentFeed()];
  else if (only516) fetchEvents();
  if (data) renderEvents();
  else localize();
});
$("#sessions").addEventListener("click", (e) => {
  const card = e.target.closest("[data-session]");
  if (!card) return;
  selectedID = card.dataset.session;
  renderDetail();
  $("#detail").showModal();
});
$("#close-detail").addEventListener("click", () => $("#detail").close());
$("#detail").addEventListener("close", () => {
  selectedID = null;
});
$("#detail").addEventListener("click", (e) => {
  if (
    e.target === $("#detail") &&
    e.clientX < $("#detail").getBoundingClientRect().left
  )
    $("#detail").close();
});
$("#detail-content").addEventListener("click", async (e) => {
  if (e.target.id !== "copy-id") return;
  const button = e.target;
  try {
    await navigator.clipboard.writeText(selectedID);
    button.textContent = t("copied");
  } catch {
    button.textContent = t("copyFailed");
  }
});
$("#chart").addEventListener("pointermove", (e) => {
  if (!chartPoints.length) return;
  const box = $("#chart").getBoundingClientRect(),
    { w, left, right } = chartGeometry,
    x = ((e.clientX - box.left) / box.width) * w;
  const index = Math.max(
      0,
      Math.min(
        chartPoints.length - 1,
        Math.round(
          ((x - left) / (w - left - right)) * (chartPoints.length - 1),
        ),
      ),
    ),
    p = chartPoints[index];
  const line = $("#chart-crosshair");
  if (line) {
    const xx =
      left + (index / Math.max(1, chartPoints.length - 1)) * (w - left - right);
    line.setAttribute("x1", xx);
    line.setAttribute("x2", xx);
    line.setAttribute("visibility", "visible");
  }
  const tip = $("#chart-tooltip");
  tip.hidden = false;
  tip.innerHTML = `${chartTime(p.at * 1000, true)}<br><span class="mint">${t("reasoningShort")} ${number(p.reasoning)}</span><br><span class="purple">${t("outputShort")} ${number(p.output)}</span>`;
  tip.style.left =
    Math.min(e.clientX + 16, window.innerWidth - tip.offsetWidth - 12) + "px";
  tip.style.top = Math.max(8, e.clientY - tip.offsetHeight - 12) + "px";
});
$("#chart").addEventListener("pointerleave", () => {
  $("#chart-tooltip").hidden = true;
  $("#chart-crosshair")?.setAttribute("visibility", "hidden");
});
const radar = $("#radar"),
  ctx = radar.getContext("2d"),
  reduceMotion = matchMedia("(prefers-reduced-motion: reduce)");
let lastFrame = -Infinity;
function drawRadar(time, schedule = true) {
  if (schedule) requestAnimationFrame(drawRadar);
  if (
    document.hidden ||
    (schedule && time - lastFrame < (reduceMotion.matches ? 1000 : 40))
  )
    return;
  lastFrame = time;
  const rect = radar.getBoundingClientRect(),
    dpr = Math.min(devicePixelRatio || 1, 2),
    w = rect.width,
    h = rect.height;
  if (!w || !h) return;
  if (
    radar.width !== Math.round(w * dpr) ||
    radar.height !== Math.round(h * dpr)
  ) {
    radar.width = Math.round(w * dpr);
    radar.height = Math.round(h * dpr);
  }
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
  ctx.clearRect(0, 0, w, h);
  const cx = w / 2,
    cy = h / 2,
    r = Math.min(w * 0.46, h * 0.46),
    turn = reduceMotion.matches ? -0.6 : time / 22000;
  const fog = ctx.createRadialGradient(cx, cy, r * 0.1, cx, cy, r * 1.22);
  fog.addColorStop(0, "#8875ff09");
  fog.addColorStop(0.65, "#8875ff19");
  fog.addColorStop(1, "#8875ff00");
  ctx.fillStyle = fog;
  ctx.fillRect(0, 0, w, h);
  function arc(radius, start, end, color, width = 1) {
    ctx.strokeStyle = color;
    ctx.lineWidth = width;
    ctx.beginPath();
    ctx.arc(cx, cy, radius, start, end);
    ctx.stroke();
  }
  for (const scale of [0.54, 0.67, 0.79, 0.95, 1.06])
    arc(r * scale, 0, Math.PI * 2, scale === 0.79 ? "#a89aff55" : "#a89aff23");
  for (let i = 0; i < 120; i++) {
    const a = (i / 120) * Math.PI * 2,
      major = i % 10 === 0;
    ctx.strokeStyle = major ? "#b8b3e49a" : "#77769855";
    ctx.lineWidth = major ? 2 : 1;
    ctx.beginPath();
    ctx.moveTo(cx + Math.cos(a) * r * 0.96, cy + Math.sin(a) * r * 0.96);
    ctx.lineTo(
      cx + Math.cos(a) * r * (major ? 1.025 : 1),
      cy + Math.sin(a) * r * (major ? 1.025 : 1),
    );
    ctx.stroke();
  }
  ctx.shadowColor = signalColor;
  ctx.shadowBlur = 12;
  arc(r * 0.86, turn, turn + 1.2, signalColor, 3);
  arc(r * 0.86, turn + Math.PI, turn + Math.PI + 1.2, signalColor + "a0", 3);
  ctx.shadowBlur = 0;
  arc(r * 0.73, -turn * 1.3 + 0.3, -turn * 1.3 + 2.2, "#a89affaa", 1.5);
  arc(r * 1.06, -turn * 0.6 + 3, -turn * 0.6 + 3.7, "#a89affb0", 2);
  ctx.setLineDash([2, 7]);
  arc(r * 0.6, 0, Math.PI * 2, "#dcff633b");
  ctx.setLineDash([]);
  // Decorative crosshairs and orbital rings are instrumentation; dots map to actual sessions.
  for (let i = 0; i < 4; i++) {
    const a = (i * Math.PI) / 2;
    ctx.strokeStyle = "#aaa1e353";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(cx + Math.cos(a) * r * 0.88, cy + Math.sin(a) * r * 0.88);
    ctx.lineTo(cx + Math.cos(a) * r * 1.12, cy + Math.sin(a) * r * 1.12);
    ctx.stroke();
  }
  const sessions = (data?.sessions || [])
    .filter((s) => s.status !== "stopped")
    .slice(0, 18);
  sessions.forEach((s, i) => {
    let hash = 2166136261;
    for (const c of s.id)
      hash = Math.imul(hash ^ c.charCodeAt(0), 16777619) >>> 0;
    hash = Math.imul(hash ^ (hash >>> 16), 2246822507) >>> 0;
    const a = (hash % 6283) / 1000,
      rr = r * (0.8 + (hash % 15) / 100),
      x = cx + Math.cos(a) * rr,
      y = cy + Math.sin(a) * rr;
    const color = s.status === "waiting" ? "#ff9966" : "#dcff63";
    const pulse = reduceMotion.matches ? 0 : (Math.sin(time / 800 + i) + 1) / 2;
    ctx.fillStyle = color + "13";
    ctx.beginPath();
    ctx.arc(x, y, 8 + pulse * 6, 0, Math.PI * 2);
    ctx.fill();
    ctx.shadowColor = color;
    ctx.shadowBlur = 14;
    ctx.fillStyle = color;
    ctx.fillRect(x - 2.5, y - 2.5, 5, 5);
    ctx.shadowBlur = 0;
  });
}
new ResizeObserver(() => {
  drawRadar(performance.now(), false);
  renderDepth(true);
}).observe(radar);
$("#filter-516").addEventListener("click", () => {
  only516 = !only516;
  paused = false;
  extraEvents = [];
  historyExhausted = false;
  filteredEvents = [];
  nextEventCursor = 0;
  feedSerial++;
  feedLoading = false;
  if (only516) fetchEvents();
  render(true);
});
$("#load-events").addEventListener("click", () => fetchEvents(true));
new ResizeObserver(() => renderChart()).observe($("#chart"));
localize();
render(true);
poll();
requestAnimationFrame(drawRadar);
