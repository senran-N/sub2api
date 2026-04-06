<template>
  <BaseDialog
    :show="true"
    :title="t('profile.totp.disableTitle')"
    width="narrow"
    close-on-click-outside
    @close="emit('close')"
  >
    <div class="totp-disable-dialog space-y-6">
      <div class="text-center">
        <div
          class="totp-disable-dialog__icon-shell mx-auto flex h-12 w-12 items-center justify-center rounded-full"
        >
          <svg
            class="totp-disable-dialog__icon h-6 w-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="1.5"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"
            />
          </svg>
        </div>
        <p class="totp-disable-dialog__description mt-4 text-sm">
          {{ t('profile.totp.disableWarning') }}
        </p>
      </div>

      <div v-if="methodLoading" class="totp-disable-dialog__loading flex items-center justify-center">
        <div class="totp-disable-dialog__spinner h-8 w-8 animate-spin rounded-full border-b-2"></div>
      </div>

      <form v-else id="totp-disable-form" class="space-y-4" @submit.prevent="handleDisable">
        <div v-if="verificationMethod === 'email'">
          <label class="input-label">{{ t('profile.totp.emailCode') }}</label>
          <div class="flex gap-2">
            <input
              v-model="form.emailCode"
              type="text"
              maxlength="6"
              inputmode="numeric"
              class="input flex-1"
              :placeholder="t('profile.totp.enterEmailCode')"
            />
            <button
              type="button"
              class="btn btn-secondary whitespace-nowrap"
              :disabled="sendingCode || codeCooldown > 0"
              @click="handleSendCode"
            >
              {{ codeCooldown > 0 ? `${codeCooldown}s` : (sendingCode ? t('common.sending') : t('profile.totp.sendCode')) }}
            </button>
          </div>
        </div>

        <div v-else>
          <label for="password" class="input-label">
            {{ t('profile.currentPassword') }}
          </label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            class="input"
            :placeholder="t('profile.totp.enterPassword')"
          />
        </div>

        <div v-if="error" class="totp-disable-dialog__error text-sm">
          {{ error }}
        </div>
      </form>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="totp-disable-form"
          class="btn btn-danger"
          :disabled="methodLoading || loading || !canSubmit"
        >
          {{ loading ? t('common.processing') : t('profile.totp.confirmDisable') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { useAppStore } from '@/stores/app'
import { totpAPI } from '@/api'
import { resolveErrorMessage } from '@/utils/errorMessage'

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()

const methodLoading = ref(true)
const verificationMethod = ref<'email' | 'password'>('password')
const loading = ref(false)
const error = ref('')
const sendingCode = ref(false)
const codeCooldown = ref(0)
const cooldownTimer = ref<ReturnType<typeof setInterval> | null>(null)
const form = ref({
  emailCode: '',
  password: ''
})

const canSubmit = computed(() => {
  if (verificationMethod.value === 'email') {
    return form.value.emailCode.length === 6
  }
  return form.value.password.length > 0
})

const loadVerificationMethod = async () => {
  methodLoading.value = true
  try {
    const method = await totpAPI.getVerificationMethod()
    verificationMethod.value = method.method
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
    emit('close')
  } finally {
    methodLoading.value = false
  }
}

const handleSendCode = async () => {
  sendingCode.value = true
  try {
    await totpAPI.sendVerifyCode()
    appStore.showSuccess(t('profile.totp.codeSent'))
    // Start cooldown
    codeCooldown.value = 60
    if (cooldownTimer.value) {
      clearInterval(cooldownTimer.value)
      cooldownTimer.value = null
    }
    cooldownTimer.value = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) {
        if (cooldownTimer.value) {
          clearInterval(cooldownTimer.value)
          cooldownTimer.value = null
        }
      }
    }, 1000)
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('profile.totp.sendCodeFailed')))
  } finally {
    sendingCode.value = false
  }
}

const handleDisable = async () => {
  if (!canSubmit.value) return

  loading.value = true
  error.value = ''

  try {
    const request = verificationMethod.value === 'email'
      ? { email_code: form.value.emailCode }
      : { password: form.value.password }

    await totpAPI.disable(request)
    appStore.showSuccess(t('profile.totp.disableSuccess'))
    emit('success')
  } catch (disableError) {
    error.value = resolveErrorMessage(disableError, t('profile.totp.disableFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadVerificationMethod()
})

onUnmounted(() => {
  if (cooldownTimer.value) {
    clearInterval(cooldownTimer.value)
    cooldownTimer.value = null
  }
})
</script>

<style scoped>
.totp-disable-dialog__icon-shell {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.totp-disable-dialog__icon {
  color: rgb(var(--theme-danger-rgb));
}

.totp-disable-dialog__description {
  color: var(--theme-page-muted);
}

.totp-disable-dialog__loading {
  padding-block: var(--theme-profile-totp-loading-padding-y);
}

.totp-disable-dialog__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 64%, transparent);
  border-bottom-color: var(--theme-accent);
}

.totp-disable-dialog__error {
  border-radius: var(--theme-button-radius);
  padding: var(--theme-profile-totp-status-padding-y);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
