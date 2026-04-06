<template>
  <div v-if="form.platform === 'anthropic'" class="border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="theme-text-default text-sm font-medium">
        {{ t('admin.groups.claudeCode.title') }}
      </label>
      <GroupSectionInfoTooltip :text="t('admin.groups.claudeCode.tooltip')" />
    </div>
    <div class="flex items-center gap-3">
      <Toggle v-model="form.claude_code_only" />
      <span class="theme-text-muted text-sm">
        {{ form.claude_code_only ? t('admin.groups.claudeCode.enabled') : t('admin.groups.claudeCode.disabled') }}
      </span>
    </div>
    <div v-if="form.claude_code_only" class="mt-3">
      <label class="input-label">{{ t('admin.groups.claudeCode.fallbackGroup') }}</label>
      <Select
        v-model="form.fallback_group_id"
        :options="fallbackGroupOptions"
        :placeholder="t('admin.groups.claudeCode.noFallback')"
      />
      <p class="input-hint">{{ t('admin.groups.claudeCode.fallbackHint') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Select from '@/components/common/Select.vue'
import type { SelectOption } from '@/components/common/Select.vue'
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'
import GroupSectionInfoTooltip from './GroupSectionInfoTooltip.vue'

defineProps<{
  form: CreateGroupForm | EditGroupForm
  fallbackGroupOptions: SelectOption[]
}>()

const { t } = useI18n()
</script>
