<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-claude-code-card__title text-lg font-semibold">
        {{ t('admin.settings.claudeCode.title') }}
      </h2>
      <p class="settings-claude-code-card__description mt-1 text-sm">
        {{ t('admin.settings.claudeCode.description') }}
      </p>
    </div>
    <div class="settings-claude-code-card__body">
      <div>
        <label class="settings-claude-code-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.claudeCode.minVersion') }}
        </label>
        <input
          v-model="form.min_claude_code_version"
          type="text"
          class="input max-w-xs font-mono text-sm"
          :placeholder="t('admin.settings.claudeCode.minVersionPlaceholder')"
        />
        <p class="settings-claude-code-card__description mt-1.5 text-xs">
          {{ t('admin.settings.claudeCode.minVersionHint') }}
        </p>
      </div>
      <div class="mt-4">
        <label class="settings-claude-code-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.claudeCode.maxVersion') }}
        </label>
        <input
          v-model="form.max_claude_code_version"
          type="text"
          class="input max-w-xs font-mono text-sm"
          :placeholder="t('admin.settings.claudeCode.maxVersionPlaceholder')"
        />
        <p class="settings-claude-code-card__description mt-1.5 text-xs">
          {{ t('admin.settings.claudeCode.maxVersionHint') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SettingsClaudeCodeFields } from './settingsForm'

defineProps<{
  form: SettingsClaudeCodeFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-claude-code-card__title,
.settings-claude-code-card__field-label {
  color: var(--theme-page-text);
}

.settings-claude-code-card__description {
  color: var(--theme-page-muted);
}

.settings-claude-code-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
