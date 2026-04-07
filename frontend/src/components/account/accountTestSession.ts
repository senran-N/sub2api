import { computed, onBeforeUnmount, ref, watch, type Ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { Account, ClaudeModel } from '@/types'
import { AUTH_TOKEN_KEY } from '@/utils/authStorage'

type Translate = (key: string, values?: Record<string, unknown>) => string

export type AccountTestStatus = 'idle' | 'connecting' | 'success' | 'error'
export type LogTone =
  | 'default'
  | 'muted'
  | 'info'
  | 'success'
  | 'danger'
  | 'accent'
  | 'warning'
  | 'highlight'
  | 'streaming'

export interface OutputLine {
  text: string
  tone: LogTone
}

export interface PreviewImage {
  url: string
  mimeType?: string
}

interface StreamEvent {
  type: string
  text?: string
  model?: string
  success?: boolean
  error?: string
  image_url?: string
  mime_type?: string
}

interface ParseSSEChunkResult {
  events: StreamEvent[]
  buffer: string
}

interface UseAccountTestSessionOptions {
  show: Ref<boolean>
  account: Ref<Account | null>
  allowCustomPrompt: Ref<boolean>
  t: Translate
  copyToClipboard: (text: string, successMessage: string) => void
  onTerminalUpdate?: () => void | Promise<void>
  fetchImpl?: typeof fetch
  getAuthToken?: () => string | null
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

export function sortTestModels(models: ClaudeModel[]): ClaudeModel[] {
  const priorityMap = new Map(prioritizedGeminiModels.map((id, index) => [id, index]))

  return [...models].sort((left, right) => {
    const leftPriority = priorityMap.get(left.id) ?? Number.MAX_SAFE_INTEGER
    const rightPriority = priorityMap.get(right.id) ?? Number.MAX_SAFE_INTEGER
    if (leftPriority !== rightPriority) return leftPriority - rightPriority
    return 0
  })
}

export function parseSSEDataChunk(
  currentBuffer: string,
  chunk: string,
  onParseError?: (error: unknown) => void
): ParseSSEChunkResult {
  const merged = currentBuffer + chunk
  const lines = merged.split('\n')
  const nextBuffer = lines.pop() || ''
  const events: StreamEvent[] = []

  for (const line of lines) {
    if (!line.startsWith('data: ')) continue

    const jsonText = line.slice(6).trim()
    if (!jsonText) continue

    try {
      events.push(JSON.parse(jsonText) as StreamEvent)
    } catch (error) {
      onParseError?.(error)
    }
  }

  return {
    events,
    buffer: nextBuffer
  }
}

export function useAccountTestSession(options: UseAccountTestSessionOptions) {
  const status = ref<AccountTestStatus>('idle')
  const outputLines = ref<OutputLine[]>([])
  const streamingContent = ref('')
  const errorMessage = ref('')
  const availableModels = ref<ClaudeModel[]>([])
  const selectedModelId = ref('')
  const testPrompt = ref('')
  const loadingModels = ref(false)
  const generatedImages = ref<PreviewImage[]>([])

  const fetchImpl = options.fetchImpl ?? fetch
  const getAuthToken = options.getAuthToken ?? (() => localStorage.getItem(AUTH_TOKEN_KEY))

  let activeRequestController: AbortController | null = null

  const isSoraAccount = computed(() => options.account.value?.platform === 'sora')
  const supportsGeminiImageTest = computed(() => {
    if (isSoraAccount.value) return false

    const modelId = selectedModelId.value.toLowerCase()
    if (!modelId.startsWith('gemini-') || !modelId.includes('-image')) return false

    return (
      options.account.value?.platform === 'gemini' ||
      (options.account.value?.platform === 'antigravity' &&
        options.account.value?.type === 'apikey')
    )
  })
  const showCustomPromptComposer = computed(
    () =>
      options.allowCustomPrompt.value &&
      !isSoraAccount.value &&
      !supportsGeminiImageTest.value
  )
  const requestPrompt = computed(() => {
    if (isSoraAccount.value) return ''
    if (supportsGeminiImageTest.value || options.allowCustomPrompt.value) {
      return testPrompt.value.trim()
    }
    return ''
  })
  const showCopyButton = computed(
    () =>
      outputLines.value.length > 0 ||
      Boolean(streamingContent.value) ||
      generatedImages.value.length > 0
  )
  const isPrimaryActionDisabled = computed(
    () =>
      status.value === 'connecting' ||
      (!isSoraAccount.value && !selectedModelId.value)
  )
  const primaryActionLabel = computed(() => {
    if (status.value === 'connecting') return options.t('admin.accounts.testing')
    if (status.value === 'idle') return options.t('admin.accounts.startTest')
    return options.t('admin.accounts.retry')
  })
  const testTargetLabel = computed(() =>
    isSoraAccount.value
      ? options.t('admin.accounts.soraTestTarget')
      : options.t('admin.accounts.testModel')
  )
  const testModeLabel = computed(() => {
    if (isSoraAccount.value) return options.t('admin.accounts.soraTestMode')
    if (supportsGeminiImageTest.value) {
      return options.t('admin.accounts.geminiImageTestMode')
    }
    if (showCustomPromptComposer.value && testPrompt.value.trim()) {
      return options.t('admin.accounts.customPromptMode')
    }
    return options.t('admin.accounts.testPrompt')
  })

  function notifyTerminalUpdate() {
    if (!options.onTerminalUpdate) return
    void options.onTerminalUpdate()
  }

  function addLine(text: string, tone: LogTone = 'default') {
    outputLines.value.push({ text, tone })
    notifyTerminalUpdate()
  }

  function cancelActiveRequest() {
    if (!activeRequestController) return

    activeRequestController.abort()
    activeRequestController = null
  }

  function resetState() {
    status.value = 'idle'
    outputLines.value = []
    streamingContent.value = ''
    errorMessage.value = ''
    generatedImages.value = []
  }

  async function loadAvailableModels() {
    if (!options.account.value) return

    if (options.account.value.platform === 'sora') {
      availableModels.value = []
      selectedModelId.value = ''
      loadingModels.value = false
      return
    }

    loadingModels.value = true
    selectedModelId.value = ''

    try {
      const models = await adminAPI.accounts.getAvailableModels(options.account.value.id)
      availableModels.value =
        options.account.value.platform === 'gemini' ||
        options.account.value.platform === 'antigravity'
          ? sortTestModels(models)
          : models

      if (availableModels.value.length === 0) return

      if (options.account.value.platform === 'gemini') {
        selectedModelId.value = availableModels.value[0].id
        return
      }

      const sonnetModel = availableModels.value.find((model) =>
        model.id.includes('sonnet')
      )
      selectedModelId.value = sonnetModel?.id || availableModels.value[0].id
    } catch (error) {
      console.error('Failed to load available models:', error)
      availableModels.value = []
      selectedModelId.value = ''
    } finally {
      loadingModels.value = false
    }
  }

  function handleEvent(event: StreamEvent) {
    switch (event.type) {
      case 'test_start':
        addLine(options.t('admin.accounts.connectedToApi'), 'success')
        if (event.model) {
          addLine(options.t('admin.accounts.usingModel', { model: event.model }), 'accent')
        }
        addLine(
          isSoraAccount.value
            ? options.t('admin.accounts.soraTestingFlow')
            : supportsGeminiImageTest.value
              ? options.t('admin.accounts.sendingGeminiImageRequest')
              : options.t('admin.accounts.sendingTestMessage'),
          'muted'
        )
        addLine('', 'default')
        addLine(options.t('admin.accounts.response'), 'warning')
        break

      case 'content':
        if (!event.text) return
        streamingContent.value += event.text
        notifyTerminalUpdate()
        break

      case 'image':
        if (!event.image_url) return
        generatedImages.value.push({
          url: event.image_url,
          mimeType: event.mime_type
        })
        addLine(
          options.t('admin.accounts.geminiImageReceived', {
            count: generatedImages.value.length
          }),
          'highlight'
        )
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

  async function startTest() {
    if (!options.account.value || (!isSoraAccount.value && !selectedModelId.value)) return

    resetState()
    cancelActiveRequest()
    activeRequestController = new AbortController()
    status.value = 'connecting'

    addLine(
      options.t('admin.accounts.startingTestForAccount', {
        name: options.account.value.name
      }),
      'info'
    )
    addLine(
      options.t('admin.accounts.testAccountTypeLabel', {
        type: options.account.value.type
      }),
      'muted'
    )
    addLine('', 'default')

    try {
      const response = await fetchImpl(
        `/api/v1/admin/accounts/${options.account.value.id}/test`,
        {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${getAuthToken()}`,
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
        }
      )

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

        const chunkText = decoder.decode(value, { stream: true })
        const parsed = parseSSEDataChunk(buffer, chunkText, (error) => {
          console.error('Failed to parse SSE event:', error)
        })
        buffer = parsed.buffer
        for (const event of parsed.events) {
          handleEvent(event)
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

  function copyOutput() {
    const lines = outputLines.value.map((line) => line.text)

    if (streamingContent.value) {
      lines.push(streamingContent.value)
    }

    if (generatedImages.value.length > 0) {
      lines.push('')
      lines.push(options.t('admin.accounts.geminiImagePreview'))
      lines.push(...generatedImages.value.map((image) => image.url))
    }

    options.copyToClipboard(lines.join('\n'), options.t('admin.accounts.outputCopied'))
  }

  watch(
    () => options.show.value,
    async (isVisible) => {
      if (isVisible && options.account.value) {
        testPrompt.value = ''
        resetState()
        await loadAvailableModels()
        return
      }

      cancelActiveRequest()
    },
    { immediate: true }
  )

  watch(selectedModelId, () => {
    if (supportsGeminiImageTest.value && !testPrompt.value.trim()) {
      testPrompt.value = options.t('admin.accounts.geminiImagePromptDefault')
    }
  })

  onBeforeUnmount(() => {
    cancelActiveRequest()
  })

  return {
    status,
    outputLines,
    streamingContent,
    errorMessage,
    availableModels,
    selectedModelId,
    testPrompt,
    loadingModels,
    generatedImages,
    isSoraAccount,
    supportsGeminiImageTest,
    showCustomPromptComposer,
    showCopyButton,
    isPrimaryActionDisabled,
    primaryActionLabel,
    testTargetLabel,
    testModeLabel,
    startTest,
    copyOutput,
    cancelActiveRequest
  }
}
