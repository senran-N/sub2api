<template>
  <div v-if="form.platform === 'antigravity'" class="border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('admin.groups.mcpXml.title') }}
      </label>
      <GroupSectionInfoTooltip :text="t('admin.groups.mcpXml.tooltip')" />
    </div>
    <div class="flex items-center gap-3">
      <Toggle v-model="form.mcp_xml_inject" />
      <span class="text-sm text-gray-500 dark:text-gray-400">
        {{ form.mcp_xml_inject ? t('admin.groups.mcpXml.enabled') : t('admin.groups.mcpXml.disabled') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from '../groupsForm'
import GroupSectionInfoTooltip from './GroupSectionInfoTooltip.vue'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>
