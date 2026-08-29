(async function prefillPatientForms() {
  try {
    const response = await fetch("/api/patient/me", {
      credentials: "include",
      cache: "no-store",
    });
    if (!response.ok) return;
    const patient = await response.json();
    document.querySelectorAll("form").forEach((form) => {
      const name = form.querySelector(
        'input[name="name"], textarea[name="name"]',
      );
      const phone = form.querySelector(
        'input[name="phone"], textarea[name="phone"]',
      );
      const email = form.querySelector(
        'input[name="email"], textarea[name="email"]',
      );
      if (name && patient.name) name.value = patient.name;
      if (phone && patient.phone) phone.value = patient.phone;
      if (email && patient.email) email.value = patient.email;
    });
  } catch (error) {

    console.debug("Patient prefill skipped:", error);
  }
})();
