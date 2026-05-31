// StockTrack – client helpers

document.addEventListener('DOMContentLoaded', () => {
  // Active nav
  document.querySelectorAll('nav a').forEach(a => {
    if (location.pathname === new URL(a.href, location.href).pathname) a.classList.add('active');
  });

  // ── XML Stock Check Console ────────────────────────────────────────────────
  const sendBtn = document.getElementById('send-xml');
  const xmlInput = document.getElementById('xml-body');
  const resultBox = document.getElementById('xml-result');

  if (sendBtn && xmlInput && resultBox) {
    sendBtn.addEventListener('click', async () => {
      const body = xmlInput.value.trim();
      if (!body) { resultBox.textContent = '// Empty request'; return; }
      resultBox.textContent = '// Sending…';
      try {
        const res = await fetch('/stock', {
          method: 'POST',
          headers: { 'Content-Type': 'application/xml' },
          body,
        });
        const text = await res.text();
        // Pretty-print JSON if possible
        try {
          resultBox.textContent = JSON.stringify(JSON.parse(text), null, 2);
        } catch {
          resultBox.textContent = text;
        }
      } catch (e) {
        resultBox.textContent = '// Network error: ' + e.message;
      }
    });
  }
});
