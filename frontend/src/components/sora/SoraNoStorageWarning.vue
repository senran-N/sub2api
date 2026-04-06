<template>
  <div class="sora-no-storage-warning">
    <span>⚠️</span>
    <div>
      <p class="sora-no-storage-title">{{ t('sora.noStorageWarningTitle') }}</p>
      <p class="sora-no-storage-desc">{{ t('sora.noStorageWarningDesc') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<style scoped>
.sora-no-storage-warning {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 14px 20px;
  background: color-mix(in srgb, var(--sora-warning) 8%, transparent);
  border: 1px solid color-mix(in srgb, var(--sora-warning) 20%, transparent);
  border-radius: 12px;
  font-size: 13px;
}

.sora-no-storage-title {
  font-weight: 600;
  color: var(--sora-warning);
  margin-bottom: 4px;
}

.sora-no-storage-desc {
  color: var(--sora-text-secondary);
  line-height: 1.5;
}
</style>
