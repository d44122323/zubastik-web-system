let patients=[]; let currentPatient=null; let activeFilter='Все';
const $=s=>document.querySelector(s);

document.addEventListener('DOMContentLoaded',()=>{ bindEvents(); loadPatients(); });
function bindEvents(){
 $('#searchInput').addEventListener('input',renderFiltered);
 $('#addPatientBtn').addEventListener('click',openCreate);
 $('#closeBtn').addEventListener('click',closeDrawer); $('#overlay').addEventListener('click',closeDrawer);
 $('#editBtn').addEventListener('click',openEdit); $('#cancelEditBtn').addEventListener('click',showView); $('#savePatientBtn').addEventListener('click',savePatient);
 document.querySelectorAll('.filter-btn').forEach(b=>b.addEventListener('click',()=>{activeFilter=b.dataset.filter;document.querySelectorAll('.filter-btn').forEach(x=>x.classList.remove('active'));b.classList.add('active');renderFiltered();}));
 document.addEventListener('keydown',e=>{if(e.key==='Escape')closeDrawer();});
}
async function loadPatients(){
 try{const r=await fetch('/api/patients'); if(!r.ok) throw new Error(); patients=await r.json(); renderFiltered();}
 catch(e){console.error(e); $('#patientsTable').innerHTML='<div class="empty">Не удалось загрузить пациентов.</div>';}
}
function renderFiltered(){
 const q=$('#searchInput').value.trim().toLowerCase();
 let list=patients.filter(p=>(!q || (p.name||'').toLowerCase().includes(q) || (p.phone||'').toLowerCase().includes(q)));
 if(activeFilter==='Новые') list=list.filter(p=>(p.visits||0)<=1);
 if(activeFilter==='Постоянные') list=list.filter(p=>(p.visits||0)>=3);
 if(activeFilter==='Были на приёме') list=list.filter(p=>(p.visits||0)>0);
 const box=$('#patientsTable'); box.innerHTML='';
 if(!list.length){box.innerHTML='<div class="empty">Пациенты не найдены.</div>';return;}
 list.forEach(p=>{
  const row=document.createElement('div'); row.className='table-row'; row.dataset.id=p.id;
  row.innerHTML=`<span class="muted">#${p.id}</span><strong>${esc(p.name||'—')}</strong><span>${esc(p.phone||'—')}</span><span>${esc(p.lastDoctorName||'Не назначен')}</span><span>${formatDateTime(p.lastVisit)}</span><span>${esc(p.lastService||'—')}</span>`;
  row.addEventListener('click',()=>openPatient(p.id)); box.appendChild(row);
 });
}
async function openPatient(id){
 try{const r=await fetch(`/api/patients/${id}`,{credentials:'same-origin'});if(!r.ok){const text=await r.text();throw new Error(text||('HTTP '+r.status));}currentPatient=await r.json();showView();fillView();openDrawer();}
 catch(e){console.error(e);alert('Не удалось загрузить карточку пациента'+(e.message?'\n\n'+e.message:''));}
}
function fillView(){
 const p=currentPatient; $('#drawerTitle').textContent=`${p.name||'Пациент'} #${p.id}`;
 $('#patientNameView').textContent=p.name||'—'; $('#patientPhoneView').textContent=p.phone||'—'; $('#patientDoctorView').textContent=p.lastDoctorName||'Не назначен';
 $('#visitsView').textContent=p.visits||0; $('#lastVisitView').textContent=formatDateTime(p.lastVisit); $('#totalView').textContent=`${p.total||0} ₽`; $('#commentView').textContent=p.comment||'—';
 renderHistory(p.history||[]); renderMedical(p.medical_records||[]);
}
function renderHistory(items){
 const box=$('#history'); box.innerHTML=''; if(!items.length){box.innerHTML='<div class="empty small">История заявок отсутствует.</div>';return;}
 items.forEach(v=>{const el=document.createElement('div');el.className='history-item';el.innerHTML=`<div><strong>${esc(v.service||'Запись')}</strong><div class="history-meta">${formatAppointment(v)} · ${esc(v.doctorName||'Не назначен')}</div></div><div class="history-right"><b>${v.price||0} ₽</b><span class="status">${esc(v.status||'—')}</span></div>`;box.appendChild(el);});
}
function renderMedical(records){
 const box=$('#medicalRecords');box.innerHTML='';if(!records.length){box.innerHTML='<div class="empty small">Медицинских записей пока нет.</div>';return;}
 records.forEach(m=>{const el=document.createElement('article');el.className='medical-card';
  const files=(m.files||[]).map(f=>`<a href="/api/admin/medical-files/${f.id}" target="_blank" rel="noopener">📄 ${esc(f.file_name)}</a>`).join('');
  el.innerHTML=`<div class="medical-head"><strong>${esc(m.appointment_date||'')} ${esc(m.appointment_time||'')}</strong><span>Запись #${m.appointment_id}</span></div>
  ${field('Жалобы',m.complaints)}${field('Диагноз',m.diagnosis)}${field('Лечение',m.treatment)}${field('Рекомендации',m.recommendations)}
  ${files?`<div class="medical-files"><span>Документы</span>${files}</div>`:''}`;box.appendChild(el);});
}
function field(title,value){return value?`<div class="medical-field"><span>${title}</span><p>${esc(value)}</p></div>`:'';}
function openCreate(){currentPatient=null;$('#drawerTitle').textContent='Добавить пациента';$('#patientName').value='';$('#patientPhone').value='';$('#patientComment').value='';$('#editBtn').style.display='none';$('#drawerView').classList.add('hidden');$('#drawerForm').classList.remove('hidden');openDrawer();}
function openEdit(){if(!currentPatient)return;$('#patientName').value=currentPatient.name||'';$('#patientPhone').value=currentPatient.phone||'';$('#patientComment').value=currentPatient.comment||'';$('#drawerView').classList.add('hidden');$('#drawerForm').classList.remove('hidden');}
function showView(){if(!currentPatient)return;$('#drawerView').classList.remove('hidden');$('#drawerForm').classList.add('hidden');$('#editBtn').style.display='inline-flex';}
async function savePatient(){
 const data={name:$('#patientName').value.trim(),phone:$('#patientPhone').value.trim(),comment:$('#patientComment').value.trim()}; if(!data.name||!data.phone){alert('Заполните имя и телефон');return;}
 try{const url=currentPatient?`/api/patients/${currentPatient.id}`:'/api/patients';const r=await fetch(url,{method:currentPatient?'PUT':'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(data)});if(!r.ok)throw new Error();closeDrawer();await loadPatients();}
 catch(e){console.error(e);alert('Не удалось сохранить пациента');}
}
function openDrawer(){$('#drawer').classList.add('active');$('#overlay').classList.add('active');}
function closeDrawer(){$('#drawer').classList.remove('active');$('#overlay').classList.remove('active');}
function formatDateTime(v){if(!v)return '—';const d=new Date(v);if(Number.isNaN(d.getTime()))return '—';return d.toLocaleString('ru-RU',{day:'2-digit',month:'2-digit',year:'numeric',hour:'2-digit',minute:'2-digit'});}
function formatAppointment(v){if(v.appointmentDate)return `${v.appointmentDate.split('-').reverse().join('.')} ${v.appointmentTime||''}`.trim();return formatDateTime(v.date);}
function esc(v){return String(v??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#039;'}[c]));}
