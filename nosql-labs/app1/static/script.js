// CipherNote – client-side logic

document.addEventListener('DOMContentLoaded', () => {
  // Active nav highlight
  document.querySelectorAll('nav a').forEach(a => {
    if (a.getAttribute('href') === location.pathname) a.classList.add('active');
  });

  initLogin();
  initSearchFocus();
});

// ── Login form ───────────────────────────────────────────────────────────────
// Sends credentials as JSON (Content-Type: application/json).
// This is a common pattern for SPA logins and is what enables VULN-B:
// an intercepting proxy can modify the JSON values to inject operators.
function initLogin() {
  const form = document.getElementById('login-form');
  if (!form) return;

  const msgBox = document.getElementById('login-msg');

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = form.querySelector('button[type=submit]');
    btn.disabled = true;
    btn.textContent = 'Signing in…';

    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;

    try {
      const res = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        // Note: values are sent as a JSON object.
        // The server decodes into interface{}, enabling operator injection.
        body: JSON.stringify({ username, password }),
      });

      const data = await res.json();

      if (res.ok) {
        msgBox.className = 'alert alert-success';
        msgBox.textContent = `Welcome back, ${data.username}. Redirecting…`;
        msgBox.style.display = 'block';
        setTimeout(() => location.href = data.redirect || '/', 800);
      } else {
        msgBox.className = 'alert alert-error';
        msgBox.textContent = data.error || 'Authentication failed.';
        msgBox.style.display = 'block';
        btn.disabled = false;
        btn.textContent = 'Sign In';
      }
    } catch {
      msgBox.className = 'alert alert-error';
      msgBox.textContent = 'Network error. Please try again.';
      msgBox.style.display = 'block';
      btn.disabled = false;
      btn.textContent = 'Sign In';
    }
  });
}

// ── Search input auto-focus ──────────────────────────────────────────────────
function initSearchFocus() {
  const input = document.querySelector('.search-hero input, .search-bar-inline input');
  if (input && !input.value) input.focus();
}
