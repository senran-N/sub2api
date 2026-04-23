<template>
  <div class="card overflow-hidden">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.authBindings.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('profile.authBindings.description') }}
      </p>
    </div>

    <div class="space-y-5 px-6 py-5">
      <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-gray-900 dark:text-white">{{ t('profile.authBindings.emailLabel') }}</h3>
              <span :class="['badge', user?.email_bound ? 'badge-success' : 'badge-gray']">
                {{ user?.email_bound ? t('profile.authBindings.status.bound') : t('profile.authBindings.status.notBound') }}
              </span>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ user?.email || t('profile.authBindings.emailPlaceholder') }}
            </p>
          </div>
        </div>
        <div class="mt-4 grid gap-2 sm:grid-cols-[minmax(0,1.4fr)_auto]">
          <input v-model.trim="emailForm.email" type="email" class="input" :placeholder="t('profile.authBindings.emailPlaceholder')" :disabled="isSendingCode || isBindingEmail" />
          <button type="button" class="btn btn-secondary btn-sm" :disabled="isSendingCode || isBindingEmail || !emailForm.email" @click="handleSendCode">
            {{ isSendingCode ? t('common.loading') : t('profile.authBindings.sendCodeAction') }}
          </button>
          <input v-model.trim="emailForm.verifyCode" type="text" inputmode="numeric" maxlength="6" class="input" :placeholder="t('profile.authBindings.codePlaceholder')" :disabled="isBindingEmail" />
          <input v-model="emailForm.password" type="password" class="input" :placeholder="user?.email_bound ? t('profile.authBindings.replaceEmailPasswordPlaceholder') : t('profile.authBindings.passwordPlaceholder')" :disabled="isBindingEmail" />
          <button type="button" class="btn btn-primary btn-sm sm:col-span-2" :disabled="isBindingEmail || !canSubmitEmail" @click="handleBindEmail">
            {{ isBindingEmail ? t('common.loading') : (user?.email_bound ? t('profile.authBindings.confirmEmailReplaceAction') : t('profile.authBindings.confirmEmailBindAction')) }}
          </button>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-gray-900 dark:text-white">LinuxDo</h3>
              <span :class="['badge', user?.linuxdo_bound ? 'badge-success' : 'badge-gray']">
                {{ user?.linuxdo_bound ? t('profile.authBindings.status.bound') : t('profile.authBindings.status.notBound') }}
              </span>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('profile.authBindings.linuxdoDescription') }}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button v-if="linuxdoEnabled && !user?.linuxdo_bound" type="button" class="btn btn-primary btn-sm" @click="handleBindLinuxDo">
              {{ t('profile.authBindings.bindLinuxDoAction') }}
            </button>
            <button v-if="user?.linuxdo_bound" type="button" class="btn btn-secondary btn-sm" :disabled="isUnbindingLinuxDo" @click="handleUnbindLinuxDo">
              {{ isUnbindingLinuxDo ? t('common.loading') : t('profile.authBindings.unbindAction') }}
            </button>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-gray-900 dark:text-white">OIDC</h3>
              <span :class="['badge', user?.oidc_bound ? 'badge-success' : 'badge-gray']">
                {{ user?.oidc_bound ? t('profile.authBindings.status.bound') : t('profile.authBindings.status.notBound') }}
              </span>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('profile.authBindings.oidcDescription') }}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button v-if="oidcEnabled && !user?.oidc_bound" type="button" class="btn btn-primary btn-sm" @click="handleBindOIDC">
              {{ t('profile.authBindings.bindOIDCAction') }}
            </button>
            <button v-if="user?.oidc_bound" type="button" class="btn btn-secondary btn-sm" :disabled="isUnbindingOIDC" @click="handleUnbindOIDC">
              {{ isUnbindingOIDC ? t('common.loading') : t('profile.authBindings.unbindAction') }}
            </button>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-gray-100 p-4 dark:border-dark-700">
        <div class="flex items-start justify-between gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <h3 class="font-medium text-gray-900 dark:text-white">WeChat</h3>
              <span :class="['badge', user?.wechat_bound ? 'badge-success' : 'badge-gray']">
                {{ user?.wechat_bound ? t('profile.authBindings.status.bound') : t('profile.authBindings.status.notBound') }}
              </span>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('profile.authBindings.wechatDescription') }}
            </p>
          </div>
          <div class="flex shrink-0 gap-2">
            <button v-if="wechatEnabled && !user?.wechat_bound" type="button" class="btn btn-primary btn-sm" @click="handleBindWeChat">
              {{ t('profile.authBindings.bindWeChatAction') }}
            </button>
            <button v-if="user?.wechat_bound" type="button" class="btn btn-secondary btn-sm" :disabled="isUnbindingWeChat" @click="handleUnbindWeChat">
              {{ isUnbindingWeChat ? t('common.loading') : t('profile.authBindings.unbindAction') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  bindEmailIdentity,
  sendEmailBindingCode,
  startLinuxDoBinding,
  startOIDCBinding,
  startWeChatBinding,
  unbindAuthIdentity
} from '@/api/user'
import { useAppStore, useAuthStore } from '@/stores'
import type { User } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = withDefaults(
  defineProps<{
    user: User | null
    linuxdoEnabled?: boolean
    oidcEnabled?: boolean
    wechatEnabled?: boolean
  }>(),
  {
    linuxdoEnabled: false,
    oidcEnabled: false,
    wechatEnabled: false
  }
)

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const emailForm = reactive({
  email: props.user?.email_bound ? props.user?.email || '' : '',
  verifyCode: '',
  password: ''
})

const isSendingCode = ref(false)
const isBindingEmail = ref(false)
const isUnbindingLinuxDo = ref(false)
const isUnbindingOIDC = ref(false)
const isUnbindingWeChat = ref(false)
const canSubmitEmail = computed(() => !!emailForm.email && !!emailForm.verifyCode && !!emailForm.password)

watch(
  () => props.user,
  (user) => {
    if (user?.email_bound && user.email) {
      emailForm.email = user.email
    }
  },
  { immediate: true }
)

async function handleSendCode() {
  isSendingCode.value = true
  try {
    await sendEmailBindingCode(emailForm.email)
    appStore.showSuccess(t('auth.codeSentSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    isSendingCode.value = false
  }
}

async function handleBindEmail() {
  isBindingEmail.value = true
  try {
    const updated = await bindEmailIdentity({
      email: emailForm.email,
      verify_code: emailForm.verifyCode,
      password: emailForm.password
    })
    authStore.user = updated
    emailForm.verifyCode = ''
    emailForm.password = ''
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    isBindingEmail.value = false
  }
}

async function handleBindLinuxDo() {
  try {
    await startLinuxDoBinding('/profile')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function handleBindOIDC() {
  try {
    await startOIDCBinding('/profile')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function handleBindWeChat() {
  try {
    await startWeChatBinding('/profile')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function handleUnbindLinuxDo() {
  isUnbindingLinuxDo.value = true
  try {
    const updated = await unbindAuthIdentity('linuxdo')
    authStore.user = updated
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    isUnbindingLinuxDo.value = false
  }
}

async function handleUnbindOIDC() {
  isUnbindingOIDC.value = true
  try {
    const updated = await unbindAuthIdentity('oidc')
    authStore.user = updated
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    isUnbindingOIDC.value = false
  }
}

async function handleUnbindWeChat() {
  isUnbindingWeChat.value = true
  try {
    const updated = await unbindAuthIdentity('wechat')
    authStore.user = updated
    appStore.showSuccess(t('common.saved'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    isUnbindingWeChat.value = false
  }
}
</script>
