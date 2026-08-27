const navLinks = document.querySelectorAll('.nav-menu a');
function scrollToBlock(selector){
    const block = document.querySelector(selector);
    if(block){
        block.scrollIntoView({
            behavior:"smooth",
            block:"start"
        });
    }
}
navLinks.forEach(link=>{
    link.addEventListener("click",(e)=>{
        const href = link.getAttribute("href");
        if(!href.startsWith("#")){
            return;
        }
        e.preventDefault();
        scrollToBlock(href);
    });
});
const btnRecord = document.querySelector(".btn-record");
if(btnRecord){
    btnRecord.addEventListener("click",()=>{
        scrollToBlock("#contacts");
    });
}
const serviceButtons = document.querySelectorAll(".btn-service");
serviceButtons.forEach(btn=>{
    btn.addEventListener("click",()=>{
        window.location.href="services.html";
    });
});
const btnHero = document.querySelector(".btn-hero");
if(btnHero){
    btnHero.addEventListener("click",()=>{
        scrollToBlock("#contacts");
    });
}
document.addEventListener("DOMContentLoaded",()=>{
    const recordButtons = document.querySelectorAll(
        ".btn-record, .btn-hero"
    );
    const contactBlock=document.getElementById("contacts");
    recordButtons.forEach(btn=>{
        btn.addEventListener("click",()=>{
            if(contactBlock){
                contactBlock.scrollIntoView({
                    behavior:"smooth"
                });
            }
        });
    });
});
// Если пациент пришёл со страницы конкретного врача, сохраняем выбранного врача в заявке.
document.addEventListener("DOMContentLoaded", () => {
    const params = new URLSearchParams(window.location.search);
    const doctorId = params.get("doctor");
    if (!doctorId) return;

    document.querySelectorAll('input[name="doctor_id"]').forEach(input => {
        input.value = doctorId;
    });
    const selectedDate = params.get("date") || "";
    const selectedTime = params.get("time") || "";
    document.querySelectorAll('input[name="appointment_date"]').forEach(input => input.value = selectedDate);
    document.querySelectorAll('input[name="appointment_time"]').forEach(input => input.value = selectedTime);

    fetch(`/api/doctors/${encodeURIComponent(doctorId)}`)
        .then(response => response.ok ? response.json() : null)
        .then(doctor => {
            if (!doctor) return;
            const doctorName = doctor.name;
            const comment = document.querySelector('input[name="comment"]');
            if (comment && !comment.value) comment.value = `Запись к врачу: ${doctorName}${selectedDate ? `, ${selectedDate}` : ''}${selectedTime ? ` в ${selectedTime}` : ''}`;
            const footerComment = document.querySelector('.footer-online input[name="comment"]');
            if (footerComment) footerComment.value = `Запись к врачу: ${doctorName}${selectedDate ? `, ${selectedDate}` : ''}${selectedTime ? ` в ${selectedTime}` : ''}`;
        })
        .catch(console.error);

});

const modal = document.getElementById("calculatorModal");
const openBtn = document.getElementById("openCalculator");
const closeBtn = document.getElementById("closeCalculator");
const stepOne = document.querySelector(".step-one");
const stepTwo = document.querySelector(".step-two");
const stepThree = document.querySelector(".step-three");
const stepFour = document.querySelector(".step-four");
const stepFive = document.querySelector(".step-five");
const stepText = document.querySelector(".step-text");
const title = document.querySelector(".calculator-title");
const backBtn =
document.querySelector(".btn-back-step");
const nextBtn =
document.querySelector(".btn-next-step");
const calculateBtn =
document.querySelector(".calculate-button");
const editBtn =
document.querySelector(".btn-edit");
const consultationBtn =
document.querySelector(".btn-record-consultation");
const categoryButtons =
document.querySelectorAll(".service-option");
const serviceList =
document.querySelector(".service-list");
const selectedServicesList =
document.querySelector(".selected-services-list");
const priceRows =
document.querySelector(".price-rows");
const totalPrice =
document.querySelector(".total-price");
const divider =
document.querySelector(".calculator-divider");
let currentStep = 1;
let selectedCategory = "";
let selectedServices = [];
let totalCost = 0;
function sendCalculatorEvent(event) {
    console.log("sendCalculatorEvent:", event);
    fetch("/api/calculator-event", {
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "event=" + encodeURIComponent(event)
    })
    .then(response => {
        console.log("Status:", response.status);
    })
    .catch(error => {
        console.error("Calculator analytics:", error);
    });
}
function resetCalculator(){
    currentStep = 1;
    selectedCategory = "";
    selectedServices = [];
    totalCost = 0;
    categoryButtons.forEach(button=>{
        button.classList.remove("active");
    });
    serviceList.innerHTML = "";
    selectedServicesList.innerHTML = "";
    priceRows.innerHTML = "";
    totalPrice.innerHTML = "≈ 0 ₽";
    stepOne.style.display="block";
    stepTwo.style.display="none";
    stepThree.style.display="none";
    stepFour.style.display="none";
    divider.style.display="none";
    backBtn.style.display="none";
    nextBtn.style.display="block";
    calculateBtn.style.display="none";
    editBtn.style.display="none";
    consultationBtn.style.display="none";
    stepText.innerText="Шаг 1 из 5";
    title.innerText="Что вас интересует?";
}
if(openBtn){
openBtn.addEventListener("click",()=>{
    modal.classList.add("active");
    resetCalculator();
    sendCalculatorEvent("open");
});
}
closeBtn.addEventListener("click",()=>{
    modal.classList.remove("active");
});
modal.addEventListener("click",(event)=>{
    if(event.target === modal){
        modal.classList.remove("active");
    }
});
categoryButtons.forEach(button=>{
    button.addEventListener("click",()=>{
        categoryButtons.forEach(item=>{
            item.classList.remove("active");
        });
        button.classList.add("active");
        selectedCategory =
        button.innerText.trim();
    });
});
const services = {
    "🦷 Консультация и диагностика":[
        "Первичный осмотр и консультация"
    ],
    "🪥 Профессиональная гигиена":[
        "Ультразвуковая чистка",
        "Air Flow",
        "Фторирование"
    ],
    "🩹 Лечение зубов":[
        "Лечение кариеса под протезирование",
        "Лечение кариеса",
        "Лечение пульпита",
        "Лечение периодонтита",
        "Реставрация зуба"
    ],
    "🦴 Удаление зубов и хирургия":[
        {
            title:"Удаление зубов",
            items:[
                "Удаление зуба под протезирование",
                "Удаление временного (молочного) зуба",
                "Удаление постоянного зуба (простое)",
                "Удаление постоянного зуба (сложное)",
                "Удаление зуба мудрости (3 моляр)"
            ]
        },
        {
            title:"Ретинированные и дистопированные зубы",
            items:[
                "Удаление ретинированного / дистопированного зуба - I степень сложности",
                "Удаление ретинированного / дистопированного зуба - II степень сложности",
                "Удаление ретинированного / дистопированного зуба - III степень сложности"
            ]
        },
        {
            title:"Дополнительные хирургические процедуры",
            items:[
                "Удаление отломка коронковой части зуба",
                "Вскрытие поднадкостничного абсцесса",
                "Пластика уздечки верхней или нижней губы",
                "Рассечение капюшона при перикоронарите",
                "Кюретаж лунки ранее удалённого зуба",
                "Наложение шва"
            ]
        }
    ],
    "👑 Коронки и виниры":[
        {
            title:"Коронки",
            items:[
                "Коронка (титановый сплав + керамика NORITAKE, Япония)",
                "Коронка (титановый сплав + керамика VITA VM, Германия)",
                "Коронка (чистый титан + керамика NORITAKE, Япония)",
                "Коронка из диоксида циркония (Китай)",
                "Коронка из диоксида циркония (Япония)",
                "Коронка из диоксида циркония (Германия)"
            ]
        },
        {
            title:"Виниры",
            items:[
                "Винир из диоксида циркония (Китай)",
                "Винир из диоксида циркония (Япония)",
                "Винир из диоксида циркония (Германия)"
            ]
        }
    ],
    "🔩 Имплантация":[
        {
            title:"Импланты",
            items:[
                "Имплант (Корея)",
                "Имплант (Япония)",
                "Имплант (Германия)",
                "Имплант (Швейцария)"
            ]
        },
        {
            title:"Дополнительные процедуры",
            items:[
                "Формирователь десны",
                "Абатмент",
                "Синус-лифтинг (без учёта расходных материалов)",
                "Забор костного трансплантата"
            ]
        }
    ]
};
let prices = {
    "Первичный осмотр и консультация":0,
    "Ультразвуковая чистка":500,
    "Air Flow":500,
    "Фторирование":500,
    "Лечение кариеса под протезирование":0,
    "Лечение кариеса":600,
    "Лечение пульпита":600,
    "Лечение периодонтита":750,
    "Реставрация зуба":950,
    "Удаление зуба под протезирование":0,
    "Удаление временного (молочного) зуба":0,
    "Удаление постоянного зуба (простое)":200,
    "Удаление постоянного зуба (сложное)":450,
    "Удаление зуба мудрости (3 моляр)":1200,
    "Удаление ретинированного / дистопированного зуба - I степень сложности":900,
    "Удаление ретинированного / дистопированного зуба - II степень сложности":200,
    "Удаление ретинированного / дистопированного зуба - III степень сложности":400,
    "Удаление отломка коронковой части зуба":600,
    "Вскрытие поднадкостничного абсцесса":200,
    "Пластика уздечки верхней или нижней губы":1500,
    "Рассечение капюшона при перикоронарите":300,
    "Кюретаж лунки ранее удалённого зуба":250,
    "Наложение шва":200,
    "Коронка (титановый сплав + керамика NORITAKE, Япония)":2700,
    "Коронка (титановый сплав + керамика VITA VM, Германия)":2900,
    "Коронка (чистый титан + керамика NORITAKE, Япония)":3100,
    "Коронка из диоксида циркония (Китай)":6250,
    "Коронка из диоксида циркония (Япония)":9500,
    "Коронка из диоксида циркония (Германия)":11000,
    "Винир из диоксида циркония (Китай)":6000,
    "Винир из диоксида циркония (Япония)":6500,
    "Винир из диоксида циркония (Германия)":7400,
    "Имплант (Корея)":25000,
    "Имплант (Япония)":25000,
    "Имплант (Германия)":25000,
    "Имплант (Швейцария)":25000,
    "Формирователь десны":2000,
    "Абатмент":4000,
    "Синус-лифтинг (без учёта расходных материалов)":5000,
    "Забор костного трансплантата":20000
};
async function loadServicePrices(){
    try{
        const response = await fetch("/api/services");
        if(!response.ok) return;
        const data = await response.json();
        const remotePrices = {};
        data.forEach(service => { remotePrices[service.name] = Number(service.price) || 0; });
        prices = { ...prices, ...remotePrices };
    }catch(error){
        console.warn("Не удалось загрузить актуальные цены", error);
    }
}
loadServicePrices();
function createServices(category){
    serviceList.innerHTML="";
    services[category].forEach(service=>{
        if(typeof service === "string"){
            addServiceButton(service);
        }else{
            const title =
            document.createElement("div");
            title.className =
            "service-category-title";
            title.innerText =
            service.title;
            serviceList.appendChild(title);
            service.items.forEach(item=>{
                addServiceButton(item);
            });
        }
    });
}
function addServiceButton(serviceName){
    const button =
    document.createElement("button");
    button.className =
    "service-list-button";
    button.innerText =
    serviceName;
    button.addEventListener("click",()=>{
        button.classList.toggle("active");
        if(button.classList.contains("active")){
            if(!selectedServices.includes(serviceName)){
                selectedServices.push(serviceName);
            }
        }else{
            selectedServices =
            selectedServices.filter(
                item=>item!==serviceName
            );
        }
    });
    serviceList.appendChild(button);
}
nextBtn.addEventListener("click",()=>{
    if(currentStep === 1){
        if(selectedCategory === ""){
            alert("Выберите категорию услуг");
            return;
        }
        currentStep = 2;
        stepOne.style.display="none";
        stepTwo.style.display="block";
        stepThree.style.display="none";
        stepFour.style.display="none";
        divider.style.display="block";
        backBtn.style.display="block";
        nextBtn.style.display="block";
        calculateBtn.style.display="none";
        editBtn.style.display="none";
        consultationBtn.style.display="none";
        stepText.innerText =
        "Шаг 2 из 5";
        title.innerText =
        "Выберите услугу";
        selectedServices=[];
        createServices(selectedCategory);
        return;
    }
    if(currentStep === 2){
        if(selectedServices.length===0){
            alert("Выберите хотя бы одну услугу");
            return;
        }
        currentStep=3;
        stepOne.style.display="none";
        stepTwo.style.display="none";
        stepThree.style.display="block";
        stepFour.style.display="none";
        nextBtn.style.display="none";
        calculateBtn.style.display="block";
        editBtn.style.display="none";
        consultationBtn.style.display="none";
        stepText.innerText =
        "Шаг 3 из 5";
        title.innerText =
        "Подтвердите выбор";
        showSelectedServices();
        return;
    }
});
function showSelectedServices(){
    const container =
    document.querySelector(".confirm-services");
    container.innerHTML="";
    selectedServices.forEach(service=>{
        const item =
        document.createElement("div");
        item.className =
        "confirm-service-item";
        item.innerHTML =
        "✔ " + service;
        container.appendChild(item);
    });
}
calculateBtn.addEventListener("click",()=>{
    currentStep=4;
    document
.querySelector(".calculator-bottom")
.classList.add("step-four-buttons");
    stepOne.style.display="none";
    stepTwo.style.display="none";
    stepThree.style.display="none";
    stepFour.style.display="block";
backBtn.style.display="none";
    nextBtn.style.display="none";
    calculateBtn.style.display="none";
    editBtn.style.display="flex";
    consultationBtn.style.display="flex";
    stepText.innerText =
    "Шаг 4 из 5";
    title.innerText =
    "Выбранные услуги";
    showCalculation();
    sendCalculatorEvent("finish");
});
function showCalculation(){
    selectedServicesList.innerHTML="";
    priceRows.innerHTML="";
    totalCost=0;
    selectedServices.forEach(service=>{
        const item =
        document.createElement("div");
        item.className =
        "selected-service";
        item.innerHTML =
        "✔ " + service;
        selectedServicesList.appendChild(item);
    });
    selectedServices.forEach(service=>{
        const row =
        document.createElement("div");
        row.className =
        "price-row";
        const price =
        prices[service] ?? 0;
        totalCost += price;
        row.innerHTML = `
            <span>
                ${service}
            </span>
            <span>
                ${
                    price === 0
                    ?
                    "Бесплатно"
                    :
                    "от " + 
                    price.toLocaleString("ru-RU")
                    +
                    " ₽"
                }
            </span>
        `;
        priceRows.appendChild(row);
    });
    totalPrice.innerHTML =
        "≈ " +
        totalCost.toLocaleString("ru-RU")
        +
        " ₽";
}
backBtn.addEventListener("click",()=>{
   if(currentStep === 4){
    currentStep = 3;
    document
    .querySelector(".calculator-bottom")
    .classList.remove("step-four-buttons");
    backBtn.style.display="block";
    nextBtn.style.display="none";
    calculateBtn.style.display="block";
    editBtn.style.display="none";
    consultationBtn.style.display="none";
    stepText.innerText =
    "Шаг 3 из 5";
    title.innerText =
    "Подтвердите выбор";
    return;
}
    if(currentStep === 3){
        currentStep = 2;
        stepThree.style.display="none";
        stepTwo.style.display="block";
        nextBtn.style.display="block";
        calculateBtn.style.display="none";
        editBtn.style.display="none";
        consultationBtn.style.display="none";
        stepText.innerText =
        "Шаг 2 из 5";
        title.innerText =
        "Выберите услугу";
        return;
    }
    if(currentStep === 2){
        currentStep = 1;
        stepTwo.style.display="none";
        stepOne.style.display="block";
        divider.style.display="none";
        backBtn.style.display="none";
        nextBtn.style.display="block";
        calculateBtn.style.display="none";
        editBtn.style.display="none";
        consultationBtn.style.display="none";
        stepText.innerText =
        "Шаг 1 из 5";
        title.innerText =
        "Что вас интересует?";
    }
});
editBtn.addEventListener("click",()=>{
    currentStep = 3;
    backBtn.style.display="block";
    document
.querySelector(".calculator-bottom")
.classList.remove("step-four-buttons");
    stepFour.style.display="none";
    stepThree.style.display="block";
    calculateBtn.style.display="block";
    editBtn.style.display="none";
    consultationBtn.style.display="none";
    stepText.innerText =
    "Шаг 3 из 5";
    title.innerText =
    "Подтвердите выбор";
    showSelectedServices();
});
consultationBtn.addEventListener("click",()=>{
    sendCalculatorEvent("request");
document.getElementById("calculatorServices").value =
    selectedServices.join(" • ")
document.getElementById("calculatorPrice").value =
    totalCost;
    currentStep = 5;
    stepFour.style.display="none";
    stepFive.style.display="block";
    stepText.innerText =
    "Шаг 5 из 5";
    title.innerText =
    "";
    divider.style.display="none";
    editBtn.style.display="none";
    consultationBtn.style.display="none";
});
document.addEventListener(
"DOMContentLoaded",
()=>{
    resetCalculator();
stepFive.style.display="none";
});
function updateCalculatorHeight(){
    const content =
    document.querySelector(".calculator-content");
    if(!content) return;
    content.style.height="auto";
}
window.addEventListener(
"resize",
()=>{
    updateCalculatorHeight();
});
updateCalculatorHeight();
[
    nextBtn,
    calculateBtn,
    editBtn,
    consultationBtn
].forEach(button=>{
    if(button){
        button.addEventListener(
            "dblclick",
            e=>{
                e.preventDefault();
            }
        );
    }
});
const recordForm = document.querySelector(".record-form");
if (recordForm) {
    recordForm.addEventListener("submit", () => {
        document.getElementById("calculatorServices").value =
            selectedServices.join(" • ")
        document.getElementById("calculatorPrice").value =
            totalCost;
    });
}
const aiInput = document.getElementById("aiInput");
const aiButton = document.getElementById("aiSend");
const aiAnswer = document.getElementById("aiAnswer");
if (aiButton && aiInput && aiAnswer) {
    async function sendAIQuestion(){
        const question = aiInput.value.trim();
        if(!question){
            return;
        }
        aiAnswer.style.display = "block";
        aiAnswer.classList.add("show");
        aiAnswer.innerHTML = 
        "ИИ думает...";
        aiInput.value = "";
        try{
const response = await fetch(
    "/chat",
    {
        method:"POST",
        headers:{
            "Content-Type":"application/json"
        },
        body:JSON.stringify({
            message: question
        })
    }
);
            if(!response.ok){
                throw new Error(
                    "Ошибка сервера"
                );
            }
            const data = await response.json();
            aiAnswer.innerHTML =
            data.answer;
        }
        catch(error){
            console.error(
                "AI ERROR:",
                error
            );
            aiAnswer.innerHTML =
            "Не удалось подключиться к ИИ";
        }
    }
    aiButton.addEventListener(
        "click",
        sendAIQuestion
    );
    aiInput.addEventListener(
        "keydown",
        function(event){
            if(event.key === "Enter"){
                event.preventDefault();
                sendAIQuestion();
            }
        }
    );
}
