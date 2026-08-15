let drawerMode = "view";
let patients = [];
let currentPatient = null;
let drawer;
let overlay;
let closeBtn;
let addPatientBtn;
let savePatientBtn;
let editBtn;
let drawerView;
let drawerForm;
let drawerTitle;
let searchInput;
let filterButtons;
document.addEventListener(
"DOMContentLoaded",
()=>{
drawer =
document.querySelector(".patient-drawer");
overlay =
document.querySelector(".overlay");
closeBtn =
document.querySelector(".close-btn");
addPatientBtn =
document.querySelector(".add-patient-btn");
savePatientBtn =
document.querySelector("#savePatientBtn");
editBtn =
document.querySelector(".edit-btn");
drawerView =
document.querySelector(".drawer-view");
drawerForm =
document.querySelector(".drawer-form");
drawerTitle =
document.querySelector(".drawer-main-title");
searchInput =
document.querySelector(".search-input");
filterButtons =
document.querySelectorAll(".filter-btn");
initEvents();
loadPatients();
});
function initEvents(){
if(addPatientBtn){
addPatientBtn.addEventListener(
"click",
openCreatePatient
);
}
if(editBtn){
editBtn.addEventListener(
"click",
editPatientMode
);
}
if(closeBtn){
closeBtn.addEventListener(
"click",
closeDrawer
);
}
if(overlay){
overlay.addEventListener(
"click",
closeDrawer
);
}
if(savePatientBtn){
savePatientBtn.addEventListener(
"click",
savePatient
);
}
if(searchInput){
searchInput.addEventListener(
"input",
searchPatients
);
}
filterButtons.forEach(button=>{
button.addEventListener(
"click",
()=>{
filterButtons.forEach(btn=>{
btn.classList.remove("active");
});
button.classList.add("active");
filterPatients(
button.textContent.trim()
);
});
});
document.addEventListener(
"keydown",
(e)=>{
if(e.key==="Escape"){
closeDrawer();
}
});
}
async function loadPatients(){
try{
const response =
await fetch("/api/patients");
if(!response.ok){
throw new Error(
"Ошибка загрузки пациентов"
);
}
patients =
await response.json();
renderPatients(
patients
);
}
catch(error){
console.error(error);
}
}
function renderPatients(list){
const table =
document.querySelector(".patients-table");
if(!table) return;
table.innerHTML = `
<div class="table-head">
    <span>ID</span>
    <span>Имя</span>
    <span>Телефон</span>
    <span>Последняя услуга</span>
    <span>Последний визит</span>
</div>
<div class="table-divider"></div>
`;
list.forEach(patient=>{
let date="—";
if(patient.lastVisit){
date =
new Date(patient.lastVisit)
.toLocaleDateString("ru-RU");
}
table.innerHTML += `
<div class="table-row patient-row"
     data-id="${patient.id}">
    <span>
        ${patient.id}
    </span>
    <span>
        ${patient.name}
    </span>
    <span>
        ${patient.phone}
    </span>
    <span>
        ${patient.lastService || "—"}
    </span>
    <span>
        ${date}
    </span>
</div>
<div class="table-divider"></div>
`;
});
bindRows();
}
function bindRows(){
document
.querySelectorAll(".patient-row")
.forEach(row=>{
row.addEventListener(
"click",
()=>{
openPatient(
row.dataset.id
);
});
});
}
async function openPatient(id){
try{
const response =
await fetch(
`/api/patients/${id}`
);
if(!response.ok){
throw new Error(
"Пациент не найден"
);
}
const patient =
await response.json();
currentPatient =
patient;
drawerMode =
"view";
drawerTitle.textContent =
"Пациент #" + patient.id;
drawerView.style.display =
"block";
drawerForm.style.display =
"none";
document.querySelector(
".patient-name-view"
).textContent =
patient.name || "—";
document.querySelector(
".patient-phone-view"
).textContent =
patient.phone || "—";
document.querySelector(
".visits-view"
).textContent =
patient.visits || 0;
document.querySelector(
".last-visit-view"
).textContent =
patient.lastVisit
?
new Date(patient.lastVisit)
.toLocaleDateString("ru-RU")
:
"—";
document.querySelector(".total-view").textContent =
(patient.totalPrice || patient.total || 0) + " ₽";
renderHistory(
patient.history || []
);
drawer.classList.add(
"active"
);
overlay.classList.add(
"active"
);
}
catch(error){
console.error(error);
}
}
function renderHistory(history){
const container =
document.querySelector(".history");
if(!container)
return;
container.innerHTML="";
if(history.length===0){
container.innerHTML=`
<p>
История посещений отсутствует
</p>
`;
return;
}
history.forEach((item,index)=>{
container.innerHTML += `
<div class="history-item">
<span>
${formatDate(item.date)}
</span>
<span>
${item.service}
</span>
<span>
${item.price || "Бесплатно"}
</span>
</div>
${index !== history.length-1
?
'<div class="history-divider"></div>'
:
''}
`;
});
}
function openCreatePatient(){
drawerMode =
"create";
currentPatient =
null;
drawerTitle.textContent =
"Добавить пациента";
drawerView.style.display =
"none";
drawerForm.style.display =
"block";
clearForm();
drawer.classList.add(
"active"
);
overlay.classList.add(
"active"
);
}
function editPatientMode(){
if(!currentPatient)
return;
drawerMode =
"edit";
drawerTitle.textContent =
"Редактировать пациента";
drawerView.style.display =
"none";
drawerForm.style.display =
"block";
document.querySelector(
"#patientName"
).value =
currentPatient.name || "";
document.querySelector(
"#patientPhone"
).value =
currentPatient.phone || "";
document.querySelector(
"#patientComment"
).value =
currentPatient.comment || "";
}
async function savePatient(){
const data = {
name:
document.querySelector("#patientName")
.value.trim(),
phone:
document.querySelector("#patientPhone")
.value.trim(),
comment:
document.querySelector("#patientComment")
.value.trim()
};
if(!data.name || !data.phone){
alert(
"Заполните имя и телефон"
);
return;
}
try{
let response;
if(drawerMode === "create"){
response =
await fetch(
"/api/patients",
{
method:"POST",
headers:{
"Content-Type":
"application/json"
},
body:
JSON.stringify(data)
});
}
if(drawerMode === "edit"){
response =
await fetch(
`/api/patients/${currentPatient.id}`,
{
method:"PUT",
headers:{
"Content-Type":
"application/json"
},
body:
JSON.stringify(data)
});
}
if(!response.ok){
throw new Error(
"Ошибка сохранения"
);
}
closeDrawer();
await loadPatients();
}
catch(error){
console.error(error);
alert(
"Не удалось сохранить пациента"
);
}
}
function searchPatients(){
const text =
searchInput.value
.toLowerCase()
.trim();
const result =
patients.filter(patient=>{
return (
patient.name
.toLowerCase()
.includes(text)
||
patient.phone
.includes(text)
);
});
renderPatients(result);
}
function filterPatients(type){
let result =
patients;
switch(type){
case "Новые":
result =
patients.filter(
patient =>
patient.visits === 1
);
break;
case "Постоянные":
result =
patients.filter(
patient =>
patient.visits >= 3
);
break;
case "Были на приёме":
result =
patients.filter(
patient =>
patient.visits > 1
);
break;
default:
result =
patients;
}
renderPatients(
result
);
}
function closeDrawer(){
drawer.classList.remove(
"active"
);
overlay.classList.remove(
"active"
);
}
function clearForm(){
document.querySelector(
"#patientName"
).value="";
document.querySelector(
"#patientPhone"
).value="";
document.querySelector(
"#patientComment"
).value="";
}
function formatDate(date){
if(!date)
return "—";
return new Date(date)
.toLocaleDateString(
"ru-RU"
);
}