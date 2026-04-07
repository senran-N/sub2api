<template>
  <div class="space-y-6">
    <SoraProfilesCard
      :profiles="soraS3Profiles"
      :loading="loadingSoraProfiles"
      :activating="activatingSoraProfile"
      :deleting="deletingSoraProfile"
      @create="startCreateSoraProfile"
      @reload="loadSoraS3Profiles"
      @edit="editSoraProfile"
      @activate="activateSoraProfile"
      @remove="removeSoraProfile"
    />
  </div>

  <SoraProfileDrawer
    :open="soraProfileDrawerOpen"
    :creating="creatingSoraProfile"
    :saving="savingSoraProfile"
    :testing="testingSoraProfile"
    :form="soraProfileForm"
    @close="closeSoraProfileDrawer"
    @test="testSoraProfileConnection"
    @save="saveSoraProfile"
  />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import SoraProfileDrawer from './dataManagement/SoraProfileDrawer.vue'
import SoraProfilesCard from './dataManagement/SoraProfilesCard.vue'
import { useDataManagementSoraProfiles } from './dataManagement/useSoraProfiles'

const { t } = useI18n()
const appStore = useAppStore()

const {
  loadingSoraProfiles,
  savingSoraProfile,
  testingSoraProfile,
  activatingSoraProfile,
  deletingSoraProfile,
  creatingSoraProfile,
  soraProfileDrawerOpen,
  soraS3Profiles,
  soraProfileForm,
  loadSoraS3Profiles,
  startCreateSoraProfile,
  editSoraProfile,
  closeSoraProfileDrawer,
  saveSoraProfile,
  testSoraProfileConnection,
  activateSoraProfile,
  removeSoraProfile
} = useDataManagementSoraProfiles({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess,
  confirm: window.confirm
})

onMounted(async () => {
  await loadSoraS3Profiles()
})
</script>
