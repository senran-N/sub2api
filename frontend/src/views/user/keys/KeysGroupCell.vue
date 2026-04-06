<template>
  <div class="group/dropdown relative">
    <button
      :ref="buttonRef"
      @click="$emit('open-selector', row)"
      class="keys-group-cell__trigger flex cursor-pointer items-center gap-2 transition-all duration-200"
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
      <span v-else class="keys-group-cell__empty text-sm">{{ noGroupLabel }}</span>
      <span class="keys-group-cell__hint text-xs">{{ selectGroupLabel }}</span>
      <Icon
        name="sort"
        size="sm"
        class="keys-group-cell__icon opacity-60 transition-opacity group-hover/dropdown:opacity-100"
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

<style scoped>
.keys-group-cell__trigger {
  margin: calc(var(--theme-key-row-action-padding) * -1);
  border-radius: var(--theme-key-row-action-radius);
  padding:
    calc(var(--theme-key-row-action-padding) - 0.125rem)
    calc(var(--theme-key-row-action-padding) + 0.125rem);
}

.keys-group-cell__trigger:hover {
  background: var(--theme-button-ghost-hover-bg);
}

.keys-group-cell__empty {
  color: color-mix(in srgb, var(--theme-page-muted) 82%, transparent);
}

.keys-group-cell__hint {
  color: var(--theme-page-muted);
}

.keys-group-cell__icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}
</style>
