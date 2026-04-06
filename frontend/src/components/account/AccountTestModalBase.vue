<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.testAccountConnection')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <div
        v-if="account"
        class="account-test-modal__account-card flex items-center justify-between gap-4 max-sm:flex-col max-sm:items-start"
      >
        <div class="flex items-center gap-3">
          <div class="account-test-modal__account-icon">
            <Icon name="play" size="md" class="account-test-modal__account-icon-symbol" :stroke-width="2" />
          </div>
          <div>
            <div class="account-test-modal__account-name">{{ account.name }}</div>
            <div class="account-test-modal__account-meta">
              <span class="theme-chip theme-chip--compact theme-chip--neutral uppercase">
                {{ account.type }}
              </span>
              <span>{{ t('admin.accounts.account') }}</span>
            </div>
          </div>
        </div>
        <span :class="getAccountStatusClasses(account.status)">
          {{ account.status }}
        </span>
      </div>

      <div v-if="!isSoraAccount" class="space-y-1.5">
        <label class="input-label">{{ t('admin.accounts.selectTestModel') }}</label>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="loadingModels || status === 'connecting'"
          value-key="id"
          label-key="display_name"
          :placeholder="loadingModels ? `${t('common.loading')}...` : t('admin.accounts.selectTestModel')"
        />
      </div>
      <div v-else class="account-test-modal__hint">
        {{ t('admin.accounts.soraTestHint') }}
      </div>

      <div v-if="supportsGeminiImageTest" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.geminiImagePromptLabel')"
          :placeholder="t('admin.accounts.geminiImagePromptPlaceholder')"
          :hint="t('admin.accounts.geminiImageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />
      </div>

      <div v-else-if="showCustomPromptComposer" class="space-y-1.5">
        <TextArea
          v-model="testPrompt"
          :label="t('admin.accounts.customPromptLabel')"
          :placeholder="t('admin.accounts.customPromptPlaceholder')"
          :hint="t('admin.accounts.customPromptHint')"
          :disabled="status === 'connecting'"
          rows="2"
        />
      </div>

      <div class="group relative">
        <div ref="terminalRef" class="account-test-modal__terminal">
          <div v-if="status === 'idle'" class="account-test-modal__terminal-state account-test-modal__terminal-state--idle">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.readyToTest') }}</span>
          </div>
          <div
            v-else-if="status === 'connecting'"
            class="account-test-modal__terminal-state account-test-modal__terminal-state--connecting"
          >
            <Icon name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            <span>{{ t('admin.accounts.connectingToApi') }}</span>
          </div>

          <div
            v-for="(line, index) in outputLines"
            :key="index"
            :class="line.className"
          >
            {{ line.text }}
          </div>

          <div v-if="streamingContent" :class="getLogLineClasses('streaming')">
            {{ streamingContent }}<span class="animate-pulse">_</span>
          </div>

          <div v-if="status === 'success'" class="account-test-modal__result account-test-modal__result--success">
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('admin.accounts.testCompleted') }}</span>
          </div>
          <div v-else-if="status === 'error'" class="account-test-modal__result account-test-modal__result--error">
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <button
          v-if="showCopyButton"
          type="button"
          class="account-test-modal__copy-button"
          :title="t('admin.accounts.copyOutput')"
          @click="copyOutput"
        >
          <Icon name="link" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="account-test-modal__section-label">
          {{ t('admin.accounts.geminiImagePreview') }}
        </div>
        <div class="grid gap-3 sm:grid-cols-2">
          <a
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            :href="image.url"
            target="_blank"
            rel="noopener noreferrer"
            class="account-test-modal__preview-card"
          >
            <img :src="image.url" :alt="`gemini-test-image-${index + 1}`" class="h-48 w-full object-cover" />
            <div class="account-test-modal__preview-meta">
              {{ image.mimeType || 'image/*' }}
            </div>
          </a>
        </div>
      </div>

      <div class="account-test-modal__test-meta">
        <div class="flex items-center gap-3">
          <span class="account-test-modal__meta-item">
            <Icon name="grid" size="sm" :stroke-width="2" />
            {{ testTargetLabel }}
          </span>
        </div>
        <span class="account-test-modal__meta-item">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{ testModeLabel }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          class="btn btn-secondary account-test-modal__footer-button"
          :disabled="status === 'connecting'"
          @click="handleClose"
        >
          {{ t('common.close') }}
        </button>
        <button
          type="button"
          :disabled="isPrimaryActionDisabled"
          :class="getPrimaryActionClasses()"
          @click="startTest"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>{{ primaryActionLabel }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import TextArea from '@/components/common/TextArea.vue'
import { Icon } from '@/components/icons'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'

type AccountTestStatus = 'idle' | 'connecting' | 'success' | 'error'
type LogTone = 'default' | 'muted' | 'info' | 'success' | 'danger' | 'accent' | 'warning' | 'highlight' | 'streaming'

interface OutputLine {
  text: string
  className: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

const prioritizedGeminiModels = [
  'gemini-3.1-flash-image',
  'gemini-2.5-flash-image',
  'gemini-2.5-flash',
  'gemini-2.5-pro',
  'gemini-3-flash-preview',
  'gemini-3-pro-preview',
  'gemini-2.0-flash'
]

const props = withDefaults(
  defineProps<{
    show: boolean
    account: Account | null
    allowCustomPrompt?: boolean
  }>(),
  {
    allowCustomPrompt: false
  }
)

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const terminalRef = ref<HTMLElement | null>(null)
const status = ref<AccountTestStatus>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
const availableModels = ref<ClaudeModel[]>([])
const selectedModelId = ref('')
const testPrompt = ref('')
const loadingModels = ref(false)
const generatedImages = ref<PreviewImage[]>([])

let activeRequestController: AbortController | null = null

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const isSoraAccount = computed(() => props.account?.platform === 'sora')
const supportsGeminiImageTest = computed(() => {
  if (isSoraAccount.value) return false

  const modelId = selectedModelId.value.toLowerCase()
  if (!modelId.startsWith('gemini-') || !modelId.includes('-image')) return false

  return props.account?.platform === 'gemini' || (props.account?.platform === 'antigravity' && props.account?.type === 'apikey')
})
const showCustomPromptComposer = computed(() => props.allowCustomPrompt && !isSoraAccount.value && !supportsGeminiImageTest.value)
const requestPrompt = computed(() => {
  if (isSoraAccount.value) return ''
  if (supportsGeminiImageTest.value || props.allowCustomPrompt) return testPrompt.value.trim()
  return ''
})
const showCopyButton = computed(() => {
  return outputLines.value.length > 0 || Boolean(streamingContent.value) || generatedImages.value.length > 0
})
const isPrimaryActionDisabled = computed(() => {
  return status.value === 'connecting' || (!isSoraAccount.value && !selectedModelId.value)
})
const primaryActionLabel = computed(() => {
  if (status.value === 'connecting') return t('admin.accounts.testing')
  if (status.value === 'idle') return t('admin.accounts.startTest')
  return t('admin.accounts.retry')
})
const testTargetLabel = computed(() => {
  return isSoraAccount.value ? t('admin.accounts.soraTestTarget') : t('admin.accounts.testModel')
})
const testModeLabel = computed(() => {
  if (isSoraAccount.value) return t('admin.accounts.soraTestMode')
  if (supportsGeminiImageTest.value) return t('admin.accounts.geminiImageTestMode')
  if (showCustomPromptComposer.value && testPrompt.value.trim()) return t('admin.accounts.customPromptMode')
  return t('admin.accounts.testPrompt')
})

const logToneClassMap: Record<LogTone, string> = {
  default: 'account-test-modal__log-line--default',
  muted: 'account-test-modal__log-line--muted',
  info: 'account-test-modal__log-line--info',
  success: 'account-test-modal__log-line--success',
  danger: 'account-test-modal__log-line--danger',
  accent: 'account-test-modal__log-line--accent',
  warning: 'account-test-modal__log-line--warning',
  highlight: 'account-test-modal__log-line--highlight',
  streaming: 'account-test-modal__log-line--streaming'
}

const getLogLineClasses = (tone: LogTone = 'default') => {
  return joinClassNames('account-test-modal__log-line', logToneClassMap[tone])
}

const getAccountStatusClasses = (accountStatus: string | undefined) => {
  return joinClassNames(
    'theme-chip account-test-modal__status-chip inline-flex items-center text-xs font-semibold capitalize',
    accountStatus === 'active' ? 'theme-chip--success' : 'theme-chip--neutral'
  )
}

const getPrimaryActionClasses = () => {
  return joinClassNames(
    'btn account-test-modal__footer-button',
    status.value === 'success'
      ? 'btn-success'
      : status.value === 'error'
        ? 'btn-warning'
        : 'btn-primary'
  )
}

const sortTestModels = (models: ClaudeModel[]) => {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((left, right) => {
    const leftPriority = priorityMap.get(left.id) ?? Number.MAX_SAFE_INTEGER
    const rightPriority = priorityMap.get(right.id) ?? Number.MAX_SAFE_INTEGER
    if (leftPriority !== rightPriority) return leftPriority - rightPriority
    return 0
  })
}

watch(
  () => props.show,
  async (isVisible) => {
    if (isVisible && props.account) {
      testPrompt.value = ''
      resetState()
      await loadAvailableModels()
      return
    }

    cancelActiveRequest()
  }
)

watch(selectedModelId, () => {
  if (supportsGeminiImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('admin.accounts.geminiImagePromptDefault')
  }
})

onBeforeUnmount(() => {
  cancelActiveRequest()
})

const loadAvailableModels = async () => {
  if (!props.account) return

  if (props.account.platform === 'sora') {
    availableModels.value = []
    selectedModelId.value = ''
    loadingModels.value = false
    return
  }

  loadingModels.value = true
  selectedModelId.value = ''

  try {
    const models = await adminAPI.accounts.getAvailableModels(props.account.id)
    availableModels.value = props.account.platform === 'gemini' || props.account.platform === 'antigravity'
      ? sortTestModels(models)
      : models

    if (availableModels.value.length === 0) return

    if (props.account.platform === 'gemini') {
      selectedModelId.value = availableModels.value[0].id
      return
    }

    const sonnetModel = availableModels.value.find((model) => model.id.includes('sonnet'))
    selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
  } catch (error) {
    console.error('Failed to load available models:', error)
    availableModels.value = []
    selectedModelId.value = ''
  } finally {
    loadingModels.value = false
  }
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
}

const handleClose = () => {
  if (status.value === 'connecting') return

  cancelActiveRequest()
  emit('close')
}

const cancelActiveRequest = () => {
  if (!activeRequestController) return

  activeRequestController.abort()
  activeRequestController = null
}

const addLine = (text: string, tone: LogTone = 'default') => {
  outputLines.value.push({ text, className: getLogLineClasses(tone) })
  scrollToBottom()
}

const scrollToBottom = async () => {
  await nextTick()
  if (!terminalRef.value) return
  terminalRef.value.scrollTop = terminalRef.value.scrollHeight
}

const startTest = async () => {
  if (!props.account || (!isSoraAccount.value && !selectedModelId.value)) return

  resetState()
  cancelActiveRequest()
  activeRequestController = new AbortController()
  status.value = 'connecting'

  addLine(t('admin.accounts.startingTestForAccount', { name: props.account.name }), 'info')
  addLine(t('admin.accounts.testAccountTypeLabel', { type: props.account.type }), 'muted')
  addLine('', 'default')

  try {
    const response = await fetch(`/api/v1/admin/accounts/${props.account.id}/test`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${localStorage.getItem('auth_token')}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(
        isSoraAccount.value
          ? {}
          : {
              model_id: selectedModelId.value,
              prompt: requestPrompt.value
            }
      ),
      signal: activeRequestController.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('No response body')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue

        const jsonText = line.slice(6).trim()
        if (!jsonText) continue

        try {
          handleEvent(JSON.parse(jsonText))
        } catch (error) {
          console.error('Failed to parse SSE event:', error)
        }
      }
    }
  } catch (error) {
    if (error instanceof Error && error.name === 'AbortError') return

    status.value = 'error'
    errorMessage.value = error instanceof Error ? error.message : 'Unknown error'
    addLine(`Error: ${errorMessage.value}`, 'danger')
  } finally {
    activeRequestController = null
  }
}

const handleEvent = (event: {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}) => {
  switch (event.type) {
    case 'test_start':
      addLine(t('admin.accounts.connectedToApi'), 'success')
      if (event.model) {
        addLine(t('admin.accounts.usingModel', { model: event.model }), 'accent')
      }
      addLine(
        isSoraAccount.value
          ? t('admin.accounts.soraTestingFlow')
          : supportsGeminiImageTest.value
            ? t('admin.accounts.sendingGeminiImageRequest')
            : t('admin.accounts.sendingTestMessage'),
        'muted'
      )
      addLine('', 'default')
      addLine(t('admin.accounts.response'), 'warning')
      break

    case 'content':
      if (!event.text) return
      streamingContent.value += event.text
      scrollToBottom()
      break

    case 'image':
      if (!event.image_url) return
      generatedImages.value.push({
        url: event.image_url,
        mimeType: event.mime_type
      })
      addLine(t('admin.accounts.geminiImageReceived', { count: generatedImages.value.length }), 'highlight')
      break

    case 'test_complete':
      if (streamingContent.value) {
        addLine(streamingContent.value, 'streaming')
        streamingContent.value = ''
      }

      if (event.success) {
        status.value = 'success'
      } else {
        status.value = 'error'
        errorMessage.value = event.error || 'Test failed'
      }
      break

    case 'error':
      status.value = 'error'
      errorMessage.value = event.error || 'Unknown error'
      if (streamingContent.value) {
        addLine(streamingContent.value, 'streaming')
        streamingContent.value = ''
      }
      break
  }
}

const copyOutput = () => {
  const lines = outputLines.value.map((line) => line.text)

  if (streamingContent.value) {
    lines.push(streamingContent.value)
  }

  if (generatedImages.value.length > 0) {
    lines.push('')
    lines.push(t('admin.accounts.geminiImagePreview'))
    lines.push(...generatedImages.value.map((image) => image.url))
  }

  copyToClipboard(lines.join('\n'), t('admin.accounts.outputCopied'))
}
</script>

<style scoped>
.account-test-modal__account-card {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--theme-surface-soft) 84%, var(--theme-surface)) 0%,
      color-mix(in srgb, var(--theme-accent-soft) 52%, var(--theme-surface)) 100%
    );
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--theme-surface-contrast) 10%, transparent),
    0 16px 32px color-mix(in srgb, var(--theme-surface-contrast) 6%, transparent);
  padding: 0.9rem 1rem;
}

.account-test-modal__account-icon {
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: linear-gradient(
    135deg,
    var(--theme-accent),
    color-mix(in srgb, var(--theme-accent-strong) 32%, var(--theme-accent) 68%)
  );
  box-shadow: 0 12px 24px color-mix(in srgb, var(--theme-accent) 24%, transparent);
}

.account-test-modal__account-icon-symbol,
.account-test-modal__footer-button {
  color: var(--theme-filled-text);
}

.account-test-modal__account-name {
  color: var(--theme-page-text);
  font-size: 0.98rem;
  font-weight: 700;
}

.account-test-modal__account-meta {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}

.account-test-modal__hint {
  border: 1px solid color-mix(in srgb, rgb(var(--theme-info-rgb)) 32%, var(--theme-card-border));
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 78%, var(--theme-page-text));
  padding: 0.65rem 0.8rem;
  font-size: 0.75rem;
}

.account-test-modal__terminal {
  max-height: 240px;
  min-height: 120px;
  overflow-y: auto;
  --account-test-modal-terminal-fg: color-mix(in srgb, var(--theme-filled-text) 90%, transparent);
  --account-test-modal-terminal-fg-strong: color-mix(in srgb, var(--theme-filled-text) 88%, transparent);
  --account-test-modal-terminal-fg-muted: color-mix(in srgb, var(--theme-filled-text) 58%, transparent);
  --account-test-modal-terminal-fg-subtle: color-mix(in srgb, var(--theme-filled-text) 48%, transparent);
  --account-test-modal-terminal-border: color-mix(
    in srgb,
    var(--theme-card-border) 68%,
    var(--theme-surface-contrast) 32%
  );
  border: 1px solid var(--account-test-modal-terminal-border);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--theme-surface-contrast) 92%, var(--theme-page-bg) 8%) 0%,
      color-mix(in srgb, var(--theme-surface-contrast) 88%, var(--theme-accent) 12%) 100%
    );
  box-shadow:
    inset 0 1px 0 color-mix(in srgb, var(--theme-filled-text) 8%, transparent),
    0 16px 32px color-mix(in srgb, var(--theme-overlay-strong) 32%, transparent);
  color: var(--account-test-modal-terminal-fg);
  font-family: var(--theme-font-mono);
  font-size: 0.88rem;
  line-height: 1.55;
  padding: 1rem;
  scrollbar-width: thin;
  scrollbar-color: var(--theme-scrollbar-thumb) transparent;
}

.account-test-modal__terminal-state,
.account-test-modal__result,
.account-test-modal__meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
}

.account-test-modal__terminal-state--idle {
  color: var(--account-test-modal-terminal-fg-subtle);
}

.account-test-modal__terminal-state--connecting {
  color: rgb(var(--theme-warning-rgb));
}

.account-test-modal__log-line {
  white-space: pre-wrap;
  word-break: break-word;
}

.account-test-modal__log-line--default {
  color: var(--account-test-modal-terminal-fg-strong);
}

.account-test-modal__log-line--muted {
  color: var(--account-test-modal-terminal-fg-muted);
}

.account-test-modal__log-line--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__log-line--success,
.account-test-modal__log-line--streaming {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 68%, var(--theme-filled-text));
}

.account-test-modal__log-line--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__log-line--accent {
  color: color-mix(in srgb, var(--theme-accent) 68%, var(--theme-filled-text));
}

.account-test-modal__log-line--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__log-line--highlight {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__result {
  margin-top: 0.75rem;
  border-top: 1px solid color-mix(in srgb, var(--theme-filled-text) 12%, transparent);
  padding-top: 0.75rem;
}

.account-test-modal__result--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__result--error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 72%, var(--theme-filled-text));
}

.account-test-modal__copy-button {
  position: absolute;
  top: 0.55rem;
  right: 0.55rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid color-mix(in srgb, var(--theme-filled-text) 12%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-contrast) 76%, transparent);
  color: color-mix(in srgb, var(--theme-filled-text) 62%, transparent);
  opacity: 0;
  padding: 0.35rem;
  transition:
    opacity 0.2s ease,
    color 0.2s ease,
    background-color 0.2s ease,
    border-color 0.2s ease;
}

.group:hover .account-test-modal__copy-button,
.account-test-modal__copy-button:focus-visible {
  opacity: 1;
}

.account-test-modal__copy-button:hover,
.account-test-modal__copy-button:focus-visible {
  background: color-mix(in srgb, var(--theme-surface-contrast) 92%, transparent);
  border-color: color-mix(in srgb, var(--theme-filled-text) 22%, transparent);
  color: var(--theme-filled-text);
  outline: none;
}

.account-test-modal__section-label {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  font-weight: 700;
}

.account-test-modal__status-chip {
  padding: calc(var(--theme-button-padding-y) * 0.45) calc(var(--theme-button-padding-x) * 0.6);
  border-radius: 9999px;
}

.account-test-modal__preview-card {
  overflow: hidden;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 4px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease;
}

.account-test-modal__preview-card:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 44%, var(--theme-card-border));
  box-shadow: var(--theme-card-shadow-hover);
  transform: translateY(-1px);
}

.account-test-modal__preview-meta {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  padding: 0.65rem 0.8rem;
}

.account-test-modal__test-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  padding: 0 0.25rem;
}

@media (max-width: 640px) {
  .account-test-modal__test-meta {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
