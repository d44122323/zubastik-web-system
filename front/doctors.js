document.querySelectorAll(".doctor-card").forEach(card => {
    const button = card.querySelector(".toggle-btn");
    const text = card.querySelector(".doctor-text");
    if (!button || !text) return;
    button.addEventListener("click", () => {
        text.classList.toggle("open");
        if (text.classList.contains("open")) {
            button.textContent = "Свернуть";
        } else {
            button.textContent = "Читать полностью";
        }
    });
});