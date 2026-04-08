<template>
  <BaseDialog :show="show" :title="t('admin.groups.rateMultipliersTitle')" width="wide" @close="handleClose">
    <div v-if="group" class="group-rate-multipliers-modal__content">
      <!-- 分组信息 -->
      <div class="group-rate-multipliers-modal__group-summary">
        <span :class="getPlatformChipClasses(group.platform)">
          <PlatformIcon :platform="group.platform" size="sm" />
          {{ t('admin.groups.platforms.' + group.platform) }}
        </span>
        <span class="group-rate-multipliers-modal__separator">|</span>
        <span class="group-rate-multipliers-modal__group-name font-medium">{{ group.name }}</span>
        <span class="group-rate-multipliers-modal__separator">|</span>
        <span class="group-rate-multipliers-modal__group-meta">
          {{ t('admin.groups.columns.rateMultiplier') }}: {{ group.rate_multiplier }}x
        </span>
      </div>

      <!-- 操作区 -->
      <div class="group-rate-multipliers-modal__control-panel">
        <!-- 添加用户 -->
        <h4 class="group-rate-multipliers-modal__section-title mb-2 text-sm font-medium">
          {{ t('admin.groups.addUserRate') }}
        </h4>
        <div class="flex items-end gap-2">
          <div class="relative flex-1">
            <input
              v-model="searchQuery"
              type="text"
              autocomplete="off"
              class="input w-full"
              :placeholder="t('admin.groups.searchUserPlaceholder')"
              @input="handleSearchUsers"
              @focus="showDropdown = true"
            />
            <div
              v-if="showDropdown && searchResults.length > 0"
              class="group-rate-multipliers-modal__search-dropdown"
            >
              <button
                v-for="user in searchResults"
                :key="user.id"
                type="button"
                class="group-rate-multipliers-modal__search-item"
                @click="selectUser(user)"
              >
                <span class="group-rate-multipliers-modal__search-id">#{{ user.id }}</span>
                <span class="group-rate-multipliers-modal__search-name">{{ user.username || user.email }}</span>
                <span v-if="user.username" class="group-rate-multipliers-modal__search-email text-xs">{{ user.email }}</span>
              </button>
            </div>
          </div>
          <div class="w-24">
            <input
              v-model.number="newRate"
              type="number"
              step="0.001"
              min="0"
              autocomplete="off"
              class="group-rate-multipliers-modal__number-input hide-spinner input w-full"
              placeholder="1.0"
            />
          </div>
          <button
            type="button"
            class="btn btn-primary shrink-0"
            :disabled="!selectedUser || !newRate"
            @click="handleAddLocal"
          >
            {{ t('common.add') }}
          </button>
        </div>

        <!-- 批量调整 + 全部清空 -->
        <div v-if="localEntries.length > 0" class="group-rate-multipliers-modal__batch-bar mt-3 flex items-center gap-3 pt-3">
          <span class="group-rate-multipliers-modal__batch-label text-xs font-medium">{{ t('admin.groups.batchAdjust') }}</span>
          <div class="flex items-center gap-1.5">
            <span class="group-rate-multipliers-modal__separator text-xs">×</span>
            <input
              v-model.number="batchFactor"
              type="number"
              step="0.1"
              min="0"
              autocomplete="off"
              class="group-rate-multipliers-modal__number-input group-rate-multipliers-modal__number-input--compact group-rate-multipliers-modal__number-input--batch hide-spinner"
              placeholder="0.5"
            />
            <button
              type="button"
              class="btn btn-primary btn-sm shrink-0 group-rate-multipliers-modal__batch-apply-button"
              :disabled="!batchFactor || batchFactor <= 0"
              @click="applyBatchFactor"
            >
              {{ t('admin.groups.applyMultiplier') }}
            </button>
          </div>
          <div class="ml-auto">
            <button
              type="button"
              class="group-rate-multipliers-modal__clear-button"
              @click="clearAllLocal"
            >
              {{ t('admin.groups.clearAll') }}
            </button>
          </div>
        </div>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="group-rate-multipliers-modal__loading">
        <Icon name="refresh" size="md" class="group-rate-multipliers-modal__loading-icon animate-spin" />
      </div>

      <!-- 已设置的用户列表 -->
      <div v-else>
        <h4 class="group-rate-multipliers-modal__section-title mb-2 text-sm font-medium">
          {{ t('admin.groups.rateMultipliers') }} ({{ localEntries.length }})
        </h4>

        <div v-if="localEntries.length === 0" class="group-rate-multipliers-modal__empty">
          {{ t('admin.groups.noRateMultipliers') }}
        </div>

        <div v-else>
          <!-- 表格 -->
          <div class="group-rate-multipliers-modal__table-shell">
            <div class="group-rate-multipliers-modal__table-scroll">
              <table class="w-full text-sm">
                <thead class="sticky top-0 z-[1]">
                  <tr class="group-rate-multipliers-modal__table-head">
                    <th class="group-rate-multipliers-modal__table-header">{{ t('admin.groups.columns.userEmail') }}</th>
                    <th class="group-rate-multipliers-modal__table-header">ID</th>
                    <th class="group-rate-multipliers-modal__table-header">{{ t('admin.groups.columns.userName') }}</th>
                    <th class="group-rate-multipliers-modal__table-header">{{ t('admin.groups.columns.userNotes') }}</th>
                    <th class="group-rate-multipliers-modal__table-header">{{ t('admin.groups.columns.userStatus') }}</th>
                    <th class="group-rate-multipliers-modal__table-header">{{ t('admin.groups.columns.rateMultiplier') }}</th>
                    <th v-if="showFinalRate" class="group-rate-multipliers-modal__table-header group-rate-multipliers-modal__table-header--accent">{{ t('admin.groups.finalRate') }}</th>
                    <th class="group-rate-multipliers-modal__table-header group-rate-multipliers-modal__table-header--icon"></th>
                  </tr>
                </thead>
                <tbody class="group-rate-multipliers-modal__table-body">
                  <tr
                    v-for="entry in paginatedLocalEntries"
                    :key="entry.user_id"
                    class="group-rate-multipliers-modal__table-row"
                  >
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__cell-muted">{{ entry.user_email }}</td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--nowrap group-rate-multipliers-modal__cell-soft">{{ entry.user_id }}</td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--nowrap group-rate-multipliers-modal__cell-strong">{{ entry.user_name || '-' }}</td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__cell-muted group-rate-multipliers-modal__table-cell--notes" :title="entry.user_notes">{{ entry.user_notes || '-' }}</td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--nowrap">
                      <span :class="getUserStatusClasses(entry.user_status)">
                        {{ entry.user_status }}
                      </span>
                    </td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--nowrap">
                      <input
                        type="number"
                        step="0.001"
                        min="0"
                        autocomplete="off"
                        :value="entry.rate_multiplier"
                        class="group-rate-multipliers-modal__number-input group-rate-multipliers-modal__number-input--compact hide-spinner"
                        @change="updateLocalRate(entry.user_id, ($event.target as HTMLInputElement).value)"
                      />
                    </td>
                    <td v-if="showFinalRate" class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--nowrap group-rate-multipliers-modal__final-rate">
                      {{ computeFinalRate(entry.rate_multiplier) }}
                    </td>
                    <td class="group-rate-multipliers-modal__table-cell group-rate-multipliers-modal__table-cell--icon">
                      <button
                        type="button"
                        :class="getActionButtonClasses('danger')"
                        @click="removeLocal(entry.user_id)"
                      >
                        <Icon name="trash" size="sm" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 分页 -->
          <Pagination
            :total="localEntries.length"
            :page="currentPage"
            :page-size="pageSize"
            :page-size-options="[10, 20, 50]"
            @update:page="currentPage = $event"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </div>

      <!-- 底部操作栏 -->
      <div class="group-rate-multipliers-modal__footer flex items-center gap-3 pt-4">
        <!-- 左侧：未保存提示 + 撤销 -->
        <template v-if="isDirty">
          <span class="group-rate-multipliers-modal__unsaved text-xs">{{ t('admin.groups.unsavedChanges') }}</span>
          <button
            type="button"
            class="group-rate-multipliers-modal__revert text-xs font-medium"
            @click="handleCancel"
          >
            {{ t('admin.groups.revertChanges') }}
          </button>
        </template>
        <!-- 右侧：关闭 / 保存 -->
        <div class="ml-auto flex items-center gap-3">
          <button type="button" class="btn btn-sm group-rate-multipliers-modal__footer-button" @click="handleClose">
            {{ t('common.close') }}
          </button>
          <button
            v-if="isDirty"
            type="button"
            class="btn btn-primary btn-sm group-rate-multipliers-modal__footer-button"
            :disabled="saving"
            @click="handleSave"
          >
            <Icon v-if="saving" name="refresh" size="sm" class="mr-1 animate-spin" />
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </BaseDialog>

</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { GroupRateMultiplierEntry } from '@/api/admin/groups'
import type { AdminGroup, AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'

interface LocalEntry extends GroupRateMultiplierEntry {}

const props = defineProps<{
  show: boolean
  group: AdminGroup | null
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)
const serverEntries = ref<GroupRateMultiplierEntry[]>([])
const localEntries = ref<LocalEntry[]>([])
const searchQuery = ref('')
const searchResults = ref<AdminUser[]>([])
const showDropdown = ref(false)
const selectedUser = ref<AdminUser | null>(null)
const newRate = ref<number | null>(null)
const currentPage = ref(1)
const pageSize = ref(10)
const batchFactor = ref<number | null>(null)

let searchTimeout: ReturnType<typeof setTimeout>

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getPlatformChipClasses = (platform: string) => {
  switch (platform) {
    case 'anthropic':
      return 'theme-chip theme-chip--regular theme-chip--brand-orange inline-flex items-center gap-1.5'
    case 'openai':
      return 'theme-chip theme-chip--regular theme-chip--success inline-flex items-center gap-1.5'
    case 'antigravity':
      return 'theme-chip theme-chip--regular theme-chip--brand-purple inline-flex items-center gap-1.5'
    default:
      return 'theme-chip theme-chip--regular theme-chip--info inline-flex items-center gap-1.5'
  }
}

const getUserStatusClasses = (status: string) => {
  return joinClassNames(
    'theme-chip theme-chip--regular inline-flex rounded-full',
    status === 'active'
      ? 'theme-chip--success'
      : 'theme-chip--neutral'
  )
}

const getActionButtonClasses = (tone: 'danger') => {
  return joinClassNames(
    'group-rate-multipliers-modal__action-button',
    tone === 'danger' && 'group-rate-multipliers-modal__action-button--danger'
  )
}

// 是否显示"最终倍率"预览列
const showFinalRate = computed(() => {
  return batchFactor.value != null && batchFactor.value > 0 && batchFactor.value !== 1
})

// 计算最终倍率预览
const computeFinalRate = (rate: number) => {
  if (!batchFactor.value) return rate
  return parseFloat((rate * batchFactor.value).toFixed(6))
}

// 检测是否有未保存的修改
const isDirty = computed(() => {
  if (localEntries.value.length !== serverEntries.value.length) return true
  const serverMap = new Map(serverEntries.value.map(e => [e.user_id, e.rate_multiplier]))
  return localEntries.value.some(e => {
    const serverRate = serverMap.get(e.user_id)
    return serverRate === undefined || serverRate !== e.rate_multiplier
  })
})

const paginatedLocalEntries = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return localEntries.value.slice(start, start + pageSize.value)
})

const cloneEntries = (entries: GroupRateMultiplierEntry[]): LocalEntry[] => {
  return entries.map(e => ({ ...e }))
}

async function loadEntries() {
  if (!props.group) return
  loading.value = true
  try {
    serverEntries.value = await adminAPI.groups.getGroupRateMultipliers(props.group.id)
    localEntries.value = cloneEntries(serverEntries.value)
    adjustPage()
  } catch (error) {
    appStore.showError(t('admin.groups.failedToLoad'))
    console.error('Error loading group rate multipliers:', error)
  } finally {
    loading.value = false
  }
}

const adjustPage = () => {
  const totalPages = Math.max(1, Math.ceil(localEntries.value.length / pageSize.value))
  if (currentPage.value > totalPages) {
    currentPage.value = totalPages
  }
}

watch(() => props.show, (val) => {
  if (val && props.group) {
    currentPage.value = 1
    batchFactor.value = null
    searchQuery.value = ''
    searchResults.value = []
    selectedUser.value = null
    newRate.value = null
    loadEntries()
  }
}, { immediate: true })

const handlePageSizeChange = (newSize: number) => {
  pageSize.value = newSize
  currentPage.value = 1
}

const handleSearchUsers = () => {
  clearTimeout(searchTimeout)
  selectedUser.value = null
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    showDropdown.value = false
    return
  }
  searchTimeout = setTimeout(async () => {
    try {
      const res = await adminAPI.users.list(1, 10, { search: searchQuery.value.trim() })
      searchResults.value = res.items
      showDropdown.value = true
    } catch {
      searchResults.value = []
    }
  }, 300)
}

const selectUser = (user: AdminUser) => {
  selectedUser.value = user
  searchQuery.value = user.email
  showDropdown.value = false
  searchResults.value = []
}

// 本地添加（或覆盖已有用户）
const handleAddLocal = () => {
  if (!selectedUser.value || !newRate.value) return
  const user = selectedUser.value
  const idx = localEntries.value.findIndex(e => e.user_id === user.id)
  const entry: LocalEntry = {
    user_id: user.id,
    user_name: user.username || '',
    user_email: user.email,
    user_notes: user.notes || '',
    user_status: user.status || 'active',
    rate_multiplier: newRate.value
  }
  if (idx >= 0) {
    localEntries.value[idx] = entry
  } else {
    localEntries.value.push(entry)
  }
  searchQuery.value = ''
  selectedUser.value = null
  newRate.value = null
  adjustPage()
}

// 本地修改倍率
const updateLocalRate = (userId: number, value: string) => {
  const num = parseFloat(value)
  if (isNaN(num)) return
  const entry = localEntries.value.find(e => e.user_id === userId)
  if (entry) {
    entry.rate_multiplier = num
  }
}

// 本地删除
const removeLocal = (userId: number) => {
  localEntries.value = localEntries.value.filter(e => e.user_id !== userId)
  adjustPage()
}

// 批量乘数应用到本地
const applyBatchFactor = () => {
  if (!batchFactor.value || batchFactor.value <= 0) return
  for (const entry of localEntries.value) {
    entry.rate_multiplier = parseFloat((entry.rate_multiplier * batchFactor.value).toFixed(6))
  }
  batchFactor.value = null
}

// 本地清空
const clearAllLocal = () => {
  localEntries.value = []
}

// 取消：恢复到服务器数据
const handleCancel = () => {
  localEntries.value = cloneEntries(serverEntries.value)
  batchFactor.value = null
  adjustPage()
}

// 保存：一次性提交所有数据
const handleSave = async () => {
  if (!props.group) return
  saving.value = true
  try {
    const entries = localEntries.value.map(e => ({
      user_id: e.user_id,
      rate_multiplier: e.rate_multiplier
    }))
    await adminAPI.groups.batchSetGroupRateMultipliers(props.group.id, entries)
    appStore.showSuccess(t('admin.groups.rateSaved'))
    emit('success')
    emit('close')
  } catch (error) {
    appStore.showError(t('admin.groups.failedToSave'))
    console.error('Error saving rate multipliers:', error)
  } finally {
    saving.value = false
  }
}

// 关闭时如果有未保存修改，先恢复
const handleClose = () => {
  if (isDirty.value) {
    localEntries.value = cloneEntries(serverEntries.value)
  }
  emit('close')
}

// 点击外部关闭下拉
const handleClickOutside = () => {
  showDropdown.value = false
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  clearTimeout(searchTimeout)
})
</script>

<style scoped>
.hide-spinner::-webkit-outer-spin-button,
.hide-spinner::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.hide-spinner {
  -moz-appearance: textfield;
}

.group-rate-multipliers-modal__group-summary,
.group-rate-multipliers-modal__control-panel,
.group-rate-multipliers-modal__table-shell {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.group-rate-multipliers-modal__content {
  display: flex;
  flex-direction: column;
  gap: var(--theme-table-layout-gap);
}

.group-rate-multipliers-modal__group-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.75rem;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  padding: calc(var(--theme-table-mobile-card-padding) * 0.625) var(--theme-table-mobile-card-padding);
  font-size: 0.875rem;
  background: color-mix(in srgb, var(--theme-surface-soft) 72%, var(--theme-surface));
}

.group-rate-multipliers-modal__control-panel {
  border-radius: calc(var(--theme-surface-radius) + 2px);
  padding: var(--theme-table-mobile-card-padding);
}

.group-rate-multipliers-modal__group-name,
.group-rate-multipliers-modal__section-title,
.group-rate-multipliers-modal__search-name,
.group-rate-multipliers-modal__cell-strong {
  color: var(--theme-page-text);
}

.group-rate-multipliers-modal__separator,
.group-rate-multipliers-modal__group-meta,
.group-rate-multipliers-modal__search-id,
.group-rate-multipliers-modal__search-email,
.group-rate-multipliers-modal__batch-label,
.group-rate-multipliers-modal__cell-muted,
.group-rate-multipliers-modal__empty {
  color: var(--theme-page-muted);
}

.group-rate-multipliers-modal__cell-soft {
  color: color-mix(in srgb, var(--theme-page-muted) 76%, transparent);
}

.group-rate-multipliers-modal__search-dropdown {
  position: absolute;
  left: 0;
  right: 0;
  top: 100%;
  z-index: 10;
  margin-top: var(--theme-floating-panel-gap);
  max-height: var(--theme-search-dropdown-max-height);
  overflow-y: auto;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.group-rate-multipliers-modal__search-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.5rem;
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.6);
  text-align: left;
  font-size: 0.875rem;
  color: var(--theme-page-text);
}

.group-rate-multipliers-modal__search-item:hover {
  background: var(--theme-dropdown-item-hover-bg);
}

.group-rate-multipliers-modal__number-input {
  border: 1px solid var(--theme-input-border);
  border-radius: calc(var(--theme-button-radius) - 2px);
  background: var(--theme-input-bg);
  color: var(--theme-input-text);
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease;
}

.group-rate-multipliers-modal__number-input::placeholder {
  color: var(--theme-input-placeholder);
}

.group-rate-multipliers-modal__number-input:focus {
  border-color: color-mix(in srgb, var(--theme-accent) 68%, var(--theme-input-border));
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 14%, transparent);
}

.group-rate-multipliers-modal__number-input--compact {
  width: 5rem;
  padding: calc(var(--theme-button-padding-y) * 0.45) calc(var(--theme-button-padding-x) * 0.35);
  text-align: center;
  font-size: 0.875rem;
  font-weight: 500;
  box-shadow: none;
}

.group-rate-multipliers-modal__number-input--batch {
  width: 5rem;
}

.group-rate-multipliers-modal__batch-bar,
.group-rate-multipliers-modal__footer {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 70%, transparent);
}

.group-rate-multipliers-modal__clear-button {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: calc(var(--theme-button-padding-y) * 0.6) calc(var(--theme-button-padding-x) * 0.75);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s ease;
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 20%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.group-rate-multipliers-modal__clear-button:hover,
.group-rate-multipliers-modal__action-button--danger:hover {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 16%, var(--theme-surface));
}

.group-rate-multipliers-modal__loading-icon,
.group-rate-multipliers-modal__final-rate,
.group-rate-multipliers-modal__revert {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.group-rate-multipliers-modal__table-head {
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: var(--theme-table-head-bg);
}

.group-rate-multipliers-modal__table-header {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
  text-align: left;
  font-size: var(--theme-table-head-font-size);
  font-weight: 500;
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
  color: var(--theme-table-head-text);
}

.group-rate-multipliers-modal__table-header--icon {
  width: 2.5rem;
}

.group-rate-multipliers-modal__table-header--accent {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.group-rate-multipliers-modal__table-scroll {
  max-height: var(--theme-balance-history-list-max-height);
  overflow-y: auto;
}

.group-rate-multipliers-modal__table-cell {
  padding: calc(var(--theme-button-padding-y) * 0.8) calc(var(--theme-button-padding-x) * 0.6);
}

.group-rate-multipliers-modal__table-cell--nowrap {
  white-space: nowrap;
}

.group-rate-multipliers-modal__table-cell--notes {
  max-width: calc(var(--theme-settings-menu-width-sm) + 1rem);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.group-rate-multipliers-modal__table-cell--icon {
  width: 2.5rem;
}

.group-rate-multipliers-modal__table-body tr + tr td {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.group-rate-multipliers-modal__table-row:hover {
  background: var(--theme-table-row-hover);
}

.group-rate-multipliers-modal__action-button {
  border-radius: calc(var(--theme-button-radius) - 2px);
  padding: 0.25rem;
  transition: color 0.2s ease, background-color 0.2s ease;
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group-rate-multipliers-modal__action-button--danger:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.group-rate-multipliers-modal__unsaved {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.group-rate-multipliers-modal__batch-apply-button,
.group-rate-multipliers-modal__footer-button {
  padding: calc(var(--theme-button-padding-y) * 0.5) calc(var(--theme-button-padding-x) * 0.8);
}

.group-rate-multipliers-modal__loading {
  display: flex;
  justify-content: center;
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
}

.group-rate-multipliers-modal__empty {
  padding: calc(var(--theme-table-mobile-empty-padding) * 0.5) 0;
  text-align: center;
  font-size: 0.875rem;
}

.group-rate-multipliers-modal__revert:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 22%, var(--theme-accent) 78%);
}
</style>
