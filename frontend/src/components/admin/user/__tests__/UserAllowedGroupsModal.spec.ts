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
})
