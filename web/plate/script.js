const plateConditions = {};
/**
 * Renders a plate preview card HTML.
 * @param {Object} entry — plate object with fields: title.ru, body.ru, action, icon_url, etc.
 * @returns {string} HTML string for preview.
 */
function renderPlateCard(entry) {
    const title = entry.title?.ru || "(без заголовка)";
    const body = entry.body?.ru || "";
    const actionTitle = entry.action_title?.ru || "Перейти";
    const actionUrl = entry.action || "#";
    const icon = entry.icon_url || "";
    return `
    <div class="plate-wrapper">
      ${icon ? `<img src="${icon}" style="width:100%; height:120px; margin-bottom:8px; border-radius:100px; object-fit:cover;">` : ""}
      <div class="plate-title">${title}</div>
      <div class="plate-body">${body}</div>
      <div class="plate-button-row">
        <button onclick="openInModal('${actionUrl}')" class="plate-button plate-button-go">${actionTitle}</button>
      </div>
    </div>`;
}
const availableConditions = {};
const platesIsActiveMap = new Map();

async function listPlateHistory() {
    // Fetch history and ensure we get an array
    let list = await fetch("/plate_history")
        .then(r => r.json())
        .catch(err => {
            console.error("listPlateHistory fetch error:", err);
            return [];
        });
    if (!Array.isArray(list)) {
        console.warn("listPlateHistory: expected array but got", list);
        list = [];
    }
    platesIsActiveMap.clear();
    list.forEach(plate => {
        platesIsActiveMap.set(plate.plate_id, plate.is_active);
    });
    return list;
}

document.addEventListener("DOMContentLoaded", () => {
    loadConditionDefaults().then();
    loadActionTypes().then();
    listPlateHistory().then();
});

async function loadActionTypes() {
    const res = await fetch('/static/actions.txt');
    const text = await res.text();
    const lines = text.split('\n');
    const select = document.getElementById("action_type_select");
    select.innerHTML = "";

    lines.forEach(line => {
        const trimmed = line.trim();
        if (!trimmed) return;
        const [label, value] = trimmed.split("|").map(s => s.trim());
        if (!label || !value) return;
        const option = new Option(label, value);
        select.appendChild(option);
    });
}

async function toggleActive(plateId, isActive) {
    const payload = {
        plate_id: plateId, is_active: isActive
    };

    try {
        const res = await fetch('/save_plate', {
            method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(payload)
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        platesIsActiveMap.set(plateId, isActive);
        fetchPlateHistory().then();
        showToast(`Плашка ${isActive ? 'включена' : 'отключена'}`);
    } catch (err) {
        console.error('Ошибка при переключении активности:', err);
    }
}

function showToast(message) {
    const toast = document.createElement('div');
    toast.textContent = message;
    Object.assign(toast.style, {
        position: 'fixed',
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        background: 'rgba(0, 0, 0, 0.9)',
        color: '#fff',
        padding: '12px 24px',
        fontSize: '16px',
        fontWeight: 'bold',
        borderRadius: '8px',
        boxShadow: '0 2px 10px rgba(0, 0, 0, 0.3)',
        zIndex: 9999
    });
    document.body.appendChild(toast);
    setTimeout(() => {
        toast.remove();
    }, 2000);
}

async function loadConditionDefaults() {
    const res = await fetch('/static/conditions.txt');
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

        availableConditions[key] = {operators, defaultVal};

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
/**
 * Вставляет в указанный контейнер поле с кодом языка.
 *
 * @param {Object} opts
 * @param {string} opts.containerId  — id контейнера, куда вставлять
 * @param {string} opts.fieldPrefix  — префикс для id нового поля (например "title" или "body")
 * @param {string} opts.labelText    — текст лейбла перед полем
 * @param {string} [opts.tag="textarea"] — HTML-тег поля ("textarea" или "input")
 * @param {string} [opts.attrs=""]   — доп. атрибуты, например 'rows="2"' или 'type="text"'
 * @param {string} [langArg]         — код языка; если не передан, спросятся через prompt
 */
function addLangField({containerId, fieldPrefix, labelText, tag = "textarea", attrs = ""}, langArg) {
    let lang = (langArg || "").trim();
    if (!lang) {
        lang = prompt("Код языка (например en, tj):")?.trim();
        if (!lang) return;
    }

    const id = `${fieldPrefix}_${lang}`;
    if (document.getElementById(id)) return; // уже есть

    const container = document.getElementById(containerId);
    if (!container) {
        console.error(`Контейнер с id="${containerId}" не найден`);
        return;
    }

    const wrapper = document.createElement("div");
    wrapper.innerHTML = `
    <label>${labelText} (${lang.toUpperCase()})</label>
    <${tag} ${attrs} id="${id}"></${tag}>
  `.trim();

    const addBtn = container.querySelector(".add-btn");
    container.insertBefore(wrapper, addBtn);
}

function addTitleLanguage(lang) {
    addLangField({
        containerId: "titleGroup", fieldPrefix: "title", labelText: "Заголовок", tag: "textarea", attrs: 'rows="2"'
    }, lang);
}

function addBodyLanguage(lang) {
    addLangField({
        containerId: "bodyGroup", fieldPrefix: "body", labelText: "Сообщение", tag: "textarea", attrs: 'rows="2"'
    }, lang);
}

function addActionTitleLanguage(lang) {
    addLangField({
        containerId: "actionLangGroup",
        fieldPrefix: "action_title",
        labelText: "Текст кнопки",
        tag: "input",
        attrs: 'type="text"'
    }, lang);
}

document.getElementById("plateForm").addEventListener("submit", async function (e) {
    e.preventDefault();

    const plateId = document.getElementById("plateId").value.trim();

    const plate = {
        plate_id: plateId,
        action: {
            type: document.getElementById("action_type_select").value,
            value: document.getElementById("action").value.trim()
        },
        icon_url: document.getElementById("icon_url").value.trim(),
        with_close_btn: document.getElementById("with_close_btn").checked,
        with_reopen_btn: document.getElementById("with_reopen_button").checked,
        should_hide_after_click: document.getElementById("should_hide_after_click").checked,
        title: {},
        body: {},
        action_title: {},
        conditions: {},
        is_active: platesIsActiveMap.get(plateId) ?? true,
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

    // Show preview modal instead of confirm dialog
    let modal = document.getElementById("submissionModal");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "submissionModal";
        Object.assign(modal.style, {
            position: "fixed",
            top: 0,
            left: 0,
            width: "100%",
            height: "100%",
            background: "rgba(0,0,0,0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 10000
        });
        modal.innerHTML = `
          <div class="submission-modal-content">
            <div id="submissionModalContent"></div>
            <div style="text-align:right; margin-top:12px;">
              <button id="submissionModalSave">Сохранить</button>
              <button id="submissionModalCancel">Отмена</button>
            </div>
          </div>`;
        document.body.appendChild(modal);
    }
    // Populate preview and show modal
    document.getElementById("submissionModalContent").innerHTML = renderPlateCard(plate);
    modal.style.display = "flex";

    // Handle modal buttons
    document.getElementById("submissionModalCancel").onclick = () => {
        modal.style.display = "none";
    };
    document.getElementById("submissionModalSave").onclick = async () => {
        modal.style.display = "none";
        try {
            const res = await fetch("/save_plate", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(plate),
            });
            if (res.ok) {
                showToast(platesIsActiveMap.has(plate.plate_id)
                    ? "🔄 Плашка успешно обновлена"
                    : "✅ Плашка успешно сохранена");
                if (isPlateHistoryVisible) await fetchPlateHistory();
            } else {
                const err = await res.text();
                alert("❌ Ошибка: " + err);
            }
        } catch (err) {
            console.error("Error saving plate:", err);
            alert("❌ Ошибка при сохранении");
        }
    };
});

function openInModal(url) {
    const modal = document.getElementById("modal");
    const frame = document.getElementById("modalFrame");
    frame.src = url;
    modal.style.display = "block";
}

let isPlateHistoryVisible = false;

async function fetchPlateHistory() {
    const panel = document.getElementById("historyPanel");
    panel.innerHTML = "Загрузка...";

    try {
        const list = await listPlateHistory()
        panel.innerHTML = '';
        list.forEach(entry => {
            if (!window.lastPlatePayloads) window.lastPlatePayloads = {};
            window.lastPlatePayloads[entry.id] = entry;

            const card = document.createElement("div");
            card.className = "plate-card";
            card.style.marginTop = "20px";
            card.id = "plate_card_" + entry.id;

            const isActive = entry.is_active ?? false;
            const plateId = entry.plate_id?.replace(/'/g, "\\'") || ""

            const title = entry.title?.ru || "(без заголовка)";
            const body = entry.body?.ru || "";
            const actionTitle = entry.action_title?.ru || "Перейти";
            const actionUrl = entry.action || "#";
            const icon = entry.icon_url || "";

            card.innerHTML = `
    <div class="plate-wrapper">
        <div class="plate-header-top-row">
        <div class="plate-active-status">
            <span class="plate-active-indicator ${isActive ? 'active' : 'inactive'}"></span>
            <span class="plate-active-text">${isActive ? 'Активна' : 'Неактивна'}</span>
        </div> 
        <select id="lang_select_${entry.id}" class="plate-lang-select" onchange="updatePlateLanguage(${entry.id}, this.value)">
            <option value="ru" selected>RU</option>
            <option value="en">EN</option>
            <option value="tj">TJ</option>
        </select>
    </div>
    
    <div class="plate-preview">        
      <div class="plate-content-row">
        <div class="plate-text-block">
          ${icon ? `<img src="${icon}" alt="icon" class="plate-icon" style="width:100%; height:220px; margin-bottom:8px; border-radius:100px; object-fit:cover;">` : ""}
          <div class="plate-title">${title}</div>
          <div class="plate-body">${body}</div>
          <div class="plate-button-row">
            <button onclick="openInModal('${actionUrl}')"" class="plate-button plate-button-go">${actionTitle}</button>
          </div>
        </div>
        <div id="modal" style="display:none; position:fixed; top:5%; left:10%; width:80%; height:90%; background:white; z-index:9999; border-radius:8px; box-shadow:0 0 10px rgba(0,0,0,0.5);">
            <div style="text-align:right; padding:10px;">
            <button onclick="document.getElementById('modal').style.display='none'">✖</button>
  </div>
  <iframe id="modalFrame" src="" style="width:100%; height:90%; border:none;"></iframe>
</div>
      </div>
    </div>
    <div class="plate-actions">
      <div class="plate-actions-left">
        <button onclick="togglePayload('plate_payload_${entry.id}', this)">Показать JSON</button>
        <button onclick='fillForm(${JSON.stringify({
                plate_id: entry.plate_id,
                title: entry.title,
                body: entry.body,
                action: entry.action,
                action_title: entry.action_title,
                icon_url: entry.icon_url,
                with_close_btn: entry.with_close_btn,
                with_reopen_btn: entry.with_reopen_btn,
                conditions: entry.conditions,
                should_hide_after_click: entry.should_hide_after_click,
            })})'>Изменить</button>
      </div>
      <div class="plate-actions-right">
        <label class="switch">
          <input type="checkbox" onchange="toggleActive('${plateId}', this.checked)" ${isActive ? "checked" : ""}>
          <span class="slider round"></span>
        </label>
        <span onclick="deletePlate('${plateId}')" class="delete-icon" title="Удалить">✖</span>
      </div>
    </div>
    <pre id="plate_payload_${entry.id}" class="plate-json">${JSON.stringify(entry, null, 2)}</pre>
  </div>
`;
            panel.appendChild(card);
        });
        isPlateHistoryVisible = true;
    } catch (e) {
        console.error("Ошибка загрузки истории:", e);
        panel.innerHTML = "Ошибка загрузки истории";
    }
}

function loadPlateHistory() {
    const panel = document.getElementById("historyPanel");
    if (isPlateHistoryVisible) {
        panel.innerHTML = "";
        isPlateHistoryVisible = false;
    } else {
        fetchPlateHistory().then();
    }
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
    if (payload.action?.type) {
        const actionSelect = document.getElementById("action_type_select");
        if (actionSelect) actionSelect.value = payload.action.type;
    }
    document.getElementById("action").value = payload.action?.value || "";
    document.getElementById("icon_url").value = payload.icon_url || "";
    document.getElementById("with_close_btn").checked = !!payload.with_close_btn;
    document.getElementById("with_reopen_button").checked = !!payload.with_reopen_btn;
    document.getElementById("should_hide_after_click").checked = !!payload.should_hide_after_click;

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
    } catch {
    }
}

window.fillForm = fillForm;

async function deletePlate(plateId) {
    if (!confirm("Удалить Plate " + plateId + "?")) return;
    const res = await fetch("/delete_plate", {
        method: "DELETE", headers: {"Content-Type": "application/json"}, body: JSON.stringify({plate_id: plateId})
    });

    if (res.ok) {
        showToast("🗑️ Плашка успешно удалена");
        await fetchPlateHistory();

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

    const titleEl = card.querySelector(".plate-title");
    if (titleEl) {
        const select = titleEl.querySelector("select");
        titleEl.innerHTML = title;
        if (select) titleEl.appendChild(select);
    }
    card.querySelector(".plate-body").textContent = body;
    card.querySelector(".plate-button").textContent = btn;
}
