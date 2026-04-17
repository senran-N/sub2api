<template>
  <BaseDialog :show="show" :title="t('admin.users.groupConfig')" width="wide" @close="$emit('close')">
    <div v-if="user" class="space-y-6">
      <!-- 用户信息头部 -->
      <div class="user-allowed-groups-modal__hero flex items-center gap-4">
        <div class="user-allowed-groups-modal__avatar flex h-14 w-14 items-center justify-center">
          <span class="user-allowed-groups-modal__avatar-text text-2xl font-semibold">{{ user.email.charAt(0).toUpperCase() }}</span>
        </div>
        <div class="flex-1">
          <p class="user-allowed-groups-modal__text-strong text-lg font-semibold">{{ user.email }}</p>
          <p class="user-allowed-groups-modal__text-muted mt-1 text-sm">{{ t('admin.users.groupConfigHint', { email: user.email }) }}</p>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="user-allowed-groups-modal__state-block flex justify-center">
        <svg class="user-allowed-groups-modal__spinner h-10 w-10 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
      </div>

      <div v-else class="space-y-6">
        <!-- 专属分组区域 -->
        <div v-if="exclusiveGroups.length > 0">
          <div class="mb-3 flex items-center gap-2">
            <div class="user-allowed-groups-modal__section-dot user-allowed-groups-modal__section-dot--exclusive"></div>
            <h4 class="user-allowed-groups-modal__section-title text-sm font-semibold">{{ t('admin.users.exclusiveGroups') }}</h4>
            <span class="user-allowed-groups-modal__text-soft text-xs">({{ exclusiveGroupConfigs.filter(c => c.isSelected).length }}/{{ exclusiveGroupConfigs.length }})</span>
          </div>
          <div class="grid gap-3">
            <div
              v-for="config in exclusiveGroupConfigs"
              :key="config.groupId"
              :class="getGroupCardClasses('exclusive', config.isSelected)"
            >
              <div class="flex items-center gap-4">
                <!-- 复选框 -->
                <div class="flex-shrink-0">
                  <label class="relative flex h-6 w-6 cursor-pointer items-center justify-center">
                    <input
                      type="checkbox"
                      :checked="config.isSelected"
                      @change="toggleExclusiveGroup(config.groupId)"
                      class="peer sr-only"
                    />
                    <div class="user-allowed-groups-modal__checkbox user-allowed-groups-modal__checkbox-shape" :class="{ 'user-allowed-groups-modal__checkbox--checked': config.isSelected }">
                      <svg v-if="config.isSelected" class="user-allowed-groups-modal__checkbox-icon h-full w-full" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                      </svg>
                    </div>
                  </label>
                </div>

                <!-- 分组信息 -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="user-allowed-groups-modal__text-strong text-base font-semibold">{{ config.groupName }}</span>
                    <span class="theme-chip theme-chip--compact theme-chip--brand-purple inline-flex items-center rounded-full">
                      {{ t('admin.groups.exclusive') }}
                    </span>
                  </div>
                  <div class="mt-1.5 flex items-center gap-3 text-sm">
                    <span class="user-allowed-groups-modal__text-muted inline-flex items-center gap-1">
                      <PlatformIcon :platform="config.platform" size="xs" />
                      <span>{{ config.platform }}</span>
                    </span>
                    <span class="user-allowed-groups-modal__text-soft">•</span>
                    <span class="user-allowed-groups-modal__text-muted">
                      {{ t('admin.users.defaultRate') }}: <span class="user-allowed-groups-modal__text-body font-medium">{{ config.defaultRate }}x</span>
                    </span>
                  </div>
                </div>

                <!-- 专属倍率输入 -->
                <div class="flex flex-shrink-0 items-center gap-3">
                  <label class="user-allowed-groups-modal__text-muted text-sm font-medium">{{ t('admin.users.customRate') }}</label>
                  <input
                    type="number"
                    step="0.001"
                    min="0"
                    :value="config.customRate ?? ''"
                    @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                    :placeholder="String(config.defaultRate)"
                    class="user-allowed-groups-modal__rate-input user-allowed-groups-modal__rate-input-width hide-spinner text-sm font-medium"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 公开分组区域 -->
        <div v-if="publicGroups.length > 0">
          <div class="mb-3 flex items-center gap-2">
            <div class="user-allowed-groups-modal__section-dot user-allowed-groups-modal__section-dot--public"></div>
            <h4 class="user-allowed-groups-modal__section-title text-sm font-semibold">{{ t('admin.users.publicGroups') }}</h4>
            <span class="user-allowed-groups-modal__text-soft text-xs">({{ publicGroupConfigs.length }})</span>
          </div>
          <div class="grid gap-3">
            <div
              v-for="config in publicGroupConfigs"
              :key="config.groupId"
              :class="getGroupCardClasses('public', true)"
            >
              <div class="flex items-center gap-4">
                <!-- 复选框（禁用状态） -->
                <div class="flex-shrink-0">
                  <div class="user-allowed-groups-modal__checkbox user-allowed-groups-modal__checkbox--checked user-allowed-groups-modal__checkbox--public user-allowed-groups-modal__checkbox-shape flex items-center justify-center">
                    <svg class="user-allowed-groups-modal__checkbox-icon h-full w-full" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="3">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>

                <!-- 分组信息 -->
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="user-allowed-groups-modal__text-strong text-base font-semibold">{{ config.groupName }}</span>
                  </div>
                  <div class="mt-1.5 flex items-center gap-3 text-sm">
                    <span class="user-allowed-groups-modal__text-muted inline-flex items-center gap-1">
                      <PlatformIcon :platform="config.platform" size="xs" />
                      <span>{{ config.platform }}</span>
                    </span>
                    <span class="user-allowed-groups-modal__text-soft">•</span>
                    <span class="user-allowed-groups-modal__text-muted">
                      {{ t('admin.users.defaultRate') }}: <span class="user-allowed-groups-modal__text-body font-medium">{{ config.defaultRate }}x</span>
                    </span>
                  </div>
                </div>

                <!-- 专属倍率输入 -->
                <div class="flex flex-shrink-0 items-center gap-3">
                  <label class="user-allowed-groups-modal__text-muted text-sm font-medium">{{ t('admin.users.customRate') }}</label>
                  <input
                    type="number"
                    step="0.001"
                    min="0"
                    :value="config.customRate ?? ''"
                    @input="updateCustomRate(config.groupId, ($event.target as HTMLInputElement).value)"
                    :placeholder="String(config.defaultRate)"
                    class="user-allowed-groups-modal__rate-input user-allowed-groups-modal__rate-input-width hide-spinner text-sm font-medium"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 无分组提示 -->
        <div v-if="groups.length === 0" class="user-allowed-groups-modal__state-block flex flex-col items-center justify-center text-center">
          <div class="user-allowed-groups-modal__empty-icon-wrap mb-4 flex h-16 w-16 items-center justify-center">
            <svg class="user-allowed-groups-modal__empty-icon h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
          <p class="user-allowed-groups-modal__text-muted">{{ t('common.noGroupsAvailable') }}</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" class="btn btn-secondary user-allowed-groups-modal__footer-action user-allowed-groups-modal__footer-action--cancel">{{ t('common.cancel') }}</button>
        <button @click="handleSave" :disabled="submitting" class="btn btn-primary user-allowed-groups-modal__footer-action user-allowed-groups-modal__footer-action--confirm">
          <svg v-if="submitting" class="-ml-1 mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ submitting ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser, Group, GroupPlatform } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface GroupRateConfig {
  groupId: number
  groupName: string
  platform: GroupPlatform
  isExclusive: boolean
  defaultRate: number
  customRate: number | null
  isSelected: boolean
}

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const groups = ref<Group[]>([])
const groupConfigs = ref<GroupRateConfig[]>([])
const originalGroupRates = ref<Record<number, number>>({}) // 记录原始专属倍率，用于检测删除
const loading = ref(false)
const submitting = ref(false)
let loadSequence = 0

// 分离专属分组和公开分组
const exclusiveGroups = computed(() => groups.value.filter((g) => g.is_exclusive))
const publicGroups = computed(() => groups.value.filter((g) => !g.is_exclusive))

const exclusiveGroupConfigs = computed(() => groupConfigs.value.filter((c) => c.isExclusive))
const publicGroupConfigs = computed(() => groupConfigs.value.filter((c) => !c.isExclusive))

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getGroupCardClasses = (scope: 'exclusive' | 'public', selected: boolean) => {
  return joinClassNames(
    'user-allowed-groups-modal__group-card user-allowed-groups-modal__group-card-control group relative overflow-hidden border-2 transition-all duration-200',
    scope === 'exclusive'
      ? selected
        ? 'user-allowed-groups-modal__group-card--selected'
        : 'user-allowed-groups-modal__group-card--idle'
      : 'user-allowed-groups-modal__group-card--public'
  )
}

watch(
  () => [props.show, props.user?.id] as const,
  ([isVisible, userId]) => {
    if (!isVisible || userId == null) {
      loadSequence += 1
      loading.value = false
      groups.value = []
      groupConfigs.value = []
      originalGroupRates.value = {}
      return
    }

    void load()
  },
  { immediate: true }
)

async function load() {
  const user = props.user
  if (!user) return

  const requestSequence = ++loadSequence
  loading.value = true
  try {
    const res = await adminAPI.groups.list(1, 1000)
    if (requestSequence !== loadSequence || !props.show || props.user?.id !== user.id) {
      return
    }

    // 只显示标准类型且活跃的分组
    groups.value = res.items.filter((g) => g.subscription_type === 'standard' && g.status === 'active')

    // 初始化配置
    const userAllowedGroups = user.allowed_groups || []
    const userGroupRates = user.group_rates || {}

    // 保存原始专属倍率，用于检测删除操作
    originalGroupRates.value = { ...userGroupRates }

    groupConfigs.value = groups.value.map((g) => ({
      groupId: g.id,
      groupName: g.name,
      platform: g.platform,
      isExclusive: g.is_exclusive,
      defaultRate: g.rate_multiplier,
      customRate: userGroupRates[g.id] ?? null,
      // 专属分组：检查是否在 allowed_groups 中
      // 公开分组：始终选中
      isSelected: g.is_exclusive ? userAllowedGroups.includes(g.id) : true,
    }))
  } catch (error) {
    if (requestSequence !== loadSequence || !props.show || props.user?.id !== user.id) {
      return
    }
    console.error('Failed to load groups:', error)
    appStore.showError(resolveRequestErrorMessage(error, t('admin.users.failedToLoadGroups')))
  } finally {
    if (requestSequence === loadSequence) {
      loading.value = false
    }
  }
}

const toggleExclusiveGroup = (groupId: number) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config && config.isExclusive) {
    config.isSelected = !config.isSelected
  }
}

const updateCustomRate = (groupId: number, value: string) => {
  const config = groupConfigs.value.find((c) => c.groupId === groupId)
  if (config) {
    if (value === '' || value === null || value === undefined) {
      config.customRate = null
    } else {
      const numValue = parseFloat(value)
      config.customRate = isNaN(numValue) ? null : numValue
    }
  }
}

const handleSave = async () => {
  if (!props.user) return
  submitting.value = true

  try {
    // 构建 allowed_groups（仅包含专属分组中被勾选的）
    const allowedGroups = groupConfigs.value.filter((c) => c.isExclusive && c.isSelected).map((c) => c.groupId)

    // 构建 group_rates
    // - 有新专属倍率: 设置为该值
    // - 原本有专属倍率但现在被清空: 设置为 null（表示删除）
    const groupRates: Record<number, number | null> = {}
    for (const c of groupConfigs.value) {
      const hadOriginalRate = originalGroupRates.value[c.groupId] !== undefined

      if (c.customRate !== null) {
        // 有专属倍率
        groupRates[c.groupId] = c.customRate
      } else if (hadOriginalRate) {
        // 原本有专属倍率，现在被清空，需要显式删除
        groupRates[c.groupId] = null
      }
    }

    await adminAPI.users.update(props.user.id, {
      allowed_groups: allowedGroups,
      group_rates: Object.keys(groupRates).length > 0 ? groupRates : undefined,
    })

    appStore.showSuccess(t('admin.users.groupConfigUpdated'))
    emit('success')
    emit('close')
  } catch (error) {
    console.error('Failed to update user group config:', error)
    appStore.showError(resolveRequestErrorMessage(error, t('admin.users.failedToUpdateAllowedGroups')))
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
/* 隐藏数字输入框的箭头按钮 */
.hide-spinner::-webkit-outer-spin-button,
.hide-spinner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.hide-spinner {
  -moz-appearance: textfield;
}

.user-allowed-groups-modal__hero {
  border-radius: calc(var(--theme-surface-radius) + 4px);
  padding: 1.25rem;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--theme-accent-soft) 88%, var(--theme-surface)) 0%,
    color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface)) 100%
  );
  border: 1px solid color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
}

.user-allowed-groups-modal__state-block {
  padding-block: var(--theme-user-allowed-groups-state-padding-y);
}

.user-allowed-groups-modal__avatar {
  border-radius: var(--theme-version-icon-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.user-allowed-groups-modal__avatar-text,
.user-allowed-groups-modal__spinner {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.user-allowed-groups-modal__text-strong,
.user-allowed-groups-modal__section-title,
.user-allowed-groups-modal__text-body {
  color: var(--theme-page-text);
}

.user-allowed-groups-modal__text-muted {
  color: var(--theme-page-muted);
}

.user-allowed-groups-modal__text-soft,
.user-allowed-groups-modal__empty-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.user-allowed-groups-modal__section-dot--exclusive {
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, transparent);
}

.user-allowed-groups-modal__section-dot--public {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, transparent);
}

.user-allowed-groups-modal__section-dot {
  width: var(--theme-user-allowed-groups-section-dot-size);
  height: var(--theme-user-allowed-groups-section-dot-size);
  border-radius: 999px;
}

.user-allowed-groups-modal__group-card {
  background: var(--theme-surface);
}

.user-allowed-groups-modal__group-card-control {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
}

.user-allowed-groups-modal__group-card--idle {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.user-allowed-groups-modal__group-card--idle:hover {
  border-color: color-mix(in srgb, var(--theme-card-border) 92%, transparent);
}

.user-allowed-groups-modal__group-card--selected {
  border-color: color-mix(in srgb, var(--theme-accent) 38%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 70%, var(--theme-surface));
  box-shadow: var(--theme-card-shadow);
}

.user-allowed-groups-modal__group-card--public {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 20%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 8%, var(--theme-surface));
}

.user-allowed-groups-modal__checkbox {
  border-color: color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  background: var(--theme-surface);
}

.user-allowed-groups-modal__checkbox-shape {
  width: var(--theme-user-allowed-groups-checkbox-size);
  height: var(--theme-user-allowed-groups-checkbox-size);
  border-width: 2px;
  border-style: solid;
  border-radius: var(--theme-user-allowed-groups-checkbox-radius);
  transition: all 0.2s ease;
}

.user-allowed-groups-modal__checkbox--checked {
  border-color: color-mix(in srgb, var(--theme-accent) 84%, transparent);
  background: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-accent-strong));
}

.user-allowed-groups-modal__checkbox--public {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, transparent);
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-accent-strong));
}

.user-allowed-groups-modal__checkbox-icon {
  color: var(--theme-filled-text);
}

.user-allowed-groups-modal__rate-input {
  border-radius: var(--theme-button-radius);
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--theme-input-border);
  background: var(--theme-input-bg);
  color: var(--theme-input-text);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.user-allowed-groups-modal__rate-input-width {
  width: var(--theme-user-allowed-groups-rate-input-width);
}

.user-allowed-groups-modal__rate-input::placeholder {
  color: var(--theme-input-placeholder);
}

.user-allowed-groups-modal__rate-input:focus {
  border-color: color-mix(in srgb, var(--theme-accent) 68%, var(--theme-input-border));
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 14%, transparent);
}

.user-allowed-groups-modal__empty-icon-wrap {
  border-radius: var(--theme-version-icon-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.user-allowed-groups-modal__footer-action--cancel {
  padding-inline: var(--theme-user-allowed-groups-footer-cancel-padding-x);
}

.user-allowed-groups-modal__footer-action--confirm {
  padding-inline: var(--theme-user-allowed-groups-footer-confirm-padding-x);
}
</style>
