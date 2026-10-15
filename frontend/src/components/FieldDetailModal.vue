<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in text-xs"
    @click.self="close"
    @keydown.esc="close"
  >
    <div
      class="w-full max-w-3xl max-h-[85vh] rounded-xl bg-white border border-slate-200 shadow-2xl overflow-hidden flex flex-col text-slate-800 dark:bg-slate-900 dark:border-slate-700 dark:text-slate-100 transition-colors duration-200"
    >
      <!-- Header -->
      <div class="px-6 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between shrink-0 bg-slate-50/70 dark:bg-slate-950/40">
        <div class="flex items-center gap-3 min-w-0">
          <div
            :class="[
              'w-8 h-8 rounded-lg flex items-center justify-center shrink-0 border',
              isVector
                ? 'bg-indigo-50 border-indigo-200 text-indigo-600 dark:bg-indigo-950/60 dark:border-indigo-700/60 dark:text-indigo-400'
                : isPrimaryKey
                ? 'bg-amber-50 border-amber-200 text-amber-700 dark:bg-amber-950/60 dark:border-amber-700/60 dark:text-amber-300'
                : 'bg-blue-50 border-blue-200 text-blue-600 dark:bg-blue-950/60 dark:border-blue-700/60 dark:text-blue-400'
            ]"
          >
            <Sparkles v-if="isVector" class="w-4 h-4" />
            <Key v-else-if="isPrimaryKey" class="w-4 h-4" />
            <FileText v-else class="w-4 h-4" />
          </div>

          <div class="min-w-0">
            <div class="flex items-center gap-2 flex-wrap">
              <span class="text-sm font-bold text-slate-900 dark:text-white tracking-tight font-mono">{{ fieldName }}</span>
              <span
                v-if="isPrimaryKey"
                class="px-1.5 py-0.5 rounded text-[10px] bg-amber-50 border border-amber-300 text-amber-700 dark:bg-amber-950/80 dark:border-amber-700/60 dark:text-amber-300 font-sans font-medium"
              >
                主键 PK
              </span>
              <span
                v-if="isVector"
                class="px-1.5 py-0.5 rounded text-[10px] bg-indigo-50 border border-indigo-200 text-indigo-700 dark:bg-indigo-950/80 dark:border-indigo-700/60 dark:text-indigo-300 font-sans font-medium"
              >
                向量 VECTOR
              </span>
              <span class="px-2 py-0.5 rounded-full text-[10px] bg-slate-100 border border-slate-200 text-slate-600 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-300 font-mono">
                {{ typeBadgeText }}
              </span>
            </div>
            <p v-if="docId" class="text-[11px] text-slate-500 dark:text-slate-400 font-mono mt-0.5 truncate">
              所属文档 ID: <span class="text-slate-700 dark:text-slate-300 font-semibold">{{ docId }}</span>
            </p>
          </div>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <!-- Toggle format for string containing JSON -->
          <div v-if="canFormatAsJson" class="flex items-center bg-slate-200/80 dark:bg-slate-800 p-0.5 rounded-lg text-[11px] mr-1">
            <button
              type="button"
              @click="viewMode = 'raw'"
              :class="[
                'px-2 py-0.5 rounded transition font-medium',
                viewMode === 'raw'
                  ? 'bg-white text-blue-600 dark:bg-slate-700 dark:text-white shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              ]"
            >
              原始文本
            </button>
            <button
              type="button"
              @click="viewMode = 'json'"
              :class="[
                'px-2 py-0.5 rounded transition font-medium',
                viewMode === 'json'
                  ? 'bg-white text-blue-600 dark:bg-slate-700 dark:text-white shadow-sm'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'
              ]"
            >
              JSON 格式化
            </button>
          </div>

          <!-- Copy Button -->
          <button
            type="button"
            @click="copyContent"
            class="px-2.5 py-1.5 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 text-[11px] font-medium transition flex items-center gap-1.5 shadow-sm"
            title="复制字段完整内容"
          >
            <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-500" />
            <Copy v-else class="w-3.5 h-3.5 text-slate-400" />
            <span>{{ copied ? "已复制" : "复制" }}</span>
          </button>

          <!-- Close Button -->
          <button
            type="button"
            @click="close"
            class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
            title="关闭 (Esc)"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Content Area -->
      <div class="flex-1 overflow-y-auto p-4 sm:p-6 bg-slate-50/60 dark:bg-slate-950 font-mono text-xs leading-relaxed text-slate-800 dark:text-slate-200 select-text">
        <div v-if="isNullOrUndefined" class="py-12 text-center text-slate-400 dark:text-slate-500 italic">
          (空值 / null)
        </div>
        <pre
          v-else
          class="whitespace-pre-wrap break-all font-mono select-text"
        >{{ displayedContent }}</pre>
      </div>

      <!-- Footer -->
      <div class="px-6 py-2.5 border-t border-slate-200 bg-slate-50/70 dark:border-slate-800 dark:bg-slate-900 flex items-center justify-between shrink-0 text-[11px] text-slate-500 dark:text-slate-400">
        <div class="flex items-center gap-3">
          <span>{{ statsSummary }}</span>
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="copyContent"
            class="px-3 py-1 rounded-md border border-slate-200 bg-white hover:bg-slate-100 text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 transition font-medium flex items-center gap-1"
          >
            <Copy class="w-3 h-3 text-slate-400" />
            <span>{{ copied ? "已复制" : "复制完整内容" }}</span>
          </button>
          <button
            type="button"
            @click="close"
            class="px-3.5 py-1 rounded-md bg-blue-600 hover:bg-blue-700 text-white font-medium transition shadow-sm"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { FileText, Sparkles, Key, Copy, Check, X } from "lucide-vue-next";

const props = defineProps<{
  isOpen: boolean;
  fieldName: string;
  fieldValue: unknown;
  docId?: string;
  isPrimaryKey?: boolean;
  isVector?: boolean;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

const copied = ref(false);
const viewMode = ref<"raw" | "json">("raw");

const isNullOrUndefined = computed(() => {
  return props.fieldValue === null || props.fieldValue === undefined;
});

// Detect whether a string value is valid JSON
const parsedJsonFromString = computed(() => {
  if (typeof props.fieldValue !== "string") return null;
  const trimmed = props.fieldValue.trim();
  if ((trimmed.startsWith("{") && trimmed.endsWith("}")) || (trimmed.startsWith("[") && trimmed.endsWith("]"))) {
    try {
      return JSON.parse(trimmed);
    } catch {
      return null;
    }
  }
  return null;
});

const canFormatAsJson = computed(() => {
  return parsedJsonFromString.value !== null;
});

// Default to JSON mode if string contains valid JSON object/array
watch(
  () => [props.isOpen, props.fieldValue],
  () => {
    if (canFormatAsJson.value) {
      viewMode.value = "json";
    } else {
      viewMode.value = "raw";
    }
    copied.value = false;
  },
  { immediate: true }
);

const typeBadgeText = computed(() => {
  const val = props.fieldValue;
  if (val === null) return "null";
  if (val === undefined) return "undefined";
  if (props.isVector || (Array.isArray(val) && val.length > 0 && typeof val[0] === "number")) {
    return `Vector [${(val as number[]).length} 维]`;
  }
  if (Array.isArray(val)) {
    return `Array [${val.length} 项]`;
  }
  if (typeof val === "object") {
    return `Object {${Object.keys(val).length} 键}`;
  }
  if (typeof val === "number") {
    return Number.isInteger(val) ? "Integer" : "Float";
  }
  if (typeof val === "boolean") {
    return "Boolean";
  }
  if (typeof val === "string") {
    return `String (${val.length} 字符)`;
  }
  return typeof val;
});

const displayedContent = computed(() => {
  const val = props.fieldValue;
  if (val === null) return "null";
  if (val === undefined) return "undefined";

  if (typeof val === "string") {
    if (viewMode.value === "json" && parsedJsonFromString.value !== null) {
      return JSON.stringify(parsedJsonFromString.value, null, 2);
    }
    return val;
  }

  if (typeof val === "object") {
    try {
      return JSON.stringify(val, null, 2);
    } catch {
      return String(val);
    }
  }

  return String(val);
});

const statsSummary = computed(() => {
  const val = props.fieldValue;
  if (val === null || val === undefined) return "无内容";

  if (typeof val === "string") {
    const lines = val.split("\n").length;
    return `共 ${val.length} 个字符 · ${lines} 行`;
  }

  if (Array.isArray(val)) {
    return `共 ${val.length} 个元素`;
  }

  if (typeof val === "object") {
    return `共 ${Object.keys(val).length} 个属性`;
  }

  return `类型: ${typeof val}`;
});

async function copyContent() {
  if (isNullOrUndefined.value) return;
  try {
    await navigator.clipboard.writeText(displayedContent.value);
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 2000);
  } catch {
    // fallback
  }
}

function close() {
  emit("close");
}

function handleKeyDown(e: KeyboardEvent) {
  if (props.isOpen && e.key === "Escape") {
    close();
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeyDown);
});
</script>
