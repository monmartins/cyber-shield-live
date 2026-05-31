document.addEventListener('DOMContentLoaded', () => {
  const params = new URLSearchParams(window.location.search);
  const genre = params.get('genre');
  document.querySelectorAll('.genre-tabs a').forEach(a => {
    const u = new URL(a.href, location.origin);
    if (u.searchParams.get('genre') === genre) a.classList.add('active');
  });
  document.querySelectorAll('.code-block').forEach(block => {
    const btn = document.createElement('button');
    btn.textContent = 'Copy';
    btn.style.cssText = 'float:right;background:#334155;color:#e2e8f0;border:none;border-radius:4px;padding:2px 8px;font-size:.75rem;cursor:pointer;margin:-2px 0 4px 4px;';
    block.parentNode.insertBefore(btn, block);
    btn.addEventListener('click', () => { navigator.clipboard.writeText(block.textContent.trim()); btn.textContent='Copied!'; setTimeout(()=>btn.textContent='Copy',1500); });
  });
});
