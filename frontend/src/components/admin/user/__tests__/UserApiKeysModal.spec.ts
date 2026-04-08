import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import UserApiKeysModal from '../UserApiKeysModal.vue'

const mockGetUserApiKeys = vi.fn()
const mockGetAllGroups = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: {
      getUserApiKeys: (...args: any[]) => mockGetUserApiKeys(...args),
    },
    groups: {
      getAll: (...args: any[]) => mockGetAllGroups(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
  }),
}))

vi.mock('@/composables/useDocumentThemeVersion', () => ({
  useDocumentThemeVersion: () => ref(0),
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
  template: '<div v-if="show" class="base-dialog-stub"><slot /></div>',
})

describe('UserApiKeysModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetUserApiKeys.mockResolvedValue({ items: [] })
    mockGetAllGroups.mockResolvedValue([])
  })

  it('loads user api keys and groups immediately when mounted open', async () => {
    mount(UserApiKeysModal, {
      props: {
        show: true,
        user: {
          id: 7,
          email: 'ops@example.com',
          username: 'ops-user',
        },
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: true,
          GroupOptionItem: true,
          Teleport: true,
        },
      },
    })

    await flushPromises()

    expect(mockGetUserApiKeys).toHaveBeenCalledWith(7)
    expect(mockGetAllGroups).toHaveBeenCalledTimes(1)
  })
})
