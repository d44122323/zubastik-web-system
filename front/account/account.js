const $ = (id) => document.getElementById(id);
let data = {};
const esc = (s) =>
  String(s ?? "").replace(
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
async function api(url, opt) {
  const r = await fetch(url, { credentials: "include", ...(opt || {}) });
  if (r.status === 401) {
    location.href = "/?account=patient";
    throw new Error("unauthorized");
  }
  if (!r.ok) throw new Error(await r.text());
  return r.json();
}
function show(name) {
  document
    .querySelectorAll(".side")
    .forEach((x) => x.classList.toggle("active", x.dataset.section === name));
  document
    .querySelectorAll(".section")
    .forEach((x) => x.classList.toggle("active", x.id === name));
}
document.querySelectorAll(".side").forEach(
  (x) =>
    (x.onclick = () => {
      show(x.dataset.section);
      loadSection(x.dataset.section);
    }),
);
$("logout").onclick = async () => {
  await fetch("/patient/logout", { method: "POST", credentials: "include" });
  location.href = "/";
};
async function loadProfile() {
  const p = await api("/api/patient/me");
  data.profile = p;
  $("userName").textContent = p.name;
  $("helloName").textContent = p.name;
  $("pName").value = p.name;
  $("pPhone").value = p.phone;
  $("pEmail").value = p.email || "";
  $("pBirth").value = p.birthDate || "";
  await loadTelegramStatus();
}

async function loadTelegramStatus() {
  const status = $("telegramStatus"),
    connect = $("telegramConnect"),
    disconnect = $("telegramDisconnect");
  if (!status || !connect) return;
  try {
    const d = await api("/api/patient/telegram/status");
    if (d.connected) {
      status.textContent = d.username
        ? `🟢 Telegram подключён: @${esc(d.username)}`
        : "🟢 Telegram подключён";
      connect.classList.add("hidden");
      disconnect && disconnect.classList.remove("hidden");
    } else {
      status.textContent = "🔴 Telegram не подключён";
      connect.classList.remove("hidden");
      disconnect && disconnect.classList.add("hidden");
    }
  } catch (e) {
    status.textContent = "Не удалось проверить Telegram";
  }
}
async function connectTelegram() {
  try {
    const d = await api("/api/patient/telegram/connect");
    if (d.url) window.open(d.url, "_blank", "noopener");
    setTimeout(loadTelegramStatus, 2500);
  } catch (e) {
    alert(e.message || "Не удалось подключить Telegram");
  }
}
async function disconnectTelegram() {
  if (!confirm("Отключить Telegram-уведомления?")) return;
  try {
    await api("/api/patient/telegram/disconnect", { method: "POST" });
    await loadTelegramStatus();
  } catch (e) {
    alert(e.message || "Не удалось отключить Telegram");
  }
}
document
  .getElementById("telegramConnect")
  ?.addEventListener("click", connectTelegram);
document
  .getElementById("telegramDisconnect")
  ?.addEventListener("click", disconnectTelegram);
async function loadAppointments() {
  const a = await api("/api/patient/appointments");
  data.apps = a;
  $("appointmentsList").innerHTML = a.length
    ? a
        .map((x) => {
          const closed = [
            "Завершена",
            "Отменена",
            "Отменено",
            "Отменено пациентом",
          ].includes(x.status);
          return `<div class="appointment"><div class="row"><div><h3>${esc(x.service)}</h3><div>${esc(x.date || "Дата не назначена")} · ${esc(x.time || "—")}</div><div class="muted">Врач: ${esc(x.doctor)}</div></div><span class="tag">${esc(x.status)}</span></div>${x.comment ? `<p class="muted">${esc(x.comment)}</p>` : ""}${closed ? "" : `<div class="actions"><button class="secondary" onclick="openReschedule(${x.id})">Перенести</button><button class="danger" onclick="cancelAppointment(${x.id})">Отменить</button></div>`}</div>`;
        })
        .join("")
    : "<div class='panel'>Записей пока нет.</div>";
  const today = new Date().toISOString().slice(0, 10);
  const upcoming = a.filter(
    (x) =>
      x.date >= today &&
      !["Отменена", "Отменено", "Отменено пациентом", "Завершена"].includes(
        x.status,
      ),
  );
  $("upcomingCount").textContent = upcoming.length;
  $("nextAppointment").innerHTML = upcoming[0]
    ? `<b>${esc(upcoming[0].service)}</b><br>${esc(upcoming[0].date)} · ${esc(upcoming[0].time)}<br>Врач: ${esc(upcoming[0].doctor)}`
    : "Нет предстоящих записей.";
}
let rescheduleAppointment = null;
async function openReschedule(id) {
  rescheduleAppointment = (data.apps || []).find((x) => x.id === id);
  if (!rescheduleAppointment || !rescheduleAppointment.doctorId) {
    alert("У записи не назначен врач");
    return;
  }
  $("rescheduleTitle").textContent = rescheduleAppointment.service;
  $("rescheduleDoctor").textContent = `Врач: ${rescheduleAppointment.doctor}`;
  const d = new Date();
  d.setDate(d.getDate() + 1);
  $("rescheduleDate").min = d.toISOString().slice(0, 10);
  $("rescheduleDate").value =
    rescheduleAppointment.date || $("rescheduleDate").min;
  $("rescheduleModal").classList.remove("hidden");
  await loadRescheduleSlots();
}
async function loadRescheduleSlots() {
  if (!rescheduleAppointment) return;
  const date = $("rescheduleDate").value;
  const slots = $("rescheduleSlots");
  const msg = $("rescheduleMessage");
  slots.innerHTML = "";
  msg.textContent = "Загружаем свободное время...";
  try {
    const r = await fetch(
      `/api/doctors/${rescheduleAppointment.doctorId}/availability?date=${encodeURIComponent(date)}`,
      { credentials: "include" },
    );
    const d = await r.json();
    if (!r.ok) throw new Error(d || "Ошибка");
    if (d.unavailable || !d.isWorking) {
      msg.textContent = d.unavailableReason
        ? `Врач недоступен: ${d.unavailableReason}`
        : "В этот день врач не принимает.";
      return;
    }
    const free = (d.slots || []).filter(
      (x) =>
        x.available &&
        !(
          date === rescheduleAppointment.date &&
          x.time === rescheduleAppointment.time
        ),
    );
    msg.textContent = free.length
      ? `${d.dayName}: ${d.startTime}–${d.endTime}`
      : "Свободных слотов нет.";
    free.forEach((x) => {
      const b = document.createElement("button");
      b.type = "button";
      b.textContent = x.time;
      b.onclick = () => saveReschedule(date, x.time);
      slots.appendChild(b);
    });
  } catch (e) {
    console.error(e);
    msg.textContent = "Не удалось загрузить свободное время.";
  }
}
async function saveReschedule(date, time) {
  if (!rescheduleAppointment) return;
  try {
    await api(
      `/api/patient/appointments/${rescheduleAppointment.id}/reschedule`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action: "reschedule", date, time }),
      },
    );
    $("rescheduleModal").classList.add("hidden");
    await loadAppointments();
    await loadNotifications();
    alert("Запись успешно перенесена.");
  } catch (e) {
    alert(e.message || "Не удалось перенести запись");
    await loadRescheduleSlots();
  }
}
function closeReschedule() {
  $("rescheduleModal").classList.add("hidden");
  rescheduleAppointment = null;
}
$("rescheduleDate").addEventListener("change", loadRescheduleSlots);
$("rescheduleClose").onclick = closeReschedule;
$("rescheduleCancel").onclick = closeReschedule;
$("rescheduleModal").addEventListener("click", (e) => {
  if (e.target === $("rescheduleModal")) closeReschedule();
});
async function cancelAppointment(id) {
  if (!confirm("Отменить запись?")) return;
  await api("/api/patient/appointments/" + id, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ action: "cancel" }),
  });
  await loadAppointments();
  await loadNotifications();
  alert("Запись отменена.");
}
async function loadMedical() {
  const a = await api("/api/patient/medical");
  $("medicalCount").textContent = a.length;
  $("medicalList").innerHTML = a.length
    ? a
        .map(
          (x) =>
            `<div class="record"><div class="row"><b>${esc(x.date)} · ${esc(x.time)}</b><span class="tag">${esc(x.doctor)}</span></div><h3>Диагноз</h3><p>${esc(x.diagnosis) || "—"}</p><h3>Жалобы</h3><p>${esc(x.complaints) || "—"}</p><h3>Лечение</h3><p>${esc(x.treatment) || "—"}</p><h3>Рекомендации</h3><p>${esc(x.recommendations) || "—"}</p></div>`,
        )
        .join("")
    : "<div class='panel'>Медицинских записей пока нет.</div>";
}
async function loadDocuments() {
  const a = await api("/api/patient/documents");
  $("documentsList").innerHTML = a.length
    ? a
        .map(
          (x) =>
            `<div class="document row"><div><b>📄 ${esc(x.name)}</b><div class="muted">${esc(x.date)} ${esc(x.time)}</div></div><a class="secondary" href="/api/patient/files/${x.id}" target="_blank">Открыть</a></div>`,
        )
        .join("")
    : "<div class='panel'>Документов пока нет.</div>";
}
async function loadNotifications() {
  const a = await api("/api/patient/notifications");
  const unread = a.filter((x) => !x.read);
  $("notificationCount").textContent = unread.length;
  $("notificationsList").innerHTML = a.length
    ? a
        .map(
          (x) =>
            `<div class="notification ${x.read ? "read" : "unread"}"><div class="row"><b>${esc(x.title)}</b><span class="tag">${x.read ? "Прочитано" : "Новое"}</span></div><p>${esc(x.text)}</p><div class="muted">${esc(x.createdAt || "")}</div></div>`,
        )
        .join("")
    : "<div class='panel'>Уведомлений пока нет.</div>";
}
async function loadPayments() {
  const a = await api("/api/patient/payments");
  const paid = new Set(
    a.filter((x) => x.status === "Оплачено").map((x) => x.requestId),
  );
  const apps = data.apps || (await api("/api/patient/appointments"));
  const items = apps
    .filter(
      (x) =>
        x.price > 0 &&
        !["Отменена", "Отменено", "Отменено пациентом"].includes(x.status),
    )
    .map(
      (x) =>
        `<div class="payment"><div class="payment-main"><b>${esc(x.service)}</b><div class="muted">${esc(x.date)} · ${esc(x.time)} · Врач: ${esc(x.doctor)}</div></div><div class="payment-side"><div class="payment-amount"><span class="price-label">Стоимость</span><strong>${x.price} ₽</strong></div>${paid.has(x.id) ? `<span class="tag paid-tag">Оплачено</span>` : `<button class="primary pay-button" onclick="pay(${x.id})">Оплатить</button>`}</div></div>`,
    );
  $("paymentsList").innerHTML = items.length
    ? items.join("")
    : "<div class='panel'>Сумм к оплате пока нет.</div>";
}
async function pay(id) {
  await api("/api/patient/payments", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ requestId: id }),
  });
  alert("Тестовая оплата успешно выполнена");
  loadPayments();
}

async function loadDoctorQuestions() {
  const [questions, apps] = await Promise.all([
    api("/api/patient/questions"),
    api("/api/patient/appointments"),
  ]);
  const completed = (apps || []).filter(
    (x) => x.status === "Завершена" && x.doctorId,
  );
  const select = $("questionAppointment");
  select.innerHTML = completed.length
    ? completed
        .map(
          (x) =>
            `<option value="${x.id}">${esc(x.date || "")} · ${esc(x.time || "—")} · ${esc(x.doctor)} · ${esc(x.service)}</option>`,
        )
        .join("")
    : '<option value="">Нет завершённых посещений</option>';
  $("sendQuestion").disabled = !completed.length;
  $("questionsList").innerHTML = questions.length
    ? questions
        .map(
          (q) =>
            `<div class="question-card"><div class="row"><div><b>${esc(q.doctor)}</b><div class="muted">${esc(q.service)}</div></div><span class="tag">${q.answer ? "Ответ получен" : "Ожидает ответа"}</span></div><div class="question-bubble"><b>Вы:</b><p>${esc(q.question)}</p></div>${q.answer ? `<div class="answer-bubble"><b>Врач:</b><p>${esc(q.answer)}</p></div>` : ""}</div>`,
        )
        .join("")
    : "<div class='panel'>Вопросов пока нет.</div>";
}
async function sendDoctorQuestion() {
  const appointmentId = Number($("questionAppointment").value);
  const question = $("questionText").value.trim();
  if (!appointmentId) {
    $("questionMsg").textContent = "Сначала завершите приём у врача.";
    return;
  }
  if (!question) {
    $("questionMsg").textContent = "Напишите вопрос.";
    return;
  }
  try {
    await api("/api/patient/questions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ appointmentId, question }),
    });
    $("questionText").value = "";
    $("questionMsg").textContent = "Вопрос отправлен врачу.";
    await loadDoctorQuestions();
    await loadNotifications();
  } catch (e) {
    $("questionMsg").textContent = e.message || "Не удалось отправить вопрос.";
  }
}
$("sendQuestion").onclick = sendDoctorQuestion;

async function loadSection(s) {
  try {
    if (s === "profile") await loadProfile();
    if (s === "appointments") {
      await loadAppointments();
    }
    if (s === "medical") await loadMedical();
    if (s === "documents") await loadDocuments();
    if (s === "notifications") {
      await loadNotifications();
      await api("/api/patient/notifications", { method: "PUT" });
      await loadNotifications();
    }
    if (s === "payments") {
      await loadAppointments();
      await loadPayments();
    }
    if (s === "doctorQuestions") await loadDoctorQuestions();
  } catch (e) {
    console.error(e);
  }
}
$("profileForm").onsubmit = async (e) => {
  e.preventDefault();
  const r = await api("/api/patient/profile", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      name: $("pName").value,
      phone: $("pPhone").value,
      email: $("pEmail").value,
      birthDate: $("pBirth").value,
    }),
  });
  $("profileMsg").textContent = "Профиль сохранён";
  loadProfile();
};
(async () => {
  await loadProfile();
  await loadAppointments();
  await loadMedical();
  await loadNotifications();
})();
