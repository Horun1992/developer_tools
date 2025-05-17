const appVersions = [];

async function loadVersionHistoryFromServer() {
    const res = await fetch("/version_history");
    const history = await res.json();
    const container = document.getElementById("versionHistoryForm");
    container.innerHTML = "";

    history.forEach(ver => {
        const checkbox = document.createElement("input");
        checkbox.type = "checkbox";
        checkbox.value = ver;

        checkbox.onchange = () => {
            if (checkbox.checked && !appVersions.includes(ver)) {
                appVersions.push(ver);
                updateAppVersionUI();
                updatePushConditions();
            } else if (!checkbox.checked) {
                const i = appVersions.indexOf(ver);
                if (i !== -1) appVersions.splice(i, 1);
                updateAppVersionUI();
                updatePushConditions();
            }
        };

        const label = document.createElement("label");
        label.style.marginRight = "10px";
        label.style.display = "flex";
        label.style.alignItems = "center";
        label.style.gap = "6px";

        label.appendChild(checkbox);
        label.appendChild(document.createTextNode(ver));
        container.appendChild(label);
    });

    updateAppVersionUI();
}

function toggleAllVersions(masterCheckbox) {
    const form = document.getElementById("versionHistoryForm");
    const checkboxes = form.querySelectorAll("input[type='checkbox']");

    appVersions.length = 0; // очищаем
    checkboxes.forEach(cb => {
        cb.checked = masterCheckbox.checked;
        if (masterCheckbox.checked) {
            appVersions.push(cb.value);
        }
    });

    updateAppVersionUI();
    updatePushConditions();
}

async function saveVersionToServer(ver) {
    await fetch("/version_history", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({version: ver})
    });
}

function openVersionPrompt() {
    const input = prompt("Введите версию (например: 2.1.6, 2.1.6.1):");
    if (!input) return;

    const parts = input.split(",").map(v => v.trim()).filter(Boolean);
    parts.forEach(p => {
        if (!appVersions.includes(p)) {
            appVersions.push(p);
            saveVersionToServer(p).then(() => {
                addVersionCheckbox(p, true);
            });
        }
    });

    updateAppVersionUI();
    updatePushConditions();
}

function addVersionCheckbox(ver, checked = false) {
    const container = document.getElementById("versionHistoryForm");

    const label = document.createElement("label");
    label.style.marginRight = "10px";
    label.style.display = "flex";
    label.style.alignItems = "center";
    label.style.gap = "6px";

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.value = ver;
    checkbox.checked = checked;

    checkbox.onchange = () => {
        if (checkbox.checked && !appVersions.includes(ver)) {
            appVersions.push(ver);
        } else if (!checkbox.checked) {
            const i = appVersions.indexOf(ver);
            if (i !== -1) appVersions.splice(i, 1);
        }
        updateAppVersionUI();
        updatePushConditions();
    };

    label.appendChild(checkbox);
    label.appendChild(document.createTextNode(ver));
    container.appendChild(label);

    updateAppVersionUI();
    updatePushConditions();
}

async function removeVersionPrompt() {
    const input = prompt("Введите точную версию для удаления:");
    if (!input) return;

    // Удаляем чекбокс из UI
    const container = document.getElementById("versionHistoryForm");
    const checkboxes = container.querySelectorAll("input[type='checkbox']");

    checkboxes.forEach(cb => {
        if (cb.value === input) {
            const label = cb.parentNode;
            label.remove(); // удаляет и чекбокс, и подпись
        }
    });

    const i = appVersions.indexOf(input);
    if (i !== -1) appVersions.splice(i, 1);

    // Удаляем версию на сервере, если нужно
    try {
        await fetch("/version_history", {
            method: "DELETE",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({version: input})
        });
    } catch (e) {
        // Можно добавить обработку ошибок, если нужно
        console.warn("Ошибка при удалении версии на сервере:", e);
    }

    updateAppVersionUI();
    updatePushConditions();
}

function updateAppVersionUI() {
    const summary = document.getElementById("versionSummary");
    const container = document.getElementById("appVersionList");
    container.innerHTML = "";

    if (appVersions.length === 0) {
        summary.innerHTML = "<span style='color:#c0392b;'>Push получат все версии</span>";
        return;
    }

    summary.innerHTML = `<span style='color:#2c3e50;'>Push получат: ${appVersions.join(", ")}</span>`;
}

function updatePushConditions() {
    const map = {};
    if (appVersions.length > 0) {
        map["version"] = appVersions.join(",");
    }
    document.getElementById("pushConditions").value = JSON.stringify(map);
}

function toggleHistory() {
    const panel = document.getElementById("historyPanel");
    panel.style.display = panel.style.display === "block" ? "none" : "block";
    if (panel.style.display === "block") loadHistory();
}

function togglePayload(id, btn) {
    const el = document.getElementById(id);
    const shown = el.style.display !== "none";
    el.style.display = shown ? "none" : "block";
    btn.textContent = shown ? "Показать payload" : "Скрыть payload";
}

function fillForm(payload) {
    // Тип сборки
    document.getElementById("buildType").value = payload.build_type || "debug";

    // Заголовки
    try {
        const titleMap = JSON.parse(payload.title || "{}");
        Object.entries(titleMap).forEach(([lang, text]) => {
            const id = "title_" + lang;
            let el = document.getElementById(id);
            if (!el) {
                addTitleLanguage(lang);
                el = document.getElementById(id);
            }
            el.value = text;
        });
    } catch (e) {
        console.warn("⚠️ Ошибка парсинга title:", e);
    }

    // Сообщения
    try {
        const bodyMap = JSON.parse(payload.body || "{}");
        Object.entries(bodyMap).forEach(([lang, text]) => {
            const id = "body_" + lang;
            let el = document.getElementById(id);
            if (!el) {
                addBodyLanguage(lang);
                el = document.getElementById(id);
            }
            el.value = text;
        });
    } catch (e) {
        console.warn("⚠️ Ошибка парсинга body:", e);
    }

    // Остальные поля
    const data = payload.data || {};
    document.getElementById("image").value = data.image || "";
    document.getElementById("includeSound").checked = (data.sound || "") !== "none";
    document.getElementById("priority").value = data.priority || "high";

    // Версии (из conditions)
    const cond = data.conditions ? JSON.parse(data.conditions) : {};
    const versionStr = cond.version || "";
    const versions = versionStr.split(",").map(v => v.trim()).filter(Boolean);

    // Сброс старых
    appVersions.length = 0;
    const allCheckboxes = document.querySelectorAll("#versionHistoryForm input[type='checkbox']");
    allCheckboxes.forEach(cb => (cb.checked = false));

    versions.forEach(ver => {
        const match = Array.from(allCheckboxes).find(cb => cb.value === ver);
        if (match) match.checked = true;
        else addVersionCheckbox(ver, true); // если версии нет — добавим
        appVersions.push(ver);
    });

    updateAppVersionUI();
    updatePushConditions();
}

async function loadHistory() {
    const panel = document.getElementById("historyPanel");
    panel.innerHTML = "Загрузка...";
    try {
        const res = await fetch("/push_history.json");
        const data = await res.json();
        if (!Array.isArray(data)) throw new Error("Некорректный формат истории");

        const limited = data.slice(-30).reverse();
        panel.innerHTML = '';
        const clearBtn = document.createElement("button");
        clearBtn.textContent = "Очистить";
        clearBtn.onclick = async () => {
            if (confirm("Очистить историю?")) {
                await fetch("/push_history.json", {
                    method: "DELETE"
                });
                toggleHistory();
            }
        };
        panel.appendChild(clearBtn);

        limited.forEach(entry => {
            const status = entry.success ? "✅" : "❌";
            const item = document.createElement("div");
            const payloadJson = JSON.stringify(entry.payload, null, 2);
            const safeTopic = (entry.topic || "x").replace(/[^a-z0-9_]/gi, "_");
            const payloadId = `payload_${entry.timestamp}_${safeTopic}`;

            item.innerHTML = `
        <div>
            <div style="margin-top: 16px;">
                <strong>${status} ${new Date(entry.timestamp * 1000).toLocaleString()}</strong>
            </div>
            <span style="color: #666;">Topic: ${entry.topic || "-"}</span><br>
            <span style="color: #666;">Title: ${entry.title || "-"}</span><br>
            <span style="color: #666;">Message ID: ${entry.message_id || "-"}</span><br>

            <div style="margin-top: 6px; display: flex; gap: 10px; flex-wrap: wrap;">
                <button onclick="togglePayload('${payloadId}', this)" style="flex: 1;">Показать payload</button>
                <button onclick='fillForm(${JSON.stringify(entry.payload)})' style="flex: 1;">Повторить</button>
            </div>

            <pre id="${payloadId}" style="display: none; background:#f9f9f9; padding: 8px; border-radius: 6px;">${payloadJson}</pre>
        </div>
       `;
            panel.appendChild(item);
        });
    } catch (e) {
        panel.innerHTML = "Ошибка загрузки истории.";
    }
}

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

function markInvalid(id, message) {
    const el = document.getElementById(id);
    el.classList.add("invalid");
    if (!window.firstInvalid) {
        window.firstInvalid = el;
    }
    console.warn("⚠️ " + message);
}

document.getElementById("pushForm").addEventListener("submit", async function (e) {
    e.preventDefault();
    window.firstInvalid = null;

    const ruTitle = document.getElementById("title_ru").value.trim();
    const ruBody = document.getElementById("body_ru").value.trim();

    document.getElementById("title_ru").classList.remove("invalid");
    document.getElementById("body_ru").classList.remove("invalid");

    if (!ruTitle) markInvalid("title_ru", "Заголовок (RU) обязателен");
    if (!ruBody) markInvalid("body_ru", "Сообщение (RU) обязательно");

    if (window.firstInvalid) {
        alert("❌ Пожалуйста, заполните обязательные поля");
        window.firstInvalid.focus();
        return;
    }

    const title = {};
    const body = {};

    document.querySelectorAll("[id^='title_']").forEach(el => {
        const lang = el.id.split("_")[1];
        title[lang] = el.value;
    });

    document.querySelectorAll("[id^='body_']").forEach(el => {
        const lang = el.id.split("_")[1];
        body[lang] = el.value;
    });

    const buildType = document.getElementById("buildType").value;
    const image = document.getElementById("image").value;
    const sound = document.getElementById("includeSound").checked ? "default" : "none";
    const pushConditions = document.getElementById("pushConditions").value;
    const priority = document.getElementById("priority").value;

    const data = {
        image,
        sound,
        priority
    };

    if (pushConditions) {
        data["conditions"] = pushConditions;
    }

    const payload = {
        build_type: buildType,
        title: JSON.stringify(title),
        body: JSON.stringify(body),
        data
    };

    const confirmed = confirm("Отправить push?\n" + JSON.stringify(payload, null, 2));
    if (!confirmed) return;

    try {
        await fetch("/send_push", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(payload),
        }).then(async (res) => {
            const result = await res.json();
            alert(res.ok ? `✅ Успешно: ${result.message}` : `❌ Ошибка: ${JSON.stringify(result)}`);
            if (document.getElementById("historyPanel").style.display === "block") await loadHistory();
        });
    } catch (err) {
        alert("❌ Ошибка сети: " + err.message);
    }
});

document.addEventListener("DOMContentLoaded", () => {
    loadVersionHistoryFromServer();
});
