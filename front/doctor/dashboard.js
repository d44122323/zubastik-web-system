async function loadDoctor() {
  const response = await fetch("/api/doctor/me", { credentials: "include" });
  if (!response.ok) {
    window.location.replace("/?account=doctor");
    return;
  }

  const doctor = await response.json();
  document.getElementById("doctorName").textContent = doctor.name;
  document.getElementById("welcome").textContent =
    `Добро пожаловать, ${doctor.name}`;
  document.getElementById("doctorId").textContent = `#${doctor.id}`;
}

async function logout() {
  await fetch("/doctor/logout", { method: "POST", credentials: "include" });
  window.location.replace("/?account=doctor");
}

document.getElementById("logout").addEventListener("click", logout);
document.querySelectorAll("[data-soon]").forEach((link) => {
  link.addEventListener("click", (event) => {
    event.preventDefault();
    alert(`${link.dataset.soon}: следующий этап разработки`);
  });
});
document.addEventListener("DOMContentLoaded", loadDoctor);
