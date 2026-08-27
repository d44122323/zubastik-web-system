const patientEl=document.getElementById('patient');
const dateEl=document.getElementById('date');
const timeEl=document.getElementById('time');
const errorEl=document.getElementById('error');
let doctor=null, services=[];
function esc(x){return String(x??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}
function showError(msg){errorEl.textContent=msg||'Произошла ошибка';errorEl.classList.remove('hidden')}
function clearError(){errorEl.textContent='';errorEl.classList.add('hidden')}
function today(){const d=new Date();const local=new Date(d.getTime()-d.getTimezoneOffset()*60000);return local.toISOString().slice(0,10)}
async function loadServices(){
 const r=await fetch('/api/services',{credentials:'include'});
 if(!r.ok)return;
 services=await r.json();
 const box=document.getElementById('serviceOptions');
 box.innerHTML=services.map(s=>`<label class="service-option"><input type="checkbox" value="${s.id}" data-name="${esc(s.name)}" data-price="${s.price}"><span>${esc(s.name)}</span><b>${Number(s.price).toLocaleString('ru-RU')} ₽</b></label>`).join('');
 box.addEventListener('change',updateServiceTotal);
 updateServiceTotal();
}
function updateServiceTotal(){let total=0;document.querySelectorAll('#serviceOptions input:checked').forEach(x=>total+=Number(x.dataset.price||0));document.getElementById('serviceTotal').textContent=total.toLocaleString('ru-RU')+' ₽'}

async function init(){
 const me=await fetch('/api/doctor/me',{credentials:'include'});if(!me.ok){location='/?account=doctor';return}doctor=await me.json();document.getElementById('doctorName').textContent=doctor.name;dateEl.min=today();await loadServices();
 const r=await fetch('/api/doctor/patients',{credentials:'include'});if(!r.ok){showError('Не удалось загрузить пациентов');return}const patients=await r.json();
 patientEl.innerHTML='<option value="">Выберите пациента</option>'+patients.map(p=>`<option value="${p.id}">${esc(p.name)} — ${esc(p.phone)}</option>`).join('');
}
async function loadSlots(){
 clearError();timeEl.disabled=true;timeEl.innerHTML='<option value="">Загрузка времени...</option>';
 if(!dateEl.value){timeEl.innerHTML='<option value="">Сначала выберите дату</option>';return}
 const r=await fetch(`/api/doctor/schedule/availability?date=${encodeURIComponent(dateEl.value)}`,{credentials:'include'});
 if(!r.ok){timeEl.innerHTML='<option value="">Не удалось загрузить время</option>';showError('Не удалось получить расписание врача');return}
 const data=await r.json();const slots=(data.slots||[]).filter(x=>x.available);
 if(!data.isWorking||!slots.length){timeEl.innerHTML='<option value="">Свободных слотов нет</option>';return}
 timeEl.innerHTML='<option value="">Выберите время</option>'+slots.map(x=>`<option value="${esc(x.time)}">${esc(x.time)}</option>`).join('');timeEl.disabled=false;
}
dateEl.addEventListener('change',loadSlots);
document.getElementById('form').addEventListener('submit',async e=>{
 e.preventDefault();clearError();
 const selected=[...document.querySelectorAll('#serviceOptions input:checked')];const body={patient_id:Number(patientEl.value),service:selected.map(x=>x.dataset.name).join(', '),date:dateEl.value,time:timeEl.value,comment:document.getElementById('comment').value.trim(),price:selected.reduce((sum,x)=>sum+Number(x.dataset.price||0),0)};
 if(!body.patient_id||!body.service||!body.date||!body.time){showError('Заполните пациента, услугу, дату и время');return}
 const r=await fetch('/api/doctor/appointments/create',{method:'POST',credentials:'include',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
 if(!r.ok){showError(await r.text());await loadSlots();return}
 const created=await r.json();alert(`Запись №${created.id} создана и подтверждена.`);location='/doctor/requests.html';
});
document.getElementById('logout').onclick=async()=>{await fetch('/doctor/logout',{method:'POST',credentials:'include'});location='/?account=doctor'};
init();
