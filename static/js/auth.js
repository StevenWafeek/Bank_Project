(() => {
  const loginForm = document.getElementById('login-form');
  const signupForm = document.getElementById('signup-form');
  const msg = document.getElementById('auth-message');

  api.get('/api/me').then(() => location.replace('/dashboard.html')).catch(() => {});

  document.querySelectorAll('.tab').forEach((tab) => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.tab').forEach((t) => t.classList.toggle('active', t === tab));
      const isLogin = tab.dataset.tab === 'login';
      loginForm.classList.toggle('hidden', !isLogin);
      signupForm.classList.toggle('hidden', isLogin);
      showMessage(msg, '');
    });
  });

  async function submit(form, handler) {
    const btn = form.querySelector('button[type=submit]');
    btn.disabled = true;
    showMessage(msg, '');
    try {
      await handler(Object.fromEntries(new FormData(form)));
      location.replace('/dashboard.html');
    } catch (err) {
      showMessage(msg, err.message);
    } finally {
      btn.disabled = false;
    }
  }

  loginForm.addEventListener('submit', (e) => {
    e.preventDefault();
    submit(loginForm, ({ email, password }) => {
      if (!email || !password) throw new Error('Please enter your email and password.');
      return api.post('/api/login', { email, password });
    });
  });

  signupForm.addEventListener('submit', (e) => {
    e.preventDefault();
    submit(signupForm, ({ name, email, password, confirm }) => {
      if (!name || !email || !password) throw new Error('Please fill in all fields.');
      if (password.length < 8) throw new Error('Password must be at least 8 characters.');
      if (password !== confirm) throw new Error('Passwords do not match.');
      return api.post('/api/signup', { name, email, password });
    });
  });
})();
