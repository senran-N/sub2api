<template>
  <BaseDialog :show="show" :title="t('admin.users.userApiKeys')" width="wide" @close="handleClose">
    <div v-if="user" class="space-y-4">
      <div class="user-api-keys-modal__user-card">
        <div class="user-api-keys-modal__avatar">
          <span class="user-api-keys-modal__avatar-text">{{ user.email.charAt(0).toUpperCase() }}</span>
        </div>
        <div>
          <p class="user-api-keys-modal__user-email">{{ user.email }}</p>
          <p class="user-api-keys-modal__user-name">{{ user.username }}</p>
        </div>
      </div>
      <div v-if="loading" class="user-api-keys-modal__state">
        <svg class="user-api-keys-modal__spinner" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
      </div>
      <div v-else-if="apiKeys.length === 0" class="user-api-keys-modal__state">
        <p class="user-api-keys-modal__muted">{{ t('admin.users.noApiKeys') }}</p>
      </div>
      <div v-else ref="scrollContainerRef" class="user-api-keys-modal__list space-y-3 overflow-y-auto" @scroll="closeGroupSelector">
        <div v-for="key in apiKeys" :key="key.id" class="user-api-keys-modal__key-card">
          <div class="flex items-start justify-between">
            <div class="min-w-0 flex-1">
              <div class="mb-1 flex items-center gap-2">
                <span class="user-api-keys-modal__key-name">{{ key.name }}</span>
                <span :class="getKeyStatusClasses(key.status)">{{ key.status }}</span>
              </div>
              <p class="user-api-keys-modal__key-value">{{ key.key.substring(0, 20) }}...{{ key.key.substring(key.key.length - 8) }}</p>
            </div>
          </div>
          <div class="user-api-keys-modal__meta-row">
            <div class="flex items-center gap-1">
              <span>{{ t('admin.users.group') }}:</span>
              <button
                :ref="(el) => setGroupButtonRef(key.id, el)"
                @click="openGroupSelector(key)"
                class="user-api-keys-modal__group-trigger"
                :disabled="updatingKeyIds.has(key.id)"
              >
                <GroupBadge
                  v-if="key.group_id && key.group"
                  :name="key.group.name"
                  :platform="key.group.platform"
                  :subscription-type="key.group.subscription_type"
                  :rate-multiplier="key.group.rate_multiplier"
                />
                <span v-else class="user-api-keys-modal__muted italic">{{ t('admin.users.none') }}</span>
                <svg v-if="updatingKeyIds.has(key.id)" class="user-api-keys-modal__mini-spinner" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                <svg v-else class="user-api-keys-modal__chevron" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 15L12 18.75 15.75 15m-7.5-6L12 5.25 15.75 9" /></svg>
              </button>
            </div>
            <div class="flex items-center gap-1"><span>{{ t('admin.users.columns.created') }}: {{ formatDateTime(key.created_at) }}</span></div>
          </div>
        </div>
      </div>
    </div>
  </BaseDialog>

  <!-- Group Selector Dropdown -->
  <Teleport to="body">
    <div
      v-if="groupSelectorKeyId !== null && dropdownPosition"
      ref="dropdownRef"
      class="user-api-keys-modal__dropdown animate-in fade-in slide-in-from-top-2 fixed z-[100000020] overflow-hidden duration-200"
      :style="{ top: dropdownPosition.top + 'px', left: dropdownPosition.left + 'px' }"
    >
      <div class="user-api-keys-modal__dropdown-panel">
        <!-- Unbind option -->
        <button
          @click="changeGroup(selectedKeyForGroup!, null)"
          :class="getDropdownOptionClasses(!selectedKeyForGroup?.group_id)"
        >
          <span class="user-api-keys-modal__muted italic">{{ t('admin.users.none') }}</span>
          <svg
            v-if="!selectedKeyForGroup?.group_id"
            class="user-api-keys-modal__dropdown-check"
            fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"
          ><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" /></svg>
        </button>
        <!-- Group options -->
        <button
          v-for="group in allGroups"
          :key="group.id"
          @click="changeGroup(selectedKeyForGroup!, group.id)"
          :class="getDropdownOptionClasses(selectedKeyForGroup?.group_id === group.id)"
        >
          <GroupOptionItem
            :name="group.name"
            :platform="group.platform"
            :subscription-type="group.subscription_type"
            :rate-multiplier="group.rate_multiplier"
            :description="group.description"
            :selected="selectedKeyForGroup?.group_id === group.id"
          />
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick, type ComponentPublicInstance } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import { formatDateTime } from '@/utils/format'
import { clampFloatingPanelPosition, readThemePixelValue } from '@/utils/floatingPanel'
import type { AdminUser, AdminGroup, ApiKey } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null }>()
const emit = defineEmits(['close'])
const { t } = useI18n()
const appStore = useAppStore()
const themeVersion = useDocumentThemeVersion()

const apiKeys = ref<ApiKey[]>([])
const allGroups = ref<AdminGroup[]>([])
const loading = ref(false)
const updatingKeyIds = ref(new Set<number>())
const groupSelectorKeyId = ref<number | null>(null)
const dropdownPosition = ref<{ top: number; left: number } | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const scrollContainerRef = ref<HTMLElement | null>(null)
const groupButtonRefs = ref<Map<number, HTMLElement>>(new Map())

const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getKeyStatusClasses = (status: string) => {
  return joinClassNames(
    'theme-chip theme-chip--compact text-xs',
    status === 'active' ? 'theme-chip--success' : 'theme-chip--danger'
  )
}

const getDropdownOptionClasses = (isSelected: boolean) => {
  return joinClassNames(
    'user-api-keys-modal__dropdown-option',
    isSelected && 'user-api-keys-modal__dropdown-option--selected'
  )
}

const getErrorMessage = (error: unknown, fallbackMessage: string) => {
  return error instanceof Error && error.message ? error.message : fallbackMessage
}

const setGroupButtonRef = (keyId: number, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) {
    groupButtonRefs.value.set(keyId, el)
  } else {
    groupButtonRefs.value.delete(keyId)
  }
}

watch(() => props.show, (v) => {
  if (v && props.user) {
    load()
    loadGroups()
  } else {
    closeGroupSelector()
  }
})

const load = async () => {
  if (!props.user) return
  loading.value = true
  groupButtonRefs.value.clear()
  try {
    const res = await adminAPI.users.getUserApiKeys(props.user.id)
    apiKeys.value = res.items || []
  } catch (error) {
    console.error('Failed to load API keys:', error)
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    allGroups.value = groups
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const getDropdownWidth = () => readThemePixelValue('--theme-user-api-keys-dropdown-width', 256)
const getDropdownHeight = () => {
  const maxHeight = readThemePixelValue('--theme-user-api-keys-dropdown-max-height', 256)
  const padding = readThemePixelValue('--theme-user-api-keys-dropdown-padding', 6)
  return maxHeight + padding * 2
}

const updateGroupSelectorPosition = () => {
  if (groupSelectorKeyId.value === null) {
    dropdownPosition.value = null
    return
  }

  const buttonEl = groupButtonRefs.value.get(groupSelectorKeyId.value)
  if (!buttonEl) {
    closeGroupSelector()
    return
  }

  const rect = buttonEl.getBoundingClientRect()
  const gap = readThemePixelValue('--theme-floating-panel-gap', 4)
  const viewportPadding = readThemePixelValue('--theme-floating-panel-viewport-padding', 8)
  const panelWidth = dropdownRef.value?.offsetWidth ?? getDropdownWidth()
  const panelHeight = dropdownRef.value?.offsetHeight ?? getDropdownHeight()
  const spaceBelow = window.innerHeight - rect.bottom
  const openUpward = spaceBelow < panelHeight && rect.top > spaceBelow
  const desiredPosition = {
    top: openUpward ? rect.top - panelHeight - gap : rect.bottom + gap,
    left: rect.left
  }

  dropdownPosition.value = clampFloatingPanelPosition(desiredPosition, {
    panelWidth,
    panelHeight,
    padding: viewportPadding
  })
}

const openGroupSelector = (key: ApiKey) => {
  if (groupSelectorKeyId.value === key.id) {
    closeGroupSelector()
  } else {
    if (!groupButtonRefs.value.has(key.id)) {
      closeGroupSelector()
      return
    }
    groupSelectorKeyId.value = key.id
  }
}

const closeGroupSelector = () => {
  groupSelectorKeyId.value = null
  dropdownPosition.value = null
}

watch(
  [groupSelectorKeyId, themeVersion],
  async ([keyId]) => {
    if (keyId === null) {
      dropdownPosition.value = null
      return
    }

    await nextTick()
    updateGroupSelectorPosition()
  },
  { immediate: true }
)

const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
  closeGroupSelector()
  if (key.group_id === newGroupId || (!key.group_id && newGroupId === null)) return

  updatingKeyIds.value.add(key.id)
  try {
    const result = await adminAPI.apiKeys.updateApiKeyGroup(key.id, newGroupId)
    // Update local data
    const idx = apiKeys.value.findIndex((k) => k.id === key.id)
    if (idx !== -1) {
      apiKeys.value[idx] = result.api_key
    }
    if (result.auto_granted_group_access && result.granted_group_name) {
      appStore.showSuccess(t('admin.users.groupChangedWithGrant', { group: result.granted_group_name }))
    } else {
      appStore.showSuccess(t('admin.users.groupChangedSuccess'))
    }
  } catch (error) {
    appStore.showError(getErrorMessage(error, t('admin.users.groupChangeFailed')))
  } finally {
    updatingKeyIds.value.delete(key.id)
  }
}

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && groupSelectorKeyId.value !== null) {
    event.stopPropagation()
    closeGroupSelector()
  }
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (dropdownRef.value && !dropdownRef.value.contains(target)) {
    // Check if the click is on one of the group trigger buttons
    for (const el of groupButtonRefs.value.values()) {
      if (el.contains(target)) return
    }
    closeGroupSelector()
  }
}

const handleClose = () => {
  closeGroupSelector()
  emit('close')
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeyDown, true)
  window.addEventListener('resize', updateGroupSelectorPosition)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeyDown, true)
  window.removeEventListener('resize', updateGroupSelectorPosition)
})
</script>

<style scoped>
.user-api-keys-modal__list {
  max-height: var(--theme-user-api-keys-list-max-height);
}

.user-api-keys-modal__user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  padding: 1rem;
}

.user-api-keys-modal__avatar {
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: color-mix(in srgb, var(--theme-accent-soft) 90%, var(--theme-surface));
}

.user-api-keys-modal__avatar-text,
.user-api-keys-modal__spinner,
.user-api-keys-modal__mini-spinner,
.user-api-keys-modal__dropdown-check {
  color: var(--theme-accent);
}

.user-api-keys-modal__user-email,
.user-api-keys-modal__key-name {
  color: var(--theme-page-text);
  font-weight: 600;
}

.user-api-keys-modal__user-name,
.user-api-keys-modal__key-value,
.user-api-keys-modal__meta-row,
.user-api-keys-modal__muted,
.user-api-keys-modal__chevron {
  color: var(--theme-page-muted);
}

.user-api-keys-modal__state {
  display: flex;
  justify-content: center;
  padding: 2rem 0;
}

.user-api-keys-modal__spinner {
  height: 2rem;
  width: 2rem;
  animation: spin 1s linear infinite;
}

.user-api-keys-modal__key-card {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-surface);
  padding: 1rem;
}

.user-api-keys-modal__key-value {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: var(--theme-font-mono);
  font-size: 0.875rem;
}

.user-api-keys-modal__meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 0.75rem;
  font-size: 0.75rem;
}

.user-api-keys-modal__group-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin: -0.125rem -0.25rem;
  border-radius: calc(var(--theme-button-radius) - 2px);
  padding: 0.125rem 0.25rem;
  transition: background-color 0.18s ease;
}

.user-api-keys-modal__group-trigger:hover:not(:disabled),
.user-api-keys-modal__group-trigger:focus-visible {
  background: color-mix(in srgb, var(--theme-button-ghost-hover-bg) 90%, transparent);
  outline: none;
}

.user-api-keys-modal__group-trigger:disabled {
  cursor: not-allowed;
}

.user-api-keys-modal__mini-spinner,
.user-api-keys-modal__chevron {
  height: 0.75rem;
  width: 0.75rem;
}

.user-api-keys-modal__mini-spinner {
  animation: spin 1s linear infinite;
}

.user-api-keys-modal__dropdown {
  width: min(
    var(--theme-user-api-keys-dropdown-width),
    calc(100vw - (var(--theme-floating-panel-viewport-padding) + var(--theme-floating-panel-viewport-padding)))
  );
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.user-api-keys-modal__dropdown-panel {
  max-height: var(--theme-user-api-keys-dropdown-max-height);
  overflow-y: auto;
  padding: var(--theme-user-api-keys-dropdown-padding);
}

.user-api-keys-modal__dropdown-option {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  transition: background-color 0.18s ease;
}

.user-api-keys-modal__dropdown-option:hover,
.user-api-keys-modal__dropdown-option:focus-visible {
  background: color-mix(in srgb, var(--theme-button-ghost-hover-bg) 90%, transparent);
  outline: none;
}

.user-api-keys-modal__dropdown-option--selected {
  background: color-mix(in srgb, var(--theme-accent-soft) 78%, var(--theme-surface));
}
</style>
