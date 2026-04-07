<template>
  <div v-if="form.platform === 'antigravity'" class="border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="theme-text-default text-sm font-medium">
        {{ t('admin.groups.mcpXml.title') }}
      </label>
      <GroupSectionInfoTooltip :text="t('admin.groups.mcpXml.tooltip')" />
    </div>
    <div class="flex items-center gap-3">
      <Toggle v-model="form.mcp_xml_inject" />
      <span class="theme-text-muted text-sm">
        {{ form.mcp_xml_inject ? t('admin.groups.mcpXml.enabled') : t('admin.groups.mcpXml.disabled') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { CreateGroupForm, EditGroupForm } from './groupsForm'
import GroupSectionInfoTooltip from './GroupSectionInfoTooltip.vue'

defineProps<{
  form: CreateGroupForm | EditGroupForm
}>()

const { t } = useI18n()
</script>
