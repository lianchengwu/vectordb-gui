<template>
  <header
    class="h-9 w-full flex items-center justify-between px-3 border-b border-slate-200 dark:border-slate-800 bg-slate-100/90 dark:bg-slate-900/90 select-none shrink-0 wails-drag transition-colors duration-200"
    @dblclick="onToggleMaximize"
  >
    <!-- Left: Brand / Window Title -->
    <div class="flex items-center gap-2 pointer-events-none">
      <div class="w-4 h-4 rounded bg-blue-600/20 text-blue-500 flex items-center justify-center">
        <Database class="w-3 h-3 text-blue-600 dark:text-blue-400" />
      </div>
      <span class="text-xs font-semibold text-slate-800 dark:text-slate-200 tracking-tight">
        Tencent Cloud VectorDB
      </span>
      <span
        v-if="activeConnectionName"
        class="text-[11px] text-slate-500 dark:text-slate-400 font-mono flex items-center gap-1 ml-1"
      >
        <span class="text-slate-400 dark:text-slate-600">/</span>
        <span class="text-slate-600 dark:text-slate-300">{{ activeConnectionName }}</span>
      </span>
    </div>

    <!-- Center: Empty drag region -->
    <div class="flex-1 h-full wails-drag"></div>

    <!-- Right: Theme toggle & Window Controls -->
    <div class="flex items-center gap-1 wails-no-drag">
      <!-- Theme Switcher -->
      <button
        @click="themeStore.toggleTheme"
        class="h-7 px-2 rounded hover:bg-slate-200/80 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition flex items-center gap-1.5 mr-1"
        :title="themeStore.isDark.value ? '切换至浅色主题' : '切换至深色主题'"
      >
        <Sun v-if="themeStore.isDark.value" class="w-3.5 h-3.5 text-amber-400" />
        <Moon v-else class="w-3.5 h-3.5 text-slate-600" />
      </button>

      <span class="h-4 w-[1px] bg-slate-200 dark:bg-slate-800 mx-0.5"></span>

      <!-- Window Minimize -->
      <button
        @click="onMinimize"
        class="w-7 h-7 rounded hover:bg-slate-200/80 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition flex items-center justify-center"
        title="最小化"
      >
        <Minus class="w-3.5 h-3.5" />
      </button>

      <!-- Window Maximize / Restore -->
      <button
        @click="onToggleMaximize"
        class="w-7 h-7 rounded hover:bg-slate-200/80 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition flex items-center justify-center"
        :title="isMaximized ? '还原' : '最大化'"
      >
        <Copy v-if="isMaximized" class="w-3 h-3 rotate-180" />
        <Square v-else class="w-3 h-3" />
      </button>

      <!-- Window Close -->
      <button
        @click="onClose"
        class="w-7 h-7 rounded hover:bg-rose-600 hover:text-white text-slate-600 dark:text-slate-400 transition flex items-center justify-center"
        title="关闭"
      >
        <X class="w-3.5 h-3.5" />
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { Database, Minus, Square, Copy, X, Sun, Moon } from "lucide-vue-next";
import { Window } from "@wailsio/runtime";
import { useThemeStore } from "../stores/theme";
import { useConnectionStore } from "../stores/connection";

const themeStore = useThemeStore();
const connStore = useConnectionStore();

const isMaximized = ref(false);

const activeConnectionName = computed(() => {
  return connStore.activeConnection.value?.name || connStore.activeConnection.value?.url || "";
});

async function updateMaximizeState() {
  try {
    if (typeof Window.IsMaximised === "function") {
      isMaximized.value = await Window.IsMaximised();
    }
  } catch {
    // Ignore fallback
  }
}

function onMinimize() {
  try {
    Window.Minimise();
  } catch (err) {
    console.error("Window.Minimise error:", err);
  }
}

async function onToggleMaximize() {
  try {
    await Window.ToggleMaximise();
    await updateMaximizeState();
  } catch (err) {
    console.error("Window.ToggleMaximise error:", err);
  }
}

function onClose() {
  try {
    Window.Close();
  } catch (err) {
    console.error("Window.Close error:", err);
  }
}
</script>
