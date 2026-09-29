<template>
  <div ref="rootEl" class="relative inline-block" :class="wrapperClass">
    <button
      type="button"
      :id="buttonId"
      :disabled="disabled"
      @click="toggle"
      @keydown.down.prevent="openMenu"
      @keydown.up.prevent="openMenu"
      @keydown.enter.prevent="toggle"
      @keydown.space.prevent="toggle"
      class="w-full flex items-center justify-between gap-2 text-left focus:outline-none focus:ring-2 focus:ring-blue-500/40 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
      :class="[triggerClass, isOpen ? 'ring-2 ring-blue-500/40' : '']"
      :title="title"
    >
      <span class="truncate">{{ selectedLabel ?? placeholder }}</span>
      <ChevronDown
        class="w-3.5 h-3.5 shrink-0 text-slate-400 transition-transform duration-150"
        :class="isOpen ? 'rotate-180' : ''"
      />
    </button>

    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <ul
        v-if="isOpen"
        class="absolute z-50 mt-1 max-h-60 overflow-y-auto rounded-lg py-1 border shadow-lg bg-white border-slate-200 dark:bg-slate-800 dark:border-slate-700 text-slate-800 dark:text-slate-200"
        :class="menuClass"
        role="listbox"
      >
        <li
          v-for="opt in options"
          :key="String(opt.value)"
          role="option"
          :aria-selected="opt.value === modelValue"
          class="px-3 py-1.5 cursor-pointer flex items-center justify-between gap-2 hover:bg-blue-50 dark:hover:bg-slate-700/80"
          :class="opt.value === modelValue ? 'text-blue-600 dark:text-blue-400 font-medium' : ''"
          @click="select(opt.value)"
          @mousemove="hoverIndex = index(opt.value)"
        >
          <span class="truncate">{{ opt.label }}</span>
          <Check v-if="opt.value === modelValue" class="w-3.5 h-3.5 shrink-0" />
        </li>
      </ul>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ChevronDown, Check } from "lucide-vue-next";

export interface SelectOption {
  value: string | number;
  label: string;
}

const props = withDefaults(
  defineProps<{
    modelValue: string | number;
    options: SelectOption[];
    placeholder?: string;
    disabled?: boolean;
    /** Classes for the trigger button (padding/bg/border/text-size), mirrors old select styling */
    triggerClass?: string;
    /** Classes for the dropdown menu panel */
    menuClass?: string;
    /** Classes for the outer wrapper */
    wrapperClass?: string;
    buttonId?: string;
    title?: string;
  }>(),
  {
    placeholder: "",
    disabled: false,
    triggerClass: "",
    menuClass: "",
    wrapperClass: "",
    buttonId: undefined,
    title: undefined,
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string | number): void;
  (e: "change", value: string | number): void;
}>();

const isOpen = ref(false);
const rootEl = ref<HTMLElement | null>(null);
const hoverIndex = ref(0);

const selectedLabel = computed(
  () => props.options.find((o) => o.value === props.modelValue)?.label
);

function index(value: string | number): number {
  return props.options.findIndex((o) => o.value === value);
}

function toggle() {
  isOpen.value ? close() : openMenu();
}

function openMenu() {
  if (props.disabled) return;
  isOpen.value = true;
  hoverIndex.value = Math.max(0, index(props.modelValue));
}

function close() {
  isOpen.value = false;
}

function select(value: string | number) {
  if (value !== props.modelValue) {
    emit("update:modelValue", value);
    emit("change", value);
  }
  close();
  rootEl.value?.querySelector("button")?.focus();
}

function onDocClick(e: MouseEvent) {
  if (isOpen.value && rootEl.value && !rootEl.value.contains(e.target as Node)) {
    close();
  }
}

function onKeydown(e: KeyboardEvent) {
  if (!isOpen.value) return;
  const opts = props.options;
  if (!opts.length) return;
  if (e.key === "Escape") {
    close();
    rootEl.value?.querySelector("button")?.focus();
  } else if (e.key === "ArrowDown") {
    e.preventDefault();
    hoverIndex.value = (hoverIndex.value + 1) % opts.length;
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    hoverIndex.value = (hoverIndex.value - 1 + opts.length) % opts.length;
  } else if (e.key === "Enter") {
    e.preventDefault();
    const opt = opts[hoverIndex.value];
    if (opt) select(opt.value);
  }
}

onMounted(() => {
  document.addEventListener("mousedown", onDocClick);
  document.addEventListener("keydown", onKeydown);
});

onBeforeUnmount(() => {
  document.removeEventListener("mousedown", onDocClick);
  document.removeEventListener("keydown", onKeydown);
});

// Keep hover in sync when value changes externally while open
watch(
  () => props.modelValue,
  () => {
    if (isOpen.value) hoverIndex.value = Math.max(0, index(props.modelValue));
  }
);
</script>
