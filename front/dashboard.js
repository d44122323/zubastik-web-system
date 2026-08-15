if(localStorage.getItem("adminAuth") !== "true"){
    window.location.href="login.html";
}
let dashboard = {};
async function loadDashboard() {
    const response = await fetch("/api/dashboard");
    dashboard = await response.json();
    renderDashboard();
}
function renderDashboard() {
    const todayRows =
        document.querySelectorAll(".today-row");
    todayRows[0].children[1].textContent =
        dashboard.newToday;
    todayRows[1].children[1].textContent =
        dashboard.confirmedToday;
    todayRows[2].children[1].textContent =
        dashboard.cancelledToday;
    const requestsSection =
        document.querySelector(".requests");
    requestsSection.innerHTML = `
        <h2>Последние заявки</h2>
    `;
    dashboard.lastRequests.forEach(request => {
        requestsSection.innerHTML += `
            <div class="request">
                <span>${request.name}</span>
                <span>${request.services || "Без услуг"}</span>
                <span>${request.phone}</span>
                <span>
                    ${new Date(request.created_at)
                        .toLocaleTimeString("ru-RU", {
                            hour: "2-digit",
                            minute: "2-digit"
                        })}
                </span>
            </div>
            <div class="divider small-divider"></div>
        `;
    });
}
document.addEventListener("DOMContentLoaded", () => {
    loadDashboard();
});