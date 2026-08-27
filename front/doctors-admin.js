let doctors = [];

const list = document.getElementById("doctorsList");
const modal = document.getElementById("doctorModal");
const form = document.getElementById("doctorForm");
const message = document.getElementById("message");
const scheduleModal = document.getElementById("scheduleModal");
const scheduleDays = document.getElementById("scheduleDays");
const accountModal = document.getElementById("accountModal");
const photoInput = document.getElementById("photo");
const photoPreview = document.getElementById("photoPreview");
const photoPreviewImage = document.getElementById("photoPreviewImage");
let scheduleDoctorId = null;
let accountDoctorId = null;
let unavailabilityDoctorId = null;

function escapeHtml(value = "") {
    return String(value)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function showMessage(text, type = "success") {
    message.textContent = text;
    message.className = `message ${type}`;
    setTimeout(() => {
        message.textContent = "";
        message.className = "message";
    }, 3500);
}

async function loadDoctors() {
    const response = await fetch("/api/doctors?all=true");

    if (response.status === 401) {
        window.location.href = "/?account=admin";
        return;
    }

    if (!response.ok) {
        showMessage("Не удалось загрузить врачей", "error");
        return;
    }

    doctors = await response.json();
    renderDoctors();
}

function renderDoctors() {
    const search = document.getElementById("searchInput").value.trim().toLowerCase();
    const filter = document.getElementById("statusFilter").value;

    const filtered = doctors.filter(doctor => {
        const matchesSearch =
            !search ||
            doctor.name.toLowerCase().includes(search) ||
            doctor.position.toLowerCase().includes(search) ||
            doctor.specialization.toLowerCase().includes(search);

        const matchesStatus =
            filter === "all" ||
            (filter === "active" && doctor.isActive) ||
            (filter === "inactive" && !doctor.isActive);

        return matchesSearch && matchesStatus;
    });

    if (!filtered.length) {
        list.innerHTML = `<div class="empty">Врачи не найдены</div>`;
        return;
    }

    list.innerHTML = filtered.map(doctor => `
        <article class="doctor-card ${doctor.isActive ? "" : "inactive"}">
            <img src="${escapeHtml(doctor.photo || "img/logo.png")}" alt="">
            <div class="doctor-info">
                <div class="doctor-top">
                    <div>
                        <h2>${escapeHtml(doctor.name)}</h2>
                        <p class="position">${escapeHtml(doctor.position)}</p>
                    </div>
                    <span class="status ${doctor.isActive ? "active" : "inactive"}">
                        ${doctor.isActive ? "Активен" : "Не принимает"}
                    </span>
                </div>

                <div class="doctor-meta">
                    <span><b>Стаж:</b> ${escapeHtml(doctor.experience || "—")}</span>
                    <span><b>Специализация:</b> ${escapeHtml(doctor.specialization || "—")}</span>
                </div>

                <div class="doctor-actions">
                    <button class="secondary-btn schedule-btn" data-id="${doctor.id}">Расписание</button>
                    <button class="secondary-btn account-btn" data-id="${doctor.id}">Учётная запись</button>
                    <button class="secondary-btn edit-btn" data-id="${doctor.id}">Изменить</button>
                    ${doctor.isActive
                        ? `<button class="danger-btn deactivate-btn" data-id="${doctor.id}">Деактивировать</button>`
                        : `<button class="activate-btn" data-id="${doctor.id}">Вернуть в работу</button>`
                    }
                </div>
            </div>
        </article>
    `).join("");

    document.querySelectorAll(".account-btn").forEach(btn => {
        btn.addEventListener("click", () => openAccount(Number(btn.dataset.id)));
    });

    document.querySelectorAll(".edit-btn").forEach(btn => {
        btn.addEventListener("click", () => openEdit(Number(btn.dataset.id)));
    });

    document.querySelectorAll(".schedule-btn").forEach(btn => {
        btn.addEventListener("click", () => openSchedule(Number(btn.dataset.id)));
    });

    document.querySelectorAll(".deactivate-btn").forEach(btn => {
        btn.addEventListener("click", () => changeStatus(Number(btn.dataset.id), "deactivate"));
    });

    document.querySelectorAll(".activate-btn").forEach(btn => {
        btn.addEventListener("click", () => changeStatus(Number(btn.dataset.id), "activate"));
    });
}

function openCreate() {
    form.reset();
    document.getElementById("doctorId").value = "";
    clearPhotoPreview();
    document.getElementById("modalTitle").textContent = "Добавить врача";
    modal.classList.remove("hidden");
}

function openEdit(id) {
    const doctor = doctors.find(item => item.id === id);
    if (!doctor) return;

    document.getElementById("doctorId").value = doctor.id;
    document.getElementById("name").value = doctor.name || "";
    photoInput.value = "";
    setPhotoPreview(doctor.photo || "");
    document.getElementById("position").value = doctor.position || "";
    document.getElementById("experience").value = doctor.experience || "";
    document.getElementById("specialization").value = doctor.specialization || "";
    document.getElementById("education").value = doctor.education || "";
    document.getElementById("description").value = (doctor.description || []).join("\n");
    document.getElementById("specializationFull").value = doctor.specializationFull || "";

    document.getElementById("modalTitle").textContent = "Редактировать врача";
    modal.classList.remove("hidden");
}

function closeModal() {
    modal.classList.add("hidden");
    photoInput.value = "";
    clearPhotoPreview();
}

function setPhotoPreview(url) {
    if (!url) {
        clearPhotoPreview();
        return;
    }
    photoPreviewImage.src = url;
    photoPreview.classList.remove("hidden");
}

function clearPhotoPreview() {
    photoPreview.classList.add("hidden");
    photoPreviewImage.removeAttribute("src");
}

photoInput.addEventListener("change", () => {
    const file = photoInput.files && photoInput.files[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
        photoInput.value = "";
        clearPhotoPreview();
        showMessage("Выберите файл изображения", "error");
        return;
    }
    if (file.size > 5 * 1024 * 1024) {
        photoInput.value = "";
        clearPhotoPreview();
        showMessage("Фото должно быть не больше 5 МБ", "error");
        return;
    }
    setPhotoPreview(URL.createObjectURL(file));
});

async function uploadDoctorPhoto(id) {
    const file = photoInput.files && photoInput.files[0];
    if (!file) return null;

    const data = new FormData();
    data.append("photo", file);
    const response = await fetch(`/api/doctors/${id}/photo`, {
        method: "POST",
        body: data
    });
    if (response.status === 401) {
        window.location.href = "/?account=admin";
        return null;
    }
    if (!response.ok) {
        let detail = "Не удалось загрузить фото";
        try {
            const text = await response.text();
            if (text) detail = text;
        } catch (_) {}
        throw new Error(detail);
    }
    return response.json();
}

async function saveDoctor(event) {
    event.preventDefault();

    const id = document.getElementById("doctorId").value;
    const payload = {
        name: document.getElementById("name").value.trim(),
        photo: id ? (doctors.find(item => item.id === Number(id))?.photo || "") : "",
        position: document.getElementById("position").value.trim(),
        experience: document.getElementById("experience").value.trim(),
        specialization: document.getElementById("specialization").value.trim(),
        education: document.getElementById("education").value.trim(),
        description: document.getElementById("description").value
            .split("\n")
            .map(item => item.trim())
            .filter(Boolean),
        specializationFull: document.getElementById("specializationFull").value.trim()
    };

    const response = await fetch(
        id ? `/api/doctors/${id}` : "/api/doctors",
        {
            method: id ? "PUT" : "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(payload)
        }
    );

    if (response.status === 401) {
        window.location.href = "/?account=admin";
        return;
    }

    if (!response.ok) {
        showMessage("Не удалось сохранить врача", "error");
        return;
    }

    const savedDoctor = await response.json();

    try {
        if (photoInput.files && photoInput.files[0]) {
            await uploadDoctorPhoto(savedDoctor.id);
        }
    } catch (error) {
        closeModal();
        await loadDoctors();
        showMessage(`Врач сохранён, но фото не загрузилось: ${error.message}`, "error");
        return;
    }

    closeModal();
    showMessage(id ? "Данные врача обновлены" : "Врач добавлен");
    await loadDoctors();
}

async function changeStatus(id, action) {
    const doctor = doctors.find(item => item.id === id);
    if (!doctor) return;

    if (action === "deactivate" &&
        !confirm(`Деактивировать врача «${doctor.name}»?\n\nОн останется в истории старых записей, но новые записи к нему будут закрыты.`)) {
        return;
    }

    const response = await fetch(`/api/doctors/${id}/${action}`, {
        method: "POST"
    });

    if (!response.ok) {
        showMessage("Не удалось изменить статус врача", "error");
        return;
    }

    showMessage(action === "deactivate"
        ? "Врач деактивирован"
        : "Врач снова принимает пациентов"
    );

    await loadDoctors();
}

async function openAccount(id) {
    const doctor = doctors.find(item => item.id === id);
    if (!doctor) return;
    accountDoctorId = id;
    document.getElementById("accountDoctorId").value = id;
    document.getElementById("accountDoctorName").textContent = doctor.name;
    document.getElementById("accountPassword").value = "";

    const response = await fetch(`/api/doctors/${id}/account`);
    if (response.status === 401) { window.location.href = "/?account=admin"; return; }
    if (!response.ok) { showMessage("Не удалось загрузить учётную запись", "error"); return; }

    const account = await response.json();
    document.getElementById("accountLogin").value = account.login || `doctor${id}`;
    document.getElementById("accountActive").checked = account.exists ? account.isActive : true;
    accountModal.classList.remove("hidden");
}

function closeAccountModal() {
    accountModal.classList.add("hidden");
    accountDoctorId = null;
}

async function saveAccount(event) {
    event.preventDefault();
    if (!accountDoctorId) return;

    const response = await fetch(`/api/doctors/${accountDoctorId}/account`, {
        method: "PUT",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({
            login: document.getElementById("accountLogin").value.trim(),
            password: document.getElementById("accountPassword").value,
            isActive: document.getElementById("accountActive").checked
        })
    });

    if (response.status === 401) { window.location.href = "/?account=admin"; return; }
    if (response.status === 409) { showMessage("Такой логин уже используется", "error"); return; }
    if (!response.ok) {
        const text = await response.text();
        showMessage(text || "Не удалось сохранить аккаунт", "error");
        return;
    }

    closeAccountModal();
    showMessage("Учётная запись сохранена");
}

function openSchedule(id) {
    const doctor = doctors.find(item => item.id === id);
    if (!doctor) return;
    scheduleDoctorId = id;
    document.getElementById("scheduleDoctorName").textContent = doctor.name;
    unavailabilityDoctorId = id;
    scheduleDays.innerHTML = '<div class="schedule-loading">Загрузка расписания...</div>';
    document.getElementById("unavailabilityList").innerHTML = '<div class="schedule-loading">Загрузка...</div>';
    document.getElementById("unavailabilityStart").value = "";
    document.getElementById("unavailabilityEnd").value = "";
    document.getElementById("unavailabilityReason").value = "";
    scheduleModal.classList.remove("hidden");

    loadUnavailability(id);
    fetch(`/api/doctors/${id}/schedule`)
        .then(async response => {
            if (response.status === 401) { window.location.href = "/?account=admin"; return null; }
            if (!response.ok) throw new Error();
            return response.json();
        })
        .then(schedule => {
            if (!schedule) return;
            scheduleDays.innerHTML = schedule.days.map(day => `
                <div class="schedule-row" data-day="${day.weekday}">
                    <div class="schedule-day">
                        <strong>${escapeHtml(day.dayName)}</strong>
                        <label class="switch-label">
                            <input type="checkbox" class="working-toggle" ${day.isWorking ? "checked" : ""}>
                            <span>Рабочий день</span>
                        </label>
                    </div>
                    <div class="schedule-times">
                        <label>Начало<input type="time" class="start-time" value="${escapeHtml(day.startTime || "09:00")}"></label>
                        <span>—</span>
                        <label>Окончание<input type="time" class="end-time" value="${escapeHtml(day.endTime || "18:00")}"></label>
                    </div>
                </div>
            `).join("");
            updateScheduleInputs();
        })
        .catch(() => showMessage("Не удалось загрузить расписание", "error"));
}

async function loadUnavailability(id) {
    try {
        const response = await fetch(`/api/doctors/${id}/unavailability`);
        if (response.status === 401) { window.location.href = "/?account=admin"; return; }
        if (!response.ok) throw new Error();
        const items = await response.json();
        const list = document.getElementById("unavailabilityList");
        list.innerHTML = items.length ? items.map(item => `
            <div class="unavailability-item">
                <div><strong>${escapeHtml(item.startDate)} — ${escapeHtml(item.endDate)}</strong><span>${escapeHtml(item.reason || "Временно недоступен")}</span></div>
                <button type="button" class="danger-btn delete-unavailability" data-id="${item.id}">Удалить</button>
            </div>`).join("") : '<div class="empty-small">Периодов недоступности нет.</div>';
        list.querySelectorAll(".delete-unavailability").forEach(btn => btn.addEventListener("click", () => deleteUnavailability(Number(btn.dataset.id))));
    } catch (_) { document.getElementById("unavailabilityList").innerHTML = '<div class="empty-small">Не удалось загрузить периоды.</div>'; }
}

async function addUnavailability(event) {
    event.preventDefault();
    if (!unavailabilityDoctorId) return;
    const start = document.getElementById("unavailabilityStart").value;
    const end = document.getElementById("unavailabilityEnd").value;
    if (end < start) { showMessage("Дата окончания не может быть раньше даты начала", "error"); return; }
    const response = await fetch(`/api/doctors/${unavailabilityDoctorId}/unavailability`, {
        method: "POST", headers: {"Content-Type":"application/json"},
        body: JSON.stringify({startDate:start,endDate:end,reason:document.getElementById("unavailabilityReason").value.trim()})
    });
    if (!response.ok) { showMessage("Не удалось добавить период недоступности", "error"); return; }
    document.getElementById("unavailabilityForm").reset();
    await loadUnavailability(unavailabilityDoctorId);
    showMessage("Период недоступности добавлен");
}

async function deleteUnavailability(id) {
    if (!confirm("Удалить этот период недоступности?")) return;
    const response = await fetch(`/api/doctors/${unavailabilityDoctorId}/unavailability/${id}`, {method:"DELETE"});
    if (!response.ok) { showMessage("Не удалось удалить период", "error"); return; }
    await loadUnavailability(unavailabilityDoctorId);
    showMessage("Период удалён");
}

function updateScheduleInputs() {
    document.querySelectorAll(".schedule-row").forEach(row => {
        const working = row.querySelector(".working-toggle").checked;
        row.classList.toggle("day-off", !working);
        row.querySelector(".start-time").disabled = !working;
        row.querySelector(".end-time").disabled = !working;
    });
}

async function saveSchedule() {
    if (!scheduleDoctorId) return;
    const days = [...document.querySelectorAll(".schedule-row")].map(row => ({
        weekday: Number(row.dataset.day),
        isWorking: row.querySelector(".working-toggle").checked,
        startTime: row.querySelector(".start-time").value,
        endTime: row.querySelector(".end-time").value
    }));

    for (const day of days) {
        if (day.isWorking && (!day.startTime || !day.endTime || day.startTime >= day.endTime)) {
            showMessage("Проверьте время начала и окончания рабочего дня", "error");
            return;
        }
    }

    const response = await fetch(`/api/doctors/${scheduleDoctorId}/schedule`, {
        method: "PUT",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify({days})
    });
    if (response.status === 401) { window.location.href = "/?account=admin"; return; }
    if (!response.ok) {
        showMessage("Не удалось сохранить расписание", "error");
        return;
    }
    closeScheduleModal();
    showMessage("Расписание сохранено");
}

function closeScheduleModal() {
    scheduleModal.classList.add("hidden");
    scheduleDoctorId = null;
    unavailabilityDoctorId = null;
}

document.getElementById("addDoctorBtn").addEventListener("click", openCreate);
document.getElementById("closeModal").addEventListener("click", closeModal);
document.getElementById("cancelBtn").addEventListener("click", closeModal);
document.getElementById("doctorForm").addEventListener("submit", saveDoctor);
document.getElementById("searchInput").addEventListener("input", renderDoctors);
document.getElementById("closeScheduleModal").addEventListener("click", closeScheduleModal);
document.getElementById("closeAccountModal").addEventListener("click", closeAccountModal);
document.getElementById("cancelAccountBtn").addEventListener("click", closeAccountModal);
document.getElementById("accountForm").addEventListener("submit", saveAccount);
document.getElementById("cancelScheduleBtn").addEventListener("click", closeScheduleModal);
document.getElementById("saveScheduleBtn").addEventListener("click", saveSchedule);
document.getElementById("unavailabilityForm").addEventListener("submit", addUnavailability);
scheduleDays.addEventListener("change", event => {
    if (event.target.classList.contains("working-toggle")) updateScheduleInputs();
});
document.getElementById("statusFilter").addEventListener("change", renderDoctors);

modal.addEventListener("click", event => {
    if (event.target === modal) closeModal();
});

scheduleModal.addEventListener("click", event => {
    if (event.target === scheduleModal) closeScheduleModal();
});

accountModal.addEventListener("click", event => {
    if (event.target === accountModal) closeAccountModal();
});

document.addEventListener("DOMContentLoaded", loadDoctors);
