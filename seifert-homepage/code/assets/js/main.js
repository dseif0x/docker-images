(function () {
  "use strict";

  var root = document.documentElement;

  // Footer year
  var year = document.getElementById("year");
  if (year) year.textContent = String(new Date().getFullYear());

  // Theme toggle: system default until the visitor picks one.
  var toggle = document.querySelector(".theme-toggle");
  if (toggle) {
    toggle.addEventListener("click", function () {
      var current = root.getAttribute("data-theme") ||
        (window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark");
      var next = current === "light" ? "dark" : "light";
      root.setAttribute("data-theme", next);
      try { localStorage.setItem("theme", next); } catch (e) { /* ignore */ }
    });
  }

  // Nav shadow once the page is scrolled
  var nav = document.getElementById("nav");
  if (nav) {
    var onScroll = function () { nav.classList.toggle("scrolled", window.scrollY > 12); };
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
  }

  // Reveal on scroll
  var reveals = document.querySelectorAll(".reveal");
  if ("IntersectionObserver" in window) {
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          e.target.classList.add("in");
          io.unobserve(e.target);
        }
      });
    }, { rootMargin: "0px 0px -8% 0px", threshold: 0.08 });
    reveals.forEach(function (el) { io.observe(el); });
  } else {
    reveals.forEach(function (el) { el.classList.add("in"); });
  }

  // Highlight the nav link of the section in view
  var links = document.querySelectorAll(".nav-links a[href^='#']");
  var byId = {};
  links.forEach(function (a) { byId[a.getAttribute("href").slice(1)] = a; });
  if ("IntersectionObserver" in window) {
    var spy = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        var link = byId[e.target.id];
        if (!link) return;
        if (e.isIntersecting) {
          links.forEach(function (l) { l.classList.remove("active"); });
          link.classList.add("active");
        }
      });
    }, { rootMargin: "-45% 0px -50% 0px" });
    document.querySelectorAll("main section[id]").forEach(function (s) { spy.observe(s); });
  }

  // Toast
  var toastEl = document.getElementById("toast");
  var toastTimer;
  function toast(msg) {
    if (!toastEl) return;
    toastEl.textContent = msg;
    toastEl.classList.add("show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { toastEl.classList.remove("show"); }, 2200);
  }

  // Copy BibTeX
  var bibtex = [
    "@inproceedings{Maroun2025StackCache,",
    "  title     = {Optimizing Stack Cache Usage Through Local Variable Promotion},",
    "  author    = {Maroun, Emad Jacob and Seifert, Daniel and Gernedl, Johannes},",
    "  booktitle = {2025 IEEE 28th International Symposium on Real-Time Distributed Computing (ISORC)},",
    "  pages     = {297--302},",
    "  year      = {2025},",
    "  month     = may,",
    "  publisher = {IEEE},",
    "  doi       = {10.1109/ISORC65339.2025.00046}",
    "}"
  ].join("\n");

  var copyBtn = document.getElementById("copy-bibtex");
  if (copyBtn) {
    copyBtn.addEventListener("click", function () {
      var done = function () { toast("BibTeX copied to clipboard"); };
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(bibtex).then(done, function () { fallbackCopy(); done(); });
      } else {
        fallbackCopy();
        done();
      }
    });
  }

  function fallbackCopy() {
    var ta = document.createElement("textarea");
    ta.value = bibtex;
    ta.setAttribute("readonly", "");
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand("copy"); } catch (e) { /* ignore */ }
    document.body.removeChild(ta);
  }
})();
