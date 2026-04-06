<template>
  <BaseDialog :show="show" :title="t('admin.usage.cleanup.title')" width="wide" @close="handleClose">
    <div class="space-y-4">
      <UsageFilters
        v-model="localFilters"
        v-model:startDate="localStartDate"
        v-model:endDate="localEndDate"
        :exporting="false"
        :show-actions="false"
        @change="noop"
      />

      <div class="usage-cleanup-dialog__warning text-sm">
        {{ t('admin.usage.cleanup.warning') }}
      </div>

      <div class="usage-cleanup-dialog__tasks-card">
        <div class="flex items-center justify-between">
          <h4 class="usage-cleanup-dialog__tasks-title text-sm font-semibold">
            {{ t('admin.usage.cleanup.recentTasks') }}
          </h4>
          <button type="button" class="btn btn-ghost btn-sm" @click="loadTasks">
            {{ t('common.refresh') }}
          </button>
        </div>

        <div class="mt-3 space-y-2">
          <div v-if="tasksLoading" class="usage-cleanup-dialog__meta text-sm">
            {{ t('admin.usage.cleanup.loadingTasks') }}
          </div>
          <div v-else-if="tasks.length === 0" class="usage-cleanup-dialog__meta text-sm">
            {{ t('admin.usage.cleanup.noTasks') }}
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="task in tasks"
              :key="task.id"
              class="usage-cleanup-dialog__task-card flex flex-col gap-2 text-sm"
            >
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                  <span :class="statusClass(task.status)">
                    {{ statusLabel(task.status) }}
                  </span>
                  <span class="usage-cleanup-dialog__subtle text-xs">#{{ task.id }}</span>
                  <button
                    v-if="canCancel(task)"
                    type="button"
                    class="usage-cleanup-dialog__cancel-button btn btn-ghost btn-xs"
                    @click="openCancelConfirm(task)"
                  >
                    {{ t('admin.usage.cleanup.cancel') }}
                  </button>
                </div>
                <div class="usage-cleanup-dialog__subtle text-xs">
                  {{ formatDateTime(task.created_at) }}
                </div>
              </div>
              <div class="usage-cleanup-dialog__meta flex flex-wrap items-center gap-4 text-xs">
                <span>{{ t('admin.usage.cleanup.range') }}: {{ formatRange(task) }}</span>
                <span>{{ t('admin.usage.cleanup.deletedRows') }}: {{ task.deleted_rows.toLocaleString() }}</span>
              </div>
              <div v-if="task.error_message" class="usage-cleanup-dialog__task-error text-xs">
                {{ task.error_message }}
              </div>
            </div>
          </div>
        </div>

        <Pagination
          v-if="tasksTotal > tasksPageSize"
          class="mt-4"
          :total="tasksTotal"
          :page="tasksPage"
          :page-size="tasksPageSize"
          :page-size-options="[5]"
          :show-page-size-selector="false"
          :show-jump="true"
          @update:page="handleTaskPageChange"
          @update:pageSize="handleTaskPageSizeChange"
        />
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-danger" :disabled="submitting" @click="openConfirm">
          {{ submitting ? t('admin.usage.cleanup.submitting') : t('admin.usage.cleanup.submit') }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <ConfirmDialog
    :show="confirmVisible"
    :title="t('admin.usage.cleanup.confirmTitle')"
    :message="t('admin.usage.cleanup.confirmMessage')"
    :confirm-text="t('admin.usage.cleanup.confirmSubmit')"
    danger
    @confirm="submitCleanup"
    @cancel="confirmVisible = false"
  />

  <ConfirmDialog
    :show="cancelConfirmVisible"
    :title="t('admin.usage.cleanup.cancelConfirmTitle')"
    :message="t('admin.usage.cleanup.cancelConfirmMessage')"
    :confirm-text="t('admin.usage.cleanup.cancelConfirm')"
    danger
    @confirm="cancelTask"
    @cancel="cancelConfirmVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import UsageFilters from '@/components/admin/usage/UsageFilters.vue'
import { adminUsageAPI } from '@/api/admin/usage'
import type { AdminUsageQueryParams, UsageCleanupTask, CreateUsageCleanupTaskRequest } from '@/api/admin/usage'
import { requestTypeToLegacyStream } from '@/utils/usageRequestType'

interface Props {
  show: boolean
  filters: AdminUsageQueryParams
  startDate: string
  endDate: string
}

const props = defineProps<Props>()
const emit = defineEmits(['close'])

const { t } = useI18n()
const appStore = useAppStore()

const localFilters = ref<AdminUsageQueryParams>({})
const localStartDate = ref('')
const localEndDate = ref('')

const tasks = ref<UsageCleanupTask[]>([])
const tasksLoading = ref(false)
const tasksPage = ref(1)
const tasksPageSize = ref(5)
const tasksTotal = ref(0)
const submitting = ref(false)
const confirmVisible = ref(false)
const cancelConfirmVisible = ref(false)
const canceling = ref(false)
const cancelTarget = ref<UsageCleanupTask | null>(null)
let pollTimer: number | null = null

const noop = () => {}

const resetFilters = () => {
  localFilters.value = { ...props.filters }
  localStartDate.value = props.startDate
  localEndDate.value = props.endDate
  localFilters.value.start_date = localStartDate.value
  localFilters.value.end_date = localEndDate.value
  tasksPage.value = 1
  tasksTotal.value = 0
}

const startPolling = () => {
  stopPolling()
  pollTimer = window.setInterval(() => {
    loadTasks()
  }, 10000)
}

const stopPolling = () => {
  if (pollTimer !== null) {
    window.clearInterval(pollTimer)
    pollTimer = null
  }
}

const handleClose = () => {
  stopPolling()
  confirmVisible.value = false
  cancelConfirmVisible.value = false
  canceling.value = false
  cancelTarget.value = null
  submitting.value = false
  emit('close')
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    pending: t('admin.usage.cleanup.status.pending'),
    running: t('admin.usage.cleanup.status.running'),
    succeeded: t('admin.usage.cleanup.status.succeeded'),
    failed: t('admin.usage.cleanup.status.failed'),
    canceled: t('admin.usage.cleanup.status.canceled')
  }
  return map[status] || status
}

const statusClass = (status: string) => {
  const map: Record<string, string> = {
    pending: 'theme-chip theme-chip--compact theme-chip--warning',
    running: 'theme-chip theme-chip--compact theme-chip--info',
    succeeded: 'theme-chip theme-chip--compact theme-chip--success',
    failed: 'theme-chip theme-chip--compact theme-chip--danger',
    canceled: 'theme-chip theme-chip--compact theme-chip--neutral'
  }
  return map[status] || 'theme-chip theme-chip--compact theme-chip--neutral'
}

const formatDateTime = (value?: string | null) => {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

const formatRange = (task: UsageCleanupTask) => {
  const start = formatDateTime(task.filters.start_time)
  const end = formatDateTime(task.filters.end_time)
  return `${start} ~ ${end}`
}

const getUserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

const loadTasks = async () => {
  if (!props.show) return
  tasksLoading.value = true
  try {
    const res = await adminUsageAPI.listCleanupTasks({
      page: tasksPage.value,
      page_size: tasksPageSize.value
    })
    tasks.value = res.items || []
    tasksTotal.value = res.total || 0
    if (res.page) {
      tasksPage.value = res.page
    }
    if (res.page_size) {
      tasksPageSize.value = res.page_size
    }
  } catch (error) {
    console.error('Failed to load cleanup tasks:', error)
    appStore.showError(t('admin.usage.cleanup.loadFailed'))
  } finally {
    tasksLoading.value = false
  }
}

const handleTaskPageChange = (page: number) => {
  tasksPage.value = page
  loadTasks()
}

const handleTaskPageSizeChange = (size: number) => {
  if (!Number.isFinite(size) || size <= 0) return
  tasksPageSize.value = size
  tasksPage.value = 1
  loadTasks()
}

const openConfirm = () => {
  confirmVisible.value = true
}

const canCancel = (task: UsageCleanupTask) => {
  return task.status === 'pending' || task.status === 'running'
}

const openCancelConfirm = (task: UsageCleanupTask) => {
  cancelTarget.value = task
  cancelConfirmVisible.value = true
}

const buildPayload = (): CreateUsageCleanupTaskRequest | null => {
  if (!localStartDate.value || !localEndDate.value) {
    appStore.showError(t('admin.usage.cleanup.missingRange'))
    return null
  }

  const payload: CreateUsageCleanupTaskRequest = {
    start_date: localStartDate.value,
    end_date: localEndDate.value,
    timezone: getUserTimezone()
  }

  if (localFilters.value.user_id && localFilters.value.user_id > 0) {
    payload.user_id = localFilters.value.user_id
  }
  if (localFilters.value.api_key_id && localFilters.value.api_key_id > 0) {
    payload.api_key_id = localFilters.value.api_key_id
  }
  if (localFilters.value.account_id && localFilters.value.account_id > 0) {
    payload.account_id = localFilters.value.account_id
  }
  if (localFilters.value.group_id && localFilters.value.group_id > 0) {
    payload.group_id = localFilters.value.group_id
  }
  if (localFilters.value.model) {
    payload.model = localFilters.value.model
  }
  if (localFilters.value.request_type) {
    payload.request_type = localFilters.value.request_type
    const legacyStream = requestTypeToLegacyStream(localFilters.value.request_type)
    if (legacyStream !== null && legacyStream !== undefined) {
      payload.stream = legacyStream
    }
  } else if (localFilters.value.stream !== null && localFilters.value.stream !== undefined) {
    payload.stream = localFilters.value.stream
  }
  if (localFilters.value.billing_type !== null && localFilters.value.billing_type !== undefined) {
    payload.billing_type = localFilters.value.billing_type
  }

  return payload
}

const submitCleanup = async () => {
  const payload = buildPayload()
  if (!payload) {
    confirmVisible.value = false
    return
  }
  submitting.value = true
  confirmVisible.value = false
  try {
    await adminUsageAPI.createCleanupTask(payload)
    appStore.showSuccess(t('admin.usage.cleanup.submitSuccess'))
    loadTasks()
  } catch (error) {
    console.error('Failed to create cleanup task:', error)
    appStore.showError(t('admin.usage.cleanup.submitFailed'))
  } finally {
    submitting.value = false
  }
}

const cancelTask = async () => {
  const task = cancelTarget.value
  if (!task) {
    cancelConfirmVisible.value = false
    return
  }
  canceling.value = true
  cancelConfirmVisible.value = false
  try {
    await adminUsageAPI.cancelCleanupTask(task.id)
    appStore.showSuccess(t('admin.usage.cleanup.cancelSuccess'))
    loadTasks()
  } catch (error) {
    console.error('Failed to cancel cleanup task:', error)
    appStore.showError(t('admin.usage.cleanup.cancelFailed'))
  } finally {
    canceling.value = false
    cancelTarget.value = null
  }
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      resetFilters()
      loadTasks()
      startPolling()
    } else {
      stopPolling()
    }
  }
)

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.usage-cleanup-dialog__warning {
  border-radius: calc(var(--theme-surface-radius) + 4px);
  padding: var(--theme-auth-callback-feedback-padding);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-warning-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.usage-cleanup-dialog__tasks-card {
  border-radius: calc(var(--theme-surface-radius) + 4px);
  padding: var(--theme-markdown-block-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
}

.usage-cleanup-dialog__tasks-title {
  color: var(--theme-page-text);
}

.usage-cleanup-dialog__task-card {
  border-radius: var(--theme-button-radius);
  padding:
    calc(var(--theme-scheduled-tests-result-card-padding) - 0.25rem)
    var(--theme-scheduled-tests-result-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  color: var(--theme-page-text);
}

.usage-cleanup-dialog__meta,
.usage-cleanup-dialog__subtle {
  color: var(--theme-page-muted);
}

.usage-cleanup-dialog__cancel-button {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 78%, var(--theme-page-text));
}

.usage-cleanup-dialog__cancel-button:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 92%, var(--theme-page-text));
}

.usage-cleanup-dialog__task-error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
