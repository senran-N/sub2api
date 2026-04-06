<template>
  <div v-if="form.platform === 'antigravity'" class="group-supported-scopes-section border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="group-supported-scopes-section__title text-sm font-medium">
        {{ t('admin.groups.supportedScopes.title') }}
      </label>
      <GroupSectionInfoTooltip :text="t('admin.groups.supportedScopes.tooltip')" />
    </div>
    <div class="space-y-2">
      <label
        v-for="scope in scopeOptions"
        :key="scope.value"
        class="group-supported-scopes-section__option flex cursor-pointer items-center gap-2"
      >
        <input
          type="checkbox"
          :checked="form.supported_model_scopes.includes(scope.value)"
          class="group-supported-scopes-section__checkbox h-4 w-4 rounded"
          @change="$emit('toggle-scope', scope.value)"
        />
        <span class="group-supported-scopes-section__label text-sm">{{ scope.label }}</span>
      </label>
    </div>
    <p class="group-supported-scopes-section__hint mt-2 text-xs">
      {{ t('admin.groups.supportedScopes.hint') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'
import GroupSectionInfoTooltip from './GroupSectionInfoTooltip.vue'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

defineEmits<{
  'toggle-scope': [scope: string]
}>()

const { t } = useI18n()

const scopeOptions = computed(() => [
  { value: 'claude', label: t('admin.groups.supportedScopes.claude') },
  { value: 'gemini_text', label: t('admin.groups.supportedScopes.geminiText') },
  { value: 'gemini_image', label: t('admin.groups.supportedScopes.geminiImage') }
])
</script>

<style scoped>
.group-supported-scopes-section {
  border-color: var(--theme-page-border);
}

.group-supported-scopes-section__title,
.group-supported-scopes-section__label {
  color: var(--theme-page-text);
}

.group-supported-scopes-section__hint {
  color: var(--theme-page-muted);
}

.group-supported-scopes-section__checkbox {
  border-color: color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: var(--theme-input-bg);
  color: var(--theme-accent);
}

.group-supported-scopes-section__checkbox:focus {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--theme-accent) 16%, transparent);
}
</style>
