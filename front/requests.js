let requests=[], doctors=[], adminServices=[], currentRequest=null, currentServiceRequest=null, currentFilter="Все", currentSearch="", currentDate="";
const $=s=>document.querySelector(s), drawer=$(".request-drawer"), overlay=$(".overlay");
const esc=v=>String(v??"").replace(/[&<>"']/g,m=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;"}[m]));
async function loadAdminServices(){
 const r=await fetch('/api/services',{credentials:'same-origin'}); if(r.ok) adminServices=await r.json();
}
async function loadDoctors(){
 const r=await fetch("/api/doctors?all=true"); if(!r.ok)return;
 doctors=await r.json(); const sel=$(".doctor-filter"), ds=$("#doctorSelect");
 doctors.forEach(d=>{sel.innerHTML+=`<option value="${d.id}">${esc(d.name)}${d.isActive?"":" (неактивен)"}</option>`;ds.innerHTML+=`<option value="${d.id}">${esc(d.name)}${d.isActive?"":" (неактивен)"}</option>`});
}
async function loadRequests(){const r=await fetch("/api/requests"); if(r.ok){requests=await r.json();renderRequests()}}
function renderRequests(){
 let list=requests.filter(x=>(currentFilter==="Все"||x.status===currentFilter));
 const doc=$(".doctor-filter").value;
 if(doc)list=list.filter(x=>String(x.doctor_id)===doc);
 if(currentDate)list=list.filter(x=>x.appointment_date===currentDate);
 if(currentSearch){const q=currentSearch.toLowerCase();list=list.filter(x=>(x.name||"").toLowerCase().includes(q)||(x.phone||"").includes(q)||String(x.id).includes(q)||(x.doctor_name||"").toLowerCase().includes(q))}
 const table=$(".requests-table");table.innerHTML=`<div class="table-header"><span>№</span><span>Запись</span><span>Пациент</span><span>Врач</span><span>Статус</span></div>`;
 list.sort((a,b)=>(a.appointment_date||"9999").localeCompare(b.appointment_date||"9999")||(a.appointment_time||"").localeCompare(b.appointment_time||""));
 list.forEach(x=>{let c={Новая:"new",Подтверждена:"confirm",Завершена:"completed",Отменена:"cancelled"}[x.status]||"new";
 table.innerHTML+=`<div class="request-row" data-id="${x.id}"><span>${x.id}</span><span>${formatAppointment(x)}</span><span><b>${esc(x.name)}</b><small>${esc(x.phone)}</small></span><span>${esc(x.doctor_name||"Не назначен")}</span><span class="status ${c}">${esc(x.status)}</span></div><div class="table-line"></div>`});
 document.querySelectorAll(".request-row").forEach(r=>r.onclick=()=>openRequest(+r.dataset.id));
}
function formatAppointment(x){if(!x.appointment_date)return"Без даты";return new Date(x.appointment_date+"T00:00:00").toLocaleDateString("ru-RU")+" "+(x.appointment_time||"");}
function serviceNames(value){return String(value||'').split(',').map(x=>x.trim()).filter(Boolean)}
function openAdminServiceModal(){
 if(!currentRequest)return;
 const selected=new Set(serviceNames(currentRequest.services));
 const box=$('#adminServiceOptions');
 box.innerHTML=adminServices.length?adminServices.map(s=>`<div class="admin-service-option"><label><input type="checkbox" value="${s.id}" data-name="${esc(s.name)}" data-price="${s.price}" ${selected.has(s.name)?'checked':''}> ${esc(s.name)}</label><span>${Number(s.price).toLocaleString('ru-RU')} ₽</span></div>`).join(''):'<div>Услуги не загружены.</div>';
 updateAdminServiceTotal(); $('#adminServiceModal').classList.remove('hidden');
}
function closeAdminServiceModal(){$('#adminServiceModal').classList.add('hidden')}
function updateAdminServiceTotal(){let total=0;document.querySelectorAll('#adminServiceOptions input:checked').forEach(x=>total+=Number(x.dataset.price||0));$('#adminServiceTotal').textContent=total.toLocaleString('ru-RU')+' ₽'}
async function saveAdminServices(){
 if(!currentRequest)return; const ids=[...document.querySelectorAll('#adminServiceOptions input:checked')].map(x=>Number(x.value)); const b=$('#adminServiceSave');b.disabled=true;
 try{const r=await fetch('/api/requests/'+currentRequest.id,{method:'PUT',credentials:'same-origin',headers:{'Content-Type':'application/json'},body:JSON.stringify({status:currentRequest.status,service_ids:ids})});if(!r.ok)throw new Error(await r.text());currentRequest=await r.json();closeAdminServiceModal();$('#priceValue').textContent=currentRequest.price?`≈ ${currentRequest.price.toLocaleString('ru-RU')} ₽`:'—';$('#servicesList').innerHTML=serviceNames(currentRequest.services).map(s=>`<p>✔ ${esc(s)}</p>`).join('')||'<p>Не указаны</p>';alert('Услуги сохранены.');loadRequests();}catch(e){alert(e.message||'Не удалось сохранить услуги')}finally{b.disabled=false}
}
async function openRequest(id){
 const r=await fetch(`/api/requests/${id}`);if(!r.ok)return;currentRequest=await r.json();
 $(".request-drawer h2").textContent=`Запись №${currentRequest.id}`;
 $("#patientName").textContent=currentRequest.name;$("#patientPhone").textContent=currentRequest.phone;
 $("#commentValue").textContent=currentRequest.comment||"Без комментариев";
 $("#priceValue").textContent=currentRequest.price?`≈ ${currentRequest.price.toLocaleString("ru-RU")} ₽`:"—";
 $("#servicesList").innerHTML=(currentRequest.services||"").split(", ").filter(Boolean).map(s=>`<p>✔ ${esc(s)}</p>`).join("")||"<p>Не указаны</p>";
 $("#doctorSelect").value=currentRequest.doctor_id||"";
 $("#appointmentDate").value=currentRequest.appointment_date||"";
 document.querySelectorAll('input[name="status"]').forEach(x=>x.checked=x.value===currentRequest.status);
 await loadSlots();
 drawer.classList.add("active");overlay.classList.add("active");
}
async function loadSlots(){
 const did=$("#doctorSelect").value,date=$("#appointmentDate").value,time=$("#appointmentTime");
 time.innerHTML='<option value="">Выберите время</option>';
 if(!did||!date){$("#slotHint").textContent="Выберите врача и дату";return}
 const r=await fetch(`/api/doctors/${did}/availability?date=${date}`);
 if(!r.ok){$("#slotHint").textContent="Не удалось получить расписание";return}
 const data=await r.json(), old=currentRequest?.appointment_time;
 const slots=(data.slots||[]).filter(s=>s.available||s.time===old);
 slots.forEach(s=>time.innerHTML+=`<option value="${s.time}">${s.time}${s.available?"":" — текущая запись"}</option>`);
 if(old)time.value=old;
 $("#slotHint").textContent=data.isWorking?`Свободных слотов: ${slots.filter(s=>s.available).length}`:"В этот день врач не принимает";
}
async function save(statusOverride){
 if(!currentRequest)return;
 const status=statusOverride||document.querySelector('input[name="status"]:checked')?.value||"Новая";
 const body={status,doctor_id:+$("#doctorSelect").value,appointment_date:$("#appointmentDate").value,appointment_time:$("#appointmentTime").value};
 const r=await fetch(`/api/requests/${currentRequest.id}`,{method:"PUT",headers:{"Content-Type":"application/json"},body:JSON.stringify(body)});
 if(!r.ok){alert(await r.text());return}
 alert(status==="Подтверждена"?"Запись подтверждена":status==="Отменена"?"Запись отменена":"Изменения сохранены");
 closeDrawer();loadRequests();
}
async function sendTelegramRequest(){
 if(!currentRequest)return;
 const b=$("#telegramRequestBtn"); b.disabled=true; b.textContent="Отправляем...";
 try{
  const r=await fetch("/api/telegram/request/"+currentRequest.id,{method:"POST"});
  if(!r.ok)throw new Error(await r.text());
  alert("Сообщение отправлено пациенту в Telegram.");
 }catch(e){alert(e.message||"Не удалось отправить сообщение в Telegram");}
 finally{b.disabled=false;b.textContent="📨 Telegram";}
}
function closeDrawer(){drawer.classList.remove("active");overlay.classList.remove("active");currentRequest=null}
$("#telegramRequestBtn").onclick=sendTelegramRequest;
$(".close-btn").onclick=closeDrawer;overlay.onclick=closeDrawer;
$("#doctorSelect").onchange=loadSlots;$("#appointmentDate").onchange=loadSlots;
$(".save-btn").onclick=()=>save();$(".quick-confirm").onclick=()=>save("Подтверждена");$(".quick-cancel").onclick=()=>save("Отменена");$(".quick-done").onclick=()=>save("Завершена");
$(".search-input").oninput=e=>{currentSearch=e.target.value.trim();renderRequests()};
$(".date-filter").onchange=e=>{currentDate=e.target.value;renderRequests()};
$(".doctor-filter").onchange=renderRequests;
document.querySelectorAll(".status-filter button").forEach(b=>b.onclick=()=>{currentFilter=b.dataset.status;document.querySelectorAll(".status-filter button").forEach(x=>x.classList.remove("active"));b.classList.add("active");renderRequests()});
document.addEventListener("keydown",e=>{if(e.key==="Escape")closeDrawer()});
document.getElementById('addServicesBtn').onclick=openAdminServiceModal;
document.getElementById('adminServiceClose').onclick=closeAdminServiceModal;document.getElementById('adminServiceCancel').onclick=closeAdminServiceModal;document.getElementById('adminServiceSave').onclick=saveAdminServices;document.getElementById('adminServiceOptions').addEventListener('change',updateAdminServiceTotal);document.getElementById('adminServiceModal').onclick=e=>{if(e.target.id==='adminServiceModal')closeAdminServiceModal()};
document.addEventListener("DOMContentLoaded",async()=>{await loadAdminServices();await loadDoctors();await loadRequests()});
