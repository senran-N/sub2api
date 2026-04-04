import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import AnnouncementActionsCell from '../announcements/AnnouncementActionsCell.vue'
import AnnouncementCreatedAtCell from '../announcements/AnnouncementCreatedAtCell.vue'
import AnnouncementEditDialog from '../announcements/AnnouncementEditDialog.vue'
import AnnouncementNotifyModeBadge from '../announcements/AnnouncementNotifyModeBadge.vue'
import AnnouncementStatusBadge from '../announcements/AnnouncementStatusBadge.vue'
import AnnouncementTargetingCell from '../announcements/AnnouncementTargetingCell.vue'
import AnnouncementTimeRangeCell from '../announcements/AnnouncementTimeRangeCell.vue'
import AnnouncementTitleCell from '../announcements/AnnouncementTitleCell.vue'
import AnnouncementsToolbar from '../announcements/AnnouncementsToolbar.vue'

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/utils/format', () => ({
  formatDateTime: (value: string) => `date:${value}`
}))

const BaseDialogStub = {
  props: ['show', 'title', 'width'],
  template: '<div><slot /><slot name="footer" /></div>'
}

const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  template: `
    <button
      class="select-stub"
      @click="
        $emit('update:modelValue', options[1]?.value ?? modelValue);
        $emit('change', options[1]?.value ?? modelValue)
      "
    >
      {{ modelValue }}
    </button>
  `
}

describe('announcements local components', () => {
  it('renders toolbar and emits search, status, refresh, and create actions', async () => {
    const wrapper = mount(AnnouncementsToolbar, {
      props: {
        searchQuery: 'notice',
        status: '',
        statusOptions: [
          { value: '', label: 'All' },
          { value: 'active', label: 'Active' }
        ],
        loading: false
      },
      global: {
        stubs: {
          Icon: true,
          Select: SelectStub
        }
      }
    })

    await wrapper.find('input').setValue('banner')
    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('update:searchQuery')?.[0]).toEqual(['banner'])
    expect(wrapper.emitted('search')?.length).toBe(1)
    expect(wrapper.emitted('update:status')?.[0]).toEqual(['active'])
    expect(wrapper.emitted('status-change')?.length).toBe(1)
    expect(wrapper.emitted('refresh')?.length).toBe(1)
    expect(wrapper.emitted('create')?.length).toBe(1)
  })

  it('renders action buttons and re-emits clicks', async () => {
    const wrapper = mount(AnnouncementActionsCell, {
      global: {
        stubs: {
          Icon: true
        }
      }
    })

    const buttons = wrapper.findAll('button')
    await buttons[0].trigger('click')
    await buttons[1].trigger('click')
    await buttons[2].trigger('click')

    expect(wrapper.emitted('read-status')?.length).toBe(1)
    expect(wrapper.emitted('edit')?.length).toBe(1)
    expect(wrapper.emitted('delete')?.length).toBe(1)
  })

  it('renders edit dialog and emits submit', async () => {
    const wrapper = mount(AnnouncementEditDialog, {
      props: {
        show: true,
        editing: false,
        saving: false,
        form: {
          title: 'Hello',
          content: 'World',
          status: 'draft',
          notify_mode: 'silent',
          starts_at_str: '',
          ends_at_str: '',
          targeting: { any_of: [] }
        },
        subscriptionGroups: [],
        statusOptions: [
          { value: 'draft', label: 'Draft' },
          { value: 'active', label: 'Active' }
        ],
        notifyModeOptions: [
          { value: 'silent', label: 'Silent' },
          { value: 'popup', label: 'Popup' }
        ]
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          Select: SelectStub,
          AnnouncementTargetingEditor: true
        }
      }
    })

    expect(wrapper.text()).toContain('admin.announcements.form.title')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')?.length).toBe(1)
  })

  it('renders announcement table cell components', () => {
    expect(mount(AnnouncementTitleCell, {
      props: {
        id: 2,
        title: 'Hello',
        createdAt: '2026-04-05T00:00:00Z'
      }
    }).text()).toContain('date:2026-04-05T00:00:00Z')

    expect(mount(AnnouncementStatusBadge, {
      props: {
        status: 'active'
      }
    }).text()).toContain('admin.announcements.statusLabels.active')

    expect(mount(AnnouncementNotifyModeBadge, {
      props: {
        notifyMode: 'popup'
      }
    }).text()).toContain('admin.announcements.notifyModeLabels.popup')

    expect(mount(AnnouncementTargetingCell, {
      props: {
        targeting: { any_of: [] }
      }
    }).text()).toContain('admin.announcements.targetingSummaryAll')

    expect(mount(AnnouncementTimeRangeCell, {
      props: {
        startsAt: '2026-04-05T00:00:00Z',
        endsAt: undefined
      }
    }).text()).toContain('admin.announcements.timeNever')

    expect(mount(AnnouncementCreatedAtCell, {
      props: {
        value: '2026-04-05T00:00:00Z'
      }
    }).text()).toContain('date:2026-04-05T00:00:00Z')
  })
})
