async function authorizedFetch(url, options = {}) {
    let user = localStorage.getItem("authUser");
    let pass = localStorage.getItem("authPass");

    if (!user || !pass || user === "null" || pass === "null") {
        user = prompt("Введите логин:");
        pass = prompt("Введите пароль:");
        if (!user || !pass) {
            alert("⛔ Авторизация обязательна.");
            throw new Error("Authorization required");
        }
        localStorage.setItem("authUser", user);
        localStorage.setItem("authPass", pass);
    }

    const headers = {
        ...options.headers,
        "Authorization": "Basic " + btoa(`${user}:${pass}`)
    };

    const response = await fetch(url, { ...options, headers });

    if (response.status === 401) {
        localStorage.removeItem("authUser");
        localStorage.removeItem("authPass");
        alert("⛔ Неверный логин или пароль. Повторите авторизацию.");
        location.reload();
        throw new Error("Unauthorized");
    }

    return response;
}

document.addEventListener("DOMContentLoaded", () => {
    const user = localStorage.getItem("authUser");
    const pass = localStorage.getItem("authPass");

    if (!user || !pass || user === "null" || pass === "null") {
        const username = prompt("Введите логин:");
        const password = prompt("Введите пароль:");
        if (username && password) {
            localStorage.setItem("authUser", username);
            localStorage.setItem("authPass", password);
        } else {
            alert("Авторизация обязательна");
            location.reload();
            return;
        }
    }

    document.getElementById("pageContent").style.display = "";
});