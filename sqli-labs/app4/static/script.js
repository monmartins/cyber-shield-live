// StaffPortal – client helpers

document.addEventListener('DOMContentLoaded', () => {
  // Generate initials avatars from name text
  document.querySelectorAll('.emp-avatar[data-name]').forEach(el => {
    const name = el.dataset.name || '';
    const parts = name.trim().split(/\s+/);
    el.textContent = parts.length >= 2
      ? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
      : name.slice(0, 2).toUpperCase();
  });

  // Active nav link
  document.querySelectorAll('nav a').forEach(a => {
    if (a.href === location.href) a.classList.add('active');
  });
});
