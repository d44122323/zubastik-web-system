async function loadDoctor() {
  const params = new URLSearchParams(window.location.search);
  const id = params.get("id");
  if (!id) {
    window.location.href = "doctors.html";
    return;
  }

  try {
    const response = await fetch(`/api/doctors/${encodeURIComponent(id)}`);
    if (!response.ok) throw new Error("Врач не найден");
    const doctor = await response.json();

    document.title = `${doctor.name} | ЗУ-БАСТИК`;
    document.getElementById("doctorPhoto").src = doctor.photo;
    document.getElementById("doctorPhoto").alt = doctor.name;
    document.getElementById("doctorName").textContent = doctor.name;
    document.getElementById("doctorPosition").textContent = doctor.position;
    document.getElementById("doctorExperience").textContent = doctor.experience;
    document.getElementById("doctorSpecialization").textContent =
      doctor.specialization;
    document.getElementById("doctorSpecializationFull").textContent =
      doctor.specializationFull;
    document.getElementById("doctorEducation").innerHTML =
      `<p>${escapeHtml(doctor.education)}</p>`;
    document.getElementById("doctorDescription").innerHTML = (
      doctor.description || []
    )
      .map((text) => `<p>${escapeHtml(text)}</p>`)
      .join("");

    const bookingUrl = `index.html?doctor=${doctor.id}#contacts`;
    ["doctorRecordTop", "doctorRecord", "doctorRecordFooter"].forEach(
      (elementId) => {
        const link = document.getElementById(elementId);
        if (link) link.href = bookingUrl;
      },
    );
    initBooking(doctor.id);
  } catch (error) {
    document.getElementById("doctorProfile").innerHTML =
      '<div class="detail-card"><h2>Врач не найден</h2><p>Вернитесь к списку специалистов.</p></div>';
    console.error(error);
  }
}

function escapeHtml(value) {
  return String(value ?? "").replace(
    /[&<>'"]/g,
    (char) =>
      ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        "'": "&#39;",
        '"': "&quot;",
      })[char],
  );
}

document.addEventListener("DOMContentLoaded", loadDoctor);

function initBooking(doctorId) {
  const dateInput = document.getElementById("bookingDate");
  const slots = document.getElementById("slots");
  const message = document.getElementById("availabilityMessage");
  if (!dateInput || !slots) return;
  const today = new Date();
  const local = new Date(today.getTime() - today.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 10);
  dateInput.min = local;
  dateInput.value = local;

  async function loadAvailability() {
    slots.innerHTML = "";
    message.textContent = "Загружаем свободные интервалы...";
    try {
      const response = await fetch(
        `/api/doctors/${doctorId}/availability?date=${encodeURIComponent(dateInput.value)}`,
      );
      const data = await response.json();
      if (!response.ok) throw new Error(data || "Ошибка загрузки");
      if (!data.isWorking) {
        message.textContent = `${data.dayName} — выходной`;
        slots.innerHTML =
          '<div class="slot-empty">В этот день врач не принимает.</div>';
        return;
      }
      const free = (data.slots || []).filter((slot) => slot.available);
      message.textContent = `${data.dayName}, ${data.startTime}–${data.endTime}`;
      if (!free.length) {
        slots.innerHTML =
          '<div class="slot-empty">Свободных мест на эту дату нет.</div>';
        return;
      }
      free.forEach((slot) => {
        const button = document.createElement("button");
        button.type = "button";
        button.className = "slot-btn";
        button.textContent = slot.time;
        button.addEventListener("click", () => {
          window.location.href = `index.html?doctor=${doctorId}&date=${encodeURIComponent(dateInput.value)}&time=${encodeURIComponent(slot.time)}#contacts`;
        });
        slots.appendChild(button);
      });
    } catch (error) {
      console.error(error);
      message.textContent = "Не удалось загрузить расписание.";
    }
  }
  dateInput.addEventListener("change", loadAvailability);
  loadAvailability();
}
