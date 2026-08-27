let allRequests = [];
let services = [];
let currentServiceRequestId = null;
let currentStatus = "";
let currentDate = "";

async function me() {
  const r = await fetch('/api/doctor/me', { credentials: 'include' });
  if (!r.ok) { location = '/?account=doctor'; return null; }
  const d = await r.json();
  document.getElementById('doctorName').textContent = d.name;
  return d;
}

async function loadServices() {
  const r = await fetch('/api/services', {credentials:'include'});
  if (!r.ok) return;
  services = await r.json();
}

function serviceNames(value){ return String(value||'').split(',').map(x=>x.trim()).filter(Boolean); }

function openServiceModal(id) {
  const request = allRequests.find(x => Number(x.id) === Number(id));
  if (!request) return;
  currentServiceRequestId = Number(id);
  const selected = new Set(serviceNames(request.services));
  const box = document.getElementById('serviceOptions');
  box.innerHTML = services.length ? services.map(s => `<div class="service-option"><label><input type="checkbox" value="${s.id}" data-name="${esc(s.name)}" data-price="${s.price}" ${selected.has(s.name)?'checked':''}> <span>${esc(s.name)}</span></label><span class="service-option-price">${Number(s.price).toLocaleString('ru-RU')} ₽</span></div>`).join('') : '<div class="empty">Услуги не загружены.</div>';
  updateServiceTotal();
  document.getElementById('serviceModal').classList.remove('hidden');
}
function closeServiceModal(){ currentServiceRequestId=null; document.getElementById('serviceModal').classList.add('hidden'); }
function updateServiceTotal(){ let total=0; document.querySelectorAll('#serviceOptions input[type=checkbox]:checked').forEach(x=>total+=Number(x.dataset.price||0)); document.getElementById('serviceTotal').textContent=total.toLocaleString('ru-RU')+' ₽'; }
async function saveServices(){
  if(!currentServiceRequestId)return;
  const ids=[...document.querySelectorAll('#serviceOptions input[type=checkbox]:checked')].map(x=>Number(x.value));
  const b=document.getElementById('serviceSave'); b.disabled=true;
  try{
    const r=await fetch('/api/doctor/requests/'+currentServiceRequestId,{method:'PUT',credentials:'include',headers:{'Content-Type':'application/json'},body:JSON.stringify({service_ids:ids})});
    if(!r.ok) throw new Error(await r.text());
    const updated=await r.json();
    const i=allRequests.findIndex(x=>Number(x.id)===currentServiceRequestId); if(i>=0) allRequests[i]=updated;
    closeServiceModal(); applyFilters();
    alert('Услуги записи сохранены.');
  }catch(e){alert(e.message||'Не удалось сохранить услуги');}
  finally{b.disabled=false;}
}

async function load() {
  // Always load the complete set first. Filters are applied locally so switching
  // tabs never filters an already filtered result set.
  const r = await fetch('/api/doctor/requests', { credentials: 'include' });
  if (!r.ok) { location = '/?account=doctor'; return; }
  allRequests = await r.json();
  applyFilters();
}

function localToday() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`;
}

function applyFilters() {
  const search = document.getElementById('search').value.trim().toLowerCase();
  const date = currentDate;
  let result = [...allRequests];

  if (currentStatus) {
    result = result.filter(x => x.status === currentStatus);
  }
  if (date) {
    result = result.filter(x => String(x.appointment_date || '') === date);
  }
  if (search) {
    result = result.filter(x => {
      const haystack = [x.name, x.phone, x.id, x.services, x.appointment_date, x.appointment_time]
        .map(v => String(v ?? '').toLowerCase()).join(' ');
      return haystack.includes(search);
    });
  }

  render(result);
}

function esc(x) {
  return String(x ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}

function render(items) {
  const box = document.getElementById('list');
  if (!items.length) {
    box.innerHTML = '<div class="empty">Заявок по выбранным условиям нет.</div>';
    return;
  }
  box.innerHTML = items.map(x => {
    let actions = '';
    if (x.status === 'Новая') actions = '<button class="primary" data-action="Подтверждена" data-id="'+x.id+'">Подтвердить</button><button class="danger" data-action="Отменена" data-id="'+x.id+'">Отклонить</button>';
    if (x.status === 'Подтверждена') actions = '<a class="primary action-link" href="/doctor/medical-record.html?appointment_id='+x.id+'">Завершить приём</a><button class="telegram-send" data-tg-id="'+x.id+'">📨 Telegram</button><button class="danger" data-action="Отменена" data-id="'+x.id+'">Отменить</button>';
    return '<article class="request"><div><h3>'+esc(x.name)+'</h3><div class="meta">'+esc(x.services||'Услуга не указана')+' '+(x.services?'<button type="button" class="service-edit" data-service-id="'+x.id+'">Изменить услуги</button>':'<button type="button" class="service-edit" data-service-id="'+x.id+'">+ Добавить услуги</button>')+'</div><div class="meta">'+esc(x.appointment_date||'Дата не назначена')+' · '+esc(x.appointment_time||'Время не назначено')+'</div><div class="meta">'+esc(x.phone)+'</div><span class="status">'+esc(x.status)+'</span></div><div class="actions">'+actions+'</div></article>';
  }).join('');
  box.querySelectorAll('[data-action]').forEach(b => b.onclick = () => changeStatus(b.dataset.id, b.dataset.action));
  box.querySelectorAll('[data-service-id]').forEach(b => b.onclick = e => { e.stopPropagation(); openServiceModal(b.dataset.serviceId); });
  box.querySelectorAll('[data-tg-id]').forEach(b => b.onclick = async () => { b.disabled=true; try { const r=await fetch('/api/doctor/telegram/request/'+b.dataset.tgId,{method:'POST',credentials:'include'}); if(!r.ok) throw new Error(await r.text()); alert('Сообщение отправлено пациенту в Telegram.'); } catch(e){ alert(e.message||'Не удалось отправить сообщение'); } finally { b.disabled=false; } });
}

async function changeStatus(id, status) {
  if (!confirm(status === 'Подтверждена' ? 'Подтвердить заявку?' : status === 'Отменена' ? 'Отклонить заявку?' : 'Завершить приём?')) return;
  const r = await fetch('/api/doctor/requests/' + id, {
    method: 'PUT', headers: {'Content-Type':'application/json'}, credentials:'include',
    body: JSON.stringify({status})
  });
  if (!r.ok) { alert(await r.text()); return; }
  await load();
}

document.getElementById('serviceOptions').addEventListener('change', updateServiceTotal);
document.getElementById('serviceModalClose').onclick = closeServiceModal;
document.getElementById('serviceCancel').onclick = closeServiceModal;
document.getElementById('serviceSave').onclick = saveServices;
document.getElementById('serviceModal').onclick = e => { if(e.target.id === 'serviceModal') closeServiceModal(); };

document.getElementById('logout').onclick = async () => {
  await fetch('/doctor/logout', {method:'POST', credentials:'include'});
  location = '/?account=doctor';
};

document.querySelectorAll('#tabs button[data-status]').forEach(b => b.onclick = () => {
  currentStatus = b.dataset.status;
  currentDate = '';
  document.getElementById('date').value = '';
  document.querySelectorAll('#tabs button').forEach(x => x.classList.remove('active'));
  b.classList.add('active');
  applyFilters();
});

document.getElementById('today').onclick = () => {
  currentDate = localToday();
  currentStatus = '';
  document.getElementById('date').value = currentDate;
  document.querySelectorAll('#tabs button').forEach(x => x.classList.remove('active'));
  document.getElementById('today').classList.add('active');
  applyFilters();
};

document.getElementById('clear').onclick = () => {
  document.getElementById('search').value = '';
  document.getElementById('date').value = '';
  currentDate = '';
  currentStatus = '';
  document.querySelectorAll('#tabs button').forEach(x => x.classList.remove('active'));
  document.querySelector('#tabs button[data-status=""]').classList.add('active');
  applyFilters();
};

document.getElementById('date').onchange = e => {
  currentDate = e.target.value;
  currentStatus = '';
  document.querySelectorAll('#tabs button').forEach(x => x.classList.remove('active'));
  applyFilters();
};

document.getElementById('search').oninput = () => {
  clearTimeout(window.t);
  window.t = setTimeout(applyFilters, 250);
};

Promise.all([me(), loadServices()]).then(() => load());
