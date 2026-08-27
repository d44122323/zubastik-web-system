// If a patient is already logged in, prefill all public booking forms with their profile.
// Comment fields are intentionally left untouched so the patient can enter a new comment each time.
(async function prefillPatientForms(){
  try {
    const response = await fetch('/api/patient/me', { credentials: 'include', cache: 'no-store' });
    if (!response.ok) return;
    const patient = await response.json();
    document.querySelectorAll('form').forEach(form => {
      const name = form.querySelector('input[name="name"], textarea[name="name"]');
      const phone = form.querySelector('input[name="phone"], textarea[name="phone"]');
      const email = form.querySelector('input[name="email"], textarea[name="email"]');
      if (name && patient.name) name.value = patient.name;
      if (phone && patient.phone) phone.value = patient.phone;
      if (email && patient.email) email.value = patient.email;
    });
  } catch (error) {
    // A logged-out visitor should see the normal empty forms.
    console.debug('Patient prefill skipped:', error);
  }
})();
