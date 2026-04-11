import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsRequestDetailsModal from '../OpsRequestDetailsModal.vue'

const mockListRequestDetails = vi.fn()
const showError = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listRequestDetails: (...args: any[]) => mockListRequestDetails(...args),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({
    showError,
    showWarning: vi.fn(),
  }),
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true),
  }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, any>) => {
        if (key === 'admin.ops.requestDetails.rangeLabel') {
          return `range:${params?.range ?? ''}`
        }
        if (key === 'admin.ops.requestDetails.rangeHours') {
          return `${params?.n ?? ''}h`
        }
        if (key === 'admin.ops.requestDetails.rangeMinutes') {
          return `${params?.n ?? ''}m`
        }
        return key
      },
    }),
  }
})

const BaseDialogStub = defineComponent({
  name: 'BaseDialogStub',
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div v-if="show" class="base-dialog-stub"><slot /></div>',
})

const PaginationStub = defineComponent({
  name: 'PaginationStub',
  template: '<div class="pagination-stub" />',
})

const sampleResponse = {
  items: [
    {
      kind: 'success' as const,
      created_at: '2026-04-08T10:00:00.000Z',
      request_id: 'req_123',
      platform: 'openai',
      model: 'gpt-4.1',
      duration_ms: 123,
      status_code: 200,
    },
  ],
  total: 1,
}

describe('OpsRequestDetailsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListRequestDetails.mockResolvedValue(sampleResponse)
  })

  it('loads request details immediately when the modal mounts open', async () => {
    mount(OpsRequestDetailsModal, {
      props: {
        modelValue: true,
        timeRange: '1h',
        preset: {
          title: '请求明细',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(mockListRequestDetails).toHaveBeenCalledWith(
      expect.objectContaining({
        page: 1,
        page_size: 10,
        kind: 'all',
        sort: 'created_at_desc',
      })
    )
  })

  it('uses the active custom time window instead of falling back to the default 1h window', async () => {
    const customStartTime = '2026-04-08T00:00:00.000Z'
    const customEndTime = '2026-04-08T02:00:00.000Z'

    const wrapper = mount(OpsRequestDetailsModal, {
      props: {
        modelValue: true,
        timeRange: 'custom',
        customStartTime,
        customEndTime,
        preset: {
          title: '请求明细',
          kind: 'error',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(mockListRequestDetails).toHaveBeenCalledWith(
      expect.objectContaining({
        start_time: customStartTime,
        end_time: customEndTime,
        kind: 'error',
      })
    )
    expect(wrapper.text()).toContain('range:')
  })

  it('hides empty status and actions columns for success-only rows without those fields', async () => {
    mockListRequestDetails.mockResolvedValueOnce({
      items: [
        {
          kind: 'success' as const,
          created_at: '2026-04-08T10:00:00.000Z',
          request_id: 'req_success_only',
          platform: 'openai',
          model: 'gpt-5.4',
          duration_ms: 456,
          status_code: null,
          error_id: null,
        },
      ],
      total: 1,
    })

    const wrapper = mount(OpsRequestDetailsModal, {
      props: {
        modelValue: true,
        timeRange: '1h',
        preset: {
          title: '请求明细',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('admin.ops.requestDetails.table.status')
    expect(wrapper.text()).not.toContain('admin.ops.requestDetails.table.actions')
    expect(wrapper.text()).toContain('req_success_only')
  })

  it('prefers backend detail when request details loading fails', async () => {
    mockListRequestDetails.mockRejectedValueOnce({
      response: {
        data: {
          detail: 'request detail error'
        }
      },
      message: 'generic request error'
    })

    mount(OpsRequestDetailsModal, {
      props: {
        modelValue: true,
        timeRange: '1h',
        preset: {
          title: '请求明细',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Pagination: PaginationStub,
        },
      },
    })

    await flushPromises()

    expect(showError).toHaveBeenCalledWith('request detail error')
  })
})
