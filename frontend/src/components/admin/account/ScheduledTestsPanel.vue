<template>
  <BaseDialog
    :show="show"
    :title="t('admin.scheduledTests.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="space-y-4">
      <!-- Add Plan Button -->
      <div class="flex items-center justify-between">
        <p class="scheduled-tests-panel__subtitle text-sm">
          {{ t('admin.scheduledTests.title') }}
        </p>
        <button
          @click="showAddForm = !showAddForm"
          class="btn btn-primary flex items-center gap-1.5 text-sm"
        >
          <Icon name="plus" size="sm" :stroke-width="2" />
          {{ t('admin.scheduledTests.addPlan') }}
        </button>
      </div>

      <!-- Add Plan Form -->
      <div
        v-if="showAddForm"
        class="scheduled-tests-panel__form-panel"
      >
        <div class="scheduled-tests-panel__section-label mb-3 text-sm font-medium">
          {{ t('admin.scheduledTests.addPlan') }}
        </div>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label class="scheduled-tests-panel__field-label mb-1 block text-xs font-medium">
              {{ t('admin.scheduledTests.model') }}
            </label>
            <Select
              v-model="newPlan.model_id"
              :options="modelOptions"
              :placeholder="t('admin.scheduledTests.model')"
              :searchable="modelOptions.length > 5"
            />
          </div>
          <div>
            <label class="scheduled-tests-panel__field-label mb-1 flex items-center gap-1 text-xs font-medium">
              {{ t('admin.scheduledTests.cronExpression') }}
              <HelpTooltip>
                <template #trigger>
                  <span class="scheduled-tests-panel__help-trigger inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-[10px] font-semibold transition-colors">
                    ?
                  </span>
                </template>
                <div class="space-y-1.5">
                  <p class="font-medium">{{ t('admin.scheduledTests.cronTooltipTitle') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipMeaning') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleEvery30Min') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleHourly') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleDaily') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipExampleWeekly') }}</p>
                  <p>{{ t('admin.scheduledTests.cronTooltipRange') }}</p>
                </div>
              </HelpTooltip>
            </label>
            <Input
              v-model="newPlan.cron_expression"
              :placeholder="'*/30 * * * *'"
              :hint="t('admin.scheduledTests.cronHelp')"
            />
          </div>
          <div>
            <label class="scheduled-tests-panel__field-label mb-1 flex items-center gap-1 text-xs font-medium">
              {{ t('admin.scheduledTests.maxResults') }}
              <HelpTooltip>
                <template #trigger>
                  <span class="scheduled-tests-panel__help-trigger inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-[10px] font-semibold transition-colors">
                    ?
                  </span>
                </template>
                <div class="space-y-1.5">
                  <p class="font-medium">{{ t('admin.scheduledTests.maxResultsTooltipTitle') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipMeaning') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipBody') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipExample') }}</p>
                  <p>{{ t('admin.scheduledTests.maxResultsTooltipRange') }}</p>
                </div>
              </HelpTooltip>
            </label>
            <Input
              v-model="newPlan.max_results"
              type="number"
              placeholder="100"
            />
          </div>
          <div class="flex items-end">
            <label class="scheduled-tests-panel__toggle-label flex items-center gap-2 text-sm">
              <Toggle v-model="newPlan.enabled" />
              {{ t('admin.scheduledTests.enabled') }}
            </label>
          </div>
          <div class="flex items-end">
            <div>
              <label class="scheduled-tests-panel__toggle-label flex items-center gap-2 text-sm">
                <Toggle v-model="newPlan.auto_recover" />
                {{ t('admin.scheduledTests.autoRecover') }}
              </label>
              <p class="scheduled-tests-panel__help-text mt-0.5 text-xs">
                {{ t('admin.scheduledTests.autoRecoverHelp') }}
              </p>
            </div>
          </div>
        </div>
        <div class="mt-3 flex justify-end gap-2">
          <button
            @click="showAddForm = false; resetNewPlan()"
            class="btn btn-secondary btn-sm scheduled-tests-panel__secondary-button"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            @click="handleCreate"
            :disabled="!newPlan.model_id || !newPlan.cron_expression || creating"
            class="btn btn-primary btn-sm flex items-center gap-1.5"
          >
            <Icon v-if="creating" name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            {{ t('common.save') }}
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <div
        v-if="loading"
        class="scheduled-tests-panel__status-state scheduled-tests-panel__status-state--primary flex items-center justify-center"
      >
        <Icon
          name="refresh"
          size="md"
          class="scheduled-tests-panel__status-icon animate-spin"
          :stroke-width="2"
        />
        <span class="scheduled-tests-panel__status-text ml-2 text-sm">{{ t('common.loading') }}...</span>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="plans.length === 0"
        class="scheduled-tests-panel__empty-state text-center"
      >
        <Icon name="calendar" size="lg" class="scheduled-tests-panel__empty-icon mx-auto mb-2" :stroke-width="1.5" />
        <p class="scheduled-tests-panel__status-text text-sm">
          {{ t('admin.scheduledTests.noPlans') }}
        </p>
      </div>

      <!-- Plans List -->
      <div v-else class="space-y-3">
        <div
          v-for="plan in plans"
          :key="plan.id"
          class="scheduled-tests-panel__plan-card transition-all"
        >
          <!-- Plan Header -->
          <div
            class="scheduled-tests-panel__plan-header flex cursor-pointer items-center justify-between"
            @click="toggleExpand(plan.id)"
          >
            <div class="flex flex-1 items-center gap-4">
              <!-- Model -->
              <div class="min-w-0">
                <div class="scheduled-tests-panel__plan-title text-sm font-medium">
                  {{ plan.model_id }}
                </div>
                <div class="scheduled-tests-panel__plan-meta mt-0.5 font-mono text-xs">
                  {{ plan.cron_expression }}
                </div>
              </div>

              <!-- Enabled Toggle -->
              <div class="flex items-center gap-1.5" @click.stop>
                <Toggle
                  :model-value="plan.enabled"
                  @update:model-value="(val: boolean) => handleToggleEnabled(plan, val)"
                />
                <span class="scheduled-tests-panel__plan-flag text-xs">
                  {{ plan.enabled ? t('admin.scheduledTests.enabled') : '' }}
                </span>
              </div>

              <!-- Auto Recover Badge -->
              <span
                v-if="plan.auto_recover"
                class="theme-chip theme-chip--regular theme-chip--success inline-flex items-center"
              >
                {{ t('admin.scheduledTests.autoRecover') }}
              </span>
            </div>

            <div class="flex items-center gap-3">
              <!-- Last Run -->
              <div v-if="plan.last_run_at" class="scheduled-tests-panel__timestamp hidden text-right text-xs sm:block">
                <div>{{ t('admin.scheduledTests.lastRun') }}</div>
                <div>{{ formatDateTime(plan.last_run_at) }}</div>
              </div>

              <!-- Next Run -->
              <div v-if="plan.next_run_at" class="scheduled-tests-panel__timestamp hidden text-right text-xs sm:block">
                <div>{{ t('admin.scheduledTests.nextRun') }}</div>
                <div>{{ formatDateTime(plan.next_run_at) }}</div>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-1" @click.stop>
                <button
                  @click="startEdit(plan)"
                  :class="getActionButtonClasses('info')"
                  :title="t('admin.scheduledTests.editPlan')"
                >
                  <Icon name="edit" size="sm" :stroke-width="2" />
                </button>
                <button
                  @click="confirmDeletePlan(plan)"
                  :class="getActionButtonClasses('danger')"
                  :title="t('admin.scheduledTests.deletePlan')"
                >
                  <Icon name="trash" size="sm" :stroke-width="2" />
                </button>
              </div>

              <!-- Expand indicator -->
              <Icon
                name="chevronDown"
                size="sm"
                :class="[
                  'scheduled-tests-panel__chevron transition-transform duration-200',
                  expandedPlanId === plan.id ? 'rotate-180' : ''
                ]"
              />
            </div>
          </div>

          <!-- Edit Form -->
          <div
            v-if="editingPlanId === plan.id"
            class="scheduled-tests-panel__editor-panel"
            @click.stop
          >
            <div class="scheduled-tests-panel__section-label mb-2 text-xs font-medium">
              {{ t('admin.scheduledTests.editPlan') }}
            </div>
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div>
                <label class="scheduled-tests-panel__field-label mb-1 block text-xs font-medium">
                  {{ t('admin.scheduledTests.model') }}
                </label>
                <Select
                  v-model="editForm.model_id"
                  :options="modelOptions"
                  :placeholder="t('admin.scheduledTests.model')"
                  :searchable="modelOptions.length > 5"
                />
              </div>
              <div>
                <label class="scheduled-tests-panel__field-label mb-1 flex items-center gap-1 text-xs font-medium">
                  {{ t('admin.scheduledTests.cronExpression') }}
                  <HelpTooltip>
                    <template #trigger>
                      <span class="scheduled-tests-panel__help-trigger inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-[10px] font-semibold transition-colors">
                        ?
                      </span>
                    </template>
                    <div class="space-y-1.5">
                      <p class="font-medium">{{ t('admin.scheduledTests.cronTooltipTitle') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipMeaning') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleEvery30Min') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleHourly') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleDaily') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipExampleWeekly') }}</p>
                      <p>{{ t('admin.scheduledTests.cronTooltipRange') }}</p>
                    </div>
                  </HelpTooltip>
                </label>
                <Input
                  v-model="editForm.cron_expression"
                  :placeholder="'*/30 * * * *'"
                  :hint="t('admin.scheduledTests.cronHelp')"
                />
              </div>
              <div>
                <label class="scheduled-tests-panel__field-label mb-1 flex items-center gap-1 text-xs font-medium">
                  {{ t('admin.scheduledTests.maxResults') }}
                  <HelpTooltip>
                    <template #trigger>
                      <span class="scheduled-tests-panel__help-trigger inline-flex h-4 w-4 cursor-help items-center justify-center rounded-full text-[10px] font-semibold transition-colors">
                        ?
                      </span>
                    </template>
                    <div class="space-y-1.5">
                      <p class="font-medium">{{ t('admin.scheduledTests.maxResultsTooltipTitle') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipMeaning') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipBody') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipExample') }}</p>
                      <p>{{ t('admin.scheduledTests.maxResultsTooltipRange') }}</p>
                    </div>
                  </HelpTooltip>
                </label>
                <Input
                  v-model="editForm.max_results"
                  type="number"
                  placeholder="100"
                />
              </div>
              <div class="flex items-end">
                <label class="scheduled-tests-panel__toggle-label flex items-center gap-2 text-sm">
                  <Toggle v-model="editForm.enabled" />
                  {{ t('admin.scheduledTests.enabled') }}
                </label>
              </div>
              <div class="flex items-end">
                <div>
                  <label class="scheduled-tests-panel__toggle-label flex items-center gap-2 text-sm">
                    <Toggle v-model="editForm.auto_recover" />
                    {{ t('admin.scheduledTests.autoRecover') }}
                  </label>
                  <p class="scheduled-tests-panel__help-text mt-0.5 text-xs">
                    {{ t('admin.scheduledTests.autoRecoverHelp') }}
                  </p>
                </div>
              </div>
            </div>
            <div class="mt-3 flex justify-end gap-2">
              <button
                @click="cancelEdit"
                class="btn btn-secondary btn-sm scheduled-tests-panel__secondary-button"
              >
                {{ t('common.cancel') }}
              </button>
              <button
                @click="handleEdit"
                :disabled="!editForm.model_id || !editForm.cron_expression || updating"
                class="btn btn-primary btn-sm flex items-center gap-1.5"
              >
                <Icon v-if="updating" name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
                {{ t('common.save') }}
              </button>
            </div>
          </div>

          <!-- Expanded Results Section -->
          <div
            v-if="expandedPlanId === plan.id"
            class="scheduled-tests-panel__results-panel"
          >
            <div class="scheduled-tests-panel__section-label mb-2 text-xs font-medium">
              {{ t('admin.scheduledTests.results') }}
            </div>

            <!-- Results Loading -->
            <div
              v-if="loadingResults"
              class="scheduled-tests-panel__status-state scheduled-tests-panel__status-state--results flex items-center justify-center"
            >
              <Icon
                name="refresh"
                size="sm"
                class="scheduled-tests-panel__status-icon animate-spin"
                :stroke-width="2"
              />
              <span class="scheduled-tests-panel__status-text ml-2 text-xs">{{ t('common.loading') }}...</span>
            </div>

            <!-- No Results -->
            <div
              v-else-if="results.length === 0"
              class="scheduled-tests-panel__status-text scheduled-tests-panel__status-text--results-empty text-center text-xs"
            >
              {{ t('admin.scheduledTests.noResults') }}
            </div>

            <!-- Results List -->
            <div v-else class="scheduled-tests-panel__results-list space-y-2 overflow-y-auto">
              <div
                v-for="result in results"
                :key="result.id"
                class="scheduled-tests-panel__result-card"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <!-- Status Badge -->
                    <span :class="getResultStatusClasses(result.status)">
                      {{ getResultStatusLabel(result.status) }}
                    </span>

                    <!-- Latency -->
                    <span v-if="result.latency_ms > 0" class="scheduled-tests-panel__metric text-xs">
                      {{ result.latency_ms }}ms
                    </span>
                  </div>

                  <!-- Started At -->
                  <span class="scheduled-tests-panel__timestamp scheduled-tests-panel__timestamp--soft text-xs">
                    {{ formatDateTime(result.started_at) }}
                  </span>
                </div>

                <!-- Response / Error (collapsible) -->
                <div v-if="result.error_message" class="mt-2">
                  <div
                    :class="getResultDetailToggleClasses('error')"
                    @click="toggleResultDetail(result.id)"
                  >
                    {{ t('admin.scheduledTests.errorMessage') }}
                    <Icon
                      name="chevronDown"
                      size="sm"
                      :class="[
                        'inline transition-transform duration-200',
                        expandedResultIds.has(result.id) ? 'rotate-180' : ''
                      ]"
                    />
                  </div>
                  <pre
                    v-if="expandedResultIds.has(result.id)"
                    class="scheduled-tests-panel__detail-preview scheduled-tests-panel__detail-preview--error mt-1 overflow-auto whitespace-pre-wrap text-xs"
                  >{{ result.error_message }}</pre>
                </div>
                <div v-else-if="result.response_text" class="mt-2">
                  <div
                    :class="getResultDetailToggleClasses('response')"
                    @click="toggleResultDetail(result.id)"
                  >
                    {{ t('admin.scheduledTests.responseText') }}
                    <Icon
                      name="chevronDown"
                      size="sm"
                      :class="[
                        'inline transition-transform duration-200',
                        expandedResultIds.has(result.id) ? 'rotate-180' : ''
                      ]"
                    />
                  </div>
                  <pre
                    v-if="expandedResultIds.has(result.id)"
                    class="scheduled-tests-panel__detail-preview scheduled-tests-panel__detail-preview--response mt-1 overflow-auto whitespace-pre-wrap text-xs"
                  >{{ result.response_text }}</pre>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('admin.scheduledTests.deletePlan')"
      :message="t('admin.scheduledTests.confirmDelete')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import Input from '@/components/common/Input.vue'
import Toggle from '@/components/common/Toggle.vue'
import { Icon } from '@/components/icons'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import type { ScheduledTestPlan, ScheduledTestResult } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  show: boolean
  accountId: number | null
  modelOptions: SelectOption[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

// State
const loading = ref(false)
const creating = ref(false)
const loadingResults = ref(false)
const plans = ref<ScheduledTestPlan[]>([])
const results = ref<ScheduledTestResult[]>([])
const expandedPlanId = ref<number | null>(null)
const expandedResultIds = reactive(new Set<number>())
const showAddForm = ref(false)
const showDeleteConfirm = ref(false)
const deletingPlan = ref<ScheduledTestPlan | null>(null)
const editingPlanId = ref<number | null>(null)
const updating = ref(false)
let plansRequestSequence = 0
let resultsRequestSequence = 0
const editForm = reactive({
  model_id: '' as string,
  cron_expression: '' as string,
  max_results: '100' as string,
  enabled: true,
  auto_recover: false
})

const newPlan = reactive({
  model_id: '' as string,
  cron_expression: '' as string,
  max_results: '100' as string,
  enabled: true,
  auto_recover: false
})

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getActionButtonClasses = (tone: 'info' | 'danger') => {
  return joinClassNames(
    'theme-action-button',
    tone === 'info'
      ? 'theme-action-button--info'
      : 'theme-action-button--danger'
  )
}

const getResultStatusClasses = (status: string) => {
  return joinClassNames(
    'theme-chip theme-chip--regular inline-flex items-center',
    status === 'success'
      ? 'theme-chip--success'
      : status === 'running'
        ? 'theme-chip--info'
        : 'theme-chip--danger'
  )
}

const getResultStatusLabel = (status: string) => {
  if (status === 'success') {
    return t('admin.scheduledTests.success')
  }
  if (status === 'running') {
    return t('admin.scheduledTests.running')
  }
  return t('admin.scheduledTests.failed')
}

const getResultDetailToggleClasses = (tone: 'error' | 'response') => {
  return joinClassNames(
    'scheduled-tests-panel__detail-toggle cursor-pointer text-xs font-medium',
    tone === 'error'
      ? 'scheduled-tests-panel__detail-toggle--error'
      : 'scheduled-tests-panel__detail-toggle--response'
  )
}

const resetNewPlan = () => {
  newPlan.model_id = ''
  newPlan.cron_expression = ''
  newPlan.max_results = '100'
  newPlan.enabled = true
  newPlan.auto_recover = false
}

const invalidatePlansRequest = () => {
  plansRequestSequence += 1
  loading.value = false
}

const invalidateResultsRequest = () => {
  resultsRequestSequence += 1
  loadingResults.value = false
}

const resetResultsState = () => {
  invalidateResultsRequest()
  results.value = []
  expandedPlanId.value = null
  expandedResultIds.clear()
}

const resetPanelState = () => {
  invalidatePlansRequest()
  resetResultsState()
  plans.value = []
  showAddForm.value = false
  showDeleteConfirm.value = false
  deletingPlan.value = null
  editingPlanId.value = null
}

// Load plans when dialog opens
watch(
  [() => props.show, () => props.accountId],
  async ([visible, accountId]) => {
    if (visible && accountId) {
      resetResultsState()
      await loadPlans(accountId)
    } else {
      resetPanelState()
    }
  },
  { immediate: true }
)

async function loadPlans(accountId = props.accountId) {
  if (!accountId) return
  const requestSequence = ++plansRequestSequence
  loading.value = true
  try {
    const nextPlans = await adminAPI.scheduledTests.listByAccount(accountId)
    if (requestSequence !== plansRequestSequence || !props.show || props.accountId !== accountId) {
      return
    }
    plans.value = nextPlans
  } catch (error: any) {
    if (requestSequence !== plansRequestSequence || !props.show || props.accountId !== accountId) {
      return
    }
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to load plans'))
  } finally {
    if (requestSequence === plansRequestSequence) {
      loading.value = false
    }
  }
}

const handleCreate = async () => {
  if (!props.accountId || !newPlan.model_id || !newPlan.cron_expression) return
  creating.value = true
  try {
    const maxResults = Number(newPlan.max_results) || 100
    await adminAPI.scheduledTests.create({
      account_id: props.accountId,
      model_id: newPlan.model_id,
      cron_expression: newPlan.cron_expression,
      enabled: newPlan.enabled,
      max_results: maxResults,
      auto_recover: newPlan.auto_recover
    })
    appStore.showSuccess(t('admin.scheduledTests.createSuccess'))
    showAddForm.value = false
    resetNewPlan()
    await loadPlans()
  } catch (error: any) {
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to create plan'))
  } finally {
    creating.value = false
  }
}

const handleToggleEnabled = async (plan: ScheduledTestPlan, enabled: boolean) => {
  try {
    const updated = await adminAPI.scheduledTests.update(plan.id, { enabled })
    const index = plans.value.findIndex((p) => p.id === plan.id)
    if (index !== -1) {
      plans.value[index] = updated
    }
    appStore.showSuccess(t('admin.scheduledTests.updateSuccess'))
  } catch (error: any) {
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to update plan'))
  }
}

const startEdit = (plan: ScheduledTestPlan) => {
  editingPlanId.value = plan.id
  editForm.model_id = plan.model_id
  editForm.cron_expression = plan.cron_expression
  editForm.max_results = String(plan.max_results)
  editForm.enabled = plan.enabled
  editForm.auto_recover = plan.auto_recover
}

const cancelEdit = () => {
  editingPlanId.value = null
}

const handleEdit = async () => {
  if (!editingPlanId.value || !editForm.model_id || !editForm.cron_expression) return
  updating.value = true
  try {
    const updated = await adminAPI.scheduledTests.update(editingPlanId.value, {
      model_id: editForm.model_id,
      cron_expression: editForm.cron_expression,
      max_results: Number(editForm.max_results) || 100,
      enabled: editForm.enabled,
      auto_recover: editForm.auto_recover
    })
    const index = plans.value.findIndex((p) => p.id === editingPlanId.value)
    if (index !== -1) {
      plans.value[index] = updated
    }
    appStore.showSuccess(t('admin.scheduledTests.updateSuccess'))
    editingPlanId.value = null
  } catch (error: any) {
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to update plan'))
  } finally {
    updating.value = false
  }
}

const confirmDeletePlan = (plan: ScheduledTestPlan) => {
  deletingPlan.value = plan
  showDeleteConfirm.value = true
}

const handleDelete = async () => {
  if (!deletingPlan.value) return
  try {
    await adminAPI.scheduledTests.delete(deletingPlan.value.id)
    appStore.showSuccess(t('admin.scheduledTests.deleteSuccess'))
    plans.value = plans.value.filter((p) => p.id !== deletingPlan.value!.id)
    if (expandedPlanId.value === deletingPlan.value.id) {
      expandedPlanId.value = null
      results.value = []
    }
  } catch (error: any) {
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to delete plan'))
  } finally {
    showDeleteConfirm.value = false
    deletingPlan.value = null
  }
}

const toggleExpand = async (planId: number) => {
  if (expandedPlanId.value === planId) {
    resetResultsState()
    return
  }

  expandedPlanId.value = planId
  const requestSequence = ++resultsRequestSequence
  expandedResultIds.clear()
  results.value = []
  loadingResults.value = true
  try {
    const nextResults = await adminAPI.scheduledTests.listResults(planId, 20)
    if (
      requestSequence !== resultsRequestSequence ||
      !props.show ||
      expandedPlanId.value !== planId
    ) {
      return
    }
    results.value = nextResults
  } catch (error: any) {
    if (
      requestSequence !== resultsRequestSequence ||
      !props.show ||
      expandedPlanId.value !== planId
    ) {
      return
    }
    appStore.showError(resolveRequestErrorMessage(error, 'Failed to load results'))
    results.value = []
  } finally {
    if (requestSequence === resultsRequestSequence) {
      loadingResults.value = false
    }
  }
}

const toggleResultDetail = (resultId: number) => {
  if (expandedResultIds.has(resultId)) {
    expandedResultIds.delete(resultId)
  } else {
    expandedResultIds.add(resultId)
  }
}
</script>

<style scoped>
.scheduled-tests-panel__subtitle,
.scheduled-tests-panel__plan-flag,
.scheduled-tests-panel__plan-meta,
.scheduled-tests-panel__status-text,
.scheduled-tests-panel__metric,
.scheduled-tests-panel__timestamp,
.scheduled-tests-panel__help-text {
  color: var(--theme-page-muted);
}

.scheduled-tests-panel__section-label,
.scheduled-tests-panel__field-label {
  color: color-mix(in srgb, var(--theme-page-text) 76%, var(--theme-page-muted));
}

.scheduled-tests-panel__toggle-label,
.scheduled-tests-panel__plan-title {
  color: var(--theme-page-text);
}

.scheduled-tests-panel__form-panel {
  padding: var(--theme-scheduled-tests-panel-padding);
  border: 1px solid color-mix(in srgb, var(--theme-accent) 22%, var(--theme-card-border));
  border-radius: var(--theme-surface-radius);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-accent-soft) 92%, var(--theme-surface)) 0%,
      color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface)) 100%
    );
  box-shadow: var(--theme-card-shadow);
}

.scheduled-tests-panel__help-trigger {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 78%, transparent);
  color: color-mix(in srgb, var(--theme-page-muted) 82%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
}

.scheduled-tests-panel__help-trigger:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 34%, var(--theme-card-border));
  color: color-mix(in srgb, var(--theme-accent) 86%, var(--theme-page-text));
}

.scheduled-tests-panel__secondary-button {
  border-color: var(--theme-button-secondary-border);
  background: var(--theme-button-secondary-bg);
  color: var(--theme-button-secondary-text);
  box-shadow: var(--theme-card-shadow);
}

.scheduled-tests-panel__secondary-button:hover {
  background: var(--theme-button-secondary-hover-bg);
  box-shadow: var(--theme-card-shadow-hover);
}

.scheduled-tests-panel__status-state {
  color: var(--theme-page-muted);
}

.scheduled-tests-panel__status-state--primary {
  padding-block: calc(var(--theme-scheduled-tests-panel-padding) * 2);
}

.scheduled-tests-panel__status-state--results {
  padding-block: var(--theme-scheduled-tests-panel-padding);
}

.scheduled-tests-panel__status-icon,
.scheduled-tests-panel__empty-icon,
.scheduled-tests-panel__chevron {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
}

.scheduled-tests-panel__empty-state {
  padding-block: calc(var(--theme-scheduled-tests-panel-padding) * 2.5);
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  border-radius: var(--theme-surface-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 74%, var(--theme-surface));
}

.scheduled-tests-panel__status-text--results-empty {
  display: block;
  padding-block: var(--theme-scheduled-tests-panel-padding);
}

.scheduled-tests-panel__plan-card {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.scheduled-tests-panel__plan-card:hover {
  box-shadow: var(--theme-card-shadow-hover);
}

.scheduled-tests-panel__plan-header {
  padding:
    calc(var(--theme-scheduled-tests-panel-padding) * 0.75)
    var(--theme-scheduled-tests-panel-padding);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--theme-surface) 94%, transparent),
    color-mix(in srgb, var(--theme-surface-soft) 44%, var(--theme-surface))
  );
}

.scheduled-tests-panel__editor-panel,
.scheduled-tests-panel__results-panel {
  padding:
    calc(var(--theme-scheduled-tests-panel-padding) * 0.75)
    var(--theme-scheduled-tests-panel-padding);
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.scheduled-tests-panel__editor-panel {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, rgb(var(--theme-info-rgb)) 7%, var(--theme-surface)) 0%,
    color-mix(in srgb, var(--theme-surface-soft) 76%, var(--theme-surface)) 100%
  );
}

.scheduled-tests-panel__results-panel {
  background: color-mix(in srgb, var(--theme-surface-soft) 62%, var(--theme-surface));
}

.scheduled-tests-panel__results-list {
  max-height: var(--theme-scheduled-tests-results-max-height);
}

.scheduled-tests-panel__result-card {
  padding: var(--theme-scheduled-tests-result-card-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-button-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 78%, var(--theme-surface));
}

.scheduled-tests-panel__timestamp--soft {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.scheduled-tests-panel__detail-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.125rem;
}

.scheduled-tests-panel__detail-toggle--error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.scheduled-tests-panel__detail-toggle--response {
  color: color-mix(in srgb, var(--theme-page-text) 70%, var(--theme-page-muted));
}

.scheduled-tests-panel__detail-preview {
  max-height: var(--theme-scheduled-tests-detail-max-height);
  padding: calc(var(--theme-scheduled-tests-result-card-padding) * 0.67);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  border-radius: var(--theme-button-radius);
}

.scheduled-tests-panel__detail-preview--error {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.scheduled-tests-panel__detail-preview--response {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-page-text) 82%, var(--theme-page-muted));
}
</style>
