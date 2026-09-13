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
const servicesEl = document.getElementById("services"),
  searchEl = document.getElementById("search"),
  categoryEl = document.getElementById("category");
function categories() {
  return [...new Set(all.map((x) => x.category).filter(Boolean))].sort((a, b) =>
    a.localeCompare(b, "ru"),
  );
}
function fillCategories() {
  categoryEl.innerHTML =
    '<option value="">Все разделы</option>' +
    categories()
      .map((x) => `<option value="${esc(x)}">${esc(x)}</option>`)
      .join("");
}
function row(x) {
  return `<div class="service-row" data-id="${x.id}"><input class="category-input" value="${esc(x.category)}" placeholder="Раздел"><input class="name-input" value="${esc(x.name)}" placeholder="Название услуги"><input class="price-input" type="number" min="0" value="${x.price}"><input class="order-input" type="number" min="0" value="${x.sortOrder}"><label class="active-toggle"><input class="active-input" type="checkbox" ${x.isActive ? "checked" : ""}> Активна</label><button class="save-btn">Сохранить</button></div>`;
}
function render() {
  const q = searchEl.value.trim().toLowerCase(),
    c = categoryEl.value;
  const list = all.filter(
    (x) =>
      (!c || x.category === c) &&
      (!q ||
        x.name.toLowerCase().includes(q) ||
        x.category.toLowerCase().includes(q)),
  );
  servicesEl.innerHTML =
    `<div class="service-head"><span>Раздел</span><span>Название</span><span>Цена, ₽</span><span>Порядок</span><span>Статус</span><span></span></div>` +
    (list.length
      ? list.map(row).join("")
      : '<div class="message">Ничего не найдено.</div>');
  servicesEl
    .querySelectorAll(".save-btn")
    .forEach((b) => (b.onclick = () => saveRow(b.closest(".service-row"))));
}
async function saveRow(r) {
  const id = r.dataset.id,
    item = all.find((x) => String(x.id) === id);
  const payload = {
    category: r.querySelector(".category-input").value.trim(),
    name: r.querySelector(".name-input").value.trim(),
    price: Number(r.querySelector(".price-input").value),
    sortOrder: Number(r.querySelector(".order-input").value),
    isActive: r.querySelector(".active-input").checked,
  };
  if (
    !payload.category ||
    !payload.name ||
    !Number.isInteger(payload.price) ||
    payload.price < 0 ||
    !Number.isInteger(payload.sortOrder) ||
    payload.sortOrder < 0
  ) {
    alert("Проверьте раздел, название, цену и порядок.");
    return;
  }
  const b = r.querySelector(".save-btn");
  b.disabled = true;
  try {
    const res = await fetch("/api/admin/services/" + id, {
      method: "PUT",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!res.ok) throw 0;
    Object.assign(item, await res.json());
    fillCategories();
    render();
  } catch (e) {
    alert(
      "Не удалось сохранить услугу. Возможно, такое название уже используется.",
    );
  } finally {
    b.disabled = false;
  }
}
async function create() {
  const payload = {
    category: document.getElementById("newCategory").value.trim(),
    name: document.getElementById("newName").value.trim(),
    price: Number(document.getElementById("newPrice").value || 0),
    sortOrder: Number(document.getElementById("newOrder").value || 0),
    isActive: document.getElementById("newActive").checked,
  };
  if (
    !payload.category ||
    !payload.name ||
    payload.price < 0 ||
    payload.sortOrder < 0
  ) {
    alert("Заполните раздел и название.");
    return;
  }
  try {
    const r = await fetch("/api/admin/services", {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!r.ok) throw 0;
    all.push(await r.json());
    ["newCategory", "newName", "newPrice", "newOrder"].forEach(
      (id) => (document.getElementById(id).value = ""),
    );
    document.getElementById("newActive").checked = true;
    fillCategories();
    render();
  } catch (e) {
    alert("Не удалось добавить услугу. Проверьте название и данные.");
  }
}
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
    servicesEl.innerHTML =
      '<div class="message error">Не удалось загрузить услуги.</div>';
  }
}
searchEl.oninput = render;
categoryEl.onchange = render;
document.getElementById("addServiceBtn").onclick = create;
document.addEventListener("DOMContentLoaded", load);
