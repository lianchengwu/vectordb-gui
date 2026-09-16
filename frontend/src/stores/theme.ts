import { ref } from "vue";

const isDark = ref(true);

export function useThemeStore() {
  function initTheme() {
    const saved = localStorage.getItem("tcvectordb_theme");
    if (saved === "light") {
      isDark.value = false;
    } else if (saved === "dark") {
      isDark.value = true;
    } else {
      // Default to dark mode for developer tooling
      isDark.value = true;
    }
    applyTheme();
  }

  function toggleTheme() {
    isDark.value = !isDark.value;
    localStorage.setItem("tcvectordb_theme", isDark.value ? "dark" : "light");
    applyTheme();
  }

  function applyTheme() {
    if (isDark.value) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }

  return {
    isDark,
    initTheme,
    toggleTheme,
  };
}
