<template>
  <KeysMoreActionsMenu
    :show="showMoreMenu"
    :position="moreMenuPosition"
    :row="moreMenuRow"
    :hide-ccs-import-button="hideCcsImportButton"
    @close="$emit('close-more-menu')"
    @import="$emit('import-to-ccswitch', moreMenuRow!)"
    @toggle-status="$emit('toggle-key-status', moreMenuRow!)"
    @delete="$emit('confirm-delete', moreMenuRow!)"
  />

  <KeysGroupSelectorDropdown
    :show="showGroupSelector"
    :position="dropdownPosition"
    :search-query="groupSearchQuery"
    :options="filteredGroupOptions"
    :selected-group-id="selectedKeyForGroup?.group_id ?? null"
    @close="$emit('close-group-selector')"
    @update:search-query="$emit('update:groupSearchQuery', $event)"
    @select="$emit('change-group', selectedKeyForGroup!, $event)"
  />
</template>

<script setup lang="ts">
import type { ApiKey } from '@/types'
import type { UserKeyGroupOption } from './keysForm'
import type { KeysOverlayPosition } from './keysOverlays'
import KeysGroupSelectorDropdown from './KeysGroupSelectorDropdown.vue'
import KeysMoreActionsMenu from './KeysMoreActionsMenu.vue'

defineProps<{
  showMoreMenu: boolean
  moreMenuPosition: { top: number; left: number } | null
  moreMenuRow: ApiKey | null
  hideCcsImportButton: boolean
  showGroupSelector: boolean
  dropdownPosition: KeysOverlayPosition | null
  groupSearchQuery: string
  filteredGroupOptions: UserKeyGroupOption[]
  selectedKeyForGroup: ApiKey | null
}>()

defineEmits<{
  'close-more-menu': []
  'import-to-ccswitch': [row: ApiKey]
  'toggle-key-status': [row: ApiKey]
  'confirm-delete': [row: ApiKey]
  'close-group-selector': []
  'update:groupSearchQuery': [value: string]
  'change-group': [row: ApiKey, groupId: number | null]
}>()
</script>
