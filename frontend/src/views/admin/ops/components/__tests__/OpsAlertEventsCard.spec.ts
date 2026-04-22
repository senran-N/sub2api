import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsAlertEventsCard from '../OpsAlertEventsCard.vue'

const mockListAlertEvents = vi.fn()
const mockGetAlertEvent = vi.fn()
const mockCreateAlertSilence = vi.fn()
const mockUpdateAlertEventStatus = vi.fn()
const showError = vi.fn()
const showSuccess = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    listAlertEvents: (...args: any[]) => mockListAlertEvents(...args),
    getAlertEvent: (...args: any[]) => mockGetAlertEvent(...args),
    createAlertSilence: (...args: any[]) => mockCreateAlertSilence(...args),
    updateAlertEventStatus: (...args: any[]) => mockUpdateAlertEventStatus(...args),
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError,
    showSuccess,
  }),
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
  template: '<div v-if="show" class="base-dialog-stub"><div class="dialog-title">{{ title }}</div><slot /><slot name="footer" /></div>',
})

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
  emits: ['change', 'update:modelValue'],
  template: '<div class="select-stub" />',
})

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span class="icon-stub" />',
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

function makeAlertEvent(id: number, title: string, overrides: Record<string, unknown> = {}) {
  return {
    id,
    rule_id: 100 + id,
    severity: 'P1',
    status: 'firing',
    title,
    fired_at: '2026-04-17T10:00:00Z',
    resolved_at: null,
    email_sent: false,
    created_at: '2026-04-17T10:00:00Z',
    dimensions: {
      platform: 'openai',
      group_id: 1,
    },
    ...overrides,
  }
}

function makeAlertList(count: number, prefix: string) {
  return Array.from({ length: count }, (_, index) => makeAlertEvent(index + 1, `${prefix} ${index + 1}`))
}

async function mountCard(initialEvents = makeAlertList(2, 'initial')) {
  mockListAlertEvents.mockResolvedValueOnce(initialEvents)
  const wrapper = mount(OpsAlertEventsCard, {
    props: {
      refreshToken: 0,
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Select: SelectStub,
        Icon: IconStub,
      },
    },
  })
  await flushPromises()
  return wrapper
}

async function changeTopLevelFilter(wrapper: ReturnType<typeof mount>, index: number, value: string) {
  const selects = wrapper.findAllComponents(SelectStub)
  await selects[index]?.vm.$emit('change', value)
  await flushPromises()
}

describe('OpsAlertEventsCard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockCreateAlertSilence.mockResolvedValue(undefined)
    mockUpdateAlertEventStatus.mockResolvedValue(undefined)
  })

  it('keeps the latest filter result when list requests overlap', async () => {
    const wrapper = await mountCard()
    const slowList = deferred<any[]>()
    const fastList = deferred<any[]>()

    mockListAlertEvents
      .mockReturnValueOnce(slowList.promise)
      .mockReturnValueOnce(fastList.promise)

    await changeTopLevelFilter(wrapper, 0, '6h')
    await changeTopLevelFilter(wrapper, 0, '30d')

    fastList.resolve([makeAlertEvent(31, 'latest filter result')])
    await flushPromises()

    slowList.resolve([makeAlertEvent(32, 'stale filter result')])
    await flushPromises()

    expect(wrapper.text()).toContain('latest filter result')
    expect(wrapper.text()).not.toContain('stale filter result')
  })

  it('ignores stale pagination results after a fresh filter reload', async () => {
    const wrapper = await mountCard(makeAlertList(10, 'page'))
    const slowMore = deferred<any[]>()
    const fastReload = deferred<any[]>()

    mockListAlertEvents
      .mockReturnValueOnce(slowMore.promise)
      .mockReturnValueOnce(fastReload.promise)

    const scrollWrapper = wrapper.find('.ops-alert-events-card__table-scroll')
    const scrollElement = scrollWrapper.element as HTMLElement
    Object.defineProperty(scrollElement, 'clientHeight', { configurable: true, value: 100 })
    Object.defineProperty(scrollElement, 'scrollHeight', { configurable: true, value: 560 })
    scrollElement.scrollTop = 500
    await scrollWrapper.trigger('scroll')

    await changeTopLevelFilter(wrapper, 0, '7d')

    fastReload.resolve([makeAlertEvent(41, 'fresh reload result')])
    await flushPromises()

    slowMore.resolve([makeAlertEvent(42, 'stale pagination result')])
    await flushPromises()

    expect(wrapper.text()).toContain('fresh reload result')
    expect(wrapper.text()).not.toContain('stale pagination result')
  })

  it('keeps the latest detail and history when switching alert selections quickly', async () => {
    const wrapper = await mountCard([
      makeAlertEvent(1, 'row one'),
      makeAlertEvent(2, 'row two'),
    ])

    const slowDetail = deferred<any>()
    const fastDetail = deferred<any>()
    const fastHistory = deferred<any[]>()
    const slowHistory = deferred<any[]>()

    mockGetAlertEvent
      .mockReturnValueOnce(slowDetail.promise)
      .mockReturnValueOnce(fastDetail.promise)
    mockListAlertEvents
      .mockReturnValueOnce(fastHistory.promise)
      .mockReturnValueOnce(slowHistory.promise)

    await wrapper.findAll('.ops-alert-events-card__table-shell tbody .ops-alert-events-card__row')[0]?.trigger('click')
    await flushPromises()
    await wrapper.findAll('.ops-alert-events-card__table-shell tbody .ops-alert-events-card__row')[1]?.trigger('click')
    await flushPromises()

    fastDetail.resolve(makeAlertEvent(2, 'detail two', { description: 'detail two only' }))
    await flushPromises()

    fastHistory.resolve([
      makeAlertEvent(202, 'history two', {
        rule_id: 102,
        metric_value: 33.33,
        threshold_value: 44.44,
      }),
    ])
    await flushPromises()

    slowDetail.resolve(makeAlertEvent(1, 'detail one', { description: 'detail one only' }))
    await flushPromises()

    slowHistory.resolve([
      makeAlertEvent(101, 'history one', {
        rule_id: 101,
        metric_value: 11.11,
        threshold_value: 22.22,
      }),
    ])
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('detail two only')
    expect(text).not.toContain('detail one only')
    expect(text).toContain('33.33 / 44.44')
    expect(text).not.toContain('11.11 / 22.22')
  })

  it('reloads alert events when the dashboard refresh token changes', async () => {
    const wrapper = await mountCard([makeAlertEvent(1, 'before refresh')])

    mockListAlertEvents.mockResolvedValueOnce([makeAlertEvent(2, 'after refresh')])

    await wrapper.setProps({ refreshToken: 1 })
    await flushPromises()

    expect(mockListAlertEvents).toHaveBeenCalledTimes(2)
  })
})
