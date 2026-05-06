async function initSearch() {
  const input = document.querySelector("#search-input");
  const results = document.querySelector("#search-results");
  if (!input || !results) return;
  let pagefind;
  let timer;
  async function ensurePagefind() {
    if (pagefind) return pagefind;
    pagefind = await import("/pagefind/pagefind.js");
    return pagefind;
  }
  input.addEventListener("input", () => {
    clearTimeout(timer);
    timer = setTimeout(async () => {
      const q = input.value.trim();
      if (q.length < 2) {
        results.innerHTML = "";
        return;
      }
      const pf = await ensurePagefind();
      const search = await pf.search(q);
      const items = await Promise.all((search.results || []).slice(0, 8).map((r) => r.data()));
      results.innerHTML = items
        .map((i) => `<a href="${i.url}" class="card" style="display:block;margin-top:10px;"><strong>${i.meta.title || i.url}</strong><p style="margin:6px 0 0;color:#666;">${i.excerpt || ""}</p></a>`)
        .join("");
    }, 180);
  });
}

document.addEventListener("DOMContentLoaded", initSearch);

