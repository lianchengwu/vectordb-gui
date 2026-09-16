<template>
  <aside class="w-72 h-full bg-white border-r border-slate-200 dark:bg-slate-900 dark:border-slate-800 flex flex-col select-none shrink-0 text-slate-700 dark:text-slate-200 transition-colors duration-200">
    <!-- Header / Brand -->
    <div class="h-14 px-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-2.5">
        <div class="w-8 h-8 rounded-lg bg-blue-50 border border-blue-200 text-blue-600 dark:bg-blue-600/20 dark:border-blue-500/30 dark:text-blue-400 flex items-center justify-center shadow-inner">
          <Database class="w-4 h-4" />
        </div>
        <div>
          <h1 class="font-semibold text-sm text-slate-900 dark:text-white tracking-tight leading-none">VectorDB Client</h1>
          <p class="text-[10px] text-slate-500 dark:text-slate-400 mt-1 font-mono leading-none">Tencent Cloud</p>
        </div>
      </div>
      <button
        @click="openCreateModal"
        class="w-7 h-7 rounded-lg bg-blue-600 hover:bg-blue-500 text-white flex items-center justify-center transition shadow-sm hover:shadow"
        title="新建连接"
      >
        <Plus class="w-4 h-4" />
      </button>
    </div>

    <!-- Connection Selector & Controls -->
    <div class="p-3 border-b border-slate-200 dark:border-slate-800/80 bg-slate-50/70 dark:bg-slate-950/40">
      <div class="text-[11px] font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-2 flex items-center justify-between">
        <span>当前连接实例</span>
        <span
          v-if="activeConnection"
          class="inline-flex items-center gap-1 text-[10px] text-emerald-600 dark:text-emerald-400 font-normal lowercase"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 dark:bg-emerald-400 animate-pulse"></span>
          已配置
        </span>
      </div>

      <div v-if="connections.length > 0" class="space-y-2">
        <div class="relative">
          <select
            :value="activeConnectionId"
            @change="onConnectionChange(($event.target as HTMLSelectElement).value)"
            class="w-full px-3 py-2 rounded-lg bg-white border border-slate-300 text-slate-900 dark:bg-slate-800/90 dark:border-slate-700 dark:text-white text-xs font-medium appearance-none focus:outline-none focus:ring-2 focus:ring-blue-500/40 transition cursor-pointer pr-8 truncate shadow-sm"
          >
            <option v-for="conn in connections" :key="conn.id" :value="conn.id">
              {{ conn.name || conn.url }}
            </option>
          </select>
          <ChevronDown class="w-3.5 h-3.5 text-slate-400 absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
        </div>

        <!-- Connection actions toolbar -->
        <div v-if="activeConnection" class="flex items-center justify-between text-xs px-1 text-slate-500 dark:text-slate-400">
          <div class="truncate text-[11px] font-mono text-slate-500 max-w-[140px]" :title="activeConnection.url">
            {{ activeConnection.url }}
          </div>
          <div class="flex items-center gap-1">
            <button
              @click="openEditModal(activeConnection)"
              class="p-1 hover:text-blue-600 hover:bg-slate-200/60 dark:hover:text-blue-400 dark:hover:bg-slate-800 rounded transition"
              title="编辑配置"
            >
              <Settings class="w-3.5 h-3.5" />
            </button>
            <button
              @click="confirmDelete(activeConnection)"
              class="p-1 hover:text-rose-600 hover:bg-rose-50 dark:hover:text-rose-400 dark:hover:bg-slate-800 rounded transition"
              title="删除连接"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      <!-- No connection state -->
      <div v-else class="text-center py-4 px-2">
        <p class="text-xs text-slate-500 dark:text-slate-400 mb-2.5">暂无已保存连接</p>
        <button
          @click="openCreateModal"
          class="w-full py-1.5 px-3 rounded-lg border border-dashed border-slate-300 hover:border-blue-500 bg-white hover:bg-blue-50/50 text-xs text-slate-600 hover:text-blue-600 dark:border-slate-700 dark:bg-transparent dark:hover:bg-slate-800/50 dark:text-slate-300 dark:hover:text-blue-400 transition flex items-center justify-center gap-1.5 shadow-sm"
        >
          <Plus class="w-3.5 h-3.5" />
          添加首个连接
        </button>
      </div>
    </div>

    <!-- Tree Navigation Slot (Task 7) -->
    <div class="flex-1 overflow-y-auto">
      <slot name="tree">
        <div class="p-4 text-xs text-slate-400 text-center">
          集合导航加载中...
        </div>
      </slot>
    </div>

    <!-- Footer with Theme Toggle -->
    <div class="h-11 px-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-[11px] text-slate-500 dark:text-slate-400 shrink-0 bg-slate-50/50 dark:bg-slate-900/50">
      <span class="flex items-center gap-1.5">
        <span class="w-2 h-2 rounded-full bg-blue-500"></span>
        Wails v3 + Go
      </span>
      <div class="flex items-center gap-2">
        <!-- Theme Switcher Button -->
        <button
          @click="themeStore.toggleTheme"
          class="p-1 rounded-md text-slate-500 hover:text-slate-900 hover:bg-slate-200/70 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800 transition flex items-center gap-1"
          :title="themeStore.isDark.value ? '切换至浅色主题' : '切换至深色主题'"
        >
          <Sun v-if="themeStore.isDark.value" class="w-3.5 h-3.5 text-amber-400" />
          <Moon v-else class="w-3.5 h-3.5 text-slate-600" />
          <span class="text-[10px]">{{ themeStore.isDark.value ? '深色' : '浅色' }}</span>
        </button>
        <span class="text-slate-400 dark:text-slate-600">|</span>
        <span class="font-mono">v0.1.0</span>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import {
  Database,
  Plus,
  ChevronDown,
  Settings,
  Trash2,
  Sun,
  Moon,
} from "lucide-vue-next";
import { useConnectionStore } from "../stores/connection";
import { useThemeStore } from "../stores/theme";
import type { ConnectionConfig } from "../types";

const emit = defineEmits<{
  (e: "connection-changed", connId: string): void;
}>();

const connStore = useConnectionStore();
const themeStore = useThemeStore();

const {
  connections,
  activeConnectionId,
  activeConnection,
  openCreateModal,
  openEditModal,
  setActiveConnection,
  remove,
} = connStore;

function onConnectionChange(id: string) {
  setActiveConnection(id);
  emit("connection-changed", id);
}

async function confirmDelete(conn: ConnectionConfig) {
  if (confirm(`确定要删除连接 "${conn.name || conn.url}" 吗？`)) {
    await remove(conn.id);
  }
}
</script>
