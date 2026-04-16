import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DataTable from '../DataTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

function stubMatchMedia(matches: boolean) {
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockImplementation(() => ({
      matches,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn()
    }))
  })
}

describe('DataTable', () => {
  beforeEach(() => {
    stubMatchMedia(true)
    Object.defineProperty(globalThis, 'ResizeObserver', {
      configurable: true,
      writable: true,
      value: class {
        observe() {}
        disconnect() {}
      }
    })
  })

  it('桌面视口只挂载桌面表格', () => {
    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data: [{ id: 1, name: 'Alice' }]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.find('.table-wrapper').exists()).toBe(true)
    expect(wrapper.find('.data-table-mobile').exists()).toBe(false)
  })

  it('移动视口不会挂载隐藏的桌面表格', () => {
    stubMatchMedia(false)

    const wrapper = mount(DataTable, {
      props: {
        columns: [{ key: 'name', label: 'Name' }],
        data: [{ id: 1, name: 'Alice' }]
      },
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    expect(wrapper.find('.data-table-mobile').exists()).toBe(true)
    expect(wrapper.find('.table-wrapper').exists()).toBe(false)
  })
})
