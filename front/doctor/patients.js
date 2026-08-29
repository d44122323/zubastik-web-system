async function init() {
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
  const q = document.getElementById("search").value.trim();
  const r = await fetch(
    "/api/doctor/patients" + (q ? "?search=" + encodeURIComponent(q) : ""),
    { credentials: "include" },
  );
  if (!r.ok) {
    location = "/?account=doctor";
    return;
  }
  render(await r.json());
}
function esc(x) {
  return String(x ?? "").replace(
    /[&<>"']/g,
    (c) =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[
        c
      ],
  );
}
function parseLocalDateTime(value) {
  if (!value) return null;
  const s = String(value);
  const m = s.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2}))?/);
  if (!m) return new Date(s);
  return new Date(
    Number(m[1]),
    Number(m[2]) - 1,
    Number(m[3]),
    Number(m[4]),
    Number(m[5]),
    Number(m[6] || 0),
  );
}
function formatVisitDate(value) {
  const d = parseLocalDateTime(value);
  if (!d || Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
function render(items) {
  const box = document.getElementById("list");
  if (!items.length) {
    box.innerHTML =
      '<div class="empty">Пациентов по вашему doctor_id не найдено.</div>';
    return;
  }
  box.innerHTML = items
    .map(
      (p) =>
        '<article class="patient" data-id="' +
        p.id +
        '"><h3>' +
        esc(p.name) +
        '</h3><div class="muted">' +
        esc(p.phone) +
        '</div><div class="muted">Посещений: ' +
        p.visits +
        '</div><div class="muted">Последнее: ' +
        (p.lastVisit ? formatVisitDate(p.lastVisit) : "—") +
        "</div></article>",
    )
    .join("");
  box
    .querySelectorAll(".patient")
    .forEach((x) => (x.onclick = () => openPatient(x.dataset.id)));
}
async function openPatient(id) {
  const r = await fetch("/api/doctor/patients/" + id, {
    credentials: "include",
  });
  if (!r.ok) {
    alert("Пациент недоступен");
    return;
  }
  const p = await r.json();
  const d = document.getElementById("details");
  d.classList.remove("hidden");
  d.innerHTML =
    "<h2>" +
    esc(p.name) +
    "</h2><p><b>Телефон:</b> " +
    esc(p.phone) +
    "</p><p><b>Комментарий:</b> " +
    esc(p.comment || "—") +
    "</p><h3>История посещений</h3>" +
    ((p.history || []).length
      ? (p.history || [])
          .map(
            (v) =>
              '<div class="visit"><b>' +
              formatVisitDate(v.date) +
              "</b><br>" +
              esc(v.service || "Услуга не указана") +
              " · " +
              v.price +
              " ₽</div>",
          )
          .join("")
      : "<p>Истории пока нет.</p>") +
    "<h3>Медицинские записи</h3>" +
    ((p.medical_records || []).length
      ? (p.medical_records || [])
          .map(
            (m) =>
              '<div class="medical-record"><b>' +
              esc(m.appointment_date || "") +
              " · " +
              esc(m.appointment_time || "") +
              "</b><p><strong>Жалобы:</strong> " +
              esc(m.complaints || "—") +
              "</p><p><strong>Диагноз:</strong> " +
              esc(m.diagnosis || "—") +
              "</p><p><strong>Лечение:</strong> " +
              esc(m.treatment || "—") +
              "</p><p><strong>Рекомендации:</strong> " +
              esc(m.recommendations || "—") +
              "</p>" +
              (m.files || [])
                .map(
                  (f) =>
                    '<a class="file" target="_blank" href="/api/doctor/medical-files/' +
                    f.id +
                    '">📎 ' +
                    esc(f.file_name) +
                    "</a>",
                )
                .join("") +
              "</div>",
          )
          .join("")
      : "<p>Медицинских записей пока нет.</p>");
  d.scrollIntoView({ behavior: "smooth" });
}
document.getElementById("logout").onclick = async () => {
  await fetch("/doctor/logout", { method: "POST", credentials: "include" });
  location = "/?account=doctor";
};
document.getElementById("search").oninput = () => {
  clearTimeout(window.t);
  window.t = setTimeout(load, 250);
};
init();
