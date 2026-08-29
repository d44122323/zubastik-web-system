const esc = (x) =>
  String(x ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
function localDate() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}
async function init() {
  const me = await fetch("/api/doctor/me", { credentials: "include" });
  if (!me.ok) {
    location = "/?account=doctor";
    return;
  }
  const d = await me.json();
  document.getElementById("doctorName").textContent = d.name;
  document.getElementById("date").value = localDate();
  await loadSchedule();
  loadDay();
}
async function loadSchedule() {
  const r = await fetch("/api/doctor/schedule", { credentials: "include" });
  if (!r.ok) {
    document.getElementById("week").innerHTML =
      '<div class="empty">Не удалось загрузить расписание.</div>';
    return;
  }
  const s = await r.json();
  document.getElementById("week").innerHTML = s.days
    .map(
      (d) =>
        `<article class="day ${d.isWorking ? "working" : "off"}"><h3>${esc(d.dayName)}</h3>${d.isWorking ? `<strong>${esc(d.startTime)} — ${esc(d.endTime)}</strong><span>Рабочий день</span>` : "<strong>Выходной</strong><span>Приём не ведётся</span>"}</article>`,
    )
    .join("");
}
async function loadDay() {
  const date = document.getElementById("date").value;
  if (!date) return;
  document.getElementById("selectedLabel").textContent = new Date(
    date + "T12:00:00",
  ).toLocaleDateString("ru-RU", {
    weekday: "long",
    day: "2-digit",
    month: "long",
    year: "numeric",
  });
  const r = await fetch(
    "/api/doctor/schedule/availability?date=" + encodeURIComponent(date),
    { credentials: "include" },
  );
  if (!r.ok) {
    document.getElementById("slots").innerHTML =
      '<div class="empty">Не удалось загрузить расписание дня.</div>';
    return;
  }
  render(await r.json());
}
function render(data) {
  const box = document.getElementById("slots");
  if (!data.isWorking) {
    box.innerHTML = '<div class="empty">В этот день врач не принимает.</div>';
    return;
  }
  if (!data.slots.length) {
    box.innerHTML =
      '<div class="empty">Рабочее время не содержит свободных слотов.</div>';
    return;
  }
  box.innerHTML = data.slots
    .map(
      (s) =>
        `<article class="slot ${s.available ? "free" : "busy"}"><div class="time">${esc(s.time)}</div><div><strong>${s.available ? "Свободно" : esc(s.patientName || "Занято")}</strong>${!s.available ? `<span>${esc(s.service || "Запись занята")}</span>` : ""}</div>${s.available ? '<a href="/doctor/book.html">Записать пациента</a>' : `<a class="view" href="/doctor/appointments.html?date=${encodeURIComponent(data.date)}">Открыть записи</a>`}</article>`,
    )
    .join("");
}
document.getElementById("date").onchange = loadDay;
document.getElementById("logout").onclick = async () => {
  await fetch("/doctor/logout", { method: "POST", credentials: "include" });
  location = "/?account=doctor";
};
init();
