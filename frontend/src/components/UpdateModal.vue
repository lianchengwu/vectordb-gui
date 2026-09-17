<template>
  <div
    v-if="isModalOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in"
    @keydown.esc="closeModal"
  >
    <div
      class="w-full max-w-lg rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200"
    >
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
        <div class="flex items-center gap-2">
          <div class="w-7 h-7 rounded-lg bg-blue-50 border border-blue-200 text-blue-600 dark:bg-blue-600/20 dark:border-blue-500/30 dark:text-blue-400 flex items-center justify-center">
            <Sparkles class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-slate-900 dark:text-white tracking-tight">检查软件更新</h3>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 font-mono">当前版本: {{ CURRENT_VERSION }}</p>
          </div>
        </div>

        <button
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
        >
          <X class="w-4 h-4" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 text-xs space-y-4">
        <!-- Case 1: Has new version -->
        <div v-if="hasUpdate && latestRelease" class="space-y-4">
          <div class="p-3.5 rounded-xl bg-blue-50 border border-blue-200 dark:bg-blue-950/40 dark:border-blue-800/60 flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-ping"></span>
              <div>
                <p class="font-semibold text-blue-950 dark:text-blue-200 text-sm">发现新版本: {{ latestRelease.tagName }}</p>
                <p class="text-[11px] text-blue-700 dark:text-blue-300/80 mt-0.5">发布时间: {{ formatTime(latestRelease.publishedAt) }}</p>
              </div>
            </div>
            <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-blue-600 text-white shadow-sm">NEW</span>
          </div>

          <!-- Release Notes -->
          <div v-if="latestRelease.notes" class="space-y-1.5">
            <h4 class="font-medium text-slate-700 dark:text-slate-300 text-[11px] uppercase tracking-wider">更新日志 (Release Notes)</h4>
            <div class="p-3 rounded-lg bg-slate-50 border border-slate-200 dark:bg-slate-950 dark:border-slate-800 max-h-48 overflow-y-auto text-slate-700 dark:text-slate-300 font-mono text-[11px] whitespace-pre-wrap leading-relaxed">
              {{ latestRelease.notes }}
            </div>
          </div>

          <!-- Download Packages list if assets present -->
          <div v-if="latestRelease.assets.length > 0" class="space-y-2">
            <h4 class="font-medium text-slate-700 dark:text-slate-300 text-[11px] uppercase tracking-wider">各平台安装包下载</h4>
            <div class="grid grid-cols-1 gap-1.5">
              <button
                v-for="asset in latestRelease.assets"
                :key="asset.name"
                type="button"
                @click="openReleaseUrl(asset.browser_download_url)"
                class="px-3 py-2 rounded-lg border border-slate-200 bg-white hover:bg-blue-50 hover:border-blue-300 text-slate-800 dark:border-slate-800 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition flex items-center justify-between shadow-sm group text-left"
              >
                <div class="flex items-center gap-2 truncate">
                  <Download class="w-3.5 h-3.5 text-blue-500 group-hover:translate-y-0.5 transition-transform shrink-0" />
                  <span class="font-mono text-xs truncate">{{ asset.name }}</span>
                </div>
                <span class="text-[10px] text-slate-400 font-mono ml-2 shrink-0">{{ formatSize(asset.size) }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Case 2: Already latest version -->
        <div v-else-if="feedbackType === 'info'" class="py-6 text-center space-y-2">
          <div class="w-12 h-12 rounded-full bg-emerald-50 border border-emerald-200 text-emerald-600 dark:bg-emerald-950/50 dark:border-emerald-700/50 dark:text-emerald-400 mx-auto flex items-center justify-center shadow-inner">
            <CheckCircle2 class="w-6 h-6" />
          </div>
          <p class="font-medium text-sm text-slate-900 dark:text-white">您当前使用的是最新版本</p>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-mono">{{ feedbackMessage }}</p>
        </div>

        <!-- Case 3: Error checking updates -->
        <div v-else-if="feedbackType === 'error'" class="py-6 text-center space-y-2">
          <div class="w-12 h-12 rounded-full bg-rose-50 border border-rose-200 text-rose-600 dark:bg-rose-950/50 dark:border-rose-700/50 dark:text-rose-400 mx-auto flex items-center justify-center shadow-inner">
            <AlertCircle class="w-6 h-6" />
          </div>
          <p class="font-medium text-sm text-slate-900 dark:text-white">检查更新失败</p>
          <p class="text-xs text-rose-600 dark:text-rose-300 font-mono">{{ feedbackMessage }}</p>
        </div>
      </div>

      <!-- Footer Actions -->
      <div class="px-6 py-3.5 bg-slate-50 dark:bg-slate-900/90 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
        <button
          v-if="hasUpdate && latestRelease"
          type="button"
          @click="openReleaseUrl(latestRelease.url)"
          class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 text-xs font-medium transition flex items-center gap-1.5 shadow-sm"
        >
          <ExternalLink class="w-3.5 h-3.5 text-blue-500" />
          打开 GitHub Release
        </button>
        <span v-else></span>

        <button
          @click="closeModal"
          class="px-4 py-1.5 rounded-lg text-xs font-medium bg-slate-200 hover:bg-slate-300 text-slate-800 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-white transition"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Sparkles,
  X,
  Download,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
} from "lucide-vue-next";
import { useUpdaterStore } from "../stores/updater";

const updaterStore = useUpdaterStore();
const {
  CURRENT_VERSION,
  hasUpdate,
  latestRelease,
  isModalOpen,
  feedbackMessage,
  feedbackType,
  openReleaseUrl,
  closeModal,
} = updaterStore;

function formatTime(isoStr: string): string {
  if (!isoStr) return "-";
  try {
    const d = new Date(isoStr);
    return d.toLocaleDateString() + " " + d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return isoStr;
  }
}

function formatSize(bytes: number): string {
  if (!bytes || bytes <= 0) return "";
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
}
</script>
