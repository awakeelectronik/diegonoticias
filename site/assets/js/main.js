document.addEventListener("DOMContentLoaded", () => {
  const toggle = document.querySelector("[data-mobile-nav]");
  if (!toggle) return;
  toggle.addEventListener("click", () => {
    document.body.classList.toggle("search-open");
  });
});

