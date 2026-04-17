import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsErrorDetailModal from '../OpsErrorDetailModal.vue'

const getRequestErrorDetailMock = vi.fn()
const getUpstreamErrorDetailMock = vi.fn()
const listRequestErrorUpstreamErrorsMock = vi.fn()
const showErrorMock = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getRequestErrorDetail: (...args: any[]) => getRequestErrorDetailMock(...args),
    getUpstreamErrorDetail: (...args: any[]) => getUpstreamErrorDetailMock(...args),
    listRequestErrorUpstreamErrors: (...args: any[]) => listRequestErrorUpstreamErrorsMock(...args)
  }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError: showErrorMock
  })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
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
  let reject!: (reason?: unknown) => void

  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })

  return { promise, resolve, reject }
}

describe('OpsErrorDetailModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    getRequestErrorDetailMock.mockResolvedValue({})
    getUpstreamErrorDetailMock.mockResolvedValue({})
    listRequestErrorUpstreamErrorsMock.mockResolvedValue({ items: [] })
  })

  it('prefers backend detail when upstream error detail loading fails', async () => {
    getUpstreamErrorDetailMock.mockRejectedValue({
      response: {
        data: {
          detail: 'ops detail error'
        }
      },
      message: 'generic ops error'
    })

    mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 9,
        errorType: 'upstream'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(getUpstreamErrorDetailMock).toHaveBeenCalledWith(9)
    expect(showErrorMock).toHaveBeenCalledWith('ops detail error')
  })

  it('renders available stage timings for request error details', async () => {
    getRequestErrorDetailMock.mockResolvedValue({
      id: 42,
      created_at: '2026-04-17T10:00:00Z',
      phase: 'request',
      error_owner: 'platform',
      status_code: 502,
      platform: 'openai',
      group_name: 'default',
      request_id: 'req_123',
      message: 'gateway timeout',
      request_type: 3,
      error_body: '{"error":"timeout"}',
      auth_latency_ms: 12,
      routing_latency_ms: 18,
      wait_user_ms: 25,
      wait_account_ms: 31,
      ws_acquire_ms: 44,
      ws_healthcheck_ms: 52,
      upstream_latency_ms: 89,
      response_latency_ms: 13,
      time_to_first_token_ms: 144
    })

    const wrapper = mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 42,
        errorType: 'request'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(getRequestErrorDetailMock).toHaveBeenCalledWith(42)
    expect(wrapper.text()).toContain('admin.ops.errorDetail.timings')
    expect(wrapper.text()).toContain('admin.ops.errorDetail.timingsHint')
    expect(wrapper.text()).toContain('admin.ops.errorDetail.waitUser')
    expect(wrapper.text()).toContain('25ms')
    expect(wrapper.text()).toContain('31ms')
    expect(wrapper.text()).toContain('44ms')
    expect(wrapper.text()).toContain('52ms')
    expect(wrapper.text()).toContain('144ms')
  })

  it('keeps the newest request detail when earlier loads resolve late', async () => {
    const firstDetail = createDeferred<Record<string, unknown>>()
    const secondDetail = createDeferred<Record<string, unknown>>()

    getRequestErrorDetailMock
      .mockReturnValueOnce(firstDetail.promise)
      .mockReturnValueOnce(secondDetail.promise)

    const wrapper = mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 101,
        errorType: 'request'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true
        }
      }
    })

    await wrapper.setProps({ errorId: 202 })

    secondDetail.resolve({
      id: 202,
      created_at: '2026-04-17T11:00:00Z',
      phase: 'request',
      error_owner: 'platform',
      status_code: 502,
      request_id: 'req_fresh',
      message: 'fresh detail',
      request_type: 1,
      error_body: '{"error":"fresh"}'
    })
    await flushPromises()

    firstDetail.resolve({
      id: 101,
      created_at: '2026-04-17T10:00:00Z',
      phase: 'request',
      error_owner: 'platform',
      status_code: 500,
      request_id: 'req_stale',
      message: 'stale detail',
      request_type: 1,
      error_body: '{"error":"stale"}'
    })
    await flushPromises()

    expect(wrapper.text()).toContain('req_fresh')
    expect(wrapper.text()).toContain('fresh detail')
    expect(wrapper.text()).not.toContain('req_stale')
    expect(wrapper.text()).not.toContain('stale detail')
  })

  it('keeps correlated upstream errors aligned with the latest request detail', async () => {
    const firstUpstreamErrors = createDeferred<{ items: Array<Record<string, unknown>> }>()
    const secondUpstreamErrors = createDeferred<{ items: Array<Record<string, unknown>> }>()

    getRequestErrorDetailMock.mockImplementation(async (id: number) => ({
      id,
      created_at: '2026-04-17T12:00:00Z',
      phase: 'request',
      error_owner: 'platform',
      status_code: 502,
      request_id: `req_${id}`,
      message: `detail ${id}`,
      request_type: 1,
      error_body: '{"error":"detail"}'
    }))

    listRequestErrorUpstreamErrorsMock
      .mockReturnValueOnce(firstUpstreamErrors.promise)
      .mockReturnValueOnce(secondUpstreamErrors.promise)

    const wrapper = mount(OpsErrorDetailModal, {
      props: {
        show: true,
        errorId: 301,
        errorType: 'request'
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Icon: true
        }
      }
    })

    await flushPromises()
    await wrapper.setProps({ errorId: 302 })
    await flushPromises()

    secondUpstreamErrors.resolve({
      items: [
        {
          id: 2,
          status_code: 503,
          request_id: 'upstream_fresh',
          message: 'fresh upstream detail'
        }
      ]
    })
    await flushPromises()

    firstUpstreamErrors.resolve({
      items: [
        {
          id: 1,
          status_code: 500,
          request_id: 'upstream_stale',
          message: 'stale upstream detail'
        }
      ]
    })
    await flushPromises()

    expect(wrapper.text()).toContain('fresh upstream detail')
    expect(wrapper.text()).not.toContain('stale upstream detail')
  })
})
