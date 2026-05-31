// ShopFlow – minimal UI helpers
document.addEventListener('DOMContentLoaded', () => {
  // Mark active nav link
  const path = window.location.pathname;
  document.querySelectorAll('nav a').forEach(a => {
    if (a.getAttribute('href') === path) a.classList.add('active');
  });

  // Mark active category tab
  const params = new URLSearchParams(window.location.search);
  const cat = params.get('category');
  document.querySelectorAll('.tabs a').forEach(a => {
    const href = new URL(a.href, location.origin);
    if (href.searchParams.get('category') === cat) a.classList.add('active');
  });

  // Auto-dismiss alerts after 6s
  document.querySelectorAll('.alert').forEach(el => {
    setTimeout(() => { el.style.opacity = '0'; el.style.transition = 'opacity .4s'; setTimeout(() => el.remove(), 400); }, 6000);
  });

  // Code block copy button
  document.querySelectorAll('.code-block').forEach(block => {
    const btn = document.createElement('button');
    btn.textContent = 'Copy';
    btn.style.cssText = 'float:right;margin:-4px 0 4px 4px;background:#334155;color:#e2e8f0;border:none;border-radius:4px;padding:2px 8px;font-size:.75rem;cursor:pointer;';
    block.parentNode.insertBefore(btn, block);
    btn.addEventListener('click', () => {
      navigator.clipboard.writeText(block.textContent.trim()).then(() => {
        btn.textContent = 'Copied!';
        setTimeout(() => { btn.textContent = 'Copy'; }, 1500);
      });
    });
  });
});
