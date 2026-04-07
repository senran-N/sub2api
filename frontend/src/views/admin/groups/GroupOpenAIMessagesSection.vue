<template>
  <div
    v-if="form.platform === 'openai'"
    class="group-openai-messages-section mt-4 border-t pt-4"
  >
    <h4 class="group-openai-messages-section__title mb-3 text-sm font-medium">
      {{ t('admin.groups.openaiMessages.title') }}
    </h4>

    <div class="flex items-center justify-between">
      <label class="group-openai-messages-section__label text-sm">
        {{ t('admin.groups.openaiMessages.allowDispatch') }}
      </label>
      <Toggle v-model="form.allow_messages_dispatch" />
    </div>
    <p class="group-openai-messages-section__hint mt-1 text-xs">
      {{ t('admin.groups.openaiMessages.allowDispatchHint') }}
    </p>

    <div v-if="form.allow_messages_dispatch" class="mt-3">
      <label class="input-label">{{ t('admin.groups.openaiMessages.defaultModel') }}</label>
      <input
        v-model="form.default_mapped_model"
        type="text"
        :placeholder="t('admin.groups.openaiMessages.defaultModelPlaceholder')"
        class="input"
      />
      <p class="input-hint">{{ t('admin.groups.openaiMessages.defaultModelHint') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from './groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>

<style scoped>
.group-openai-messages-section {
  border-color: var(--theme-page-border);
}

.group-openai-messages-section__title,
.group-openai-messages-section__label {
  color: var(--theme-page-text);
}

.group-openai-messages-section__hint {
  color: var(--theme-page-muted);
}
</style>
