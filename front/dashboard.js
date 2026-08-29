function esc(v) {
  return String(v ?? "").replace(
    /[&<>"']/g,
    (m) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#039;",
      })[m],
  );
}
function statusClass(s) {
  return s === "Новая"
    ? "new"
    : s === "Подтверждена"
      ? "confirm"
      : s === "Завершена"
        ? "completed"
        : s === "Отменена"
          ? "cancelled"
          : "";
}
async function loadNotifications() {
  try {
    const r = await fetch("/api/admin/notifications", {
      credentials: "same-origin",
    });
    if (!r.ok) throw new Error();
    const items = await r.json();
    notificationsList.innerHTML = items.length
      ? items
          .map(
            (n) =>
              `<a class="notification-item" href="${esc(n.link || "requests.html")}"><span class="notification-dot"></span><div><strong>${esc(n.title)}</strong><p>${esc(n.text)}</p></div></a>`,
          )
          .join("")
      : "<div class='empty'>Новых уведомлений нет.</div>";
  } catch (e) {
    notificationsList.innerHTML =
      "<div class='error'>Не удалось загрузить уведомления.</div>";
  }
}

async function loadDashboard() {
  try {
    const r = await fetch("/api/dashboard", { credentials: "same-origin" });
    if (!r.ok) throw new Error("dashboard");
    const d = await r.json();
    newToday.textContent = d.newToday;
    confirmedToday.textContent = d.confirmedToday;
    cancelledToday.textContent = d.cancelledToday;
    completedToday.textContent = d.completedToday;
    appointmentsToday.textContent = d.appointmentsToday;
    doctorsToday.textContent = `${d.busyDoctors}/${d.activeDoctors}`;
    doctorHint.textContent = `заняты / активны · свободны: ${d.freeDoctors}`;
    todayAppointments.innerHTML = d.todayAppointments?.length
      ? d.todayAppointments
          .map(
            (a) =>
              `<div class="appointment-row"><span class="time">${esc(a.time || "—")}</span><span>${esc(a.patient)}</span><span>${esc(a.service)}</span><span>${esc(a.doctor)}</span><span class="status ${statusClass(a.status)}">${esc(a.status)}</span></div>`,
          )
          .join("")
      : "<div class='empty'>На сегодня записей нет.</div>";
    lastRequests.innerHTML = d.lastRequests?.length
      ? d.lastRequests
          .map((x) => {
            const t = x.created_at
              ? new Date(x.created_at).toLocaleTimeString("ru-RU", {
                  hour: "2-digit",
                  minute: "2-digit",
                })
              : "";
            return `<div class="request-item"><div class="request-top"><span class="request-name">${esc(x.name)}</span><span class="request-time">${esc(t)}</span></div><div class="request-service">${esc(x.services || "Без услуг")}</div><div class="request-meta"><span>${esc(x.phone)}</span><span>${esc(x.status || "")}</span></div></div>`;
          })
          .join("")
      : "<div class='empty'>Заявок пока нет.</div>";
  } catch (e) {
    document.getElementById("todayAppointments").innerHTML =
      '<div class="error">Не удалось загрузить данные Dashboard.</div>';
  }
}
document.addEventListener("DOMContentLoaded", () => {
  loadDashboard();
  loadNotifications();
});
