<template>
  <BaseDialog :show="show" :title="t('usage.exporting')" width="narrow" @close="handleCancel">
    <div class="export-progress-dialog space-y-4">
      <div class="export-progress-dialog__meta text-sm">
        {{ t('usage.exportingProgress') }}
      </div>
      <div class="export-progress-dialog__stats flex items-center justify-between text-sm">
        <span>{{ t('usage.exportedCount', { current, total }) }}</span>
        <span class="export-progress-dialog__value font-medium">{{ normalizedProgress }}%</span>
      </div>
      <div class="export-progress-dialog__track h-2 w-full rounded-full">
        <div
          role="progressbar"
          :aria-valuenow="normalizedProgress"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-label="`${t('usage.exportingProgress')}: ${normalizedProgress}%`"
          class="export-progress-dialog__bar h-2 rounded-full transition-all"
          :style="{ width: `${normalizedProgress}%` }"
        ></div>
      </div>
      <div v-if="estimatedTime" class="export-progress-dialog__hint text-xs" aria-live="polite" aria-atomic="true">
        {{ t('usage.estimatedTime', { time: estimatedTime }) }}
      </div>
    </div>

    <template #footer>
      <button
        @click="handleCancel"
        type="button"
        class="btn btn-secondary btn-md"
      >
        {{ t('usage.cancelExport') }}
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from './BaseDialog.vue'

interface Props {
  show: boolean
  progress: number
  current: number
  total: number
  estimatedTime: string
}

interface Emits {
  (e: 'cancel'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()
const { t } = useI18n()

const normalizedProgress = computed(() => {
  const value = Number.isFinite(props.progress) ? props.progress : 0
  return Math.min(100, Math.max(0, Math.round(value)))
})

const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.export-progress-dialog__meta,
.export-progress-dialog__stats,
.export-progress-dialog__hint {
  color: var(--theme-page-muted);
}

.export-progress-dialog__value {
  color: var(--theme-page-text);
}

.export-progress-dialog__track {
  background: color-mix(in srgb, var(--theme-page-border) 84%, transparent);
}

.export-progress-dialog__bar {
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%)
  );
}
</style>
