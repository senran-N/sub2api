import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import AnnouncementsView from '../AnnouncementsView.vue'

const deleteAnnouncementMock = vi.fn()
const loadAnnouncementsMock = vi.fn()
const loadSubscriptionGroupsMock = vi.fn()
const showErrorMock = vi.fn()
const showSuccessMock = vi.fn()

vi.mock('@/api/admin', () => ({
  adminAPI: {
    announcements: {
      delete: (...args: any[]) => deleteAnnouncementMock(...args)
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: showErrorMock,
    showSuccess: showSuccessMock
  })
}))

vi.mock('../announcements/useAnnouncementsViewData', () => ({
  useAnnouncementsViewData: () => ({
    announcements: ref([
      {
        id: 1,
        title: 'Maintenance',
        status: 'active',
        notify_mode: 'popup',
        targeting: { any_of: [] },
        created_at: '2026-04-11T00:00:00Z'
      }
    ]),
    loading: ref(false),
    filters: ref({ status: '' }).value,
    searchQuery: ref(''),
    pagination: ref({ total: 1, page: 1, page_size: 20 }).value,
    loadAnnouncements: loadAnnouncementsMock,
    handlePageChange: vi.fn(),
    handlePageSizeChange: vi.fn(),
    handleStatusChange: vi.fn(),
    handleSearch: vi.fn(),
    dispose: vi.fn()
  })
}))

vi.mock('../announcements/useAnnouncementsViewEditor', () => ({
  useAnnouncementsViewEditor: () => ({
    showEditDialog: ref(false),
    saving: ref(false),
    isEditing: ref(false),
    form: ref({}).value,
    subscriptionGroups: ref([]),
    loadSubscriptionGroups: loadSubscriptionGroupsMock,
    openCreateDialog: vi.fn(),
    openEditDialog: vi.fn(),
    closeEdit: vi.fn(),
    handleSave: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const DataTableStub = defineComponent({
  name: 'DataTableStub',
  props: {
    data: {
      type: Array,
      default: () => []
    }
  },
  template: `
    <div>
      <div v-for="row in data" :key="row.id">
        <slot name="cell-actions" :row="row" />
      </div>
    </div>
  `
})

const AppLayoutStub = defineComponent({
  name: 'AppLayoutStub',
  template: '<div><slot /></div>'
})

const TablePageLayoutStub = defineComponent({
  name: 'TablePageLayoutStub',
  template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
})

const ConfirmDialogStub = defineComponent({
  name: 'ConfirmDialogStub',
  props: {
    show: { type: Boolean, default: false }
  },
  emits: ['confirm', 'cancel'],
  template: '<button v-if="show" class="confirm-delete" @click="$emit(\'confirm\')">confirm</button>'
})

describe('AnnouncementsView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    loadAnnouncementsMock.mockResolvedValue(undefined)
    loadSubscriptionGroupsMock.mockResolvedValue(undefined)
  })

  it('prefers backend detail when deleting an announcement fails', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    deleteAnnouncementMock.mockRejectedValue({
      response: {
        data: {
          detail: 'announcement delete detail'
        }
      },
      message: 'generic delete error'
    })

    const wrapper = mount(AnnouncementsView, {
      global: {
        stubs: {
          AppLayout: AppLayoutStub,
          TablePageLayout: TablePageLayoutStub,
          DataTable: DataTableStub,
          Pagination: true,
          ConfirmDialog: ConfirmDialogStub,
          EmptyState: true,
          AnnouncementReadStatusDialog: true,
          AnnouncementEditDialog: true,
          AnnouncementActionsCell: {
            emits: ['delete', 'edit', 'read-status'],
            template: '<button class="delete-row" @click="$emit(\'delete\')">delete</button>'
          },
          AnnouncementCreatedAtCell: true,
          AnnouncementNotifyModeBadge: true,
          AnnouncementStatusBadge: true,
          AnnouncementTargetingCell: true,
          AnnouncementTimeRangeCell: true,
          AnnouncementTitleCell: true,
          AnnouncementsToolbar: true
        }
      }
    })

    await wrapper.get('.delete-row').trigger('click')
    await wrapper.get('.confirm-delete').trigger('click')
    await flushPromises()

    expect(deleteAnnouncementMock).toHaveBeenCalledWith(1)
    expect(showErrorMock).toHaveBeenCalledWith('announcement delete detail')
    expect(consoleSpy).toHaveBeenCalledTimes(1)
    consoleSpy.mockRestore()
  })
})
