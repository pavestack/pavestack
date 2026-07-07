(function () {
  'use strict';

  var STORAGE_KEY = 'pavestack-theme';
  var root = document.documentElement;
  var themeToggle = document.getElementById('themeToggle');
  var navToggle = document.getElementById('navToggle');
  var navMenu = document.getElementById('navMenu');

  function currentTheme() {
    return root.getAttribute('data-theme') === 'light' ? 'light' : 'dark';
  }

  function applyToggleLabel(theme) {
    if (!themeToggle) return;
    var next = theme === 'light' ? 'dark' : 'light';
    themeToggle.setAttribute('aria-label', 'Switch to ' + next + ' theme');
  }

  applyToggleLabel(currentTheme());

  if (themeToggle) {
    themeToggle.addEventListener('click', function () {
      var next = currentTheme() === 'light' ? 'dark' : 'light';
      root.setAttribute('data-theme', next);
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch (e) {
        /* localStorage unavailable (private mode, etc.) — theme just won't persist */
      }
      applyToggleLabel(next);
    });
  }

  if (navToggle && navMenu) {
    navToggle.addEventListener('click', function () {
      var isOpen = navMenu.classList.toggle('is-open');
      navToggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    });

    navMenu.querySelectorAll('a').forEach(function (link) {
      link.addEventListener('click', function () {
        navMenu.classList.remove('is-open');
        navToggle.setAttribute('aria-expanded', 'false');
      });
    });
  }
})();
