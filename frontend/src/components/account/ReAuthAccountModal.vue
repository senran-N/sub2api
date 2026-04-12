<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.reAuthorizeAccount')"
    width="normal"
    @close="handleClose"
  >
    <div v-if="account" class="space-y-4">
      <div class="re-auth-account-modal__account-card">
        <div class="flex items-center gap-3">
          <div :class="accountIconClass">
            <Icon name="sparkles" size="md" class="re-auth-account-modal__platform-icon-symbol" />
          </div>
          <div>
            <span class="re-auth-account-modal__account-name block font-semibold">{{ account.name }}</span>
            <span class="re-auth-account-modal__account-type text-sm">
              {{ accountTypeLabel }}
            </span>
          </div>
        </div>
      </div>

      <fieldset v-if="isAnthropic" class="re-auth-account-modal__method-fieldset">
        <legend class="input-label">{{ t('admin.accounts.oauth.authMethod') }}</legend>
        <div class="mt-2 flex gap-4">
          <label class="re-auth-account-modal__radio-row flex cursor-pointer items-center">
            <input
              v-model="addMethod"
              type="radio"
              value="oauth"
              class="re-auth-account-modal__radio-input mr-2"
            />
            <span class="re-auth-account-modal__radio-label text-sm">{{ t('admin.accounts.types.oauth') }}</span>
          </label>
          <label class="re-auth-account-modal__radio-row flex cursor-pointer items-center">
            <input
              v-model="addMethod"
              type="radio"
              value="setup-token"
              class="re-auth-account-modal__radio-input mr-2"
            />
            <span class="re-auth-account-modal__radio-label text-sm">{{ t('admin.accounts.setupTokenLongLived') }}</span>
          </label>
        </div>
      </fieldset>

      <div v-if="isGemini" class="re-auth-account-modal__oauth-type-card">
        <div class="re-auth-account-modal__oauth-type-label mb-2 text-sm font-medium">
          {{ t('admin.accounts.oauth.gemini.oauthTypeLabel') }}
        </div>
        <div class="flex items-center gap-3">
          <div :class="geminiOAuthTypeIconClass">
            <Icon v-if="geminiOAuthType === 'google_one'" name="user" size="sm" />
            <Icon v-else-if="geminiOAuthType === 'code_assist'" name="cloud" size="sm" />
            <Icon v-else name="sparkles" size="sm" />
          </div>
          <div>
            <span class="re-auth-account-modal__oauth-type-name block text-sm font-medium">
              {{ geminiOAuthTypeTitle }}
            </span>
            <span class="re-auth-account-modal__oauth-type-description text-xs">
              {{ geminiOAuthTypeDescription }}
            </span>
          </div>
        </div>
      </div>

      <OAuthAuthorizationFlow
        ref="oauthFlowRef"
        :add-method="addMethod"
        :auth-url="currentAuthUrl"
        :session-id="currentSessionId"
        :loading="currentLoading"
        :error="currentError"
        :show-help="isAnthropic"
        :show-proxy-warning="isAnthropic"
        :show-cookie-option="isAnthropic"
        :allow-multiple="false"
        :method-label="t('admin.accounts.inputMethod')"
        :platform="isOpenAI ? 'openai' : isGemini ? 'gemini' : isAntigravity ? 'antigravity' : 'anthropic'"
        :show-project-id="isGemini && geminiOAuthType === 'code_assist'"
        @generate-url="handleGenerateUrl"
        @cookie-auth="handleCookieAuth"
      />

    </div>

    <template #footer>
      <div v-if="account" class="flex justify-between gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.cancel') }}
        </button>
        <button
          v-if="isManualInputMethod"
          type="button"
          :disabled="!canExchangeCode"
          class="btn btn-primary"
          @click="handleExchangeCode"
        >
          <svg
            v-if="currentLoading"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{
            currentLoading
              ? t('admin.accounts.oauth.verifying')
              : t('admin.accounts.oauth.completeAuth')
          }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import {
  useAccountOAuth,
  type AddMethod,
  type AuthInputMethod
} from '@/composables/useAccountOAuth'
import { useOpenAIOAuth } from '@/composables/useOpenAIOAuth'
import { useGeminiOAuth } from '@/composables/useGeminiOAuth'
import { useAntigravityOAuth } from '@/composables/useAntigravityOAuth'
import type { Account } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import OAuthAuthorizationFlow from './OAuthAuthorizationFlow.vue'

interface OAuthFlowExposed {
  authCode: string
  oauthState: string
  projectId: string
  sessionKey: string
  inputMethod: AuthInputMethod
  reset: () => void
}

interface Props {
  show: boolean
  account: Account | null
}

type ReAuthTone = 'accent' | 'success' | 'info' | 'brand'

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
  reauthorized: [account?: Account]
}>()

const appStore = useAppStore()
const { t } = useI18n()

// OAuth composables
const claudeOAuth = useAccountOAuth()
const openaiOAuth = useOpenAIOAuth()
const geminiOAuth = useGeminiOAuth()
const antigravityOAuth = useAntigravityOAuth()

// Refs
const oauthFlowRef = ref<OAuthFlowExposed | null>(null)

// State
const addMethod = ref<AddMethod>('oauth')
const geminiOAuthType = ref<'code_assist' | 'google_one' | 'ai_studio'>('code_assist')

const isOpenAI = computed(() => props.account?.platform === 'openai')
const isGemini = computed(() => props.account?.platform === 'gemini')
const isAnthropic = computed(() => props.account?.platform === 'anthropic')
const isAntigravity = computed(() => props.account?.platform === 'antigravity')
const accountCredentials = computed(() => {
  return (props.account?.credentials ?? {}) as Record<string, unknown>
})
const accountTone = computed<ReAuthTone>(() => {
  if (isOpenAI.value) return 'success'
  if (isGemini.value) return 'info'
  if (isAntigravity.value) return 'brand'
  return 'accent'
})
const accountTypeLabel = computed(() => {
  if (isOpenAI.value) return t('admin.accounts.openaiAccount')
  if (isGemini.value) return t('admin.accounts.geminiAccount')
  if (isAntigravity.value) return t('admin.accounts.antigravityAccount')
  return t('admin.accounts.claudeCodeAccount')
})
const accountIconClass = computed(() => [
  're-auth-account-modal__platform-icon',
  `re-auth-account-modal__platform-icon--${accountTone.value}`
])
const geminiOAuthTypeTone = computed<ReAuthTone>(() => {
  if (geminiOAuthType.value === 'google_one') return 'brand'
  if (geminiOAuthType.value === 'code_assist') return 'info'
  return 'accent'
})
const geminiOAuthTypeIconClass = computed(() => [
  're-auth-account-modal__platform-icon',
  're-auth-account-modal__platform-icon--small',
  `re-auth-account-modal__platform-icon--${geminiOAuthTypeTone.value}`
])
const geminiOAuthTypeTitle = computed(() => {
  if (geminiOAuthType.value === 'google_one') return 'Google One'
  if (geminiOAuthType.value === 'code_assist') return t('admin.accounts.gemini.oauthType.builtInTitle')
  return t('admin.accounts.gemini.oauthType.customTitle')
})
const geminiOAuthTypeDescription = computed(() => {
  if (geminiOAuthType.value === 'google_one') return '个人账号'
  if (geminiOAuthType.value === 'code_assist') return t('admin.accounts.gemini.oauthType.builtInDesc')
  return t('admin.accounts.gemini.oauthType.customDesc')
})

const currentAuthUrl = computed(() => {
  if (isOpenAI.value) return openaiOAuth.authUrl.value
  if (isGemini.value) return geminiOAuth.authUrl.value
  if (isAntigravity.value) return antigravityOAuth.authUrl.value
  return claudeOAuth.authUrl.value
})
const currentSessionId = computed(() => {
  if (isOpenAI.value) return openaiOAuth.sessionId.value
  if (isGemini.value) return geminiOAuth.sessionId.value
  if (isAntigravity.value) return antigravityOAuth.sessionId.value
  return claudeOAuth.sessionId.value
})
const currentLoading = computed(() => {
  if (isOpenAI.value) return openaiOAuth.loading.value
  if (isGemini.value) return geminiOAuth.loading.value
  if (isAntigravity.value) return antigravityOAuth.loading.value
  return claudeOAuth.loading.value
})
const currentError = computed(() => {
  if (isOpenAI.value) return openaiOAuth.error.value
  if (isGemini.value) return geminiOAuth.error.value
  if (isAntigravity.value) return antigravityOAuth.error.value
  return claudeOAuth.error.value
})

const isManualInputMethod = computed(() => {
  return isOpenAI.value || isGemini.value || isAntigravity.value || oauthFlowRef.value?.inputMethod === 'manual'
})

const canExchangeCode = computed(() => {
  const authCode = oauthFlowRef.value?.authCode || ''
  const sessionId = currentSessionId.value
  const loading = currentLoading.value
  return authCode.trim() && sessionId && !loading
})

const getErrorMessage = (error: unknown, fallbackMessage: string) => {
  if (typeof error === 'object' && error !== null && 'response' in error) {
    const response = (error as { response?: { data?: { detail?: string } } }).response
    if (typeof response?.data?.detail === 'string' && response.data.detail.trim()) {
      return response.data.detail
    }
  }
  if (error instanceof Error && error.message.trim()) {
    return error.message
  }
  return fallbackMessage
}

watch(
  () => props.show,
  (newVal) => {
    if (newVal && props.account) {
      if (
        isAnthropic.value &&
        (props.account.type === 'oauth' || props.account.type === 'setup-token')
      ) {
        addMethod.value = props.account.type as AddMethod
      }
      if (isGemini.value) {
        geminiOAuthType.value =
          accountCredentials.value.oauth_type === 'google_one'
            ? 'google_one'
            : accountCredentials.value.oauth_type === 'ai_studio'
              ? 'ai_studio'
              : 'code_assist'
      }
    } else {
      resetState()
    }
  }
)

const resetState = () => {
  addMethod.value = 'oauth'
  geminiOAuthType.value = 'code_assist'
  claudeOAuth.resetState()
  openaiOAuth.resetState()
  geminiOAuth.resetState()
  antigravityOAuth.resetState()
  oauthFlowRef.value?.reset()
}

const handleClose = () => {
  emit('close')
}

const handleGenerateUrl = async () => {
  if (!props.account) return

  if (isOpenAI.value) {
    await openaiOAuth.generateAuthUrl(props.account.proxy_id)
  } else if (isGemini.value) {
    const creds = (props.account.credentials || {}) as Record<string, unknown>
    const tierId = typeof creds.tier_id === 'string' ? creds.tier_id : undefined
    const projectId = geminiOAuthType.value === 'code_assist' ? oauthFlowRef.value?.projectId : undefined
    await geminiOAuth.generateAuthUrl(props.account.proxy_id, projectId, geminiOAuthType.value, tierId)
  } else if (isAntigravity.value) {
    await antigravityOAuth.generateAuthUrl(props.account.proxy_id)
  } else {
    await claudeOAuth.generateAuthUrl(addMethod.value, props.account.proxy_id)
  }
}

const handleExchangeCode = async () => {
  if (!props.account) return

  const authCode = oauthFlowRef.value?.authCode || ''
  if (!authCode.trim()) return

  if (isOpenAI.value) {
    const oauthClient = openaiOAuth
    const sessionId = oauthClient.sessionId.value
    if (!sessionId) return
    const stateToUse = (oauthFlowRef.value?.oauthState || oauthClient.oauthState.value || '').trim()
    if (!stateToUse) {
      oauthClient.error.value = t('admin.accounts.oauth.authFailed')
      appStore.showError(oauthClient.error.value)
      return
    }

    const tokenInfo = await oauthClient.exchangeAuthCode(
      authCode.trim(),
      sessionId,
      stateToUse,
      props.account.proxy_id
    )
    if (!tokenInfo) return

    const credentials = oauthClient.buildCredentials(tokenInfo)
    const extra = oauthClient.buildExtraInfo(tokenInfo)

    try {
      await adminAPI.accounts.update(props.account.id, {
        type: 'oauth',
        credentials,
        extra
      })

      const updatedAccount = await adminAPI.accounts.clearError(props.account.id)

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error) {
      oauthClient.error.value = getErrorMessage(error, t('admin.accounts.oauth.authFailed'))
      appStore.showError(oauthClient.error.value)
    }
  } else if (isGemini.value) {
    const sessionId = geminiOAuth.sessionId.value
    if (!sessionId) return

    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || geminiOAuth.state.value
    if (!stateToUse) return

    const tokenInfo = await geminiOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId,
      state: stateToUse,
      proxyId: props.account.proxy_id,
      oauthType: geminiOAuthType.value,
      tierId: typeof accountCredentials.value.tier_id === 'string' ? accountCredentials.value.tier_id : undefined
    })
    if (!tokenInfo) return

    const credentials = geminiOAuth.buildCredentials(tokenInfo)

    try {
      await adminAPI.accounts.update(props.account.id, {
        type: 'oauth',
        credentials
      })
      const updatedAccount = await adminAPI.accounts.clearError(props.account.id)
      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error) {
      geminiOAuth.error.value = getErrorMessage(error, t('admin.accounts.oauth.authFailed'))
      appStore.showError(geminiOAuth.error.value)
    }
  } else if (isAntigravity.value) {
    const sessionId = antigravityOAuth.sessionId.value
    if (!sessionId) return

    const stateFromInput = oauthFlowRef.value?.oauthState || ''
    const stateToUse = stateFromInput || antigravityOAuth.state.value
    if (!stateToUse) return

    const tokenInfo = await antigravityOAuth.exchangeAuthCode({
      code: authCode.trim(),
      sessionId,
      state: stateToUse,
      proxyId: props.account.proxy_id
    })
    if (!tokenInfo) return

    const credentials = antigravityOAuth.buildCredentials(tokenInfo)

    try {
      await adminAPI.accounts.update(props.account.id, {
        type: 'oauth',
        credentials
      })
      const updatedAccount = await adminAPI.accounts.clearError(props.account.id)
      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error) {
      antigravityOAuth.error.value = getErrorMessage(error, t('admin.accounts.oauth.authFailed'))
      appStore.showError(antigravityOAuth.error.value)
    }
  } else {
    const sessionId = claudeOAuth.sessionId.value
    if (!sessionId) return

    claudeOAuth.loading.value = true
    claudeOAuth.error.value = ''

    try {
      const proxyConfig = props.account.proxy_id ? { proxy_id: props.account.proxy_id } : {}
      const endpoint =
        addMethod.value === 'oauth'
          ? '/admin/accounts/exchange-code'
          : '/admin/accounts/exchange-setup-token-code'

      const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
        session_id: sessionId,
        code: authCode.trim(),
        ...proxyConfig
      })

      const extra = claudeOAuth.buildExtraInfo(tokenInfo)

      await adminAPI.accounts.update(props.account.id, {
        type: addMethod.value,
        credentials: tokenInfo,
        extra
      })

      const updatedAccount = await adminAPI.accounts.clearError(props.account.id)

      appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
      emit('reauthorized', updatedAccount)
      handleClose()
    } catch (error) {
      claudeOAuth.error.value = getErrorMessage(error, t('admin.accounts.oauth.authFailed'))
      appStore.showError(claudeOAuth.error.value)
    } finally {
      claudeOAuth.loading.value = false
    }
  }
}

const handleCookieAuth = async (sessionKey: string) => {
  if (!props.account || isOpenAI.value) return

  claudeOAuth.loading.value = true
  claudeOAuth.error.value = ''

  try {
    const proxyConfig = props.account.proxy_id ? { proxy_id: props.account.proxy_id } : {}
    const endpoint =
      addMethod.value === 'oauth'
        ? '/admin/accounts/cookie-auth'
        : '/admin/accounts/setup-token-cookie-auth'

    const tokenInfo = await adminAPI.accounts.exchangeCode(endpoint, {
      session_id: '',
      code: sessionKey.trim(),
      ...proxyConfig
    })

    const extra = claudeOAuth.buildExtraInfo(tokenInfo)

    await adminAPI.accounts.update(props.account.id, {
      type: addMethod.value,
      credentials: tokenInfo,
      extra
    })

    const updatedAccount = await adminAPI.accounts.clearError(props.account.id)

    appStore.showSuccess(t('admin.accounts.reAuthorizedSuccess'))
    emit('reauthorized', updatedAccount)
    handleClose()
  } catch (error) {
    claudeOAuth.error.value = getErrorMessage(error, t('admin.accounts.oauth.cookieAuthFailed'))
    appStore.showError(claudeOAuth.error.value)
  } finally {
    claudeOAuth.loading.value = false
  }
}
</script>

<style scoped>
.re-auth-account-modal__account-card,
.re-auth-account-modal__oauth-type-card {
  padding: var(--theme-settings-card-panel-padding);
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 76%, transparent);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.re-auth-account-modal__method-fieldset {
  min-inline-size: 0;
  border: 0;
  padding: 0;
}

.re-auth-account-modal__platform-icon {
  --re-auth-tone-start: var(--theme-accent);
  --re-auth-tone-end: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-surface-contrast));
  display: flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border-radius: calc(var(--theme-surface-radius) + 2px);
  background: linear-gradient(135deg, var(--re-auth-tone-start), var(--re-auth-tone-end));
  color: var(--theme-filled-text);
}

.re-auth-account-modal__platform-icon--small {
  height: 2rem;
  width: 2rem;
  flex-shrink: 0;
}

.re-auth-account-modal__platform-icon--success {
  --re-auth-tone-start: rgb(var(--theme-success-rgb));
  --re-auth-tone-end: color-mix(in srgb, rgb(var(--theme-success-rgb)) 82%, var(--theme-surface-contrast));
}

.re-auth-account-modal__platform-icon--info {
  --re-auth-tone-start: rgb(var(--theme-info-rgb));
  --re-auth-tone-end: color-mix(in srgb, rgb(var(--theme-info-rgb)) 82%, var(--theme-surface-contrast));
}

.re-auth-account-modal__platform-icon--brand {
  --re-auth-tone-start: rgb(var(--theme-brand-purple-rgb));
  --re-auth-tone-end: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 82%, var(--theme-surface-contrast));
}

.re-auth-account-modal__account-name,
.re-auth-account-modal__radio-label,
.re-auth-account-modal__oauth-type-label,
.re-auth-account-modal__oauth-type-name {
  color: var(--theme-page-text);
}

.re-auth-account-modal__account-type,
.re-auth-account-modal__oauth-type-description {
  color: var(--theme-page-muted);
}

.re-auth-account-modal__radio-input {
  accent-color: var(--theme-accent);
}
</style>
