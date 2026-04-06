<template>
  <div class="card">
    <div class="profile-edit-form__header border-b">
      <h2 class="theme-text-strong text-lg font-medium">
        {{ t('profile.editProfile') }}
      </h2>
    </div>
    <div class="profile-edit-form__body">
      <form @submit.prevent="handleUpdateProfile" class="space-y-4">
        <div>
          <label for="username" class="input-label">
            {{ t('profile.username') }}
          </label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="input"
            :placeholder="t('profile.enterUsername')"
          />
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? t('profile.updating') : t('profile.updateProfile') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'
import { resolveErrorMessage } from '@/utils/errorMessage'

const props = defineProps<{
  initialUsername: string
}>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const username = ref(props.initialUsername)
const loading = ref(false)

watch(() => props.initialUsername, (val) => {
  username.value = val
})

const handleUpdateProfile = async () => {
  if (!username.value.trim()) {
    appStore.showError(t('profile.usernameRequired'))
    return
  }

  loading.value = true
  try {
    const updatedUser = await userAPI.updateProfile({
      username: username.value
    })
    authStore.user = updatedUser
    appStore.showSuccess(t('profile.updateSuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('profile.updateFailed')))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.profile-edit-form__header {
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  padding: var(--theme-profile-totp-header-padding-y) var(--theme-profile-totp-header-padding-x);
}

.profile-edit-form__body {
  padding: var(--theme-profile-totp-body-padding);
}
</style>
