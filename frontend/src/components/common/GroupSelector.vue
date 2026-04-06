<template>
  <div>
    <label class="input-label">
      {{ t('admin.users.groups') }}
      <span class="group-selector__count">{{
        t('common.selectedCount', { count: modelValue.length })
      }}</span>
    </label>
    <div class="group-selector__grid">
      <label
        v-for="group in filteredGroups"
        :key="group.id"
        class="group-selector__item"
        :title="t('admin.groups.rateAndAccounts', { rate: group.rate_multiplier, count: group.account_count || 0 })"
      >
        <input
          type="checkbox"
          :value="group.id"
          :checked="modelValue.includes(group.id)"
          @change="handleChange(group.id, ($event.target as HTMLInputElement).checked)"
          class="theme-checkbox group-selector__checkbox"
        />
        <GroupBadge
          :name="group.name"
          :platform="group.platform"
          :subscription-type="group.subscription_type"
          :rate-multiplier="group.rate_multiplier"
          class="min-w-0 flex-1"
        />
        <span class="group-selector__meta">{{ group.account_count || 0 }}</span>
      </label>
      <div
        v-if="filteredGroups.length === 0"
        class="group-selector__empty"
      >
        {{ t('common.noGroupsAvailable') }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from './GroupBadge.vue'
import type { AdminGroup, GroupPlatform } from '@/types'

const { t } = useI18n()

interface Props {
  modelValue: number[]
  groups: AdminGroup[]
  platform?: GroupPlatform // Optional platform filter
  mixedScheduling?: boolean // For antigravity accounts: allow anthropic/gemini groups
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:modelValue': [value: number[]]
}>()

// Filter groups by platform if specified
const filteredGroups = computed(() => {
  if (!props.platform) {
    return props.groups
  }
  // antigravity 账户启用混合调度后，可选择 anthropic/gemini 分组
  if (props.platform === 'antigravity' && props.mixedScheduling) {
    return props.groups.filter(
      (g) => g.platform === 'antigravity' || g.platform === 'anthropic' || g.platform === 'gemini'
    )
  }
  // 默认：只能选择同 platform 的分组
  return props.groups.filter((g) => g.platform === props.platform)
})

const handleChange = (groupId: number, checked: boolean) => {
  const newValue = checked
    ? [...props.modelValue, groupId]
    : props.modelValue.filter((id) => id !== groupId)
  emit('update:modelValue', newValue)
}
</script>

<style scoped>
.group-selector__count,
.group-selector__meta,
.group-selector__empty {
  color: var(--theme-page-muted);
}

.group-selector__count {
  @apply font-normal;
}

.group-selector__grid {
  @apply grid grid-cols-2 gap-1 overflow-y-auto;
  border: 1px solid var(--theme-card-border);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, transparent);
  max-height: var(--theme-group-selector-max-height);
  padding: var(--theme-group-selector-padding);
}

.group-selector__item {
  @apply flex cursor-pointer items-center gap-2 transition-colors;
  border-radius: var(--theme-button-radius);
  padding:
    var(--theme-user-groups-dropdown-item-padding-y)
    var(--theme-user-groups-dropdown-item-padding-x);
}

.group-selector__item:hover {
  background: var(--theme-surface);
}

.group-selector__checkbox {
  @apply h-3.5 w-3.5 shrink-0;
}

.group-selector__meta {
  @apply shrink-0 text-xs;
}

.group-selector__empty {
  @apply col-span-2 text-center text-sm;
  padding-block: var(--theme-group-selector-empty-padding-y);
}
</style>
