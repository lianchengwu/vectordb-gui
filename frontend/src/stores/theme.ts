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
    const root = document.documentElement;
    if (isDark.value) {
      root.classList.add("dark");
      root.style.colorScheme = "dark";
    } else {
      root.classList.remove("dark");
      root.style.colorScheme = "light";
    }
  }

  return {
    isDark,
    initTheme,
    toggleTheme,
  };
}
