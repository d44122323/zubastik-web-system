(function () {
  "use strict";

  const css = `
    .site-captcha-box{margin:10px 0 4px;padding:12px 14px;border:1px solid #e1e7eb;border-radius:16px;background:#f8fbfc}
    .site-captcha-title{font-size:13px;font-weight:600;margin:0 0 7px;color:#334}
    .site-captcha-row{display:flex;gap:8px;align-items:center}
    .site-captcha-question{flex:1;font-weight:600;color:#168fbd}
    .site-captcha-input{width:110px;height:40px;border:1px solid #d8e0e5;border-radius:12px;padding:0 10px;font-size:16px;box-sizing:border-box}
    .site-captcha-refresh{height:40px;border:0;border-radius:12px;background:#eef4f6;cursor:pointer;padding:0 11px;font-size:16px}
    .site-captcha-refresh:hover{background:#e1edf1}
    .site-captcha-error{margin:6px 0 0;color:#b32626;font-size:12px;min-height:15px}
    .site-auth-gate{position:fixed;inset:0;background:rgba(0,0,0,.42);display:flex;align-items:center;justify-content:center;z-index:99999;padding:20px;box-sizing:border-box}
    .site-auth-gate-window{width:min(430px,100%);background:#fff;border-radius:24px;padding:28px;box-shadow:0 18px 60px rgba(0,0,0,.22);text-align:center;position:relative}
    .site-auth-gate-window h3{margin:0 0 10px;font-size:23px;color:#222}
    .site-auth-gate-window p{margin:0 0 20px;color:#666;line-height:1.5}
    .site-auth-gate-actions{display:flex;gap:10px;justify-content:center;flex-wrap:wrap}
    .site-auth-gate-actions button{border:0;border-radius:22px;padding:11px 18px;font-size:15px;cursor:pointer}
    .site-auth-primary{background:#00d1ff;color:#fff}.site-auth-secondary{background:#eef3f5;color:#333}
    .site-auth-gate-close{position:absolute;right:14px;top:10px;border:0;background:transparent;font-size:25px;cursor:pointer;color:#888}
  `;
  const style = document.createElement("style");
  style.textContent = css;
  document.head.appendChild(style);

  async function getCaptcha() {
    const r = await fetch("/api/captcha", {
      credentials: "include",
      cache: "no-store",
    });
    if (!r.ok)
      throw new Error((await r.text()) || "Не удалось получить проверку");
    return r.json();
  }

  function makeCaptchaBox() {
    const box = document.createElement("div");
    box.className = "site-captcha-box";
    box.innerHTML = `
      <p class="site-captcha-title">Проверка на бота</p>
      <div class="site-captcha-row">
        <span class="site-captcha-question">Загрузка...</span>
        <input class="site-captcha-input" inputmode="numeric" autocomplete="off" placeholder="Ответ" aria-label="Ответ CAPTCHA">
        <button type="button" class="site-captcha-refresh" aria-label="Обновить проверку">↻</button>
      </div>
      <p class="site-captcha-error" aria-live="polite"></p>
    `;
    const state = { token: "", answer: "" };
    box._captchaState = state;
    const load = async () => {
      box.querySelector(".site-captcha-error").textContent = "";
      box.querySelector(".site-captcha-question").textContent = "Загрузка...";
      try {
        const c = await getCaptcha();
        state.token = c.token;
        state.answer = "";
        box.querySelector(".site-captcha-question").textContent = c.question;
        box.querySelector(".site-captcha-input").value = "";
      } catch (e) {
        box.querySelector(".site-captcha-question").textContent =
          "Ошибка загрузки";
        box.querySelector(".site-captcha-error").textContent =
          e.message || "Не удалось получить проверку";
      }
    };
    box._reloadCaptcha = load;
    box.querySelector(".site-captcha-refresh").addEventListener("click", load);
    box.querySelector(".site-captcha-input").addEventListener("input", (e) => {
      state.answer = e.target.value.trim();
    });
    load();
    return box;
  }

  function authGate() {
    const old = document.querySelector(".site-auth-gate");
    if (old) old.remove();
    const gate = document.createElement("div");
    gate.className = "site-auth-gate";
    gate.innerHTML = `
      <div class="site-auth-gate-window">
        <button type="button" class="site-auth-gate-close" aria-label="Закрыть">×</button>
        <h3>Для записи нужен аккаунт</h3>
        <p>Войдите в личный кабинет или зарегистрируйтесь, чтобы отправить заявку на приём.</p>
        <div class="site-auth-gate-actions">
          <button type="button" class="site-auth-primary">Войти</button>
          <button type="button" class="site-auth-secondary">Регистрация</button>
        </div>
      </div>`;
    const go = () => {
      location.href = "/?account=patient&record=1";
    };
    gate.querySelector(".site-auth-primary").onclick = go;
    gate.querySelector(".site-auth-secondary").onclick = () => {
      location.href = "/?account=register&record=1";
    };
    gate.querySelector(".site-auth-gate-close").onclick = () => gate.remove();
    gate.addEventListener("click", (e) => {
      if (e.target === gate) gate.remove();
    });
    document.body.appendChild(gate);
  }

  function installRegistrationCaptcha() {
    const form = document.getElementById("homeRegisterForm");
    if (!form || form.dataset.captchaReady === "1") return;
    form.dataset.captchaReady = "1";
    const box = makeCaptchaBox();
    const button = form.querySelector('button[type="submit"]');
    form.insertBefore(box, button);

    document.addEventListener(
      "submit",
      async function (event) {
        if (event.target !== form) return;
        event.preventDefault();
        event.stopImmediatePropagation();
        const state = box._captchaState;
        const answer = box.querySelector(".site-captcha-input").value.trim();
        if (!state.token || !answer) {
          box.querySelector(".site-captcha-error").textContent =
            "Ответьте на проверку";
          return;
        }
        const payload = {
          name: document.getElementById("homeRegName").value,
          phone: document.getElementById("homeRegPhone").value,
          email: document.getElementById("homeRegEmail").value,
          birthDate: document.getElementById("homeRegBirth").value,
          password: document.getElementById("homeRegPassword").value,
          privacyConsent:
            !!document.getElementById("homePrivacyConsent").checked,
          captchaToken: state.token,
          captchaAnswer: answer,
        };
        const msg = document.getElementById("homeAuthMsg");
        if (!payload.privacyConsent) {
          msg.textContent =
            "Для регистрации необходимо согласиться на обработку персональных данных и с Политикой конфиденциальности.";
          return;
        }
        msg.textContent = "Создание аккаунта...";
        try {
          const r = await fetch("/patient/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            credentials: "include",
            body: JSON.stringify(payload),
          });
          if (!r.ok) {
            msg.textContent = await r.text();
            await box._reloadCaptcha();
            return;
          }
          location.href = "/account/";
        } catch (e) {
          msg.textContent = "Не удалось создать аккаунт";
          await box._reloadCaptcha();
        }
      },
      true,
    );
  }

  async function installRequestGuard(form) {
    if (form.dataset.captchaGuard === "1") return;
    form.dataset.captchaGuard = "1";
    document.addEventListener(
      "submit",
      async function (event) {
        if (event.target !== form || form.dataset.captchaBypass === "1") return;
        event.preventDefault();
        event.stopImmediatePropagation();

        try {
          const auth = await fetch("/api/patient/me", {
            credentials: "include",
            cache: "no-store",
          });
          if (!auth.ok) {
            authGate();
            return;
          }
        } catch (_) {
          authGate();
          return;
        }

        const box = makeCaptchaBox();
        const windowBox = document.createElement("div");
        windowBox.className = "site-auth-gate";
        windowBox.innerHTML = `<div class="site-auth-gate-window"><button type="button" class="site-auth-gate-close" aria-label="Закрыть">×</button><h3>Проверка заявки</h3><p>Подтвердите, что заявку отправляет человек.</p></div>`;
        const inner = windowBox.querySelector(".site-auth-gate-window");
        inner.appendChild(box);
        const actions = document.createElement("div");
        actions.className = "site-auth-gate-actions";
        actions.innerHTML =
          '<button type="button" class="site-auth-primary">Отправить заявку</button><button type="button" class="site-auth-secondary">Отмена</button>';
        inner.appendChild(actions);
        document.body.appendChild(windowBox);
        const close = () => windowBox.remove();
        windowBox.querySelector(".site-auth-gate-close").onclick = close;
        actions.children[1].onclick = close;
        windowBox.addEventListener("click", (e) => {
          if (e.target === windowBox) close();
        });
        actions.children[0].onclick = () => {
          const state = box._captchaState;
          const answer = box.querySelector(".site-captcha-input").value.trim();
          if (!state.token || !answer) {
            box.querySelector(".site-captcha-error").textContent =
              "Ответьте на проверку";
            return;
          }
          let token = form.querySelector('input[name="captcha_token"]');
          if (!token) {
            token = document.createElement("input");
            token.type = "hidden";
            token.name = "captcha_token";
            form.appendChild(token);
          }
          let answerInput = form.querySelector('input[name="captcha_answer"]');
          if (!answerInput) {
            answerInput = document.createElement("input");
            answerInput.type = "hidden";
            answerInput.name = "captcha_answer";
            form.appendChild(answerInput);
          }
          token.value = state.token;
          answerInput.value = answer;
          form.dataset.captchaBypass = "1";
          close();
          HTMLFormElement.prototype.submit.call(form);
        };
      },
      true,
    );
  }

  document.addEventListener("DOMContentLoaded", () => {
    installRegistrationCaptcha();
    document.querySelectorAll("form").forEach((form) => {
      const action = (form.getAttribute("action") || "").trim();
      if (action === "/submit" || /\/submit(?:\?|$)/.test(action))
        installRequestGuard(form);
    });
  });
})();
