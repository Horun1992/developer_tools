const plateConditions = {};
const availableConditions = {};

document.addEventListener("DOMContentLoaded", () => {
  loadConditionDefaults();
});

async function loadConditionDefaults() {
  const res = await authorizedFetch('/static/conditions.txt');
  const text = await res.text();
  const lines = text.split('\n');
  const selKey = document.getElementById("plateConditionKey");
  selKey.innerHTML = "";

  lines.forEach(line => {
    const trimmed = line.trim();
    if (!trimmed) return;

    const [key, raw] = splitOnce(trimmed, '=').map(s => s.trim());
    if (!key || !raw) return;

    const match = raw.match(/^\(([^)]+)\):(.*)$/);
    const operators = match ? match[1].split("|") : null;
    const defaultVal = match ? match[2] : raw;

    availableConditions[key] = { operators, defaultVal };

    const option = new Option(key, key);
    selKey.appendChild(option);
  });

  selKey.addEventListener("change", updateConditionControls);
  updateConditionControls();
}

function updateConditionControls() {
  const key = document.getElementById("plateConditionKey").value;
  const cfg = availableConditions[key];
  const selOp = document.getElementById("plateConditionOperator");
  const valInput = document.getElementById("plateConditionValue");

  if (cfg?.operators) {
    selOp.innerHTML = cfg.operators.map(op => `<option value="${op}">${op}</option>`).join("");
    selOp.style.display = "inline-block";
  } else {
    selOp.style.display = "none";
  }

  valInput.value = cfg?.defaultVal || "";
}

function addPlateCondition() {
  const key = document.getElementById("plateConditionKey").value;
  const op = document.getElementById("plateConditionOperator");
  const val = document.getElementById("plateConditionValue").value.trim();
  if (!val) return alert("Введите значение");
  plateConditions[key] = (op.style.display !== "none" ? op.value + ":" : "") + val;
  document.getElementById("plateConditions").value = JSON.stringify(plateConditions);
  updateConditionUI();
}

function updateConditionUI() {
  const container = document.getElementById("plateConditionList");
  container.innerHTML = "";
  Object.entries(plateConditions).forEach(([k, v]) => {
    const div = document.createElement("div");
    div.innerHTML = `${k}: ${v} <button onclick=\"removePlateCondition('${k}')\">Удалить</button>`;
    container.appendChild(div);
  });
}

function removePlateCondition(key) {
  delete plateConditions[key];
  document.getElementById("plateConditions").value = JSON.stringify(plateConditions);
  updateConditionUI();
}

function splitOnce(str, sep) {
  const idx = str.indexOf(sep);
  if (idx < 0) return [str, ""];
  return [str.slice(0, idx), str.slice(idx + 1)];
}

// --- Добавление языков ---
function addTitleLanguage(lang) {
  if (!lang) lang = prompt("Код языка (например en, tj):");
  const id = "title_" + lang;
  if (document.getElementById(id)) return;
  const container = document.getElementById("titleGroup");
  const div = document.createElement("div");
  div.innerHTML = `<label>Заголовок (${lang.toUpperCase()})</label><textarea rows="2" id="${id}"></textarea>`;
  container.insertBefore(div, container.querySelector(".add-btn"));
}

function addBodyLanguage(lang) {
  if (!lang) lang = prompt("Код языка (например en, tj):");
  const id = "body_" + lang;
  if (document.getElementById(id)) return;
  const container = document.getElementById("bodyGroup");
  const div = document.createElement("div");
  div.innerHTML = `<label>Сообщение (${lang.toUpperCase()})</label><textarea rows="2" id="${id}"></textarea>`;
  container.insertBefore(div, container.querySelector(".add-btn"));
}

function addActionTitleLanguage(lang) {
  if (!lang) lang = prompt("Код языка (например en, tj):");
  const id = "action_title_" + lang;
  if (document.getElementById(id)) return;
  const container = document.getElementById("actionLangGroup");
  const div = document.createElement("div");
  div.innerHTML = `<label>Текст кнопки (${lang.toUpperCase()})</label><input type="text" id="${id}">`;
  container.insertBefore(div, container.querySelector(".add-btn"));
}

document.getElementById("plateForm").addEventListener("submit", async function (e) {
  e.preventDefault();

  const plate = {
    plate_id: document.getElementById("plateId").value.trim(),
    action: document.getElementById("action").value.trim(),
    icon_url: document.getElementById("icon_url").value.trim(),
    with_close_btn: document.getElementById("withCloseBtn").checked,
    title: {},
    body: {},
    action_title: {},
    conditions: {}
  };

  document.querySelectorAll("[id^='title_']").forEach(el => {
    const lang = el.id.split("_")[1];
    plate.title[lang] = el.value;
  });

  document.querySelectorAll("[id^='body_']").forEach(el => {
    const lang = el.id.split("_")[1];
    plate.body[lang] = el.value;
  });

  document.querySelectorAll("[id^='action_title_']").forEach(el => {
    const lang = el.id.split("_")[2];
    plate.action_title[lang] = el.value;
  });

  try {
    plate.conditions = JSON.parse(document.getElementById("plateConditions").value);
  } catch (err) {
    alert("❌ Невалидный JSON в условиях");
    return;
  }

  const confirmed = confirm("Отправить Plate?\n" + JSON.stringify(plate, null, 2));
  if (!confirmed) return;

  const res = await authorizedFetch("/save_plate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(plate),
  });

  if (res.ok) {
    alert("✅ Plate успешно сохранён");
  } else {
    const err = await res.text();
    alert("❌ Ошибка: " + err);
  }
});

function loadPlateHistory() {
  const panel = document.getElementById("historyPanel");
  panel.innerHTML = "Загрузка...";

  authorizedFetch("/plate_history")
    .then(res => res.json())
    .then(list => {
      if (!list) return;
      panel.innerHTML = '';
      list.forEach(entry => {
        if (!window.lastPlatePayloads) window.lastPlatePayloads = {};
        window.lastPlatePayloads[entry.id] = entry;

        const card = document.createElement("div");
        card.className = "plate-card";
        card.style.marginTop = "20px";
        card.id = "plate_card_" + entry.id;

        const title = entry.title?.ru || "(без заголовка)";
        const body = entry.body?.ru || "";
        const actionTitle = entry.action_title?.ru || "Перейти";
        const actionUrl = entry.action || "#";
        const icon = entry.icon_url || "";

        card.innerHTML = `
  <div style="position:relative;">
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px;">
      <span style="font-weight: bold;">Язык:</span>
      <div style="display:flex; align-items:center; gap:6px;">
        <select id="lang_select_${entry.id}" onchange="updatePlateLanguage(${entry.id}, this.value)" style="width:auto; padding:6px; border-radius:6px;">
          <option value="ru" selected>RU</option>
          <option value="en">EN</option>
          <option value="tj">TJ</option>
        </select>
        <span onclick="deletePlate('${entry.plate_id?.replace(/'/g, "\\'") || ""}')" style="cursor:pointer; font-size:18px; color:#888;" title="Удалить">🗑️</span>
      </div>
    </div>

    <div style="display:flex; gap: 12px;">
      ${icon ? `<img src="${icon}" alt="icon" style="width:64px;height:64px;border-radius:8px;">` : ""}
      <div style="flex:1;">
        <div class="plate-title" style="font-weight:bold; font-size:18px;">${title}</div>
        <div class="plate-body" style="margin-top:6px;">${body}</div>
      </div>
    </div>

    <div class="plate-button" style="margin-top:10px; font-weight: 500;">${actionTitle}</div>

    <div class="history-item-buttons" style="margin-top:10px;">
      <button onclick="togglePayload('plate_payload_${entry.id}', this)">Показать JSON</button>
    </div>

    <pre id="plate_payload_${entry.id}" style="display:none; background:#f9f9f9; padding:8px; border-radius:6px;">${JSON.stringify(entry, null, 2)}</pre>

    <div class="history-item-buttons" style="margin-top:10px;">
      <button onclick='fillForm(${JSON.stringify({
        plate_id: entry.plate_id,
        title: entry.title,
        body: entry.body,
        action: entry.action,
        action_title: entry.action_title,
        icon_url: entry.icon_url,
        with_close_btn: entry.with_close_btn,
        conditions: entry.conditions
      })})'>Изменить</button>
    </div>
  </div>
`;

        panel.appendChild(card);
      });
    })
    .catch(e => {
      panel.innerHTML = "Ошибка загрузки истории";
    });
}

function togglePayload(id, btn) {
  const el = document.getElementById(id);
  if (!el) return;
  const visible = el.style.display === "block";
  el.style.display = visible ? "none" : "block";
  btn.textContent = visible ? "Показать JSON" : "Скрыть JSON";
}

window.loadPlateHistory = loadPlateHistory;

function fillForm(payload) {
  document.getElementById("plateId").value = payload.plate_id || "";
  document.getElementById("action").value = payload.action || "";
  document.getElementById("icon_url").value = payload.icon_url || "";
  document.getElementById("withCloseBtn").checked = !!payload.with_close_btn;

  for (const [lang, val] of Object.entries(payload.title || {})) {
    const id = "title_" + lang;
    if (!document.getElementById(id)) addTitleLanguage(lang);
    document.getElementById(id).value = val;
  }

  for (const [lang, val] of Object.entries(payload.body || {})) {
    const id = "body_" + lang;
    if (!document.getElementById(id)) addBodyLanguage(lang);
    document.getElementById(id).value = val;
  }

  for (const [lang, val] of Object.entries(payload.action_title || {})) {
    const id = "action_title_" + lang;
    if (!document.getElementById(id)) addActionTitleLanguage(lang);
    document.getElementById(id).value = val;
  }

  try {
    document.getElementById("plateConditions").value = JSON.stringify(payload.conditions || {});
    Object.assign(plateConditions, payload.conditions || {});
    updateConditionUI();
  } catch {}
}
window.fillForm = fillForm;

async function deletePlate(plateId) {
  if (!confirm("Удалить Plate " + plateId + "?")) return;
  const res = await authorizedFetch("/delete_plate", {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ plate_id: plateId })
  });
  if (res.ok) {
    loadPlateHistory();
  } else {
    const err = await res.text();
    alert("Ошибка удаления: " + err);
  }
}

function updatePlateLanguage(id, lang) {
  const payload = window.lastPlatePayloads?.[id];
  if (!payload) return;

  const title = payload.title?.[lang] || "(без заголовка)";
  const body = payload.body?.[lang] || "";
  const btn = payload.action_title?.[lang] || "Перейти";

  const card = document.getElementById(`plate_card_${id}`);
  if (!card) return;

  card.querySelector(".plate-title").textContent = title;
  card.querySelector(".plate-body").textContent = body;
  card.querySelector(".plate-button").textContent = btn;
}
