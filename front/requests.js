let requests = [];
let currentRequestID = null;
let currentFilter = "Все";
let currentSearch = "";
const drawer = document.querySelector(".request-drawer");
const overlay = document.querySelector(".overlay");
const closeBtn = document.querySelector(".close-btn");
const saveBtn = document.querySelector(".save-btn");
async function loadRequests() {
    const response = await fetch("/api/requests");
    requests = await response.json();
    renderRequests();
}
function renderRequests() {
    const table = document.querySelector(".requests-table");
    let filtered = requests;
if (currentFilter !== "Все") {
    filtered = filtered.filter(request =>
        request.status === currentFilter
    );
}
if (currentSearch !== "") {
    const text = currentSearch.toLowerCase();
    filtered = filtered.filter(request =>
        request.name.toLowerCase().includes(text) ||
        request.phone.includes(text) ||
        String(request.id).includes(text)
    );
}
    table.innerHTML = `
        <div class="table-header">
            <span>№</span>
            <span>Дата</span>
            <span>Имя</span>
            <span>Телефон</span>
            <span>Статус</span>
        </div>
        <div class="table-line"></div>
    `;
    filtered.forEach(request => {
        let statusClass = "new";
        if (request.status === "Подтверждена")
            statusClass = "confirm";
        if (request.status === "Завершена")
            statusClass = "completed";
        if (request.status === "Отменена")
            statusClass = "cancelled";
        table.innerHTML += `
            <div class="request-row" data-id="${request.id}">
                <span>${request.id}</span>
                <span>${new Date(request.created_at).toLocaleDateString("ru-RU")}</span>
                <span>${request.name}</span>
                <span>${request.phone}</span>
                <span class="status ${statusClass}">
                    ${request.status}
                </span>
            </div>
            <div class="table-line"></div>
        `;
    });
    bindRows();
}
function bindRows() {
    document.querySelectorAll(".request-row").forEach(row => {
        row.addEventListener("click", () => {
            const id = row.dataset.id;
            openRequest(id);
        });
    });
}
document.addEventListener("DOMContentLoaded", () => {
    loadRequests();
});
async function openRequest(id) {
    const response = await fetch(`/api/requests/${id}`);
    const request = await response.json();
    currentRequestID = request.id;
    document.querySelector(
        ".request-drawer h2"
    ).textContent = `Заявка №${request.id}`;
    const drawerBlocks = document.querySelectorAll(".drawer-value");
    drawerBlocks[0].textContent =
        new Date(request.created_at).toLocaleString("ru-RU");
    drawerBlocks[1].textContent =
        request.name;
    drawerBlocks[2].textContent =
        request.phone;
    drawerBlocks[3].textContent =
        request.comment || "Без комментариев";
document.querySelector(".price-value").textContent =
    request.price
        ? `≈ ${request.price.toLocaleString("ru-RU")} ₽`
        : "≈ —";
    const servicesBlock =
        document.querySelector(".services-list");
servicesBlock.innerHTML = "";
if (request.services) {
    request.services.split(", ").forEach(service => {
        servicesBlock.innerHTML += `
            <p>✔ ${service}</p>
        `;
    });
}
    const radios =
        document.querySelectorAll(".status-options input");
    radios.forEach(r => r.checked = false);
    switch (request.status) {
        case "Новая":
            radios[0].checked = true;
            break;
        case "Подтверждена":
            radios[1].checked = true;
            break;
        case "Завершена":
            radios[2].checked = true;
            break;
        case "Отменена":
            radios[3].checked = true;
            break;
    }
    drawer.classList.add("active");
    overlay.classList.add("active");
}
function closeDrawer(){
    drawer.classList.remove("active");
    overlay.classList.remove("active");
}
closeBtn.addEventListener(
    "click",
    closeDrawer
);
overlay.addEventListener(
    "click",
    closeDrawer
);
document.addEventListener(
    "keydown",
    (event)=>{
        if(event.key==="Escape"){
            closeDrawer();
        }
    }
);
saveBtn.addEventListener("click", async () => {
    if (!currentRequestID) return;
    const radios = document.querySelectorAll(".status-options input");
    let status = "Новая";
    if (radios[1].checked)
        status = "Подтверждена";
    if (radios[2].checked)
        status = "Завершена";
    if (radios[3].checked)
        status = "Отменена";
    const response = await fetch(`/api/requests/${currentRequestID}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            status: status
        })
    });
    if (response.ok) {
        alert("Статус сохранён");
        closeDrawer();
        loadRequests();
    } else {
        alert("Ошибка сохранения");
    }
});
document.querySelector(".filter-all").addEventListener("click", () => {
    currentFilter = "Все";
    renderRequests();
});
document.querySelector(".filter-new").addEventListener("click", () => {
    currentFilter = "Новая";
    renderRequests();
});
document.querySelector(".filter-confirm").addEventListener("click", () => {
    currentFilter = "Подтверждена";
    renderRequests();
});
document.querySelector(".filter-done").addEventListener("click", () => {
    currentFilter = "Завершена";
    renderRequests();
});
document.querySelector(".filter-cancel").addEventListener("click", () => {
    currentFilter = "Отменена";
    renderRequests();
});
document
.querySelectorAll(".status-filter button")
.forEach(button => {
    button.addEventListener("click", () => {
        document
        .querySelectorAll(".status-filter button")
        .forEach(btn => btn.classList.remove("active"));
        button.classList.add("active");
    });
});
const searchInput =
    document.querySelector(".search-input");
searchInput.addEventListener("input", () => {
    currentSearch =
        searchInput.value.trim();
    renderRequests();
});