import { setupDashboard, updateDashboard } from "./dashboard.js";

const el = (id) => document.getElementById(id);
const settingsForm = el("settings");
const rules = el("rules");
const fields = {
  probeUrl: el("probe-url"), probeInterval: el("interval"), timeout: el("timeout"),
  tolerance: el("tolerance"), refreshHours: el("refresh-hours"), defaultAction: el("default-action")
};
let initialized = false;
let busy = false;
let requestNumber = 0;
let catalog;
let renderedSources = "";

async function api(path, method = "GET", data) {
  const options = data === undefined ? { method } : {
    method, headers: { "Content-Type": "application/json" }, body: JSON.stringify(data)
  };
  const res = await fetch(path, options);
  if (!res.ok) {
    const details = await res.json().catch(() => ({}));
    throw new Error(details.error || `Ошибка HTTP ${res.status}`);
  }
  return method === "GET" ? res.json() : null;
}

function notice(message, error = false) {
  el("notice").textContent = message;
  el("notice").classList.toggle("error", error);
  el("notice").hidden = !message;
}

function timestamp(value) {
  return !value || value.startsWith("0001-") ? "" : new Date(value).toLocaleString("ru");
}

function updateOrder() {
  [...rules.children].forEach((row, index) => {
    row.querySelector(".up").disabled = index === 0;
    row.querySelector(".down").disabled = index === rules.children.length - 1;
  });
  el("add-rule").disabled = rules.children.length >= 128;
  el("add-service").disabled = !catalog?.length || rules.children.length >= 128;
}

function addRule(rule = { match: "DOMAIN-SUFFIX", value: "", action: "VPN" }) {
  const row = el("rule-template").content.firstElementChild.cloneNode(true);
  const match = row.querySelector(".match");
  const value = row.querySelector(".value");
  const service = row.querySelector(".service");
  const preview = row.querySelector(".preview");
  match.value = rule.match;
  for (const preset of catalog) service.add(new Option(preset.label, preset.id));
  if (rule.match === "SERVICE") service.value = rule.value;
  else value.value = rule.value;
  row.querySelector(".action").value = rule.action;
  row.querySelector(".up").addEventListener("click", () => {
    if (row.previousElementSibling) rules.insertBefore(row, row.previousElementSibling);
    updateOrder();
  });
  row.querySelector(".down").addEventListener("click", () => {
    if (row.nextElementSibling) rules.insertBefore(row.nextElementSibling, row);
    updateOrder();
  });
  row.querySelector(".delete").addEventListener("click", () => { row.remove(); updateOrder(); });
  function showCondition() {
    const isService = match.value === "SERVICE";
    const preset = catalog.find((item) => item.id === service.value);
    row.querySelector(".manual").hidden = isService;
    row.querySelector(".preset").hidden = !isService;
    value.disabled = isService;
    value.placeholder = match.value.includes("CIDR") ? "192.0.2.0/24" : "example.com";
    const domains = preset?.domains ?? [];
    preview.hidden = !isService || !domains.length;
    preview.querySelector("summary").textContent = `Домены (${domains.length})`;
    preview.querySelector(".domains").textContent = domains.join(" · ");
  }
  match.addEventListener("change", showCondition);
  service.addEventListener("change", showCondition);
  showCondition();
  rules.append(row);
}

function loadSettings(settings) {
  for (const [key, input] of Object.entries(fields)) input.value = settings[key];
  el("dns-mapping").checked = settings.dnsMapping;
  el("dns-help").hidden = el("dns-option").hidden || !settings.dnsMapping;
  rules.replaceChildren();
  settings.rules.forEach(addRule);
  updateOrder();
  initialized = true;
  el("save").disabled = false;
}

function renderSources(sources) {
  const signature = JSON.stringify(sources);
  if (signature === renderedSources) return;
  renderedSources = signature;
  el("subscription-status").textContent = sources.length ? "" : "Источники пока не добавлены";
  el("sources").replaceChildren(...sources.map((source, index) => {
    const row = document.createElement("div");
    row.className = "source-row";
    const info = document.createElement("div");
    const title = document.createElement("strong");
    const caption = document.createElement("p");
    const actions = document.createElement("div");
    actions.className = "actions";
    caption.className = "chart-caption";
    const kind = source.remote ? "Подписка" : "Ключ";
    title.textContent = source.label || `${kind} ${index + 1}`;
    const date = timestamp(source.updatedAt);
    caption.textContent = source.remote
      ? [`${kind} · ${source.nodeCount} в списке`, source.error || (date && `Обновлено ${date}`)].filter(Boolean).join(" · ")
      : kind;
    info.append(title, caption);
    row.append(info, actions);
    const path = `/api/sources/${encodeURIComponent(source.id)}`;
    const addButton = (label, handler) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "secondary";
      button.textContent = label;
      button.setAttribute("aria-label", `${label} · ${title.textContent}`);
      button.addEventListener("click", () => handler(button));
      actions.append(button);
    };
    if (source.remote) {
      addButton("Обновить", (button) => mutate(button, path + "/refresh", "POST", {}, "Подписка обновлена."));
    }
    addButton("Удалить", (button) => {
      if (confirm(`Удалить источник «${title.textContent}»?`)) {
        mutate(button, path, "DELETE", {}, "Источник удалён.");
      }
    });
    return row;
  }));
}

async function loadStatus() {
  const request = ++requestNumber;
  const data = await api("/api/status");
  if (request !== requestNumber) return;
  const selected = data.nodes.find((node) => node.id === data.selected);
  const selectedUdp = data.nodes.find((node) => node.id === data.selectedUdp);
  el("mode").textContent = data.tun ? "Шлюз RouterOS · IPv4" : "Локальный прокси · IPv4";
  el("active").textContent = data.coreError ? "Нет связи с ядром" : selected?.label ?? "Напрямую";
  el("active-udp").textContent = data.coreError ? "—" : selectedUdp?.label ?? "Напрямую";
  const checked = timestamp(data.checkedAt);
  el("health").textContent = data.coreError ? "Состояние маршрута неизвестно" : checked
    ? `Проверено ${checked}${selected ? "" : " · доступных VPN-серверов нет"}` : "Ожидается проверка узлов";
  el("health-dot").classList.toggle("online", Boolean(!data.coreError && selected));
  updateDashboard(data, selected, selectedUdp);
  if (data.telegram) {
    const { cidrs, error, updatedAt } = data.telegram;
    const updated = timestamp(updatedAt);
    const summary = `${cidrs.length} диапазонов IPv4 · ${updated ? `обновлено ${updated}` : "встроенный список"}`;
    el("telegram-status").textContent = error ? `${summary}. ${error}.` : summary;
    el("telegram-cidrs").textContent = cidrs.join(" · ");
  }
  renderSources(data.sources);
  el("count").textContent = data.nodes.length;
  const list = data.nodes.map((node) => {
    const item = document.createElement("div");
    const name = document.createElement("span");
    const health = document.createElement("span");
    item.className = node.id === data.selected || node.id === data.selectedUdp ? "server active" : "server";
    name.className = "server-name";
    health.className = "server-info";
    name.textContent = node.label;
    health.textContent = `${node.kind.toUpperCase()}${node.udp ? " · UDP" : ""} · ${node.delay ? `${node.delay} мс` : checked ? "нет ответа" : "ожидание"}`;
    item.append(name, health);
    return item;
  });
  if (!list.length) {
    const empty = document.createElement("p");
    empty.textContent = "Добавьте подписку или ключ, чтобы увидеть доступные серверы.";
    list.push(empty);
  }
  el("servers").replaceChildren(...list);
  if (!initialized && !busy) {
    catalog ??= await api("/api/services");
    if (request !== requestNumber) return;
    el("dns-option").hidden = !data.tun;
    loadSettings(data.settings);
  }
}

async function mutate(button, path, method, data, message, after) {
  if (busy) return;
  busy = true;
  button.disabled = true;
  notice("Выполняется…");
  try {
    await api(path, method, data);
    after?.();
    notice(message);
  } catch (err) {
    notice(err.message, true);
  } finally {
    busy = false;
    button.disabled = false;
  }
  await loadStatus().catch((err) => notice(err.message, true));
}

function showFeed(visible) {
  el("feed-url").type = visible ? "text" : "password";
  el("show-feed").textContent = visible ? "Скрыть" : "Показать";
}
el("show-feed").addEventListener("click", () => showFeed(el("feed-url").type === "password"));
el("subscription").addEventListener("submit", (event) => {
  event.preventDefault();
  mutate(event.submitter, "/api/sources", "POST", { input: el("feed-url").value }, "Источник добавлен.", () => {
    el("feed-url").value = "";
    showFeed(false);
  });
});
el("probe").addEventListener("click", (event) => mutate(event.currentTarget, "/api/probe", "POST", {}, "Проверка запущена."));
el("refresh-telegram").addEventListener("click", (event) => mutate(event.currentTarget, "/api/services/telegram/refresh", "POST", {}, "Список CIDR Telegram обновлён."));
el("add-rule").addEventListener("click", () => { addRule(); updateOrder(); });
el("add-service").addEventListener("click", () => {
  addRule({ match: "SERVICE", value: catalog[0].id, action: "VPN" });
  updateOrder();
});
el("dns-mapping").addEventListener("change", (event) => { el("dns-help").hidden = !event.target.checked; });
settingsForm.addEventListener("submit", (event) => {
  event.preventDefault();
  const data = Object.fromEntries(Object.entries(fields).map(([key, input]) => [
    key, input.type === "number" ? input.valueAsNumber : input.value.trim()
  ]));
  data.dnsMapping = el("dns-mapping").checked;
  data.rules = [...rules.children].map((row) => {
    const match = row.querySelector(".match").value;
    const value = row.querySelector(match === "SERVICE" ? ".service" : ".value").value.trim();
    return { match, value, action: row.querySelector(".action").value };
  });
  mutate(event.submitter, "/api/settings", "PUT", data, "Настройки сохранены.");
});
setupDashboard();
loadStatus().catch((err) => notice(err.message, true));
setInterval(() => { if (!busy && !document.hidden) loadStatus().catch(() => {}); }, 4000);
