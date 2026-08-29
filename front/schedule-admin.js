let doctors = [];
let selectedDoctorId = "";
let selectedDate = "";
let currentAvailability = null;
let selectedRequest = null;
const $ = (id) => document.getElementById(id);
function esc(v = "") {
  return String(v)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}
function localDate() {
  const d = new Date();
  return new Date(d.getTime() - d.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 10);
}
function addDays(s, n) {
  const d = new Date(`${s}T12:00:00`);
  d.setDate(d.getDate() + n);
  return d.toISOString().slice(0, 10);
}
function fmtDate(s) {
  if (!s) return "—";
  return new Date(`${s}T12:00:00`).toLocaleDateString("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}
function showMessage(t, error = false) {
  $("message").textContent = t || "";
  $("message").style.color = error ? "#d00000" : "#078c4f";
  setTimeout(() => {
    $("message").textContent = "";
  }, 3500);
}
async function loadDoctors() {
  const r = await fetch("/api/doctors?all=true", {
    credentials: "same-origin",
  });
  if (!r.ok) throw new Error("Не удалось загрузить врачей");
  doctors = await r.json();
  const opts =
    '<option value="">Все врачи</option>' +
    doctors
      .map(
        (d) =>
          `<option value="${d.id}">${esc(d.name)}${d.isActive ? "" : " — не принимает"}</option>`,
      )
      .join("");
  $("doctorSelect").innerHTML = opts;
  $("rescheduleDoctor").innerHTML = doctors
    .filter((d) => d.isActive)
    .map((d) => `<option value="${d.id}">${esc(d.name)}</option>`)
    .join("");
}
async function loadSchedule() {
  selectedDoctorId = $("doctorSelect").value;
  selectedDate = $("dateInput").value;
  const grid = $("scheduleGrid");
  grid.innerHTML = '<div class="empty">Загрузка расписания…</div>';
  let ids = selectedDoctorId
    ? [Number(selectedDoctorId)]
    : doctors.filter((d) => d.isActive).map((d) => d.id);
  if (!ids.length) {
    grid.innerHTML = '<div class="empty">Нет активных врачей.</div>';
    return;
  }
  try {
    const results = await Promise.all(
      ids.map((id) =>
        fetch(
          `/api/doctors/${id}/availability?date=${encodeURIComponent(selectedDate)}`,
          { credentials: "same-origin" },
        ).then((r) => {
          if (!r.ok) throw new Error("availability");
          return r.json();
        }),
      ),
    );
    renderSchedule(results);
  } catch (e) {
    grid.innerHTML =
      '<div class="empty">Не удалось загрузить расписание.</div>';
    showMessage("Не удалось загрузить расписание", true);
  }
}
function renderSchedule(results) {
  const grid = $("scheduleGrid");
  const doctorName = selectedDoctorId
    ? doctors.find((d) => d.id === Number(selectedDoctorId))?.name || "Врач"
    : "Все врачи";
  $("scheduleTitle").textContent = doctorName;
  $("scheduleSubtitle").textContent =
    `${fmtDate(selectedDate)} · ${results.reduce((n, x) => n + (x.isWorking ? x.slots.length : 0), 0)} слотов`;
  const cards = [];
  results.forEach((a) => {
    const d = doctors.find((x) => x.id === a.doctorId);
    if (!a.isWorking) {
      if (selectedDoctorId) {
        const reason = a.unavailableReason
          ? ` Причина: ${esc(a.unavailableReason)}`
          : "";
        cards.push(
          `<div class="empty schedule-unavailable"><strong>${esc(d?.name || "Врач")} не принимает ${fmtDate(selectedDate)}.</strong>${a.unavailable ? `<span>Врач недоступен.${reason}</span>` : ""}</div>`,
        );
      }
      return;
    }
    a.slots.forEach((s) => {
      cards.push(
        `<button class="slot ${s.available ? "free" : "busy"}" data-doctor="${a.doctorId}" data-time="${esc(s.time)}" data-request="${s.requestId || 0}"><div class="slot-time">${esc(s.time)}</div><div class="slot-status">${s.available ? "Свободно" : esc(d?.name || "Врач")}</div>${s.available ? "" : `<div class="slot-name">${esc(s.patientName || "Пациент")}</div><div class="slot-service">${esc(s.service || "")}</div>`}</button>`,
      );
    });
  });
  grid.innerHTML = cards.length
    ? cards.join("")
    : '<div class="empty">На выбранную дату свободных/занятых слотов нет.</div>';
  grid.querySelectorAll(".slot").forEach((b) =>
    b.addEventListener("click", () => {
      const req = Number(b.dataset.request);
      if (req) openAppointment(req);
    }),
  );
}
async function openAppointment(id) {
  const r = await fetch(`/api/requests/${id}`, { credentials: "same-origin" });
  if (!r.ok) {
    showMessage("Не удалось открыть запись", true);
    return;
  }
  selectedRequest = await r.json();
  $("appointmentTitle").textContent = selectedRequest.name || "Запись";
  $("appointmentInfo").innerHTML =
    `<div class="info-row"><span>Услуга</span><strong>${esc(selectedRequest.services || "—")}</strong></div><div class="info-row"><span>Врач</span><strong>${esc(selectedRequest.doctor_name || "Не назначен")}</strong></div><div class="info-row"><span>Дата и время</span><strong>${esc(selectedRequest.appointment_date || "—")} · ${esc(selectedRequest.appointment_time || "—")}</strong></div><div class="info-row"><span>Статус</span><strong>${esc(selectedRequest.status || "—")}</strong></div><div class="info-row"><span>Телефон</span><strong>${esc(selectedRequest.phone || "—")}</strong></div>`;
  $("overlay").classList.add("active");
  $("appointmentDrawer").classList.add("active");
}
function closeDrawer() {
  $("overlay").classList.remove("active");
  $("appointmentDrawer").classList.remove("active");
}
function openReschedule() {
  if (!selectedRequest) return;
  $("reschedulePatient").textContent =
    `${selectedRequest.name} · ${selectedRequest.services || "Без услуги"}`;
  $("rescheduleDoctor").value =
    selectedRequest.doctor_id || doctors.find((d) => d.isActive)?.id || "";
  $("rescheduleDate").value = selectedRequest.appointment_date || selectedDate;
  $("rescheduleModal").classList.remove("hidden");
  loadRescheduleTimes();
}
async function loadRescheduleTimes() {
  const id = Number($("rescheduleDoctor").value);
  const date = $("rescheduleDate").value;
  const sel = $("rescheduleTime");
  sel.innerHTML = "<option>Загрузка…</option>";
  if (!id || !date) return;
  const r = await fetch(`/api/doctors/${id}/availability?date=${date}`, {
    credentials: "same-origin",
  });
  if (!r.ok) {
    sel.innerHTML = '<option value="">Ошибка загрузки</option>';
    return;
  }
  const a = await r.json();
  if (a.unavailable) {
    sel.innerHTML = `<option value="">Врач недоступен${a.unavailableReason ? ` — ${a.unavailableReason}` : ""}</option>`;
    return;
  }
  const times = a.slots || [];
  let html = times
    .filter(
      (s) =>
        s.available ||
        (s.time === selectedRequest?.appointment_time &&
          id === selectedRequest?.doctor_id &&
          date === selectedRequest?.appointment_date),
    )
    .map(
      (s) =>
        `<option value="${s.time}">${s.time}${s.time === selectedRequest?.appointment_time && id === selectedRequest?.doctor_id && date === selectedRequest?.appointment_date ? " · текущее" : ""}</option>`,
    )
    .join("");
  sel.innerHTML = html || '<option value="">Нет свободного времени</option>';
  if (
    selectedRequest?.appointment_time &&
    [...sel.options].some((o) => o.value === selectedRequest.appointment_time)
  )
    sel.value = selectedRequest.appointment_time;
}
async function saveReschedule() {
  if (!selectedRequest) return;
  const body = {
    status: selectedRequest.status || "Подтверждена",
    doctor_id: Number($("rescheduleDoctor").value),
    appointment_date: $("rescheduleDate").value,
    appointment_time: $("rescheduleTime").value,
  };
  const r = await fetch(`/api/requests/${selectedRequest.id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "same-origin",
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    showMessage(t || "Не удалось перенести запись", true);
    return;
  }
  $("rescheduleModal").classList.add("hidden");
  closeDrawer();
  showMessage("Запись изменена");
  await loadSchedule();
}
async function cancelAppointment() {
  if (!selectedRequest || !confirm("Отменить эту запись?")) return;
  const body = {
    status: "Отменена",
    doctor_id: selectedRequest.doctor_id,
    appointment_date: selectedRequest.appointment_date,
    appointment_time: selectedRequest.appointment_time,
  };
  const r = await fetch(`/api/requests/${selectedRequest.id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "same-origin",
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    showMessage("Не удалось отменить запись", true);
    return;
  }
  closeDrawer();
  showMessage("Запись отменена");
  await loadSchedule();
}
async function openScheduleSettings() {
  const id = Number($("doctorSelect").value);
  if (!id) {
    showMessage("Сначала выберите врача", true);
    return;
  }
  const doctor = doctors.find((d) => d.id === id);
  $("scheduleDoctorModalTitle").textContent =
    `Рабочее расписание · ${doctor?.name || "Врач"}`;
  const r = await fetch(`/api/doctors/${id}/schedule`, {
    credentials: "same-origin",
  });
  if (!r.ok) {
    showMessage("Не удалось загрузить рабочее расписание", true);
    return;
  }
  const s = await r.json();
  $("scheduleDays").innerHTML = s.days
    .map(
      (d) =>
        `<div class="day-row"><strong>${esc(d.dayName)}</strong><input type="checkbox" data-working="${d.weekday}" ${d.isWorking ? "checked" : ""}><input type="time" data-start="${d.weekday}" value="${d.startTime || "09:00"}"><input type="time" data-end="${d.weekday}" value="${d.endTime || "18:00"}"></div>`,
    )
    .join("");
  $("scheduleModal").classList.remove("hidden");
  toggleScheduleInputs();
}
function toggleScheduleInputs() {
  document.querySelectorAll(".day-row").forEach((row) => {
    const c = row.querySelector("input[type=checkbox]");
    row.querySelectorAll("input[type=time]").forEach((x) => {
      x.disabled = !c.checked;
      x.classList.toggle("disabled-time", !c.checked);
    });
  });
}
async function saveSchedule() {
  const id = Number($("doctorSelect").value);
  const days = [...document.querySelectorAll(".day-row")].map((row) => {
    const c = row.querySelector("input[type=checkbox]"),
      s = row.querySelector("input[data-start]"),
      e = row.querySelector("input[data-end]");
    return {
      weekday: Number(c.dataset.working),
      isWorking: c.checked,
      startTime: s.value || "09:00",
      endTime: e.value || "18:00",
    };
  });
  const r = await fetch(`/api/doctors/${id}/schedule`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "same-origin",
    body: JSON.stringify({ doctorId: id, days }),
  });
  if (!r.ok) {
    showMessage("Не удалось сохранить расписание", true);
    return;
  }
  $("scheduleModal").classList.add("hidden");
  showMessage("Рабочее расписание сохранено");
  await loadSchedule();
}
function init() {
  selectedDate = localDate();
  $("dateInput").value = selectedDate;
  loadDoctors()
    .then(loadSchedule)
    .catch((e) => showMessage(e.message, true));
  $("doctorSelect").addEventListener("change", loadSchedule);
  $("dateInput").addEventListener("change", loadSchedule);
  $("prevDay").onclick = () => {
    $("dateInput").value = addDays($("dateInput").value, -1);
    loadSchedule();
  };
  $("nextDay").onclick = () => {
    $("dateInput").value = addDays($("dateInput").value, 1);
    loadSchedule();
  };
  $("todayBtn").onclick = () => {
    $("dateInput").value = localDate();
    loadSchedule();
  };
  $("closeDrawer").onclick = closeDrawer;
  $("overlay").onclick = closeDrawer;
  $("rescheduleBtn").onclick = openReschedule;
  $("cancelAppointmentBtn").onclick = cancelAppointment;
  $("closeReschedule").onclick = () =>
    $("rescheduleModal").classList.add("hidden");
  $("cancelReschedule").onclick = () =>
    $("rescheduleModal").classList.add("hidden");
  $("rescheduleDoctor").onchange = loadRescheduleTimes;
  $("rescheduleDate").onchange = loadRescheduleTimes;
  $("saveReschedule").onclick = saveReschedule;
  $("scheduleSettingsBtn").onclick = openScheduleSettings;
  $("closeSchedule").onclick = () => $("scheduleModal").classList.add("hidden");
  $("cancelSchedule").onclick = () =>
    $("scheduleModal").classList.add("hidden");
  $("saveSchedule").onclick = saveSchedule;
  document.addEventListener("change", (e) => {
    if (e.target.matches('.day-row input[type="checkbox"]'))
      toggleScheduleInputs();
  });
}
document.addEventListener("DOMContentLoaded", init);
