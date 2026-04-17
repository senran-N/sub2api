import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsErrorDetailsModal from '../OpsErrorDetailsModal.vue'

const listRequestErrorsMock = vi.fn()
const listUpstreamErrorsMock = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listRequestErrors: (...args: any[]) => listRequestErrorsMock(...args),
    listUpstreamErrors: (...args: any[]) => listUpstreamErrorsMock(...args),
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
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
  template: '<div v-if="show"><slot /></div>',
})

const SelectStub = defineComponent({
  name: 'SelectStub',
  props: ['modelValue', 'options'],
  emits: ['update:modelValue'],
  template: '<div class="select-stub" />',
})

const OpsErrorLogTableStub = defineComponent({
  name: 'OpsErrorLogTableStub',
  props: {
    rows: { type: Array, default: () => [] },
    total: { type: Number, default: 0 },
    loading: { type: Boolean, default: false },
    page: { type: Number, default: 1 },
    pageSize: { type: Number, default: 10 },
  },
  emits: ['openErrorDetail', 'update:page', 'update:pageSize'],
  template: '<div class="ops-error-log-table-stub">{{ Array.isArray(rows) ? rows.map((row) => row.request_id).join("|") : "" }}</div>',
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

describe('OpsErrorDetailsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    listRequestErrorsMock.mockResolvedValue({ items: [], total: 0 })
    listUpstreamErrorsMock.mockResolvedValue({ items: [], total: 0 })
  })

  it('keeps the latest error log result when earlier loads resolve late', async () => {
    const firstResponse = createDeferred<{ items: Array<Record<string, unknown>>, total: number }>()
    const secondResponse = createDeferred<{ items: Array<Record<string, unknown>>, total: number }>()

    listRequestErrorsMock
      .mockReturnValueOnce(firstResponse.promise)
      .mockReturnValueOnce(secondResponse.promise)

    const wrapper = mount(OpsErrorDetailsModal, {
      props: {
        show: true,
        timeRange: '1h',
        errorType: 'request',
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          OpsErrorLogTable: OpsErrorLogTableStub,
        },
      },
    })

    await wrapper.setProps({ platform: 'openai' })

    secondResponse.resolve({
      items: [
        {
          id: 2,
          request_id: 'fresh_log',
        },
      ],
      total: 1,
    })
    await flushPromises()

    firstResponse.resolve({
      items: [
        {
          id: 1,
          request_id: 'stale_log',
        },
      ],
      total: 1,
    })
    await flushPromises()

    expect(wrapper.text()).toContain('fresh_log')
    expect(wrapper.text()).not.toContain('stale_log')
  })
})
