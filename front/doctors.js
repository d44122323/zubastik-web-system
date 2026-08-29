async function loadDoctors() {
  const homepage = document.querySelector("#homepage-specialists");
  const container = homepage || document.querySelector(".doctors-wrapper");
  if (!container) return;

  try {
    const response = await fetch("/api/doctors", {
      credentials: "same-origin",
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const doctors = await response.json();
    if (!Array.isArray(doctors)) throw new Error("Некорректный ответ API");

    if (homepage) {
      homepage.innerHTML = "";
      const carousel = document.querySelector("#homepage-specialists-carousel");
      const prev = carousel?.querySelector(".specialists-arrow-prev");
      const next = carousel?.querySelector(".specialists-arrow-next");

      if (doctors.length === 0) {
        homepage.innerHTML =
          '<p class="specialists-loading">Специалисты пока не добавлены.</p>';
        if (prev) prev.hidden = true;
        if (next) next.hidden = true;
        return;
      }

      doctors.forEach((doctor) => {
        const card = document.createElement("article");
        card.className = "specialist-card";
        const description = Array.isArray(doctor.description)
          ? doctor.description
          : [];
        const bio = description[0] || doctor.specialization || "";
        const photo = doctor.photo || "img/logo.png";
        card.innerHTML = `
                    <img src="${escapeHtml(photo)}" class="specialist-img" alt="${escapeHtml(doctor.name)}">
                    <h3 class="specialist-name">${escapeHtml(doctor.name)}</h3>
                    <p class="specialist-position">${escapeHtml(doctor.position || doctor.specialization || "")}</p>
                    <p class="specialist-bio">${escapeHtml(bio)}</p>
                    <a href="doctor.html?id=${encodeURIComponent(doctor.id)}" class="btn-specialist">Подробнее о враче</a>
                `;
        homepage.appendChild(card);
      });

      let page = 0;
      const getPerPage = () =>
        window.innerWidth <= 768 ? 1 : window.innerWidth <= 1100 ? 2 : 3;
      const renderCarousel = () => {
        const perPage = getPerPage();
        const pages = Math.max(1, Math.ceil(doctors.length / perPage));
        page = Math.min(page, pages - 1);
        const offset = page * perPage;
        Array.from(homepage.children).forEach((card, index) => {
          if (!card.classList.contains("specialist-card")) return;
          card.style.display =
            index >= offset && index < offset + perPage ? "flex" : "none";
        });
        if (prev) {
          prev.hidden = pages <= 1;
          prev.disabled = page === 0;
        }
        if (next) {
          next.hidden = pages <= 1;
          next.disabled = page === pages - 1;
        }
      };

      prev?.addEventListener("click", () => {
        if (page > 0) {
          page--;
          renderCarousel();
        }
      });
      next?.addEventListener("click", () => {
        const perPage = getPerPage();
        if (page < Math.ceil(doctors.length / perPage) - 1) {
          page++;
          renderCarousel();
        }
      });
      window.addEventListener("resize", renderCarousel);
      renderCarousel();
      return;
    }
    const oldCards = container.querySelectorAll(".doctor-card");
    oldCards.forEach((card) => card.remove());

    doctors.forEach((doctor) => {
      const card = document.createElement("article");
      card.className = "doctor-card";
      const description = Array.isArray(doctor.description)
        ? doctor.description
        : [];
      const shortDescription = description.slice(0, 3);
      const hidden = description.slice(3);
      card.innerHTML = `
                <div class="doctor-photo">
                    <img src="${escapeHtml(doctor.photo)}" alt="${escapeHtml(doctor.name)}">
                </div>
                <div class="doctor-info">
                    <div class="doctor-main">
                        <h2 class="doctor-name">${escapeHtml(doctor.name)}</h2>
                        <div class="doctor-text">
                            ${shortDescription.map((p) => `<p>${escapeHtml(p)}</p>`).join("")}
                            ${hidden.length ? `<div class="hidden-text">${hidden.map((p) => `<p>${escapeHtml(p)}</p>`).join("")}</div>` : ""}
                        </div>
                        <div class="doctor-actions">
                            ${hidden.length ? '<button class="toggle-btn">Читать полностью</button>' : ""}
                            <a href="doctor.html?id=${encodeURIComponent(doctor.id)}" class="doctor-btn">Подробнее о враче</a>
                        </div>
                    </div>
                    <div class="doctor-badges">
                        <div class="doctor-badge">
                            <span class="badge-title">Должность</span>
                            <p>${escapeHtml(doctor.position)}</p>
                        </div>
                        <div class="doctor-badge small">
                            <span class="badge-title">Стаж работы</span>
                            <p>${escapeHtml(doctor.experience)}</p>
                        </div>
                    </div>
                </div>`;
      container.appendChild(card);
    });

    container.querySelectorAll(".toggle-btn").forEach((button) => {
      button.addEventListener("click", () => {
        const text = button
          .closest(".doctor-main")
          .querySelector(".doctor-text");
        text.classList.toggle("open");
        button.textContent = text.classList.contains("open")
          ? "Свернуть"
          : "Читать полностью";
      });
    });
  } catch (error) {
    console.error("Ошибка загрузки специалистов:", error);
    if (homepage) {
      homepage.innerHTML =
        '<p class="specialists-loading">Не удалось загрузить специалистов.</p>';
    }
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

document.addEventListener("DOMContentLoaded", loadDoctors);
