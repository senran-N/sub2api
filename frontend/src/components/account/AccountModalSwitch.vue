<template>
  <button
    type="button"
    :class="trackClasses"
    :disabled="disabled"
    :aria-label="ariaLabel"
    :aria-pressed="modelValue"
    @click="emit('update:modelValue', !modelValue)"
  >
    <span :class="thumbClasses" />
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    disabled?: boolean;
    ariaLabel?: string;
  }>(),
  {
    disabled: false,
    ariaLabel: undefined,
  },
);

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
}>();

const trackClasses = computed(() => [
  "account-modal-switch",
  "relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none",
  props.modelValue
    ? "account-modal-switch--enabled"
    : "account-modal-switch--disabled",
  props.disabled && "account-modal-switch--disabled-control",
]);

const thumbClasses = computed(() => [
  "account-modal-switch__thumb",
  "pointer-events-none inline-block h-5 w-5 transform rounded-full shadow ring-0 transition duration-200 ease-in-out",
  props.modelValue ? "translate-x-5" : "translate-x-0",
]);
</script>

<style scoped>
.account-modal-switch {
  box-shadow: 0 0 0 1px
    color-mix(in srgb, var(--theme-page-border) 40%, transparent);
}

.account-modal-switch:focus-visible {
  box-shadow:
    0 0 0 2px color-mix(in srgb, var(--theme-accent) 22%, transparent),
    0 0 0 4px color-mix(in srgb, var(--theme-accent) 12%, transparent);
}

.account-modal-switch--enabled {
  background: var(--theme-accent);
}

.account-modal-switch--disabled {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 76%,
    var(--theme-surface)
  );
}

.account-modal-switch--disabled-control {
  cursor: not-allowed;
  opacity: 0.62;
}

.account-modal-switch__thumb {
  background: var(--theme-surface-contrast);
}
</style>
