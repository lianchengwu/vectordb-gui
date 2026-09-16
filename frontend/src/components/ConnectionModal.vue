<template>
  <div
    v-if="isModalOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 animate-fade-in"
  >
    <div class="w-full max-w-lg rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <h3 class="text-lg font-semibold tracking-tight text-slate-900 dark:text-white flex items-center gap-2">
          <Database class="w-5 h-5 text-blue-500" />
          {{ isEditing ? '编辑连接配置' : '新建 VectorDB 连接' }}
        </h3>
        <button
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Form -->
      <div class="px-6 py-5 space-y-4 text-sm">
        <!-- Name -->
        <div>
          <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">连接名称</label>
          <input
            v-model="form.name"
            type="text"
            placeholder="例如: 测试环境集群 / Prod Cluster"
            class="w-full px-3.5 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition shadow-inner"
          />
        </div>

        <!-- URL -->
        <div>
          <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
            实例地址 (URL) <span class="text-rose-500">*</span>
          </label>
          <input
            v-model="form.url"
            type="text"
            placeholder="http://10.0.x.x:80 或公网域名:端口"
            class="w-full px-3.5 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition font-mono text-xs shadow-inner"
          />
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">例如: http://lb-xxxx.clb.ap-guangzhou.tencentclb.com:50000</p>
        </div>

        <!-- Username & Timeout Row -->
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
              账号 (Username) <span class="text-rose-500">*</span>
            </label>
            <input
              v-model="form.username"
              type="text"
              placeholder="默认 root"
              class="w-full px-3.5 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition shadow-inner"
            />
          </div>
          <div>
            <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">超时时间 (秒)</label>
            <input
              v-model.number="form.timeout"
              type="number"
              min="1"
              max="300"
              class="w-full px-3.5 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition shadow-inner"
            />
          </div>
        </div>

        <!-- API Key -->
        <div>
          <label class="block font-medium text-slate-700 dark:text-slate-300 mb-1.5">
            API Key (密码凭证) <span class="text-rose-500">*</span>
          </label>
          <div class="relative">
            <input
              v-model="form.apiKey"
              :type="showApiKey ? 'text' : 'password'"
              placeholder="腾讯云控制台获取的 API Key"
              class="w-full pl-3.5 pr-10 py-2 rounded-lg bg-slate-50 border border-slate-300 text-slate-900 placeholder-slate-400 dark:bg-slate-800/80 dark:border-slate-700 dark:text-white dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition font-mono text-xs shadow-inner"
            />
            <button
              type="button"
              @click="showApiKey = !showApiKey"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-white p-1"
            >
              <EyeOff v-if="showApiKey" class="w-4 h-4" />
              <Eye v-else class="w-4 h-4" />
            </button>
          </div>
        </div>

        <!-- Test result banner -->
        <div
          v-if="testResult"
          :class="[
            'p-3 rounded-lg border text-xs flex items-start gap-2.5 transition shadow-sm',
            testResult.success
              ? 'bg-emerald-50 border-emerald-200 text-emerald-800 dark:bg-emerald-950/40 dark:border-emerald-800/60 dark:text-emerald-300'
              : 'bg-rose-50 border-rose-200 text-rose-800 dark:bg-rose-950/40 dark:border-rose-800/60 dark:text-rose-300',
          ]"
        >
          <CheckCircle2 v-if="testResult.success" class="w-4 h-4 mt-0.5 shrink-0 text-emerald-500 dark:text-emerald-400" />
          <AlertCircle v-else class="w-4 h-4 mt-0.5 shrink-0 text-rose-500 dark:text-rose-400" />
          <div class="flex-1 break-all">
            <span class="font-semibold">{{ testResult.success ? '测试连接成功' : '测试连接失败' }}:</span>
            {{ testResult.message }}
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="px-6 py-4 bg-slate-50 dark:bg-slate-900/90 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
        <button
          type="button"
          @click="onTest"
          :disabled="isTesting || !form.url || !form.apiKey"
          class="px-3.5 py-2 rounded-lg text-xs font-medium border border-slate-300 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-750 dark:text-slate-200 transition flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
        >
          <Loader2 v-if="isTesting" class="w-3.5 h-3.5 animate-spin" />
          <Radio v-else class="w-3.5 h-3.5 text-blue-500" />
          测试连通性
        </button>

        <div class="flex items-center gap-2.5">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 rounded-lg text-xs font-medium text-slate-600 hover:text-slate-900 hover:bg-slate-200/60 dark:text-slate-400 dark:hover:text-white dark:hover:bg-slate-800 transition"
          >
            取消
          </button>
          <button
            type="button"
            @click="onSave"
            :disabled="isSaving || !form.url || !form.apiKey"
            class="px-4 py-2 rounded-lg text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white transition flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
          >
            <Loader2 v-if="isSaving" class="w-3.5 h-3.5 animate-spin" />
            <Save v-else class="w-3.5 h-3.5" />
            保存连接
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import {
  Database,
  X,
  Eye,
  EyeOff,
  Radio,
  Save,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from "lucide-vue-next";
import { useConnectionStore } from "../stores/connection";
import type { ConnectionConfig, TestResult } from "../types";

const store = useConnectionStore();
const { isModalOpen, editingConnection, closeModal, save, test } = store;

const showApiKey = ref(false);
const isTesting = ref(false);
const isSaving = ref(false);
const testResult = ref<TestResult | null>(null);

const form = ref<ConnectionConfig>({
  id: "",
  name: "",
  url: "http://",
  username: "root",
  apiKey: "",
  timeout: 10,
});

const isEditing = computed(() => !!editingConnection.value?.id);

watch(
  () => editingConnection.value,
  (val) => {
    if (val) {
      form.value = {
        id: val.id || "",
        name: val.name || "",
        url: val.url || "http://",
        username: val.username || "root",
        apiKey: val.apiKey || "",
        timeout: val.timeout || 10,
      };
      testResult.value = null;
      showApiKey.value = false;
    }
  },
  { immediate: true }
);

async function onTest() {
  isTesting.value = true;
  testResult.value = null;
  try {
    const res = await test(form.value);
    testResult.value = res;
  } finally {
    isTesting.value = false;
  }
}

async function onSave() {
  if (!form.value.url || !form.value.apiKey) return;
  if (!form.value.name.trim()) {
    form.value.name = form.value.url.replace(/^https?:\/\//, "");
  }
  isSaving.value = true;
  try {
    await save(form.value);
  } finally {
    isSaving.value = false;
  }
}
</script>
