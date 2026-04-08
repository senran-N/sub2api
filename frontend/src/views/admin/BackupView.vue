<template>
  <div class="space-y-6">
    <BackupS3ConfigCard
      :form="s3Form"
      :secret-configured="s3SecretConfigured"
      :saving="savingS3"
      :testing="testingS3"
      @open-guide="showR2Guide = true"
      @test="testS3"
      @save="saveS3Config"
    />

    <BackupScheduleCard
      :form="scheduleForm"
      :saving="savingSchedule"
      @save="saveSchedule"
    />

    <BackupOperationsCard
      :backups="backups"
      :loading="loadingBackups"
      :creating="creatingBackup"
      :restoring-id="restoringId"
      :manual-expire-days="manualExpireDays"
      @update:manual-expire-days="manualExpireDays = $event"
      @create="createBackup"
      @refresh="loadBackups"
      @download="downloadBackup"
      @restore="restoreBackup"
      @remove="removeBackup"
    />
  </div>

  <BackupR2GuideModal v-if="showR2Guide" :show="showR2Guide" @close="showR2Guide = false" />
</template>

<script setup lang="ts">
import { defineAsyncComponent, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import BackupOperationsCard from './backup/BackupOperationsCard.vue'
import BackupS3ConfigCard from './backup/BackupS3ConfigCard.vue'
import BackupScheduleCard from './backup/BackupScheduleCard.vue'
import { useBackupViewConfig } from './backup/useBackupViewConfig'
import { useBackupViewOperations } from './backup/useBackupViewOperations'

const BackupR2GuideModal = defineAsyncComponent(() => import('./backup/BackupR2GuideModal.vue'))

const { t } = useI18n()
const appStore = useAppStore()

const showR2Guide = ref(false)

const {
  s3Form,
  s3SecretConfigured,
  savingS3,
  testingS3,
  scheduleForm,
  savingSchedule,
  loadS3Config,
  saveS3Config,
  testS3,
  loadSchedule,
  saveSchedule
} = useBackupViewConfig({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess
})

const {
  backups,
  loadingBackups,
  creatingBackup,
  restoringId,
  manualExpireDays,
  loadBackups,
  createBackup,
  downloadBackup,
  restoreBackup,
  removeBackup,
  initialize,
  dispose
} = useBackupViewOperations({
  t,
  showSuccess: appStore.showSuccess,
  showError: appStore.showError,
  showWarning: appStore.showWarning,
  confirm: (message: string) => window.confirm(message),
  prompt: (message: string) => window.prompt(message),
  openUrl: (url: string) => {
    window.open(url, '_blank')
  }
})

onMounted(async () => {
  await Promise.all([loadS3Config(), loadSchedule(), initialize()])
})

onBeforeUnmount(() => {
  dispose()
})
</script>
