import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsSystemLogTable from '../OpsSystemLogTable.vue'

const mockListSystemLogs = vi.fn()
const mockGetSystemLogSinkHealth = vi.fn()
const mockGetRuntimeLogConfig = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listSystemLogs: (...args: any[]) => mockListSystemLogs(...args),
    getSystemLogSinkHealth: (...args: any[]) => mockGetSystemLogSinkHealth(...args),
    getRuntimeLogConfig: (...args: any[]) => mockGetRuntimeLogConfig(...args),
    updateRuntimeLogConfig: vi.fn(),
    resetRuntimeLogConfig: vi.fn(),
    cleanupSystemLogs: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
}))

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: {
    modelValue: {
      type: [String, Number, Boolean, Object],
      default: '',
    },
    options: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue', 'change'],
  template: '<div class="select-stub" />',
})

const PaginationStub = defineComponent({
  name: 'PaginationStub',
  template: '<div class="pagination-stub" />',
})

function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

function makeLog(id: number, message: string) {
  return {
    id,
    level: 'info',
    message,
    created_at: '2026-04-17T00:00:00Z',
    request_id: '',
    client_request_id: '',
    user_id: null,
    account_id: null,
    platform: '',
    model: '',
    extra: {},
  }
}

describe('OpsSystemLogTable', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListSystemLogs.mockResolvedValue({
      items: [],
      total: 0,
    })
    mockGetSystemLogSinkHealth.mockResolvedValue({
      queue_depth: 0,
      queue_capacity: 0,
      dropped_count: 0,
      write_failed_count: 0,
      written_count: 0,
      avg_write_delay_ms: 0,
      last_error: '',
    })
  })

  it('surfaces backend detail when runtime log config loading fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockGetRuntimeLogConfig.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'runtime log config detail error',
        },
      },
      message: 'generic runtime log config error',
    })

    mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('runtime log config detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest log query result when searches are triggered back to back', async () => {
    mockGetRuntimeLogConfig.mockResolvedValue({
      level: 'info',
      enable_sampling: false,
      sampling_initial: 100,
      sampling_thereafter: 100,
      caller: true,
      stacktrace_level: 'error',
      retention_days: 30,
    })

    const wrapper = mount(OpsSystemLogTable, {
      global: {
        stubs: {
          Select: SelectStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    const slowResponse = deferred<any>()
    const fastResponse = deferred<any>()
    mockListSystemLogs
      .mockReturnValueOnce(slowResponse.promise)
      .mockReturnValueOnce(fastResponse.promise)

    const queryInput = wrapper.get('input[placeholder="消息/request_id"]')
    const queryButton = wrapper
      .findAll('button')
      .find((button) => button.text() === '查询')

    expect(queryButton).toBeTruthy()
    await queryInput.setValue('first')
    await queryButton!.trigger('click')
    await queryInput.setValue('second')
    await queryButton!.trigger('click')

    fastResponse.resolve({
      items: [makeLog(2, 'second-hit')],
      total: 1,
    })
    await flushPromises()

    slowResponse.resolve({
      items: [makeLog(1, 'first-hit')],
      total: 1,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('second-hit')
    expect(wrapper.text()).not.toContain('first-hit')
  })
})
