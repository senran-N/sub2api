import { computed, reactive, ref } from 'vue'
import { adminAPI } from '@/api/admin'
import type { AdminGroup, Announcement } from '@/types'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildCreateAnnouncementRequest,
  buildUpdateAnnouncementRequest,
  createDefaultAnnouncementForm,
  fillAnnouncementForm,
  filterAnnouncementSubscriptionGroups,
  resetAnnouncementForm,
  validateAnnouncementTargeting
} from './announcementsForm'

interface AnnouncementsViewEditorOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showSuccess: (message: string) => void
  showError: (message: string) => void
  reloadAnnouncements: () => Promise<void>
}

export function useAnnouncementsViewEditor(options: AnnouncementsViewEditorOptions) {
  const showEditDialog = ref(false)
  const saving = ref(false)
  const editingAnnouncement = ref<Announcement | null>(null)
  const form = reactive(createDefaultAnnouncementForm())
  const subscriptionGroups = ref<AdminGroup[]>([])

  const isEditing = computed(() => editingAnnouncement.value !== null)

  const loadSubscriptionGroups = async () => {
    try {
      const groups = await adminAPI.groups.getAll()
      subscriptionGroups.value = filterAnnouncementSubscriptionGroups(groups || [])
    } catch (error) {
      console.error('Error loading groups:', error)
    }
  }

  const openCreateDialog = () => {
    editingAnnouncement.value = null
    resetAnnouncementForm(form)
    showEditDialog.value = true
  }

  const openEditDialog = (announcement: Announcement) => {
    editingAnnouncement.value = announcement
    fillAnnouncementForm(form, announcement)
    showEditDialog.value = true
  }

  const closeEdit = () => {
    showEditDialog.value = false
    editingAnnouncement.value = null
  }

  const handleSave = async () => {
    if (!validateAnnouncementTargeting(form.targeting)) {
      options.showError(options.t('admin.announcements.failedToCreate'))
      return
    }

    saving.value = true
    try {
      if (!editingAnnouncement.value) {
        await adminAPI.announcements.create(buildCreateAnnouncementRequest(form))
        options.showSuccess(options.t('common.success'))
        showEditDialog.value = false
        await options.reloadAnnouncements()
        return
      }

      await adminAPI.announcements.update(
        editingAnnouncement.value.id,
        buildUpdateAnnouncementRequest(form, editingAnnouncement.value)
      )
      options.showSuccess(options.t('common.success'))
      showEditDialog.value = false
      editingAnnouncement.value = null
      await options.reloadAnnouncements()
    } catch (error: any) {
      console.error('Failed to save announcement:', error)
      options.showError(
        resolveRequestErrorMessage(
          error,
          editingAnnouncement.value
            ? options.t('admin.announcements.failedToUpdate')
            : options.t('admin.announcements.failedToCreate')
        )
      )
    } finally {
      saving.value = false
    }
  }

  return {
    showEditDialog,
    saving,
    editingAnnouncement,
    isEditing,
    form,
    subscriptionGroups,
    loadSubscriptionGroups,
    openCreateDialog,
    openEditDialog,
    closeEdit,
    handleSave
  }
}
