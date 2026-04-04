<template>
  <BaseDialog
    :show="show"
    :title="editing ? t('admin.announcements.editAnnouncement') : t('admin.announcements.createAnnouncement')"
    width="wide"
    @close="emit('close')"
  >
    <form id="announcement-form" class="space-y-4" @submit.prevent="emit('submit')">
      <div>
        <label class="input-label">{{ t('admin.announcements.form.title') }}</label>
        <input v-model="form.title" type="text" class="input" required />
      </div>

      <div>
        <label class="input-label">{{ t('admin.announcements.form.content') }}</label>
        <textarea v-model="form.content" rows="6" class="input" required></textarea>
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div>
          <label class="input-label">{{ t('admin.announcements.form.status') }}</label>
          <Select v-model="form.status" :options="statusOptions" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.announcements.form.notifyMode') }}</label>
          <Select v-model="form.notify_mode" :options="notifyModeOptions" />
          <p class="input-hint">{{ t('admin.announcements.form.notifyModeHint') }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div>
          <label class="input-label">{{ t('admin.announcements.form.startsAt') }}</label>
          <input v-model="form.starts_at_str" type="datetime-local" class="input" />
          <p class="input-hint">{{ t('admin.announcements.form.startsAtHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.announcements.form.endsAt') }}</label>
          <input v-model="form.ends_at_str" type="datetime-local" class="input" />
          <p class="input-hint">{{ t('admin.announcements.form.endsAtHint') }}</p>
        </div>
      </div>

      <AnnouncementTargetingEditor
        v-model="form.targeting"
        :groups="subscriptionGroups"
      />
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="submit" form="announcement-form" :disabled="saving" class="btn btn-primary">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { AdminGroup, AnnouncementNotifyMode, AnnouncementStatus } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import AnnouncementTargetingEditor from '@/components/admin/announcements/AnnouncementTargetingEditor.vue'
import type { AnnouncementFormState } from '../announcementsForm'

defineProps<{
  show: boolean
  editing: boolean
  saving: boolean
  form: AnnouncementFormState
  subscriptionGroups: AdminGroup[]
  statusOptions: Array<{ value: AnnouncementStatus; label: string }>
  notifyModeOptions: Array<{ value: AnnouncementNotifyMode; label: string }>
}>()

const emit = defineEmits<{
  close: []
  submit: []
}>()

const { t } = useI18n()
</script>
