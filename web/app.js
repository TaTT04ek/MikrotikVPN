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
    preview.hidden = !isService || !preset;
    preview.querySelector("summary").textContent = preset ? `Домены (${preset.domains.length})` : "Домены";
    preview.querySelector(".domains").textContent = preset?.domains.join(" · ") ?? "";
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
  const updated = timestamp(data.updatedAt);
  el("subscription-status").textContent = data.refreshError || (data.subscribed
    ? updated ? `Обновлено ${updated}` : "Подписка сохранена" : "Подписка не добавлена");
  el("refresh").disabled = !data.subscribed;
  el("remove").disabled = !data.subscribed;
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
    empty.textContent = "Добавьте подписку, чтобы увидеть доступные серверы.";
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
    if (path.startsWith("/api/subscription")) el("subscription-help").hidden = false;
  } finally {
    busy = false;
    button.disabled = false;
  }
  await loadStatus().catch((err) => notice(err.message, true));
}

function showFeed(visible) {
  el("feed-url").type = visible ? "url" : "password";
  el("show-feed").textContent = visible ? "Скрыть" : "Показать";
}
el("show-feed").addEventListener("click", () => showFeed(el("feed-url").type === "password"));
el("subscription").addEventListener("submit", (event) => {
  event.preventDefault();
  el("subscription-help").hidden = true;
  mutate(event.submitter, "/api/subscription", "PUT", { url: el("feed-url").value }, "Подписка загружена.", () => {
    el("feed-url").value = "";
    showFeed(false);
  });
});
el("feed-url").addEventListener("input", () => { el("subscription-help").hidden = true; });
el("refresh").addEventListener("click", (event) => mutate(event.currentTarget, "/api/subscription/refresh", "POST", {}, "Подписка обновлена."));
el("remove").addEventListener("click", (event) => {
  if (confirm("Удалить подписку и её серверы?")) {
    mutate(event.currentTarget, "/api/subscription", "PUT", { url: "" }, "Подписка удалена.");
  }
});
el("probe").addEventListener("click", (event) => mutate(event.currentTarget, "/api/probe", "POST", {}, "Проверка запущена."));
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
