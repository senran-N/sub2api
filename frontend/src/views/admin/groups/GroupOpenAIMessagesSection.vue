<template>
  <div
    v-if="form.platform === 'openai'"
    class="mt-4 border-t border-gray-200 pt-4 dark:border-dark-400"
  >
    <h4 class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ t('admin.groups.openaiMessages.title') }}
    </h4>

    <div class="flex items-center justify-between">
      <label class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('admin.groups.openaiMessages.allowDispatch') }}
      </label>
      <Toggle v-model="form.allow_messages_dispatch" />
    </div>
    <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
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
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>
