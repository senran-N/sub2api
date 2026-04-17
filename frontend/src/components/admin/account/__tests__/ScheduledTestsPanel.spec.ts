import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import ScheduledTestsPanel from '../ScheduledTestsPanel.vue'
import type { ScheduledTestPlan, ScheduledTestResult } from '@/types'

const {
  showErrorMock,
  showSuccessMock,
  listByAccountMock,
  createMock,
  updateMock,
  deleteMock,
  listResultsMock
} = vi.hoisted(() => ({
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  listByAccountMock: vi.fn(),
  createMock: vi.fn(),
  updateMock: vi.fn(),
  deleteMock: vi.fn(),
  listResultsMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    scheduledTests: {
      listByAccount: listByAccountMock,
      create: createMock,
      update: updateMock,
      delete: deleteMock,
      listResults: listResultsMock
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' }
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>'
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createPlan(overrides: Partial<ScheduledTestPlan> = {}): ScheduledTestPlan {
  return {
    id: 1,
    account_id: 7,
    model_id: 'claude-3-7-sonnet',
    cron_expression: '0 * * * *',
    enabled: true,
    max_results: 100,
    auto_recover: false,
    last_run_at: null,
    next_run_at: null,
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z',
    ...overrides
  }
}

function createResult(overrides: Partial<ScheduledTestResult> = {}): ScheduledTestResult {
  return {
    id: 1,
    plan_id: 1,
    status: 'success',
    response_text: 'ok',
    error_message: '',
    latency_ms: 120,
    started_at: '2026-04-17T00:00:00Z',
    finished_at: '2026-04-17T00:00:01Z',
    created_at: '2026-04-17T00:00:01Z',
    ...overrides
  }
}

function mountPanel(props: { show?: boolean; accountId?: number | null } = {}) {
  return mount(ScheduledTestsPanel, {
    props: {
      show: props.show ?? true,
      accountId: props.accountId ?? 7,
      modelOptions: []
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: true,
        HelpTooltip: true,
        Select: true,
        Input: true,
        Toggle: true,
        Icon: true
      }
    }
  })
}

describe('ScheduledTestsPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listByAccountMock.mockResolvedValue([])
    createMock.mockResolvedValue({})
    updateMock.mockResolvedValue({})
    deleteMock.mockResolvedValue(undefined)
    listResultsMock.mockResolvedValue([])
  })

  it('prefers backend detail when loading plans fails', async () => {
    listByAccountMock.mockRejectedValue({
      response: {
        data: {
          detail: 'scheduled tests detail'
        }
      },
      message: 'generic scheduled tests error'
    })

    mountPanel()

    await flushPromises()

    expect(listByAccountMock).toHaveBeenCalledWith(7)
    expect(showErrorMock).toHaveBeenCalledWith('scheduled tests detail')
  })

  it('keeps the latest plans when account switches before the previous load resolves', async () => {
    const firstPlans = createDeferred<ScheduledTestPlan[]>()
    const secondPlans = createDeferred<ScheduledTestPlan[]>()

    listByAccountMock
      .mockImplementationOnce(() => firstPlans.promise)
      .mockImplementationOnce(() => secondPlans.promise)

    const wrapper = mountPanel({ accountId: 7 })
    await wrapper.setProps({ accountId: 8 })

    secondPlans.resolve([createPlan({ id: 2, account_id: 8, model_id: 'gpt-5-latest' })])
    await flushPromises()

    firstPlans.resolve([createPlan({ id: 1, account_id: 7, model_id: 'claude-old' })])
    await flushPromises()

    expect(listByAccountMock).toHaveBeenNthCalledWith(1, 7)
    expect(listByAccountMock).toHaveBeenNthCalledWith(2, 8)
    expect((wrapper.vm as any).plans).toEqual([
      expect.objectContaining({ account_id: 8, model_id: 'gpt-5-latest' })
    ])
    expect(wrapper.text()).toContain('gpt-5-latest')
    expect(wrapper.text()).not.toContain('claude-old')
  })

  it('keeps the latest expanded plan results when result requests resolve out of order', async () => {
    const firstResults = createDeferred<ScheduledTestResult[]>()
    const secondResults = createDeferred<ScheduledTestResult[]>()

    listByAccountMock.mockResolvedValue([
      createPlan({ id: 1, model_id: 'claude-plan' }),
      createPlan({ id: 2, model_id: 'gpt-plan' })
    ])
    listResultsMock
      .mockImplementationOnce(() => firstResults.promise)
      .mockImplementationOnce(() => secondResults.promise)

    const wrapper = mountPanel()
    await flushPromises()

    const headers = wrapper.findAll('.scheduled-tests-panel__plan-header')
    await headers[0]!.trigger('click')
    await headers[1]!.trigger('click')

    secondResults.resolve([
      createResult({ id: 2, plan_id: 2, response_text: 'latest response', started_at: '2026-04-17T01:00:00Z' })
    ])
    await flushPromises()

    firstResults.resolve([
      createResult({ id: 1, plan_id: 1, response_text: 'stale response', started_at: '2026-04-17T00:00:00Z' })
    ])
    await flushPromises()

    expect(listResultsMock).toHaveBeenNthCalledWith(1, 1, 20)
    expect(listResultsMock).toHaveBeenNthCalledWith(2, 2, 20)
    expect((wrapper.vm as any).expandedPlanId).toBe(2)
    expect((wrapper.vm as any).results).toEqual([
      expect.objectContaining({ plan_id: 2, response_text: 'latest response' })
    ])
    expect((wrapper.vm as any).loadingResults).toBe(false)
  })
})
