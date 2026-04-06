<template>
  <div class="card">
    <div class="profile-password-form__header border-b">
      <h2 class="theme-text-strong text-lg font-medium">
        {{ t('profile.changePassword') }}
      </h2>
    </div>
    <div class="profile-password-form__body">
      <form @submit.prevent="handleChangePassword" class="space-y-4">
        <div>
          <label for="old_password" class="input-label">
            {{ t('profile.currentPassword') }}
          </label>
          <input
            id="old_password"
            v-model="form.old_password"
            type="password"
            required
            autocomplete="current-password"
            class="input"
          />
        </div>

        <div>
          <label for="new_password" class="input-label">
            {{ t('profile.newPassword') }}
          </label>
          <input
            id="new_password"
            v-model="form.new_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
          <p class="input-hint">
            {{ t('profile.passwordHint') }}
          </p>
        </div>

        <div>
          <label for="confirm_password" class="input-label">
            {{ t('profile.confirmNewPassword') }}
          </label>
          <input
            id="confirm_password"
            v-model="form.confirm_password"
            type="password"
            required
            autocomplete="new-password"
            class="input"
          />
          <p
            v-if="form.new_password && form.confirm_password && form.new_password !== form.confirm_password"
            class="input-error-text"
          >
            {{ t('profile.passwordsNotMatch') }}
          </p>
        </div>

        <div class="flex justify-end pt-4">
          <button type="submit" :disabled="loading" class="btn btn-primary">
            {{ loading ? t('profile.changingPassword') : t('profile.changePasswordButton') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { userAPI } from '@/api'
import { resolveErrorMessage } from '@/utils/errorMessage'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const form = ref({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const handleChangePassword = async () => {
  if (form.value.new_password !== form.value.confirm_password) {
    appStore.showError(t('profile.passwordsNotMatch'))
    return
  }

  if (form.value.new_password.length < 8) {
    appStore.showError(t('profile.passwordTooShort'))
    return
  }

  loading.value = true
  try {
    await userAPI.changePassword(form.value.old_password, form.value.new_password)
    form.value = { old_password: '', new_password: '', confirm_password: '' }
    appStore.showSuccess(t('profile.passwordChangeSuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('profile.passwordChangeFailed')))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.profile-password-form__header {
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  padding: var(--theme-profile-totp-header-padding-y) var(--theme-profile-totp-header-padding-x);
}

.profile-password-form__body {
  padding: var(--theme-profile-totp-body-padding);
}
</style>
