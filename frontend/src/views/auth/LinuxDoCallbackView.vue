<template>
  <AuthLayout>
    <div class="space-y-6">
      <div class="text-center">
        <h2 class="linuxdo-callback-view__title">
          {{ t('auth.linuxdo.callbackTitle') }}
        </h2>
        <p class="linuxdo-callback-view__description mt-2 text-sm">
          {{ isProcessing ? t('auth.linuxdo.callbackProcessing') : t('auth.linuxdo.callbackHint') }}
        </p>
      </div>

      <transition name="fade">
        <div v-if="needsInvitation" class="space-y-4">
          <p class="linuxdo-callback-view__body text-sm">
            {{ t('auth.linuxdo.invitationRequired') }}
          </p>
          <div>
            <input
              v-model="invitationCode"
              type="text"
              class="input w-full"
              :placeholder="t('auth.invitationCodePlaceholder')"
              :disabled="isSubmitting"
              @keyup.enter="handleSubmitInvitation"
            />
          </div>
          <transition name="fade">
            <p v-if="invitationError" class="linuxdo-callback-view__error-text text-sm">
              {{ invitationError }}
            </p>
          </transition>
          <button
            class="btn btn-primary w-full"
            :disabled="isSubmitting || !invitationCode.trim()"
            @click="handleSubmitInvitation"
          >
            {{ isSubmitting ? t('auth.linuxdo.completing') : t('auth.linuxdo.completeRegistration') }}
          </button>
        </div>
      </transition>

      <transition name="fade">
        <div
          v-if="errorMessage"
          class="linuxdo-callback-view__error-card"
        >
          <div class="flex items-start gap-3">
            <div class="flex-shrink-0">
              <Icon name="exclamationCircle" size="md" class="linuxdo-callback-view__error-icon" />
            </div>
            <div class="space-y-2">
              <p class="linuxdo-callback-view__error-text text-sm">
                {{ errorMessage }}
              </p>
              <router-link to="/login" class="btn btn-primary">
                {{ t('auth.linuxdo.backToLogin') }}
              </router-link>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </AuthLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { AuthLayout } from '@/components/layout'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore, useAppStore } from '@/stores'
import { completeLinuxDoOAuthRegistration } from '@/api/auth'
import { resolveErrorMessage } from '@/utils/errorMessage'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const isProcessing = ref(true)
const errorMessage = ref('')

// Invitation code flow state
const needsInvitation = ref(false)
const pendingOAuthToken = ref('')
const invitationCode = ref('')
const isSubmitting = ref(false)
const invitationError = ref('')
const redirectTo = ref('/dashboard')

function parseFragmentParams(): URLSearchParams {
  const raw = typeof window !== 'undefined' ? window.location.hash : ''
  const hash = raw.startsWith('#') ? raw.slice(1) : raw
  return new URLSearchParams(hash)
}

function sanitizeRedirectPath(path: string | null | undefined): string {
  if (!path) return '/dashboard'
  if (!path.startsWith('/')) return '/dashboard'
  if (path.startsWith('//')) return '/dashboard'
  if (path.includes('://')) return '/dashboard'
  if (path.includes('\n') || path.includes('\r')) return '/dashboard'
  return path
}

async function handleSubmitInvitation() {
  invitationError.value = ''
  if (!invitationCode.value.trim()) return

  isSubmitting.value = true
  try {
    const tokenData = await completeLinuxDoOAuthRegistration(
      pendingOAuthToken.value,
      invitationCode.value.trim()
    )
    if (tokenData.refresh_token) {
      localStorage.setItem('refresh_token', tokenData.refresh_token)
    }
    if (tokenData.expires_in) {
      localStorage.setItem('token_expires_at', String(Date.now() + tokenData.expires_in * 1000))
    }
    await authStore.setToken(tokenData.access_token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirectTo.value)
  } catch (e: unknown) {
    invitationError.value = resolveErrorMessage(e, t('auth.linuxdo.completeRegistrationFailed'))
  } finally {
    isSubmitting.value = false
  }
}

onMounted(async () => {
  const params = parseFragmentParams()

  const token = params.get('access_token') || ''
  const refreshToken = params.get('refresh_token') || ''
  const expiresInStr = params.get('expires_in') || ''
  const redirect = sanitizeRedirectPath(
    params.get('redirect') || (route.query.redirect as string | undefined) || '/dashboard'
  )
  const error = params.get('error')
  const errorDesc = params.get('error_description') || params.get('error_message') || ''

  if (error) {
    if (error === 'invitation_required') {
      pendingOAuthToken.value = params.get('pending_oauth_token') || ''
      redirectTo.value = sanitizeRedirectPath(params.get('redirect'))
      if (!pendingOAuthToken.value) {
        errorMessage.value = t('auth.linuxdo.invalidPendingToken')
        appStore.showError(errorMessage.value)
        isProcessing.value = false
        return
      }
      needsInvitation.value = true
      isProcessing.value = false
      return
    }
    errorMessage.value = errorDesc || error
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  if (!token) {
    errorMessage.value = t('auth.linuxdo.callbackMissingToken')
    appStore.showError(errorMessage.value)
    isProcessing.value = false
    return
  }

  try {
    // Store refresh token and expires_at (convert to timestamp) if provided
    if (refreshToken) {
      localStorage.setItem('refresh_token', refreshToken)
    }
    if (expiresInStr) {
      const expiresIn = parseInt(expiresInStr, 10)
      if (!isNaN(expiresIn)) {
        localStorage.setItem('token_expires_at', String(Date.now() + expiresIn * 1000))
      }
    }

    await authStore.setToken(token)
    appStore.showSuccess(t('auth.loginSuccess'))
    await router.replace(redirect)
  } catch (e: unknown) {
    errorMessage.value = resolveErrorMessage(e, t('auth.loginFailed'))
    appStore.showError(errorMessage.value)
    isProcessing.value = false
  }
})
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.linuxdo-callback-view__title {
  font-family: var(--theme-auth-section-title-font);
  font-size: var(--theme-auth-section-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-auth-section-title-letter-spacing);
  color: var(--theme-page-text);
}

.linuxdo-callback-view__description {
  color: var(--theme-page-muted);
}

.linuxdo-callback-view__body {
  color: color-mix(in srgb, var(--theme-page-text) 82%, transparent);
}

.linuxdo-callback-view__error-card {
  padding: var(--theme-auth-callback-feedback-padding);
  border-radius: var(--theme-auth-feedback-radius);
  border: 1px solid color-mix(in srgb, rgb(var(--theme-danger-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.linuxdo-callback-view__error-icon,
.linuxdo-callback-view__error-text {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
