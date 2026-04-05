<template>
  <div class="group/dropdown relative">
    <button
      :ref="buttonRef"
      @click="$emit('open-selector', row)"
      class="-mx-2 -my-1 flex cursor-pointer items-center gap-2 rounded-lg px-2 py-1 transition-all duration-200 hover:bg-gray-100 dark:hover:bg-dark-700"
      :title="clickToChangeTitle"
    >
      <GroupBadge
        v-if="row.group"
        :name="row.group.name"
        :platform="row.group.platform"
        :subscription-type="row.group.subscription_type"
        :rate-multiplier="row.group.rate_multiplier"
        :user-rate-multiplier="row.group ? userGroupRates[row.group.id] : null"
      />
      <span v-else class="text-sm text-gray-400 dark:text-dark-500">{{ noGroupLabel }}</span>
      <span class="text-xs text-gray-500 dark:text-gray-400">{{ selectGroupLabel }}</span>
      <Icon
        name="sort"
        size="sm"
        class="text-gray-400 opacity-60 transition-opacity group-hover/dropdown:opacity-100"
        :stroke-width="2"
      />
    </button>
  </div>
</template>

<script setup lang="ts">
import type { ComponentPublicInstance } from 'vue'
import type { ApiKey } from '@/types'
import GroupBadge from '@/components/common/GroupBadge.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  row: ApiKey
  userGroupRates: Record<number, number>
  clickToChangeTitle: string
  noGroupLabel: string
  selectGroupLabel: string
  buttonRef: (el: Element | ComponentPublicInstance | null) => void
}>()

defineEmits<{
  'open-selector': [row: ApiKey]
}>()
</script>
