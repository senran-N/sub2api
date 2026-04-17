import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserAllowedGroupsModal from '../UserAllowedGroupsModal.vue'

const mockListGroups = vi.fn()
const mockUpdateUser = vi.fn()
const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      list: (...args: any[]) => mockListGroups(...args)
    },
    users: {
      update: (...args: any[]) => mockUpdateUser(...args)
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock
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
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createGroup(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    name: 'Exclusive A',
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: true,
    status: 'active',
    subscription_type: 'standard',
    ...overrides
  }
}

describe('UserAllowedGroupsModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockListGroups.mockResolvedValue({ items: [] })
    mockUpdateUser.mockResolvedValue({})
  })

  it('loads groups immediately when mounted open and surfaces request details', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    mockListGroups.mockRejectedValue({
      response: {
        data: {
          detail: 'allowed groups detail error'
        }
      },
      message: 'generic groups error'
    })

    mount(UserAllowedGroupsModal, {
      props: {
        show: true,
        user: {
          id: 9,
          email: 'user@example.com',
          allowed_groups: [],
          group_rates: {}
        }
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          PlatformIcon: true
        }
      }
    })

    await flushPromises()

    expect(mockListGroups).toHaveBeenCalledWith(1, 1000)
    expect(showErrorMock).toHaveBeenCalledWith('allowed groups detail error')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })

  it('keeps the latest user group config when requests resolve out of order', async () => {
    const firstLoad = createDeferred<{ items: ReturnType<typeof createGroup>[] }>()
    const secondLoad = createDeferred<{ items: ReturnType<typeof createGroup>[] }>()
    mockListGroups
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const wrapper = mount(UserAllowedGroupsModal, {
      props: {
        show: true,
        user: {
          id: 9,
          email: 'first@example.com',
          allowed_groups: [],
          group_rates: {}
        }
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          PlatformIcon: true
        }
      }
    })

    await wrapper.setProps({
      user: {
        id: 10,
        email: 'second@example.com',
        allowed_groups: [1],
        group_rates: { 1: 2.5 }
      }
    })

    secondLoad.resolve({ items: [createGroup()] })
    await flushPromises()
    expect((wrapper.find('input[type="checkbox"]').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.find('input[type="number"]').element as HTMLInputElement).value).toBe('2.5')

    firstLoad.resolve({ items: [createGroup()] })
    await flushPromises()
    expect((wrapper.find('input[type="checkbox"]').element as HTMLInputElement).checked).toBe(true)
    expect((wrapper.find('input[type="number"]').element as HTMLInputElement).value).toBe('2.5')
  })

  it('ignores a stale save result after the modal closes and reopens', async () => {
    const saveRequest = createDeferred<Record<string, never>>()
    mockListGroups
      .mockResolvedValueOnce({ items: [createGroup()] })
      .mockResolvedValueOnce({ items: [createGroup({ id: 2, name: 'Exclusive B' })] })
    mockUpdateUser.mockReturnValueOnce(saveRequest.promise)

    const wrapper = mount(UserAllowedGroupsModal, {
      props: {
        show: true,
        user: {
          id: 9,
          email: 'first@example.com',
          allowed_groups: [],
          group_rates: {}
        }
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          PlatformIcon: true
        }
      }
    })

    await flushPromises()

    const saveButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('common.save'))
    expect(saveButton).toBeTruthy()
    await saveButton!.trigger('click')
    await flushPromises()

    await wrapper.setProps({ show: false })
    await flushPromises()
    await wrapper.setProps({
      show: true,
      user: {
        id: 10,
        email: 'second@example.com',
        allowed_groups: [2],
        group_rates: { 2: 1.8 }
      }
    })
    await flushPromises()

    saveRequest.resolve({})
    await flushPromises()

    expect(wrapper.text()).toContain('Exclusive B')
    expect(wrapper.text()).not.toContain('Exclusive A')
    expect(showSuccessMock).not.toHaveBeenCalledWith('admin.users.groupConfigUpdated')
    expect(wrapper.emitted('success')).toBeFalsy()
    expect(wrapper.emitted('close')).toBeFalsy()
  })
})
