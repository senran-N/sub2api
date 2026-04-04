<template>
  <div v-if="form.platform === 'antigravity'" class="border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('admin.groups.supportedScopes.title') }}
      </label>
      <GroupSectionInfoTooltip :text="t('admin.groups.supportedScopes.tooltip')" />
    </div>
    <div class="space-y-2">
      <label
        v-for="scope in scopeOptions"
        :key="scope.value"
        class="flex cursor-pointer items-center gap-2"
      >
        <input
          type="checkbox"
          :checked="form.supported_model_scopes.includes(scope.value)"
          class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700"
          @change="$emit('toggle-scope', scope.value)"
        />
        <span class="text-sm text-gray-700 dark:text-gray-300">{{ scope.label }}</span>
      </label>
    </div>
    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
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
