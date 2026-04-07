import { ref } from 'vue'
import { adminAPI } from '@/api'
import type { SoraS3Profile } from '@/api/admin/settings'
import { resolveRequestErrorMessage } from '@/utils/requestError'
import {
  buildCreateSoraS3ProfileRequest,
  buildTestSoraS3ConnectionRequest,
  buildUpdateSoraS3ProfileRequest,
  createDefaultSoraS3ProfileForm,
  getPreferredSoraProfileID,
  validateSoraS3ProfileForm,
  type SoraS3ProfileForm
} from './dataManagementHelpers'

interface DataManagementSoraProfilesOptions {
  t: (key: string, params?: Record<string, unknown>) => string
  showError: (message: string) => void
  showSuccess: (message: string) => void
  confirm: (message: string) => boolean
}

export function useDataManagementSoraProfiles(options: DataManagementSoraProfilesOptions) {
  const loadingSoraProfiles = ref(false)
  const savingSoraProfile = ref(false)
  const testingSoraProfile = ref(false)
  const activatingSoraProfile = ref(false)
  const deletingSoraProfile = ref(false)
  const creatingSoraProfile = ref(false)
  const soraProfileDrawerOpen = ref(false)

  const soraS3Profiles = ref<SoraS3Profile[]>([])
  const selectedSoraProfileID = ref('')
  const soraProfileForm = ref<SoraS3ProfileForm>(createDefaultSoraS3ProfileForm())

  const syncSoraProfileFormWithSelection = () => {
    const profile = soraS3Profiles.value.find(
      (item) => item.profile_id === selectedSoraProfileID.value
    )
    soraProfileForm.value = createDefaultSoraS3ProfileForm(profile)
  }

  const loadSoraS3Profiles = async () => {
    loadingSoraProfiles.value = true
    try {
      const result = await adminAPI.settings.listSoraS3Profiles()
      soraS3Profiles.value = result.items || []
      if (!creatingSoraProfile.value) {
        const stillExists = selectedSoraProfileID.value
          ? soraS3Profiles.value.some((item) => item.profile_id === selectedSoraProfileID.value)
          : false
        if (!stillExists) {
          selectedSoraProfileID.value = getPreferredSoraProfileID(soraS3Profiles.value)
        }
        syncSoraProfileFormWithSelection()
      }
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      loadingSoraProfiles.value = false
    }
  }

  const startCreateSoraProfile = () => {
    creatingSoraProfile.value = true
    selectedSoraProfileID.value = ''
    soraProfileForm.value = createDefaultSoraS3ProfileForm()
    soraProfileDrawerOpen.value = true
  }

  const editSoraProfile = (profileID: string) => {
    selectedSoraProfileID.value = profileID
    creatingSoraProfile.value = false
    syncSoraProfileFormWithSelection()
    soraProfileDrawerOpen.value = true
  }

  const closeSoraProfileDrawer = () => {
    soraProfileDrawerOpen.value = false
    if (creatingSoraProfile.value) {
      creatingSoraProfile.value = false
      selectedSoraProfileID.value = getPreferredSoraProfileID(soraS3Profiles.value)
      syncSoraProfileFormWithSelection()
    }
  }

  const saveSoraProfile = async () => {
    const validationErrorKey = validateSoraS3ProfileForm(soraProfileForm.value, {
      creating: creatingSoraProfile.value,
      selectedProfileID: selectedSoraProfileID.value
    })
    if (validationErrorKey) {
      options.showError(options.t(validationErrorKey))
      return
    }

    savingSoraProfile.value = true
    try {
      if (creatingSoraProfile.value) {
        const created = await adminAPI.settings.createSoraS3Profile(
          buildCreateSoraS3ProfileRequest(soraProfileForm.value)
        )
        selectedSoraProfileID.value = created.profile_id
        creatingSoraProfile.value = false
        soraProfileDrawerOpen.value = false
        options.showSuccess(options.t('admin.settings.soraS3.profileCreated'))
      } else {
        await adminAPI.settings.updateSoraS3Profile(
          selectedSoraProfileID.value,
          buildUpdateSoraS3ProfileRequest(soraProfileForm.value)
        )
        soraProfileDrawerOpen.value = false
        options.showSuccess(options.t('admin.settings.soraS3.profileSaved'))
      }

      await loadSoraS3Profiles()
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      savingSoraProfile.value = false
    }
  }

  const testSoraProfileConnection = async () => {
    testingSoraProfile.value = true
    try {
      const result = await adminAPI.settings.testSoraS3Connection(
        buildTestSoraS3ConnectionRequest(
          soraProfileForm.value,
          creatingSoraProfile.value ? undefined : selectedSoraProfileID.value
        )
      )
      options.showSuccess(result.message || options.t('admin.settings.soraS3.testSuccess'))
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      testingSoraProfile.value = false
    }
  }

  const activateSoraProfile = async (profileID: string) => {
    activatingSoraProfile.value = true
    try {
      await adminAPI.settings.setActiveSoraS3Profile(profileID)
      options.showSuccess(options.t('admin.settings.soraS3.profileActivated'))
      await loadSoraS3Profiles()
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      activatingSoraProfile.value = false
    }
  }

  const removeSoraProfile = async (profileID: string) => {
    if (!options.confirm(options.t('admin.settings.soraS3.deleteConfirm', { profileID }))) {
      return
    }

    deletingSoraProfile.value = true
    try {
      await adminAPI.settings.deleteSoraS3Profile(profileID)
      if (selectedSoraProfileID.value === profileID) {
        selectedSoraProfileID.value = ''
      }
      options.showSuccess(options.t('admin.settings.soraS3.profileDeleted'))
      await loadSoraS3Profiles()
    } catch (error) {
      options.showError(resolveRequestErrorMessage(error, options.t('errors.networkError')))
    } finally {
      deletingSoraProfile.value = false
    }
  }

  return {
    loadingSoraProfiles,
    savingSoraProfile,
    testingSoraProfile,
    activatingSoraProfile,
    deletingSoraProfile,
    creatingSoraProfile,
    soraProfileDrawerOpen,
    soraS3Profiles,
    selectedSoraProfileID,
    soraProfileForm,
    loadSoraS3Profiles,
    startCreateSoraProfile,
    editSoraProfile,
    closeSoraProfileDrawer,
    saveSoraProfile,
    testSoraProfileConnection,
    activateSoraProfile,
    removeSoraProfile
  }
}
