(() => {
  const page = document.getElementById("servicesPage");
  const images = ["y1.png","y2.png","y6.png","y5.png","y4.png","y3.png"];
  const esc = (v) => String(v ?? "").replace(/[&<>"']/g, m => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;", "'":"&#039;"}[m]));
  async function load() {
    try {
      const r = await fetch("/api/services");
      if (!r.ok) throw new Error();
      const list = (await r.json()).filter(x => x.isActive);
      const groups = [];
      list.forEach(x => { let g = groups.find(v => v.name === x.category); if (!g) { g={name:x.category,items:[]};groups.push(g); } g.items.push(x); });
      page.innerHTML = groups.length ? groups.map((g,i) => `<div class="service-big-card"><img src="img/${images[i % images.length]}" alt="" class="service-big-img"/><h2 class="service-big-title">${esc(g.name)}</h2><div class="service-buttons">${g.items.map(x => `<a href="index.html#contacts" class="service-link" data-service-id="${x.id}">${esc(x.name)}</a>`).join("")}</div></div>`).join("") : '<div class="services-loading">Активных услуг пока нет.</div>';
    } catch(e) { page.innerHTML='<div class="services-loading">Не удалось загрузить услуги.</div>'; }
  }
  load();
})();
