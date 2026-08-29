let period = "";
const initialDate = new URLSearchParams(location.search).get("date") || "";
const esc = (x) =>
  String(x ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
async function init() {
  if (initialDate) document.getElementById("date").value = initialDate;
  const r = await fetch("/api/doctor/me", { credentials: "include" });
  if (!r.ok) {
    location = "/?account=doctor";
    return;
  }
  const d = await r.json();
  document.getElementById("doctorName").textContent = d.name;
  load();
}
async function load() {
  const q = new URLSearchParams();
  if (period) q.set("period", period);
  if (document.getElementById("date").value)
    q.set("date", document.getElementById("date").value);
  if (document.getElementById("search").value.trim())
    q.set("search", document.getElementById("search").value.trim());
  const r = await fetch("/api/doctor/appointments?" + q, {
    credentials: "include",
  });
  if (!r.ok) {
    location = "/?account=doctor";
    return;
  }
  render(await r.json());
}
function render(items) {
  const box = document.getElementById("list");
  if (!items.length) {
    box.innerHTML =
      '<div class="empty">Записей по выбранным условиям нет.</div>';
    return;
  }
  box.innerHTML = items
    .map(
      (x) =>
        `<article class="appointment"><div><div class="date">${esc(x.appointment_date || "Дата не назначена")} · <b>${esc(x.appointment_time || "Время не назначено")}</b></div><h3>${esc(x.name)}</h3><div class="meta">${esc(x.services || "Услуга не указана")}</div><div class="meta">${esc(x.phone)}</div>${x.appointment_comment ? `<div class="comment">${esc(x.appointment_comment)}</div>` : ""}<span class="status">${esc(x.status)}</span></div><div class="actions"><a href="/doctor/patients.html?patient_id=${x.patient_id}">Пациент</a>${x.status === "Подтверждена" ? `<a class="primary action-link" href="/doctor/medical-record.html?appointment_id=${x.id}">Завершить приём</a><button class="danger" data-id="${x.id}" data-status="Отменена">Отменить</button>` : ""}</div></article>`,
    )
    .join("");
  box
    .querySelectorAll("[data-status]")
    .forEach(
      (b) => (b.onclick = () => changeStatus(b.dataset.id, b.dataset.status)),
    );
}
async function changeStatus(id, status) {
  if (
    !confirm(status === "Завершена" ? "Завершить приём?" : "Отменить запись?")
  )
    return;
  const r = await fetch("/api/doctor/requests/" + id, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ status }),
  });
  if (!r.ok) {
    alert(await r.text());
    return;
  }
  load();
}
document.querySelectorAll("[data-period]").forEach(
  (b) =>
    (b.onclick = () => {
      period = b.dataset.period;
      document
        .querySelectorAll("[data-period]")
        .forEach((x) => x.classList.remove("active"));
      b.classList.add("active");
      load();
    }),
);
document.getElementById("date").onchange = load;
document.getElementById("search").oninput = () => {
  clearTimeout(window.t);
  window.t = setTimeout(load, 250);
};
document.getElementById("clear").onclick = () => {
  period = "";
  document.getElementById("date").value = "";
  document.getElementById("search").value = "";
  document
    .querySelectorAll("[data-period]")
    .forEach((x) => x.classList.remove("active"));
  document.querySelector('[data-period=""]').classList.add("active");
  load();
};
document.getElementById("logout").onclick = async () => {
  await fetch("/doctor/logout", { method: "POST", credentials: "include" });
  location = "/?account=doctor";
};
init();
