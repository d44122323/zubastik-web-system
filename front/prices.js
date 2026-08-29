(async () => {
  try {
    const r = await fetch("/api/services");
    if (!r.ok) return;
    const data = await r.json();
    const map = new Map(data.map((x) => [x.name, x.price]));
    document.querySelectorAll(".table-row[data-service]").forEach((row) => {
      const name = row.dataset.service;
      if (!map.has(name)) return;
      const priceEl = row.querySelector(".price");
      if (!priceEl) return;
      const value = map.get(name);
      priceEl.innerHTML =
        '<img src="img/moneyc.png" alt=""> ' +
        (value === 0
          ? "от Бесплатно"
          : "от " + value.toLocaleString("ru-RU") + " ₽");
    });
  } catch (e) {}
})();
