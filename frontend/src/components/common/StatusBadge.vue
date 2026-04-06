<template>
  <div class="status-badge">
    <span
      :class="[
        'status-badge__dot',
        variantClass
      ]"
    ></span>
    <span class="status-badge__label">
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  label: string
}>()

const variantClass = computed(() => {
  switch (props.status) {
    case 'active':
    case 'success':
      return 'status-badge__dot--success'
    case 'disabled':
    case 'inactive':
    case 'warning':
      return 'status-badge__dot--warning'
    case 'error':
    case 'danger':
      return 'status-badge__dot--danger'
    default:
      return 'status-badge__dot--neutral'
  }
})
</script>

<style scoped>
.status-badge {
  @apply flex items-center gap-1.5;
}

.status-badge__dot {
  @apply inline-block h-2 w-2 rounded-full;
}

.status-badge__dot--success {
  background: rgb(var(--theme-success-rgb));
}

.status-badge__dot--warning {
  background: rgb(var(--theme-warning-rgb));
}

.status-badge__dot--danger {
  background: rgb(var(--theme-danger-rgb));
}

.status-badge__dot--neutral {
  background: color-mix(in srgb, var(--theme-page-muted) 66%, transparent);
}

.status-badge__label {
  @apply text-sm;
  color: var(--theme-page-text);
}
</style>
