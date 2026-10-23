// Runs before first paint: apply the saved theme (if any) and mark JS as available.
(function () {
  var root = document.documentElement;
  root.classList.remove("no-js");
  try {
    var saved = localStorage.getItem("theme");
    if (saved === "light" || saved === "dark") root.setAttribute("data-theme", saved);
  } catch (e) { /* storage blocked: follow the system setting */ }
})();
