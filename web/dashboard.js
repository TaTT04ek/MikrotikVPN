const el = (id) => document.getElementById(id);
const prefersDark = matchMedia("(prefers-color-scheme: dark)");
const speed = [];
const latency = [];
const visibleCharts = new Set();
const decimal = new Intl.NumberFormat("ru", { maximumFractionDigits: 1 });
const clock = new Intl.DateTimeFormat("ru", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
let previousTraffic;
let lastCheck;

function bitrate(bytesPerSecond) {
  let value = bytesPerSecond * 8;
  const units = ["бит/с", "кбит/с", "Мбит/с", "Гбит/с"];
  let i = 0;
  while (value >= 1000 && i < units.length - 1) {
    value /= 1000;
    i++;
  }
  return `${decimal.format(value)} ${units[i]}`;
}

function drawChart(canvas, samples, unit, minimum) {
  if (!visibleCharts.has(canvas) || document.hidden) return;
  const width = canvas.clientWidth;
  const height = canvas.clientHeight;
  if (!width || !height) return;
  const ctx = canvas.getContext("2d");
  if (!ctx) return;
  const scale = Math.min(devicePixelRatio || 1, 2);
  const pixelWidth = Math.round(width * scale);
  const pixelHeight = Math.round(height * scale);
  if (canvas.width !== pixelWidth) canvas.width = pixelWidth;
  if (canvas.height !== pixelHeight) canvas.height = pixelHeight;
  ctx.setTransform(scale, 0, 0, scale, 0, 0);
  ctx.clearRect(0, 0, width, height);
  const style = getComputedStyle(canvas);
  const bottom = height - 24;
  const top = 9;
  const right = width - 8;
  const values = samples.flatMap(({ a, b }) => [a, b]).filter(Number.isFinite);
  const limit = Math.max(minimum, ...values);

  ctx.font = "11px system-ui";
  ctx.textAlign = "left";
  ctx.lineWidth = 1;
  const left = Math.ceil(ctx.measureText(unit(limit)).width) + 8;
  ctx.fillStyle = style.getPropertyValue("--muted");
  ctx.strokeStyle = style.getPropertyValue("--line");
  for (const fraction of [0, .5, 1]) {
    const y = bottom - (bottom - top) * fraction;
    ctx.fillText(unit(limit * fraction), 0, y + 4);
    ctx.beginPath();
    ctx.moveTo(left, y);
    ctx.lineTo(right, y);
    ctx.stroke();
  }
  if (!values.length) return;

  const first = samples[0].at;
  const last = samples.at(-1).at;
  ctx.fillText(clock.format(first), left, height - 4);
  if (last !== first && right - left > 118) {
    ctx.textAlign = "right";
    ctx.fillText(clock.format(last), right, height - 4);
  }
  ctx.lineWidth = 2;
  ctx.lineJoin = "round";
  for (const [key, cssColor] of [["a", "--chart-rx"], ["b", "--chart-tx"]]) {
    ctx.fillStyle = ctx.strokeStyle = style.getPropertyValue(cssColor);
    let previous;
    for (const sample of samples) {
      const value = sample[key];
      if (!Number.isFinite(value)) {
        previous = undefined;
        continue;
      }
      const x = first === last ? left : left + (right - left) * (sample.at - first) / (last - first);
      const y = bottom - (bottom - top) * value / limit;
      ctx.beginPath();
      if (previous) {
        ctx.moveTo(...previous);
        ctx.lineTo(x, y);
        ctx.stroke();
      } else {
        ctx.arc(x, y, 2.2, 0, Math.PI * 2);
        ctx.fill();
      }
      previous = [x, y];
    }
  }
}

function drawTraffic() {
  drawChart(el("traffic-chart"), speed, bitrate, 125);
}

function drawLatency() {
  drawChart(el("latency-chart"), latency, (value) => `${decimal.format(value)} мс`, 10);
}

function draw() {
  drawTraffic();
  drawLatency();
}

export function setupDashboard() {
  const theme = el("theme");
  try {
    const saved = localStorage.getItem("mikrotik-theme");
    if (["auto", "light", "dark"].includes(saved)) theme.value = saved;
  } catch { /* Private browsing can block storage. */ }
  const applyTheme = () => {
    document.documentElement.dataset.theme = theme.value === "auto" ? (prefersDark.matches ? "dark" : "light") : theme.value;
    draw();
  };
  theme.addEventListener("change", () => {
    try { localStorage.setItem("mikrotik-theme", theme.value); } catch { /* The theme still applies locally. */ }
    applyTheme();
  });
  prefersDark.addEventListener("change", applyTheme);
  const visibility = new IntersectionObserver((entries) => {
    for (const { target, isIntersecting } of entries) {
      if (isIntersecting) visibleCharts.add(target);
      else visibleCharts.delete(target);
    }
    draw();
  });
  const resize = new ResizeObserver(draw);
  for (const id of ["traffic-chart", "latency-chart"]) {
    visibility.observe(el(id));
    resize.observe(el(id));
  }
  document.addEventListener("visibilitychange", () => {
    if (document.hidden) previousTraffic = undefined;
    else draw();
  });
  const updateNavigation = () => {
    const destination = location.hash || "#overview";
    for (const link of document.querySelectorAll("nav a")) {
      if (link.hash === destination) link.setAttribute("aria-current", "location");
      else link.removeAttribute("aria-current");
    }
  };
  addEventListener("hashchange", updateNavigation);
  updateNavigation();
  applyTheme();
}

export function updateDashboard(data, tcp, udp) {
  el("available").textContent = data.coreError ? "—" : `${data.nodes.filter(({ delay }) => delay > 0).length} / ${data.nodes.length}`;
  el("delay").textContent = !data.coreError && tcp?.delay ? `${tcp.delay} мс` : "—";
  const checked = Date.parse(data.checkedAt);
  if (checked > 0 && checked !== lastCheck) {
    const a = !data.coreError && tcp?.delay || null;
    const b = !data.coreError && udp?.delay || null;
    latency.push({ at: checked, a, b });
    if (latency.length > 40) latency.shift();
    lastCheck = checked;
    el("latency-caption").textContent = latency.some(({ a, b }) => a !== null || b !== null)
      ? "HTTP-проверки выбранных выходов."
      : "Нет успешных проверок выбранных VPN-узлов.";
    drawLatency();
  }

  const traffic = data.traffic;
  el("traffic-iface").textContent = traffic?.interface ?? "";
  if (traffic?.sampledAt !== previousTraffic?.sampledAt || !traffic) {
    let received = null;
    let sent = null;
    if (traffic && previousTraffic?.interface === traffic.interface) {
      const seconds = (traffic.sampledAt - previousTraffic.sampledAt) / 1000;
      if (seconds > 0 && seconds <= 25 && traffic.received >= previousTraffic.received && traffic.sent >= previousTraffic.sent) {
        received = (traffic.received - previousTraffic.received) / seconds;
        sent = (traffic.sent - previousTraffic.sent) / seconds;
      }
    }
    el("rx-rate").textContent = received === null ? "—" : bitrate(received);
    el("tx-rate").textContent = sent === null ? "—" : bitrate(sent);
    el("traffic-caption").textContent = !traffic ? "Счётчики интерфейса недоступны." : received === null
      ? "Ожидаем следующий замер." : "Весь трафик интерфейса контейнера.";
    if (traffic) {
      speed.push({ at: traffic.sampledAt, a: received, b: sent });
      if (speed.length > 60) speed.shift();
      drawTraffic();
    }
    previousTraffic = traffic;
  }
}
