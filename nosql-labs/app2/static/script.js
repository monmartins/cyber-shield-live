// PeopleDir – client helpers

document.addEventListener('DOMContentLoaded', () => {
  // Generate initials for avatar elements
  document.querySelectorAll('[data-initials]').forEach(el => {
    const name = el.dataset.initials || '';
    const parts = name.trim().split(/\s+/);
    el.textContent = parts.length >= 2
      ? (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
      : name.slice(0, 2).toUpperCase();
  });

  // Active nav
  document.querySelectorAll('nav a').forEach(a => {
    if (a.getAttribute('href') === location.pathname) a.classList.add('active');
  });

  // Employee card click → profile page
  document.querySelectorAll('.emp-card[data-username]').forEach(card => {
    card.addEventListener('click', () => {
      location.href = '/employee?id=' + encodeURIComponent(card.dataset.username);
    });
  });
});
