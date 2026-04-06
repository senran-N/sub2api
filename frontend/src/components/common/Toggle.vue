<template>
  <button
    type="button"
    @click="toggle"
    class="toggle-switch relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
    :class="[modelValue ? 'toggle-switch--active' : 'toggle-switch--inactive']"
    role="switch"
    :aria-checked="modelValue"
  >
    <span
      class="toggle-switch__thumb pointer-events-none inline-block h-5 w-5 transform rounded-full ring-0 transition duration-200 ease-in-out"
      :class="[modelValue ? 'translate-x-5' : 'translate-x-0']"
    />
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

function toggle() {
  emit('update:modelValue', !props.modelValue)
}
</script>

<style scoped>
.toggle-switch {
  border-radius: 9999px;
}

.toggle-switch:focus {
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent-soft) 88%, transparent);
}

.toggle-switch--active {
  background: var(--theme-accent);
}

.toggle-switch--inactive {
  background: color-mix(in srgb, var(--theme-page-border) 84%, transparent);
}

.toggle-switch__thumb {
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}
</style>
