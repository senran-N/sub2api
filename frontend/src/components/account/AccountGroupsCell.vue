<template>
  <div v-if="groups && groups.length > 0" class="account-groups-cell relative">
    <!-- 分组容器：固定最大宽度，最多显示2行 -->
    <div class="account-groups-cell__inline-list flex flex-wrap gap-1 overflow-hidden">
      <GroupBadge
        v-for="group in displayGroups"
        :key="group.id"
        :name="group.name"
        :platform="group.platform"
        :subscription-type="group.subscription_type"
        :rate-multiplier="group.rate_multiplier"
        :show-rate="false"
        class="account-groups-cell__badge"
      />
      <!-- 更多数量徽章 -->
      <button
        v-if="hiddenCount > 0"
        ref="moreButtonRef"
        @click.stop="showPopover = !showPopover"
        class="account-groups-cell__more-button inline-flex cursor-pointer items-center gap-0.5 whitespace-nowrap text-xs font-medium transition-colors"
      >
        <span>+{{ hiddenCount }}</span>
      </button>
    </div>

    <!-- Popover 显示完整列表 -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="opacity-0 scale-95"
        enter-to-class="opacity-100 scale-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="opacity-100 scale-100"
        leave-to-class="opacity-0 scale-95"
      >
        <div
          v-if="showPopover"
          ref="popoverRef"
          class="account-groups-cell__popover fixed z-50"
          :style="popoverStyle"
        >
          <div class="mb-2 flex items-center justify-between">
            <span class="account-groups-cell__count text-xs font-medium">
              {{ t('admin.accounts.groupCountTotal', { count: groups.length }) }}
            </span>
            <button
              @click="showPopover = false"
              class="account-groups-cell__close"
            >
              <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div class="account-groups-cell__popover-list flex flex-wrap gap-1.5 overflow-y-auto">
            <GroupBadge
              v-for="group in groups"
              :key="group.id"
              :name="group.name"
              :platform="group.platform"
              :subscription-type="group.subscription_type"
              :rate-multiplier="group.rate_multiplier"
              :show-rate="false"
            />
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- 点击外部关闭 popover -->
    <div
      v-if="showPopover"
      class="fixed inset-0 z-40"
      @click="showPopover = false"
    />
  </div>
  <span v-else class="account-groups-cell__empty text-sm">-</span>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import GroupBadge from '@/components/common/GroupBadge.vue'
import type { Group } from '@/types'

interface Props {
  groups: Group[] | null | undefined
  maxDisplay?: number
}

const props = withDefaults(defineProps<Props>(), {
  maxDisplay: 4
})

const { t } = useI18n()

const moreButtonRef = ref<HTMLElement | null>(null)
const popoverRef = ref<HTMLElement | null>(null)
const showPopover = ref(false)

// 显示的分组（最多显示 maxDisplay 个）
const displayGroups = computed(() => {
  if (!props.groups) return []
  if (props.groups.length <= props.maxDisplay) {
    return props.groups
  }
  // 留一个位置给 +N 按钮
  return props.groups.slice(0, props.maxDisplay - 1)
})

// 隐藏的数量
const hiddenCount = computed(() => {
  if (!props.groups) return 0
  if (props.groups.length <= props.maxDisplay) return 0
  return props.groups.length - (props.maxDisplay - 1)
})

// Popover 位置样式
const popoverStyle = computed(() => {
  if (!moreButtonRef.value) return {}
  const rect = moreButtonRef.value.getBoundingClientRect()
  const viewportHeight = window.innerHeight
  const viewportWidth = window.innerWidth

  let top = rect.bottom + 8
  let left = rect.left

  // 如果下方空间不足，显示在上方
  if (top + 280 > viewportHeight) {
    top = Math.max(8, rect.top - 280)
  }

  // 如果右侧空间不足，向左偏移
  if (left + 384 > viewportWidth) {
    left = Math.max(8, viewportWidth - 392)
  }

  return {
    top: `${top}px`,
    left: `${left}px`
  }
})

// 关闭 popover 的键盘事件
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    showPopover.value = false
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.account-groups-cell {
  max-width: var(--theme-account-groups-cell-max-width);
}

.account-groups-cell__inline-list {
  max-height: var(--theme-account-groups-inline-max-height);
}

.account-groups-cell__badge {
  max-width: var(--theme-account-groups-badge-max-width);
}

.account-groups-cell__more-button {
  border-radius: var(--theme-button-radius);
  padding:
    var(--theme-account-usage-action-padding-y)
    var(--theme-account-usage-action-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 76%, transparent);
}

.account-groups-cell__more-button:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 74%, var(--theme-surface));
  color: var(--theme-page-text);
}

.account-groups-cell__popover {
  min-width: var(--theme-user-groups-dropdown-min-width);
  max-width: var(--theme-account-groups-popover-max-width);
  border-radius: var(--theme-user-groups-dropdown-radius);
  padding: var(--theme-group-replace-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.account-groups-cell__popover-list {
  max-height: var(--theme-account-groups-list-max-height);
}

.account-groups-cell__count,
.account-groups-cell__empty {
  color: var(--theme-page-muted);
}

.account-groups-cell__close {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-inline-button-padding);
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
  transition: color 0.2s ease, background-color 0.2s ease;
}

.account-groups-cell__close:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
}
</style>
