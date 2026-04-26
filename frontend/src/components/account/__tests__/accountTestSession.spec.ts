import { describe, expect, it, vi, beforeEach } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, ref } from 'vue'
import type { Account, ClaudeModel } from '@/types'

const { getAvailableModels } = vi.hoisted(() => ({
  getAvailableModels: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getAvailableModels
    }
  }
}))

import {
  parseSSEDataChunk,
  sortTestModels,
  useAccountTestSession
} from '../accountTestSession'

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: 'Account Alpha',
    platform: 'openai',
    type: 'oauth',
    proxy_id: null,
    concurrency: 1,
    priority: 1,
    status: 'active',
    error_message: null,
    last_used_at: null,
    expires_at: null,
    auto_pause_on_expired: true,
    created_at: '2026-04-01T00:00:00Z',
    updated_at: '2026-04-01T00:00:00Z',
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null,
    ...overrides
  }
}

function createModel(id: string, displayName = id): ClaudeModel {
  return {
    id,
    display_name: displayName
  } as ClaudeModel
}

function mountSessionHarness(initialAccount = createAccount()) {
  return mount(defineComponent({
    setup(_, { expose }) {
      const show = ref(true)
      const account = ref<Account | null>(initialAccount)
      const allowCustomPrompt = ref(false)
      const session = useAccountTestSession({
        show,
        account,
        allowCustomPrompt,
        t: (key) => key,
        copyToClipboard: vi.fn()
      })

      expose({
        show,
        account,
        ...session
      })

      return () => null
    }
  }))
}

describe('accountTestSession helpers', () => {
  beforeEach(() => {
    getAvailableModels.mockReset()
  })

  it('sorts gemini models by predefined priority', () => {
    const sorted = sortTestModels([
      { id: 'gemini-2.0-flash', display_name: 'Gemini 2.0 Flash' } as ClaudeModel,
      { id: 'gemini-3.1-flash-image', display_name: 'Gemini 3.1 Flash Image' } as ClaudeModel,
      { id: 'gemini-2.5-pro', display_name: 'Gemini 2.5 Pro' } as ClaudeModel,
      { id: 'custom-model', display_name: 'Custom Model' } as ClaudeModel
    ])

    expect(sorted.map((model) => model.id)).toEqual([
      'gemini-3.1-flash-image',
      'gemini-2.5-pro',
      'gemini-2.0-flash',
      'custom-model'
    ])
  })

  it('enables image test mode for OpenAI GPT image models', async () => {
    getAvailableModels.mockResolvedValue([
      createModel('gpt-5.5', 'GPT-5.5'),
      createModel('gpt-image-2', 'GPT Image 2')
    ])

    const wrapper = mountSessionHarness(createAccount({ platform: 'openai', type: 'apikey' }))
    await flushPromises()

    ;(wrapper.vm as any).selectedModelId = 'gpt-image-2'
    await flushPromises()

    expect((wrapper.vm as any).supportsImageTest).toBe(true)
    expect((wrapper.vm as any).supportsOpenAIImageTest).toBe(true)
    expect((wrapper.vm as any).testPrompt).toBe('admin.accounts.imagePromptDefault')
    expect((wrapper.vm as any).testModeLabel).toBe('admin.accounts.imageTestMode')
  })

  it('parses sse chunks while preserving an unfinished trailing buffer', () => {
    const first = parseSSEDataChunk(
      '',
      'data: {"type":"test_start","model":"gemini-2.5"}\ndata: {"type":"content","text":"hel'
    )
    expect(first.events).toEqual([{ type: 'test_start', model: 'gemini-2.5' }])

    const second = parseSSEDataChunk(
      first.buffer,
      'lo"}\ndata: {"type":"image","image_url":"data:image/png;base64,AA=="}\n'
    )
    expect(second.events).toEqual([
      { type: 'content', text: 'hello' },
      { type: 'image', image_url: 'data:image/png;base64,AA==' }
    ])
    expect(second.buffer).toBe('')
  })

  it('reports json parse failures without breaking other events', () => {
    const onParseError = vi.fn()
    const result = parseSSEDataChunk(
      '',
      'data: {"type":"test_start"}\ndata: {"type":}\ndata: {"type":"error","error":"boom"}\n',
      onParseError
    )

    expect(onParseError).toHaveBeenCalledTimes(1)
    expect(result.events).toEqual([
      { type: 'test_start' },
      { type: 'error', error: 'boom' }
    ])
  })

  it('ignores stale availableModels after close-reopen', async () => {
    const firstLoad = createDeferred<ClaudeModel[]>()
    const secondLoad = createDeferred<ClaudeModel[]>()
    getAvailableModels
      .mockReturnValueOnce(firstLoad.promise)
      .mockReturnValueOnce(secondLoad.promise)

    const wrapper = mountSessionHarness()
    await flushPromises()

    ;(wrapper.vm as any).show = false
    await flushPromises()
    ;(wrapper.vm as any).show = true
    await flushPromises()

    secondLoad.resolve([
      createModel('claude-3-5-sonnet', 'Claude 3.5 Sonnet'),
      createModel('claude-3-haiku', 'Claude 3 Haiku')
    ])
    await flushPromises()

    expect((wrapper.vm as any).availableModels.map((model: ClaudeModel) => model.id)).toEqual([
      'claude-3-5-sonnet',
      'claude-3-haiku'
    ])
    expect((wrapper.vm as any).selectedModelId).toBe('claude-3-5-sonnet')

    firstLoad.resolve([
      createModel('stale-model', 'Stale Model')
    ])
    await flushPromises()

    expect((wrapper.vm as any).availableModels.map((model: ClaudeModel) => model.id)).toEqual([
      'claude-3-5-sonnet',
      'claude-3-haiku'
    ])
    expect((wrapper.vm as any).selectedModelId).toBe('claude-3-5-sonnet')
    expect((wrapper.vm as any).loadingModels).toBe(false)
  })

  it('reloads models on account switch and ignores the older response', async () => {
    const firstLoad = createDeferred<ClaudeModel[]>()
    const secondLoad = createDeferred<ClaudeModel[]>()
    getAvailableModels
      .mockReturnValueOnce(firstLoad.promise)
      .mockReturnValueOnce(secondLoad.promise)

    const wrapper = mountSessionHarness(createAccount({ id: 1, name: 'Account One' }))
    await flushPromises()

    ;(wrapper.vm as any).account = createAccount({
      id: 2,
      name: 'Account Two',
      platform: 'gemini'
    })
    await flushPromises()

    secondLoad.resolve([
      createModel('gemini-2.5-pro', 'Gemini 2.5 Pro'),
      createModel('gemini-2.0-flash', 'Gemini 2.0 Flash')
    ])
    await flushPromises()

    expect(getAvailableModels).toHaveBeenNthCalledWith(1, 1)
    expect(getAvailableModels).toHaveBeenNthCalledWith(2, 2)
    expect((wrapper.vm as any).availableModels.map((model: ClaudeModel) => model.id)).toEqual([
      'gemini-2.5-pro',
      'gemini-2.0-flash'
    ])
    expect((wrapper.vm as any).selectedModelId).toBe('gemini-2.5-pro')

    firstLoad.resolve([
      createModel('claude-3-5-sonnet', 'Claude 3.5 Sonnet')
    ])
    await flushPromises()

    expect((wrapper.vm as any).availableModels.map((model: ClaudeModel) => model.id)).toEqual([
      'gemini-2.5-pro',
      'gemini-2.0-flash'
    ])
    expect((wrapper.vm as any).selectedModelId).toBe('gemini-2.5-pro')
    expect((wrapper.vm as any).loadingModels).toBe(false)
  })
})
