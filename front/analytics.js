let requestsChart = null;
let revenueChart = null;
let statusChart = null;
let sourceChart = null;
let funnelChart = null;
async function loadAnalytics() {
  try {
    const response = await fetch("/api/analytics");
    if (!response.ok) {
      throw new Error("Ошибка загрузки аналитики");
    }
    const data = await response.json();
    renderKPI(data);
    renderRequestsChart(data.chart || []);
    renderRevenueChart(data.revenueChart || []);
    renderStatusChart(data.statusChart || []);
    renderSourceChart(data.sourceChart || []);
    renderFunnelChart(data.funnel || {});
    renderCalculator(data.calculator || {});
    renderTopServices(data.topServices || []);
    renderRecentRequests(data.recentRequests || []);
    renderAI(data);
  } catch (error) {
    console.error(error);
  }
}
function renderKPI(data) {
  const cards = document.querySelectorAll(".kpi-card");
  cards[0].querySelector("strong").textContent = data.newPatients;
  cards[1].querySelector("strong").textContent = data.totalRequests;
  cards[2].querySelector("strong").textContent =
    data.averagePrice.toLocaleString("ru-RU") + " ₽";
  cards[3].querySelector("strong").textContent = data.conversion + " %";
}
function renderRequestsChart(chartData) {
  const ctx = document.getElementById("requestsChart").getContext("2d");
  if (requestsChart) {
    requestsChart.destroy();
  }
  requestsChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: chartData.map((item) => item.date),
      datasets: [
        {
          label: "Заявки",
          data: chartData.map((item) => item.count),
          borderColor: "#00D1FF",
          backgroundColor: "rgba(0,209,255,.15)",
          fill: true,
          tension: 0.35,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
      },
    },
  });
}
function renderRevenueChart(chartData) {
  const canvas = document.getElementById("revenueChart");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (revenueChart) {
    revenueChart.destroy();
  }
  revenueChart = new Chart(ctx, {
    type: "line",
    data: {
      labels: chartData.map((item) => item.month),
      datasets: [
        {
          label: "Доход",
          data: chartData.map((item) => item.amount),
          borderColor: "#00D1FF",
          backgroundColor: "rgba(0,209,255,.15)",
          fill: true,
          tension: 0.35,
          borderWidth: 4,
          pointRadius: 10,
          pointHoverRadius: 12,
          pointBackgroundColor: "#00D1FF",
          pointBorderColor: "#FFFFFF",
          pointBorderWidth: 3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
      },
      scales: {
        y: {
          ticks: {
            callback(value) {
              return value.toLocaleString("ru-RU") + " ₽";
            },
          },
        },
      },
    },
  });
}
function renderStatusChart(chartData) {
  const canvas = document.getElementById("statusChart");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (statusChart) {
    statusChart.destroy();
  }
  statusChart = new Chart(ctx, {
    type: "pie",
    data: {
      labels: chartData.map((item) => item.status),
      datasets: [
        {
          data: chartData.map((item) => item.count),
          backgroundColor: [
            "#00D1FF",
            "#39C16C",
            "#FFC857",
            "#FF6B6B",
            "#8B5CF6",
            "#64748B",
          ],
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: "bottom",
          labels: {
            font: {
              family: "Inter",
              size: 16,
            },
          },
        },
      },
    },
  });
}
function renderSourceChart(chartData) {
  const canvas = document.getElementById("sourceChart");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (sourceChart) {
    sourceChart.destroy();
  }
  sourceChart = new Chart(ctx, {
    type: "doughnut",
    data: {
      labels: chartData.map((item) => item.source),
      datasets: [
        {
          data: chartData.map((item) => item.count),
          backgroundColor: [
            "#00D1FF",
            "#39C16C",
            "#FFC857",
            "#FF6B6B",
            "#8B5CF6",
            "#64748B",
            "#94A3B8",
          ],
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: "bottom",
          labels: {
            font: {
              family: "Inter",
              size: 16,
            },
          },
        },
      },
    },
  });
}
function renderFunnelChart(data) {
  const canvas = document.getElementById("funnelChart");
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  if (funnelChart) {
    funnelChart.destroy();
  }
  funnelChart = new Chart(ctx, {
    type: "bar",
    data: {
      labels: ["Новые", "Подтверждены", "Завершены", "Отменены"],
      datasets: [
        {
          data: [
            data.new || 0,
            data.confirmed || 0,
            data.completed || 0,
            data.cancelled || 0,
          ],
          backgroundColor: ["#00D1FF", "#39C16C", "#FFC857", "#FF6B6B"],
          borderRadius: 12,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
      },
      scales: {
        y: {
          beginAtZero: true,
        },
      },
    },
  });
}
function renderCalculator(data) {
  document.getElementById("calcOpen").textContent = data.open || 0;
  document.getElementById("calcFinish").textContent = data.finish || 0;
  document.getElementById("calcRequest").textContent = data.request || 0;
}
function renderTopServices(services) {
  const card = document.getElementById("topServices");

  if (!card) {
    return;
  }

  card.innerHTML = `

        <h3>
            ТОП услуг
        </h3>

    `;

  services.forEach((service) => {
    const row = document.createElement("div");

    row.className = "service-row";

    row.innerHTML = `

            <span>
                ${service.name}
            </span>

            <strong>
                ${service.count}
            </strong>

        `;

    card.appendChild(row);
  });
}
function renderRecentRequests(requests) {
  const card = document.querySelector(".recent-card");
  if (!card) return;
  card.innerHTML = `
        <h3>Последние заявки</h3>
        <div class="request-header">
            <span>Пациент</span>
            <span>Услуга</span>
            <span>Дата</span>
            <span>Статус</span>
        </div>
    `;
  requests.forEach((request) => {
    const row = document.createElement("div");
    row.className = "request-row";
    const date = new Date(request.date).toLocaleDateString("ru-RU");
    row.innerHTML = `
            <span>${request.name}</span>
            <span class="services-list">
             ${(request.service || "—")
               .split(",")
               .map((s) => s.trim())
               .join("<br>")}
            </span>
            <span>${date}</span>
            <span>${request.status}</span>
        `;
    card.appendChild(row);
  });
}
document.addEventListener("DOMContentLoaded", () => {
  loadAnalytics();
});

function renderAI(data) {
  const total = document.getElementById("aiTotal");

  if (total) {
    total.textContent = data.aiRequests || 0;
  }

  const box = document.getElementById("aiQuestions");

  if (!box) {
    return;
  }

  box.innerHTML = `

<h3>
Популярные запросы
</h3>

`;

  (data.aiTopQuestions || []).forEach((item) => {
    const row = document.createElement("div");

    row.className = "service-row";

    row.innerHTML = `

<span>
${item.question}
</span>

<strong>
${item.count}
</strong>

`;

    box.appendChild(row);
  });
}
