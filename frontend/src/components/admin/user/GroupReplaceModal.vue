<template>
  <BaseDialog :show="show" :title="t('admin.users.replaceGroupTitle')" width="narrow" @close="handleClose">
    <div v-if="oldGroup" class="space-y-4">
      <p class="group-replace-modal__description text-sm">
        {{ t('admin.users.replaceGroupHint', { old: oldGroup.name }) }}
      </p>

      <div class="group-replace-modal__current-group">
        <div class="flex items-center gap-2">
          <Icon name="shield" size="sm" class="group-replace-modal__old-group-icon" />
          <span class="group-replace-modal__group-name font-medium">{{ oldGroup.name }}</span>
          <Icon name="arrowRight" size="sm" class="group-replace-modal__arrow ml-auto" />
          <span v-if="selectedGroupId" class="group-replace-modal__selected-group font-medium">
            {{ availableGroups.find(g => g.id === selectedGroupId)?.name }}
          </span>
          <span v-else class="group-replace-modal__placeholder text-sm">?</span>
        </div>
      </div>

      <div v-if="availableGroups.length > 0" class="group-replace-modal__list space-y-2 overflow-y-auto">
        <label
          v-for="group in availableGroups"
          :key="group.id"
          class="group-replace-modal__group-option flex cursor-pointer items-center gap-3 border-2 transition-all"
          :class="selectedGroupId === group.id
            ? 'group-replace-modal__group-option--selected'
            : 'group-replace-modal__group-option--idle'"
        >
          <input
            type="radio"
            :value="group.id"
            v-model="selectedGroupId"
            class="sr-only"
          />
          <div
            class="group-replace-modal__radio flex h-5 w-5 items-center justify-center rounded-full border-2 transition-all"
            :class="selectedGroupId === group.id
              ? 'group-replace-modal__radio--selected'
              : 'group-replace-modal__radio--idle'"
          >
            <div v-if="selectedGroupId === group.id" class="group-replace-modal__radio-dot h-2 w-2 rounded-full"></div>
          </div>
          <div class="flex-1">
            <span class="group-replace-modal__group-name font-medium">{{ group.name }}</span>
            <span class="group-replace-modal__platform ml-2 text-xs">{{ group.platform }}</span>
          </div>
        </label>
      </div>

      <div v-else class="group-replace-modal__empty text-center text-sm">
        {{ t('admin.users.noOtherGroups') }}
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="$emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button
          @click="handleReplace"
          :disabled="!selectedGroupId || submitting"
          class="btn btn-primary"
        >
          <svg v-if="submitting" class="-ml-1 mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ submitting ? t('common.saving') : t('admin.users.replaceGroupConfirm') }}
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
import type { AdminUser, AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

interface Props {
  show: boolean
  user: AdminUser | null
  oldGroup: { id: number; name: string } | null
  allGroups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const selectedGroupId = ref<number | null>(null)
const submitting = ref(false)
let replaceRequestSequence = 0

const getErrorMessage = (error: unknown, fallbackMessage: string) => {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { detail?: string } } }).response
    if (typeof response?.data?.detail === 'string' && response.data.detail.trim()) {
      return response.data.detail
    }
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallbackMessage
}

const availableGroups = computed(() => {
  if (!props.oldGroup) return []
  return props.allGroups.filter(
    g => g.status === 'active' && g.is_exclusive && g.subscription_type === 'standard' && g.id !== props.oldGroup!.id
  )
})

watch(() => props.show, (v) => {
  if (!v) {
    replaceRequestSequence += 1
    submitting.value = false
    selectedGroupId.value = null
    return
  }

  selectedGroupId.value = null
}, { immediate: true })

watch(
  () => [props.user?.id, props.oldGroup?.id] as const,
  () => {
    replaceRequestSequence += 1
    submitting.value = false
    selectedGroupId.value = null
  }
)

const handleReplace = async () => {
  if (!props.user || !props.oldGroup || !selectedGroupId.value) return
  const requestSequence = ++replaceRequestSequence
  submitting.value = true

  try {
    const result = await adminAPI.users.replaceGroup(props.user.id, props.oldGroup.id, selectedGroupId.value)
    if (
      requestSequence !== replaceRequestSequence ||
      !props.show ||
      props.user == null ||
      props.oldGroup == null
    ) {
      return
    }
    appStore.showSuccess(t('admin.users.replaceGroupSuccess', { count: result.migrated_keys }))
    emit('success')
    emit('close')
  } catch (error) {
    if (
      requestSequence !== replaceRequestSequence ||
      !props.show ||
      props.user == null ||
      props.oldGroup == null
    ) {
      return
    }
    appStore.showError(getErrorMessage(error, t('admin.users.replaceGroupFailed')))
  } finally {
    if (requestSequence === replaceRequestSequence) {
      submitting.value = false
    }
  }
}

const handleClose = () => {
  replaceRequestSequence += 1
  submitting.value = false
  selectedGroupId.value = null
  emit('close')
}
</script>

<style scoped>
.group-replace-modal__description,
.group-replace-modal__platform,
.group-replace-modal__placeholder,
.group-replace-modal__empty,
.group-replace-modal__arrow {
  color: var(--theme-page-muted);
}

.group-replace-modal__current-group {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  padding: var(--theme-group-replace-card-padding);
}

.group-replace-modal__old-group-icon {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.group-replace-modal__group-name {
  color: var(--theme-page-text);
}

.group-replace-modal__selected-group {
  color: var(--theme-accent);
}

.group-replace-modal__group-option--selected {
  border-color: color-mix(in srgb, var(--theme-accent) 46%, var(--theme-card-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
}

.group-replace-modal__list {
  max-height: var(--theme-group-replace-list-max-height);
}

.group-replace-modal__group-option {
  border-radius: calc(var(--theme-surface-radius) + 2px);
  padding: var(--theme-group-replace-card-padding);
}

.group-replace-modal__empty {
  padding-block: var(--theme-user-attributes-state-padding-y);
}

.group-replace-modal__group-option--idle {
  border-color: color-mix(in srgb, var(--theme-card-border) 74%, transparent);
}

.group-replace-modal__group-option--idle:hover {
  border-color: color-mix(in srgb, var(--theme-card-border) 92%, transparent);
}

.group-replace-modal__radio--selected {
  border-color: var(--theme-accent);
  background: var(--theme-accent);
}

.group-replace-modal__radio--idle {
  border-color: var(--theme-input-border);
}

.group-replace-modal__radio-dot {
  background: var(--theme-filled-text);
}
</style>
