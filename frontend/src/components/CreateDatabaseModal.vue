<template>
  <Teleport to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-xs select-none"
      @click.self="handleClose"
    >
      <div class="w-full max-w-md rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200">
        <!-- Header -->
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0">
          <div class="flex items-center gap-2 font-semibold text-sm text-slate-900 dark:text-white">
            <Database class="w-4 h-4 text-blue-500" />
            <span>新建数据库</span>
          </div>
          <button
            type="button"
            @click="handleClose"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Form -->
        <div class="p-5 space-y-4">
          <!-- Database Name -->
          <div class="space-y-1.5">
            <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">
              数据库名称 <span class="text-rose-500">*</span>
            </label>
            <input
              ref="nameInputRef"
              v-model="name"
              type="text"
              placeholder="如 my_database"
              class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/40 focus:border-blue-500 transition"
              :disabled="loading"
              @keydown.enter="handleCreate"
            />
            <p class="text-[10px] text-slate-400 dark:text-slate-500">
              1-128 个字符，字母或数字开头，仅含字母、数字、下划线或连字符
            </p>
          </div>

          <!-- Database Type -->
          <div class="space-y-1.5">
            <label class="block text-[11px] font-medium text-slate-600 dark:text-slate-300">数据库类型</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                :class="[
                  'px-3 py-2.5 rounded-lg border text-left transition',
                  dbType === 'base'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/40 ring-1 ring-blue-500/40'
                    : 'border-slate-300 dark:border-slate-700 hover:border-slate-400 dark:hover:border-slate-600',
                ]"
                :disabled="loading"
                @click="dbType = 'base'"
              >
                <div class="flex items-center gap-1.5 font-medium text-slate-800 dark:text-slate-100">
                  <Database class="w-3.5 h-3.5 text-blue-500" />
                  <span>基础向量库</span>
                </div>
                <p class="mt-0.5 text-[10px] text-slate-500 dark:text-slate-400">BASE · 手动定义字段与索引</p>
              </button>
              <button
                type="button"
                :class="[
                  'px-3 py-2.5 rounded-lg border text-left transition',
                  dbType === 'ai'
                    ? 'border-purple-500 bg-purple-50 dark:bg-purple-950/40 ring-1 ring-purple-500/40'
                    : 'border-slate-300 dark:border-slate-700 hover:border-slate-400 dark:hover:border-slate-600',
                ]"
                :disabled="loading"
                @click="dbType = 'ai'"
              >
                <div class="flex items-center gap-1.5 font-medium text-slate-800 dark:text-slate-100">
                  <BookOpen class="w-3.5 h-3.5 text-purple-500" />
                  <span>AI 知识库</span>
                </div>
                <p class="mt-0.5 text-[10px] text-slate-500 dark:text-slate-400">AI_DB · 知识库文件集合</p>
              </button>
            </div>
          </div>

          <!-- Error -->
          <div
            v-if="error"
            class="p-2.5 rounded-lg bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-950/30 dark:border-rose-800/50 dark:text-rose-300 flex items-start gap-1.5"
          >
            <AlertCircle class="w-3.5 h-3.5 shrink-0 mt-0.5" />
            <span class="break-all">{{ error }}</span>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-950/30 flex items-center justify-end gap-2">
          <button
            type="button"
            @click="handleClose"
            class="px-4 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
            :disabled="loading"
          >
            取消
          </button>
          <button
            type="button"
            @click="handleCreate"
            class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium transition shadow-sm flex items-center gap-1.5 disabled:opacity-60"
            :disabled="loading || !name.trim()"
          >
            <Loader2 v-if="loading" class="w-3.5 h-3.5 animate-spin" />
            <span>{{ loading ? "创建中..." : "创建" }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted, onUnmounted } from "vue";
import { Database, BookOpen, X, AlertCircle, Loader2 } from "lucide-vue-next";
import { useVectorDBStore } from "../stores/vectordb";
import { useConnectionStore } from "../stores/connection";

const vdbStore = useVectorDBStore();
const connStore = useConnectionStore();

const name = ref("");
const dbType = ref("base");
const loading = ref(false);
const error = ref("");
const nameInputRef = ref<HTMLInputElement | null>(null);

const emit = defineEmits<{
  (e: "closed"): void;
  (e: "created", database: string, dbType: string): void;
}>();

function handleClose() {
  if (loading.value) return;
  emit("closed");
}

const NAME_RE = /^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/;

async function handleCreate() {
  const dbName = name.value.trim();
  if (!dbName || loading.value) return;

  if (!NAME_RE.test(dbName)) {
    error.value = "名称须为 1-128 个字符，字母或数字开头，仅含字母、数字、下划线或连字符";
    return;
  }

  loading.value = true;
  error.value = "";
  try {
    await vdbStore.createDatabase(connStore.activeConnectionId?.value || "", dbName, dbType.value);
    emit("created", dbName, dbType.value);
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === "Escape" && !loading.value) {
    handleClose();
  }
}

onMounted(async () => {
  window.addEventListener("keydown", handleKeyDown);
  await nextTick();
  nameInputRef.value?.focus();
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeyDown);
});
</script>
