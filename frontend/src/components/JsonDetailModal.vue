<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-fade-in"
    @keydown.esc="close"
  >
    <div
      class="w-full max-w-3xl max-h-[85vh] rounded-xl bg-slate-900 border border-slate-700 shadow-2xl overflow-hidden flex flex-col text-slate-100"
    >
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between shrink-0">
        <div class="flex items-center gap-2.5">
          <div class="w-7 h-7 rounded-lg bg-blue-600/20 border border-blue-500/30 flex items-center justify-center text-blue-400">
            <FileCode class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-white tracking-tight">文档详情 JSON</h3>
            <p v-if="docId" class="text-[11px] text-slate-400 font-mono mt-0.5">ID: {{ docId }}</p>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <button
            @click="copyJson"
            class="px-3 py-1.5 rounded-lg border border-slate-700 hover:border-slate-600 bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-200 transition flex items-center gap-1.5"
          >
            <Check v-if="copied" class="w-3.5 h-3.5 text-emerald-400" />
            <Copy v-else class="w-3.5 h-3.5 text-slate-400" />
            <span>{{ copied ? '已复制' : '复制 JSON' }}</span>
          </button>
          <button
            @click="close"
            class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition"
          >
            <X class="w-5 h-5" />
          </button>
        </div>
      </div>

      <!-- Vector Dimensions Pill bar if vector fields present -->
      <div v-if="vectorFields.length > 0" class="px-6 py-2 bg-slate-950/60 border-b border-slate-800/80 flex items-center gap-2 flex-wrap shrink-0">
        <span class="text-[11px] text-slate-400">包含向量字段:</span>
        <span
          v-for="vf in vectorFields"
          :key="vf.name"
          class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md bg-indigo-950/50 border border-indigo-700/50 text-indigo-300 text-[11px] font-mono"
        >
          <span class="font-medium text-indigo-200">{{ vf.name }}</span>
          <span class="text-[10px] text-indigo-400">({{ vf.dim }} 维)</span>
        </span>
      </div>

      <!-- JSON Content -->
      <div class="flex-1 overflow-y-auto p-6 bg-slate-950 font-mono text-xs leading-relaxed text-slate-300">
        <pre class="whitespace-pre-wrap break-all select-text">{{ formattedJson }}</pre>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-800 bg-slate-900 flex justify-end shrink-0">
        <button
          @click="close"
          class="px-4 py-1.5 rounded-lg text-xs font-medium bg-slate-800 hover:bg-slate-700 text-white transition"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { FileCode, Copy, Check, X } from "lucide-vue-next";

const props = defineProps<{
  isOpen: boolean;
  document: Record<string, unknown> | null;
}>();

const emit = defineEmits<{
  (e: "close"): void;
}>();

const copied = ref(false);

const docId = computed(() => {
  if (!props.document) return "";
  const candidate = props.document["id"] || props.document["_id"];
  return candidate ? String(candidate) : "";
});

const vectorFields = computed(() => {
  if (!props.document) return [];
  const fields: { name: string; dim: number }[] = [];
  for (const [key, val] of Object.entries(props.document)) {
    if (Array.isArray(val) && val.length > 0 && typeof val[0] === "number") {
      fields.push({ name: key, dim: val.length });
    }
  }
  return fields;
});

const formattedJson = computed(() => {
  if (!props.document) return "";
  return JSON.stringify(props.document, null, 2);
});

async function copyJson() {
  if (!formattedJson.value) return;
  try {
    await navigator.clipboard.writeText(formattedJson.value);
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
</script>
