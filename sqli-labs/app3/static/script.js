// NewsHub – client-side helpers

document.addEventListener('DOMContentLoaded', () => {
  // Highlight active nav link
  const links = document.querySelectorAll('nav a');
  links.forEach(l => {
    if (l.href === location.href) l.classList.add('active');
  });

  // Auto-dismiss flash messages after 5 s
  const flash = document.querySelector('.flash');
  if (flash) setTimeout(() => flash.style.opacity = '0', 5000);

  // Tag cloud: random slight size variation for visual effect
  document.querySelectorAll('.tag-chip').forEach(chip => {
    const sizes = ['0.78rem','0.82rem','0.88rem','0.92rem','0.96rem'];
    chip.style.fontSize = sizes[Math.floor(Math.random() * sizes.length)];
  });
});
