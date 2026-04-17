import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import GroupRateMultipliersModal from '../group/GroupRateMultipliersModal.vue'
import type { GroupRateMultiplierEntry } from '@/api/admin/groups'
import type { AdminGroup, AdminUser } from '@/types'

const getGroupRateMultipliersMock = vi.fn()
const listUsersMock = vi.fn()
const batchSetGroupRateMultipliersMock = vi.fn()
const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      getGroupRateMultipliers: (...args: any[]) => getGroupRateMultipliersMock(...args),
      batchSetGroupRateMultipliers: (...args: any[]) => batchSetGroupRateMultipliersMock(...args),
    },
    users: {
      list: (...args: any[]) => listUsersMock(...args),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock,
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
    width: { type: String, default: '' },
  },
  emits: ['close'],
  template: '<div v-if="show"><slot /></div>',
})

const PaginationStub = defineComponent({
  name: 'PaginationStub',
  props: {
    total: { type: Number, default: 0 },
    page: { type: Number, default: 1 },
    pageSize: { type: Number, default: 10 },
  },
  emits: ['update:page', 'update:pageSize'],
  template: '<div />',
})

const IconStub = defineComponent({
  name: 'IconStub',
  template: '<span />',
})

const PlatformIconStub = defineComponent({
  name: 'PlatformIconStub',
  template: '<span />',
})

function createDeferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((res) => {
    resolve = res
  })

  return { promise, resolve }
}

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'group-a',
    description: null,
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    sora_image_price_360: null,
    sora_image_price_540: null,
    sora_video_price_per_request: null,
    sora_video_price_per_request_hd: null,
    sora_storage_quota_bytes: 0,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    allow_messages_dispatch: false,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z',
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: true,
    supported_model_scopes: [],
    account_count: 0,
    active_account_count: 0,
    rate_limited_account_count: 0,
    default_mapped_model: '',
    sort_order: 0,
    ...overrides,
  }
}

function createUser(overrides: Partial<AdminUser> = {}): AdminUser {
  return {
    id: 1,
    email: 'user@example.com',
    username: 'user',
    role: 'user',
    balance: 0,
    status: 'active',
    allowed_groups: [],
    created_at: '2026-04-17T00:00:00Z',
    updated_at: '2026-04-17T00:00:00Z',
    notes: '',
    group_rates: {},
    current_concurrency: 0,
    sora_storage_quota_bytes: 0,
    sora_storage_used_bytes: 0,
    concurrency: 1,
    ...overrides,
  } as AdminUser
}

function createEntry(overrides: Partial<GroupRateMultiplierEntry> = {}): GroupRateMultiplierEntry {
  return {
    user_id: 1,
    user_name: 'user',
    user_email: 'user@example.com',
    user_notes: '',
    user_status: 'active',
    rate_multiplier: 1.5,
    ...overrides,
  }
}

function mountModal(props: { show?: boolean; group?: AdminGroup | null } = {}) {
  return mount(GroupRateMultipliersModal, {
    props: {
      show: props.show ?? true,
      group: props.group ?? createGroup(),
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Pagination: PaginationStub,
        Icon: IconStub,
        PlatformIcon: PlatformIconStub,
      },
    },
  })
}

describe('GroupRateMultipliersModal', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.clearAllMocks()
    getGroupRateMultipliersMock.mockResolvedValue([])
    listUsersMock.mockResolvedValue({ items: [] })
    batchSetGroupRateMultipliersMock.mockResolvedValue(undefined)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('keeps the latest entries when the group changes before the previous load resolves', async () => {
    const firstLoad = createDeferred<GroupRateMultiplierEntry[]>()
    const secondLoad = createDeferred<GroupRateMultiplierEntry[]>()
    const groupA = createGroup({ id: 1, name: 'group-a' })
    const groupB = createGroup({ id: 2, name: 'group-b' })

    getGroupRateMultipliersMock
      .mockImplementationOnce(() => firstLoad.promise)
      .mockImplementationOnce(() => secondLoad.promise)

    const wrapper = mountModal({ group: groupA })
    await wrapper.setProps({ group: groupB })

    secondLoad.resolve([createEntry({ user_id: 2, user_email: 'new@example.com' })])
    await flushPromises()

    firstLoad.resolve([createEntry({ user_id: 1, user_email: 'old@example.com' })])
    await flushPromises()

    expect((wrapper.vm as any).localEntries).toEqual([
      expect.objectContaining({ user_id: 2, user_email: 'new@example.com' })
    ])
    expect((wrapper.vm as any).loading).toBe(false)
  })

  it('keeps the latest user search results when overlapping searches resolve out of order', async () => {
    const firstSearch = createDeferred<{ items: AdminUser[] }>()
    const secondSearch = createDeferred<{ items: AdminUser[] }>()

    const wrapper = mountModal()
    await flushPromises()

    listUsersMock
      .mockImplementationOnce(() => firstSearch.promise)
      .mockImplementationOnce(() => secondSearch.promise)

    const searchInput = wrapper.get('input[placeholder="admin.groups.searchUserPlaceholder"]')

    await searchInput.setValue('alice')
    vi.advanceTimersByTime(300)
    await flushPromises()

    await searchInput.setValue('bob')
    vi.advanceTimersByTime(300)
    await flushPromises()

    secondSearch.resolve({
      items: [createUser({ id: 2, email: 'bob@example.com', username: 'bob' })]
    })
    await flushPromises()

    firstSearch.resolve({
      items: [createUser({ id: 1, email: 'alice@example.com', username: 'alice' })]
    })
    await flushPromises()

    expect((wrapper.vm as any).searchResults).toEqual([
      expect.objectContaining({ email: 'bob@example.com', username: 'bob' })
    ])
    expect((wrapper.vm as any).showDropdown).toBe(true)
  })
})
