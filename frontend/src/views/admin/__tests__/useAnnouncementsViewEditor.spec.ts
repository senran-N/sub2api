import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAnnouncementsViewEditor } from '../announcements/useAnnouncementsViewEditor'
import type { AdminGroup, Announcement } from '@/types'

const { getAllGroups, createAnnouncement, updateAnnouncement } = vi.hoisted(() => ({
  getAllGroups: vi.fn(),
  createAnnouncement: vi.fn(),
  updateAnnouncement: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: {
      getAll: getAllGroups
    },
    announcements: {
      create: createAnnouncement,
      update: updateAnnouncement
    }
  }
}))

function createGroup(overrides: Partial<AdminGroup> = {}): AdminGroup {
  return {
    id: 1,
    name: 'Pro',
    description: 'subscription plan',
    platform: 'openai',
    rate_multiplier: 1.5,
    status: 'active',
    subscription_type: 'subscription',
    user_count: 0,
    account_count: 0,
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  } as AdminGroup
}

function createSourceAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    id: 10,
    title: 'Maintenance',
    content: 'Window',
    status: 'draft',
    notify_mode: 'silent',
    targeting: { any_of: [] },
    created_at: '2026-04-04T00:00:00Z',
    updated_at: '2026-04-04T00:00:00Z',
    ...overrides
  }
}

describe('useAnnouncementsViewEditor', () => {
  beforeEach(() => {
    getAllGroups.mockReset()
    createAnnouncement.mockReset()
    updateAnnouncement.mockReset()
    getAllGroups.mockResolvedValue([
      createGroup(),
      createGroup({ id: 2, name: 'Standard', subscription_type: 'standard' })
    ])
    createAnnouncement.mockResolvedValue(createSourceAnnouncement({ id: 11 }))
    updateAnnouncement.mockResolvedValue(createSourceAnnouncement({ id: 10, title: 'Updated' }))
  })

  it('loads subscription groups and opens create mode with reset form state', async () => {
    const reloadAnnouncements = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const editor = useAnnouncementsViewEditor({
      t: (key: string) => key,
      showSuccess,
      showError,
      reloadAnnouncements
    })

    await editor.loadSubscriptionGroups()
    expect(editor.subscriptionGroups.value).toEqual([createGroup()])

    editor.form.title = 'Dirty'
    editor.openCreateDialog()
    expect(editor.showEditDialog.value).toBe(true)
    expect(editor.isEditing.value).toBe(false)
    expect(editor.form.title).toBe('')
  })

  it('validates targeting and creates announcements', async () => {
    const reloadAnnouncements = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const editor = useAnnouncementsViewEditor({
      t: (key: string) => key,
      showSuccess,
      showError,
      reloadAnnouncements
    })

    editor.openCreateDialog()
    editor.form.targeting = {
      any_of: Array.from({ length: 51 }, () => ({ all_of: [] }))
    }
    await editor.handleSave()
    expect(showError).toHaveBeenCalledWith('admin.announcements.failedToCreate')
    expect(createAnnouncement).not.toHaveBeenCalled()

    editor.form.title = 'Maintenance'
    editor.form.content = 'Window'
    editor.form.status = 'active'
    editor.form.notify_mode = 'popup'
    editor.form.targeting = { any_of: [] }
    await editor.handleSave()
    expect(createAnnouncement).toHaveBeenCalledWith({
      title: 'Maintenance',
      content: 'Window',
      status: 'active',
      notify_mode: 'popup',
      targeting: { any_of: [] },
      starts_at: undefined,
      ends_at: undefined
    })
    expect(showSuccess).toHaveBeenCalledWith('common.success')
    expect(reloadAnnouncements).toHaveBeenCalledTimes(1)
  })

  it('fills edit state and updates announcements', async () => {
    const reloadAnnouncements = vi.fn().mockResolvedValue(undefined)
    const showSuccess = vi.fn()
    const showError = vi.fn()
    const source = createSourceAnnouncement({
      starts_at: '2026-04-04T08:30:00Z',
      ends_at: '2026-04-05T09:45:00Z'
    })
    const editor = useAnnouncementsViewEditor({
      t: (key: string) => key,
      showSuccess,
      showError,
      reloadAnnouncements
    })

    editor.openEditDialog(source)
    expect(editor.isEditing.value).toBe(true)
    expect(editor.form.title).toBe('Maintenance')

    editor.form.title = 'Updated'
    editor.form.ends_at_str = ''
    await editor.handleSave()
    expect(updateAnnouncement).toHaveBeenCalledWith(10, {
      title: 'Updated',
      ends_at: 0
    })
    expect(showSuccess).toHaveBeenCalledWith('common.success')
    expect(reloadAnnouncements).toHaveBeenCalledTimes(1)
  })
})
