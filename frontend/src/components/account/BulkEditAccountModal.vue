<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.bulkEdit.title')"
    width="wide"
    @close="handleClose"
  >
    <form id="bulk-edit-account-form" class="space-y-5" @submit.prevent="handleSubmit">
      <BulkEditContextNoticeSection
        :count="accountIds.length"
        :platforms="selectedPlatforms"
      />

      <BulkEditOpenAIOptionsSection
        v-model:passthrough-apply-enabled="enableOpenAIPassthrough"
        v-model:passthrough-enabled="openaiPassthroughEnabled"
        v-model:ws-mode-apply-enabled="enableOpenAIWSMode"
        v-model:ws-mode="openaiOAuthResponsesWebSocketV2Mode"
        :show-passthrough="allOpenAIPassthroughCapable"
        :show-ws-mode="allOpenAIOAuth"
        :ws-mode-options="openAIWSModeOptions"
        :ws-mode-concurrency-hint-key="openAIWSModeConcurrencyHintKey"
      />

      <BulkEditBaseUrlSection
        v-model:enabled="enableBaseUrl"
        v-model:base-url="baseUrl"
      />

      <BulkEditApplySection
        id="bulk-edit-model-restriction"
        v-model:enabled="enableModelRestriction"
        label-key="admin.accounts.modelRestriction"
      >
        <ModelRestrictionSection
          v-model:mode="modelRestrictionMode"
          v-model:allowed-models="allowedModels"
          :platform="bulkModelRestrictionPlatform"
          :platforms="selectedPlatforms"
          :mappings="modelMappings"
          :preset-mappings="filteredPresets"
          :mapping-key="getModelMappingKey"
          :disabled="isOpenAIModelRestrictionDisabled"
          :framed="false"
          @add-mapping="addModelMapping"
          @remove-mapping="removeModelMapping"
          @add-preset="addPresetMapping"
          @update-mapping="updateModelMapping"
        />
      </BulkEditApplySection>

      <CustomErrorCodesSection
        v-model:enabled="enableCustomErrorCodes"
        v-model:input-value="customErrorCodeInput"
        :selected-codes="selectedErrorCodes"
        @toggle-code="toggleErrorCode"
        @add-code="addCustomErrorCode"
        @remove-code="removeErrorCode"
      />

      <BulkEditApplySection
        id="bulk-edit-intercept-warmup"
        v-model:enabled="enableInterceptWarmup"
        label-key="admin.accounts.interceptWarmupRequests"
        hint-key="admin.accounts.interceptWarmupRequestsDesc"
      >
        <WarmupSection
          v-model:enabled="interceptWarmupRequests"
          :framed="false"
        />
      </BulkEditApplySection>

      <BulkEditProxySection
        v-model:enabled="enableProxy"
        v-model:proxy-id="proxyId"
        :proxies="proxies"
      />

      <BulkEditScalarFieldsSection
        v-model:enable-concurrency="enableConcurrency"
        v-model:enable-load-factor="enableLoadFactor"
        v-model:enable-priority="enablePriority"
        v-model:enable-rate-multiplier="enableRateMultiplier"
        v-model:concurrency="concurrency"
        v-model:load-factor="loadFactor"
        v-model:priority="priority"
        v-model:rate-multiplier="rateMultiplier"
      />

      <BulkEditStatusSection
        v-model:enabled="enableStatus"
        v-model:status="status"
        :status-options="statusOptions"
      />

      <BulkEditApplySection
        v-if="allAnthropicOAuthOrSetupToken"
        id="bulk-edit-rpm-limit"
        v-model:enabled="enableRpmLimit"
        label-key="admin.accounts.quotaControl.rpmLimit.label"
      >
        <RpmLimitControlSection
          v-model:enabled="rpmLimitEnabled"
          v-model:base-rpm="bulkBaseRpm"
          v-model:strategy="bulkRpmStrategy"
          v-model:sticky-buffer="bulkRpmStickyBuffer"
          :user-msg-queue-mode="userMsgQueueMode || ''"
          :user-msg-queue-mode-options="umqModeOptions"
          @update:user-msg-queue-mode="toggleUserMsgQueueMode"
        />
      </BulkEditApplySection>

      <BulkEditGroupsSection
        v-model:enabled="enableGroups"
        v-model:group-ids="groupIds"
        :groups="groups"
      />
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="bulk-edit-account-form"
          :disabled="submitting"
          class="btn btn-primary"
        >
          <svg
            v-if="submitting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
          {{
            submitting ? t('admin.accounts.bulkEdit.updating') : t('admin.accounts.bulkEdit.submit')
          }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <ConfirmDialog
    :show="showMixedChannelWarning"
    :title="t('admin.accounts.mixedChannelWarningTitle')"
    :message="mixedChannelWarningMessage"
    :confirm-text="t('common.confirm')"
    :cancel-text="t('common.cancel')"
    :danger="true"
    @confirm="handleMixedChannelConfirm"
    @cancel="handleMixedChannelCancel"
  />
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Proxy as ProxyConfig, AdminGroup, AccountPlatform, AccountType } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BulkEditApplySection from '@/components/account/BulkEditApplySection.vue'
import BulkEditBaseUrlSection from '@/components/account/BulkEditBaseUrlSection.vue'
import BulkEditContextNoticeSection from '@/components/account/BulkEditContextNoticeSection.vue'
import BulkEditGroupsSection from '@/components/account/BulkEditGroupsSection.vue'
import BulkEditOpenAIOptionsSection from '@/components/account/BulkEditOpenAIOptionsSection.vue'
import BulkEditProxySection from '@/components/account/BulkEditProxySection.vue'
import BulkEditScalarFieldsSection from '@/components/account/BulkEditScalarFieldsSection.vue'
import BulkEditStatusSection from '@/components/account/BulkEditStatusSection.vue'
import CustomErrorCodesSection from '@/components/account/CustomErrorCodesSection.vue'
import ModelRestrictionSection from '@/components/account/ModelRestrictionSection.vue'
import RpmLimitControlSection from '@/components/account/RpmLimitControlSection.vue'
import WarmupSection from '@/components/account/WarmupSection.vue'
import {
  buildAccountOpenAIWSModeOptions,
  buildAccountUmqModeOptions,
  needsMixedChannelCheck
} from '@/components/account/accountModalShared'
import { buildBulkAccountMutationPayload } from '@/components/account/accountMutationPayload'
import {
  ensureModelCatalogLoaded,
  getPresetMappingsByPlatform
} from '@/composables/useModelWhitelist'
import {
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey
} from '@/utils/openaiWsMode'
import type { OpenAIWSMode } from '@/utils/openaiWsMode'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
interface Props {
  show: boolean
  accountIds: number[]
  selectedPlatforms: AccountPlatform[]
  selectedTypes: AccountType[]
  proxies: ProxyConfig[]
  groups: AdminGroup[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  updated: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const allOpenAIPassthroughCapable = computed(() => {
  return (
    props.selectedPlatforms.length === 1 &&
    props.selectedPlatforms[0] === 'openai' &&
    props.selectedTypes.length > 0 &&
    props.selectedTypes.every(t => t === 'oauth' || t === 'apikey')
  )
})

const allOpenAIOAuth = computed(() => {
  return (
    props.selectedPlatforms.length === 1 &&
    props.selectedPlatforms[0] === 'openai' &&
    props.selectedTypes.length > 0 &&
    props.selectedTypes.every(t => t === 'oauth')
  )
})

// 是否全部为 Anthropic OAuth/SetupToken（RPM 配置仅在此条件下显示）
const allAnthropicOAuthOrSetupToken = computed(() => {
  return (
    props.selectedPlatforms.length === 1 &&
    props.selectedPlatforms[0] === 'anthropic' &&
    props.selectedTypes.every(t => t === 'oauth' || t === 'setup-token')
  )
})

watch(
  () => props.selectedPlatforms,
  (platforms) => {
    for (const platform of platforms) {
      if (platform === 'grok') {
        void ensureModelCatalogLoaded(platform)
      }
    }
  },
  { immediate: true }
)

const filteredPresets = computed(() => {
  if (props.selectedPlatforms.length === 0) return []

  const dedupedPresets = new Map<string, ReturnType<typeof getPresetMappingsByPlatform>[number]>()
  for (const platform of props.selectedPlatforms) {
    for (const preset of getPresetMappingsByPlatform(platform)) {
      const key = `${preset.from}=>${preset.to}`
      if (!dedupedPresets.has(key)) {
        dedupedPresets.set(key, preset)
      }
    }
  }

  return Array.from(dedupedPresets.values())
})

// Model mapping type
interface ModelMapping {
  from: string
  to: string
}

type BulkEditStatus = 'active' | 'inactive'

// State - field enable flags
const enableBaseUrl = ref(false)
const enableModelRestriction = ref(false)
const enableCustomErrorCodes = ref(false)
const enableInterceptWarmup = ref(false)
const enableProxy = ref(false)
const enableConcurrency = ref(false)
const enableLoadFactor = ref(false)
const enablePriority = ref(false)
const enableRateMultiplier = ref(false)
const enableStatus = ref(false)
const enableGroups = ref(false)
const enableOpenAIPassthrough = ref(false)
const enableOpenAIWSMode = ref(false)
const enableRpmLimit = ref(false)

// State - field values
const submitting = ref(false)
const showMixedChannelWarning = ref(false)
const mixedChannelWarningMessage = ref('')
const pendingUpdatesForConfirm = ref<Record<string, unknown> | null>(null)
const baseUrl = ref('')
const modelRestrictionMode = ref<'whitelist' | 'mapping'>('whitelist')
const allowedModels = ref<string[]>([])
const modelMappings = ref<ModelMapping[]>([])
const selectedErrorCodes = ref<number[]>([])
const customErrorCodeInput = ref<number | null>(null)
const interceptWarmupRequests = ref(false)
const proxyId = ref<number | null>(null)
const concurrency = ref(1)
const loadFactor = ref<number | null>(null)
const priority = ref(1)
const rateMultiplier = ref(1)
const status = ref<BulkEditStatus>('active')
const groupIds = ref<number[]>([])
const openaiPassthroughEnabled = ref(false)
const openaiOAuthResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF)
const rpmLimitEnabled = ref(false)
const bulkBaseRpm = ref<number | null>(null)
const bulkRpmStrategy = ref<'tiered' | 'sticky_exempt'>('tiered')
const bulkRpmStickyBuffer = ref<number | null>(null)
const userMsgQueueMode = ref<string | null>(null)
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t))
let bulkEditRequestSequence = 0

const statusOptions = computed<Array<{ value: BulkEditStatus; label: string }>>(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])
const isOpenAIModelRestrictionDisabled = computed(
  () =>
    allOpenAIPassthroughCapable.value &&
    enableOpenAIPassthrough.value &&
    openaiPassthroughEnabled.value
)
const bulkModelRestrictionPlatform = computed(
  () => props.selectedPlatforms[0] || 'openai'
)

const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t))
const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiOAuthResponsesWebSocketV2Mode.value)
)

const appendEmptyModelMapping = (target: ModelMapping[]) => {
  target.push({ from: '', to: '' })
}

const removeModelMappingAt = (target: ModelMapping[], index: number) => {
  target.splice(index, 1)
}

const appendPresetModelMapping = (target: ModelMapping[], from: string, to: string) => {
  if (target.some((mapping) => mapping.from === from)) {
    appStore.showInfo(t('admin.accounts.mappingExists', { model: from }))
    return
  }
  target.push({ from, to })
}

const getModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>('bulk-edit-model-mapping')

const confirmCustomErrorCodeSelection = (code: number) => {
  if (code === 429) {
    return confirm(t('admin.accounts.customErrorCodes429Warning'))
  }
  if (code === 529) {
    return confirm(t('admin.accounts.customErrorCodes529Warning'))
  }
  return true
}

const clearMixedChannelState = () => {
  showMixedChannelWarning.value = false
  mixedChannelWarningMessage.value = ''
  pendingUpdatesForConfirm.value = null
  mixedChannelConfirmed.value = false
}

const invalidateBulkEditRequests = () => {
  bulkEditRequestSequence += 1
  submitting.value = false
  clearMixedChannelState()
}

const isActiveBulkEditRequest = (requestSequence: number) => (
  requestSequence === bulkEditRequestSequence && props.show
)

const resetBulkEditFormState = () => {
  enableBaseUrl.value = false
  enableModelRestriction.value = false
  enableCustomErrorCodes.value = false
  enableInterceptWarmup.value = false
  enableProxy.value = false
  enableConcurrency.value = false
  enableLoadFactor.value = false
  enablePriority.value = false
  enableRateMultiplier.value = false
  enableStatus.value = false
  enableGroups.value = false
  enableOpenAIPassthrough.value = false
  enableOpenAIWSMode.value = false
  enableRpmLimit.value = false

  baseUrl.value = ''
  openaiPassthroughEnabled.value = false
  modelRestrictionMode.value = 'whitelist'
  allowedModels.value = []
  modelMappings.value = []
  selectedErrorCodes.value = []
  customErrorCodeInput.value = null
  interceptWarmupRequests.value = false
  proxyId.value = null
  concurrency.value = 1
  loadFactor.value = null
  priority.value = 1
  rateMultiplier.value = 1
  status.value = 'active'
  groupIds.value = []
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF
  rpmLimitEnabled.value = false
  bulkBaseRpm.value = null
  bulkRpmStrategy.value = 'tiered'
  bulkRpmStickyBuffer.value = null
  userMsgQueueMode.value = null

  clearMixedChannelState()
}

const hasAnyBulkEditFieldEnabled = () =>
  enableBaseUrl.value ||
  enableOpenAIPassthrough.value ||
  enableModelRestriction.value ||
  enableCustomErrorCodes.value ||
  enableInterceptWarmup.value ||
  enableProxy.value ||
  enableConcurrency.value ||
  enableLoadFactor.value ||
  enablePriority.value ||
  enableRateMultiplier.value ||
  enableStatus.value ||
  enableGroups.value ||
  enableOpenAIWSMode.value ||
  enableRpmLimit.value ||
  userMsgQueueMode.value !== null

// Model mapping helpers
const addModelMapping = () => {
  appendEmptyModelMapping(modelMappings.value)
}

const removeModelMapping = (index: number) => {
  removeModelMappingAt(modelMappings.value, index)
}

const updateModelMapping = (
  index: number,
  field: keyof ModelMapping,
  value: string
) => {
  const mapping = modelMappings.value[index]
  if (!mapping) {
    return
  }
  modelMappings.value[index] = {
    ...mapping,
    [field]: value
  }
}

const addPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(modelMappings.value, from, to)
}

const toggleUserMsgQueueMode = (mode: string) => {
  userMsgQueueMode.value = userMsgQueueMode.value === mode ? null : mode
}

// Error code helpers
const toggleErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index === -1) {
    if (!confirmCustomErrorCodeSelection(code)) {
      return
    }
    selectedErrorCodes.value.push(code)
  } else {
    selectedErrorCodes.value.splice(index, 1)
  }
}

const addCustomErrorCode = () => {
  const code = customErrorCodeInput.value
  if (code === null || code < 100 || code > 599) {
    appStore.showError(t('admin.accounts.invalidErrorCode'))
    return
  }
  if (selectedErrorCodes.value.includes(code)) {
    appStore.showInfo(t('admin.accounts.errorCodeExists'))
    return
  }
  if (!confirmCustomErrorCodeSelection(code)) {
    return
  }
  selectedErrorCodes.value.push(code)
  customErrorCodeInput.value = null
}

const removeErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code)
  if (index !== -1) {
    selectedErrorCodes.value.splice(index, 1)
  }
}

const buildUpdatePayload = (): Record<string, unknown> | null => {
  return buildBulkAccountMutationPayload({
    baseUrl: {
      enabled: enableBaseUrl.value,
      value: baseUrl.value
    },
    customErrorCodes: {
      enabled: enableCustomErrorCodes.value,
      selectedErrorCodes: selectedErrorCodes.value
    },
    groups: {
      enabled: enableGroups.value,
      groupIds: groupIds.value
    },
    interceptWarmup: {
      enabled: enableInterceptWarmup.value,
      value: interceptWarmupRequests.value
    },
    loadFactor: {
      enabled: enableLoadFactor.value,
      value: loadFactor.value
    },
    modelRestriction: {
      allowedModels: allowedModels.value,
      disabledByOpenAIPassthrough: isOpenAIModelRestrictionDisabled.value,
      enabled: enableModelRestriction.value,
      mode: modelRestrictionMode.value,
      modelMappings: modelMappings.value
    },
    openAI: {
      passthroughEnabled: enableOpenAIPassthrough.value,
      passthroughValue: openaiPassthroughEnabled.value,
      wsModeEnabled: enableOpenAIWSMode.value,
      wsModeValue: openaiOAuthResponsesWebSocketV2Mode.value
    },
    proxy: {
      enabled: enableProxy.value,
      proxyId: proxyId.value
    },
    rpmLimit: {
      baseRpm: bulkBaseRpm.value,
      enabled: enableRpmLimit.value,
      rpmEnabled: rpmLimitEnabled.value,
      stickyBuffer: bulkRpmStickyBuffer.value,
      strategy: bulkRpmStrategy.value
    },
    scalars: {
      concurrency: concurrency.value,
      enableConcurrency: enableConcurrency.value,
      enablePriority: enablePriority.value,
      enableRateMultiplier: enableRateMultiplier.value,
      enableStatus: enableStatus.value,
      priority: priority.value,
      rateMultiplier: rateMultiplier.value,
      status: status.value
    },
    userMsgQueueMode: userMsgQueueMode.value
  })
}

const mixedChannelConfirmed = ref(false)

// 是否需要预检查：改了分组 + 全是单一的 antigravity 或 anthropic 平台
// 多平台混合的情况由 submitBulkUpdate 的 409 catch 兜底
const canPreCheck = () =>
  enableGroups.value &&
  groupIds.value.length > 0 &&
  props.selectedPlatforms.length === 1 &&
  needsMixedChannelCheck(props.selectedPlatforms[0])

const handleClose = () => {
  invalidateBulkEditRequests()
  emit('close')
}

// 预检查：提交前调接口检测，有风险就弹窗阻止，返回 false 表示需要用户确认
const preCheckMixedChannelRisk = async (
  built: Record<string, unknown>,
  requestSequence: number
): Promise<boolean> => {
  if (!canPreCheck()) return true
  if (mixedChannelConfirmed.value) return true

  try {
    const platform = props.selectedPlatforms[0]
    const selectedGroupIds = [...groupIds.value]
    const result = await adminAPI.accounts.checkMixedChannelRisk({
      platform,
      group_ids: selectedGroupIds
    })
    if (!isActiveBulkEditRequest(requestSequence)) {
      return false
    }
    if (!result.has_risk) return true

    pendingUpdatesForConfirm.value = built
    mixedChannelWarningMessage.value = result.message || t('admin.accounts.bulkEdit.failed')
    showMixedChannelWarning.value = true
    return false
  } catch (error: any) {
    if (!isActiveBulkEditRequest(requestSequence)) {
      return false
    }
    appStore.showError(resolveRequestErrorMessage(error, t('admin.accounts.bulkEdit.failed')))
    return false
  }
}

const handleSubmit = async () => {
  if (props.accountIds.length === 0) {
    appStore.showError(t('admin.accounts.bulkEdit.noSelection'))
    return
  }

  if (!hasAnyBulkEditFieldEnabled()) {
    appStore.showError(t('admin.accounts.bulkEdit.noFieldsSelected'))
    return
  }

  const built = buildUpdatePayload()
  if (!built) {
    appStore.showError(t('admin.accounts.bulkEdit.noFieldsSelected'))
    return
  }

  const requestSequence = ++bulkEditRequestSequence
  const canContinue = await preCheckMixedChannelRisk(built, requestSequence)
  if (!canContinue || !isActiveBulkEditRequest(requestSequence)) return

  await submitBulkUpdate(built, requestSequence, mixedChannelConfirmed.value)
}

const submitBulkUpdate = async (
  baseUpdates: Record<string, unknown>,
  requestSequence: number,
  confirmMixedChannelRisk = mixedChannelConfirmed.value
) => {
  if (!isActiveBulkEditRequest(requestSequence)) {
    return
  }

  // 无论是预检查确认还是 409 兜底确认，只要 mixedChannelConfirmed 为 true 就带上 flag
  const updates = confirmMixedChannelRisk
    ? { ...baseUpdates, confirm_mixed_channel_risk: true }
    : baseUpdates

  submitting.value = true

  try {
    const res = await adminAPI.accounts.bulkUpdate(props.accountIds, updates)
    if (!isActiveBulkEditRequest(requestSequence)) {
      return
    }
    const success = res.success || 0
    const failed = res.failed || 0

    if (success > 0 && failed === 0) {
      appStore.showSuccess(t('admin.accounts.bulkEdit.success', { count: success }))
    } else if (success > 0) {
      appStore.showError(t('admin.accounts.bulkEdit.partialSuccess', { success, failed }))
    } else {
      appStore.showError(t('admin.accounts.bulkEdit.failed'))
    }

    if (success > 0) {
      pendingUpdatesForConfirm.value = null
      emit('updated')
      handleClose()
    }
  } catch (error: any) {
    if (!isActiveBulkEditRequest(requestSequence)) {
      return
    }
    // 兜底：多平台混合场景下，预检查跳过，由后端 409 触发确认框
    if (error.status === 409 && error.error === 'mixed_channel_warning') {
      pendingUpdatesForConfirm.value = baseUpdates
      mixedChannelWarningMessage.value = error.message
      showMixedChannelWarning.value = true
    } else {
      appStore.showError(resolveRequestErrorMessage(error, t('admin.accounts.bulkEdit.failed')))
      console.error('Error bulk updating accounts:', error)
    }
  } finally {
    if (requestSequence === bulkEditRequestSequence) {
      submitting.value = false
    }
  }
}

const handleMixedChannelConfirm = async () => {
  if (!props.show || !pendingUpdatesForConfirm.value) {
    return
  }

  const requestSequence = ++bulkEditRequestSequence
  showMixedChannelWarning.value = false
  mixedChannelConfirmed.value = true
  await submitBulkUpdate(pendingUpdatesForConfirm.value, requestSequence, true)
}

const handleMixedChannelCancel = () => {
  clearMixedChannelState()
}

// Reset form when modal closes
watch(
  () => props.show,
  (newShow) => {
    if (!newShow) {
      invalidateBulkEditRequests()
      resetBulkEditFormState()
    }
  }
)

watch(
  () => [
    props.accountIds.join(','),
    props.selectedPlatforms.join(','),
    props.selectedTypes.join(',')
  ] as const,
  () => {
    if (!props.show) {
      return
    }
    invalidateBulkEditRequests()
    resetBulkEditFormState()
  }
)
</script>
