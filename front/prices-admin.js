let all = [];
const esc = (v) =>
  String(v ?? "").replace(
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
async function load() {
  try {
    const r = await fetch("/api/admin/services", {
      credentials: "same-origin",
    });
    if (!r.ok) throw 0;
    all = await r.json();
    fillCategories();
    render();
  } catch (e) {
    services.innerHTML =
      '<div class="message error">Не удалось загрузить цены.</div>';
  }
}
function fillCategories() {
  const set = [...new Set(all.map((x) => x.category).filter(Boolean))];
  category.innerHTML =
    '<option value="">Все категории</option>' +
    set.map((x) => `<option>${esc(x)}</option>`).join("");
}
function render() {
  const q = search.value.trim().toLowerCase(),
    c = category.value;
  const list = all.filter(
    (x) => (!c || x.category === c) && (!q || x.name.toLowerCase().includes(q)),
  );
  services.innerHTML =
    '<div class="service-head"><span>Категория</span><span>Услуга</span><span>Цена, ₽</span><span></span></div>' +
    (list.length
      ? list
          .map(
            (x) =>
              `<div class="service-row" data-id="${x.id}"><div>${esc(x.category)}</div><div>${esc(x.name)}</div><div><input class="price-input" type="number" min="0" value="${x.price}"></div><div><button class="save-btn">Сохранить</button></div></div>`,
          )
          .join("")
      : '<div class="message">Ничего не найдено.</div>');
  services.querySelectorAll(".save-btn").forEach((btn) =>
    btn.addEventListener("click", async () => {
      const row = btn.closest(".service-row"),
        id = row.dataset.id,
        item = all.find((x) => String(x.id) === id),
        price = Number(row.querySelector(".price-input").value);
      if (!Number.isInteger(price) || price < 0) {
        alert("Введите корректную цену");
        return;
      }
      btn.disabled = true;
      try {
        const r = await fetch("/api/admin/services/" + id, {
          method: "PUT",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            category: item.category,
            name: item.name,
            price,
            isActive: item.isActive,
          }),
        });
        if (!r.ok) throw 0;
        const updated = await r.json();
        Object.assign(item, updated);
        btn.textContent = "Сохранено";
        setTimeout(() => (btn.textContent = "Сохранить"), 900);
      } catch (e) {
        alert("Не удалось сохранить цену");
      } finally {
        btn.disabled = false;
      }
    }),
  );
}
search.addEventListener("input", render);
category.addEventListener("change", render);
document.addEventListener("DOMContentLoaded", load);
