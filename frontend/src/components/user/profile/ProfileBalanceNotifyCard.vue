<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('profile.balanceNotify.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('profile.balanceNotify.description') }}
      </p>
    </div>

    <div class="space-y-6 px-6 py-6">
      <div class="flex items-center justify-between gap-4">
        <label class="input-label mb-0">{{ t('profile.balanceNotify.enabled') }}</label>
        <input v-model="notifyEnabled" type="checkbox" class="toggle" @change="saveEnabled" />
      </div>

      <template v-if="notifyEnabled">
        <div>
          <label class="input-label">{{ t('profile.balanceNotify.threshold') }}</label>
          <div class="flex gap-2">
            <input
              v-model.number="customThreshold"
              type="number"
              min="0"
              step="0.01"
              class="input flex-1"
              :placeholder="thresholdPlaceholder"
            />
            <button type="button" class="btn btn-primary" :disabled="savingThreshold" @click="saveThreshold">
              {{ savingThreshold ? t('common.saving') : t('common.save') }}
            </button>
          </div>
        </div>

        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('profile.balanceNotify.extraEmails') }}</label>
            <span class="text-xs text-gray-400">{{ t('profile.balanceNotify.maxEmails') }}</span>
          </div>

          <div v-if="entries.length > 0" class="space-y-2">
            <div
              v-for="entry in entries"
              :key="entry.email"
              class="flex items-center justify-between gap-3 rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-700"
            >
              <div class="min-w-0 flex-1">
                <div class="truncate text-sm text-gray-700 dark:text-gray-300">{{ entry.email }}</div>
                <div class="text-xs" :class="entry.verified ? 'text-green-500' : 'text-yellow-500'">
                  {{ entry.verified ? t('profile.balanceNotify.verified') : t('profile.balanceNotify.unverified') }}
                </div>
              </div>
              <div class="flex items-center gap-2">
                <input
                  :checked="!entry.disabled"
                  type="checkbox"
                  class="toggle"
                  :disabled="!entry.verified"
                  @change="toggleEntry(entry)"
                />
                <button
                  v-if="!entry.verified"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  @click="startVerifyExisting(entry.email)"
                >
                  {{ t('profile.balanceNotify.verify') }}
                </button>
                <button type="button" class="btn btn-danger btn-sm" @click="removeEntry(entry.email)">
                  {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>

          <div v-if="verifyingEmail" class="rounded-lg border border-yellow-200 bg-yellow-50 p-3 dark:border-yellow-900/60 dark:bg-yellow-950/20">
            <div class="mb-2 text-sm font-medium">{{ verifyingEmail }}</div>
            <div class="flex gap-2">
              <input v-model="verifyCode" maxlength="6" class="input flex-1" :placeholder="t('profile.balanceNotify.codePlaceholder')" />
              <button type="button" class="btn btn-primary" :disabled="sendingCode" @click="sendCodeForExisting">
                {{ sendingCode ? t('common.loading') : t('profile.balanceNotify.sendCode') }}
              </button>
              <button type="button" class="btn btn-secondary" :disabled="verifying" @click="verifyExisting">
                {{ verifying ? t('common.loading') : t('profile.balanceNotify.verify') }}
              </button>
            </div>
          </div>

          <div class="rounded-lg border border-dashed border-gray-300 p-3 dark:border-dark-500">
            <div class="grid gap-2 sm:grid-cols-[1fr_auto_auto]">
              <input v-model.trim="pendingEmail" type="email" class="input" :placeholder="t('profile.balanceNotify.emailPlaceholder')" />
              <button type="button" class="btn btn-secondary" :disabled="sendingCode || !pendingEmail" @click="sendCodeForNew">
                {{ sendingCode ? t('common.loading') : t('profile.balanceNotify.sendCode') }}
              </button>
              <button type="button" class="btn btn-primary" :disabled="verifying || !pendingEmail || verifyCode.length !== 6" @click="verifyNew">
                {{ verifying ? t('common.loading') : t('profile.balanceNotify.verify') }}
              </button>
            </div>
            <input v-model="verifyCode" maxlength="6" class="input mt-2" :placeholder="t('profile.balanceNotify.codePlaceholder')" />
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { userAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { NotifyEmailEntry } from '@/types'
import { resolveErrorMessage } from '@/utils/errorMessage'

const props = defineProps<{
  enabled: boolean
  threshold: number | null
  thresholdType?: string
  extraEmails: NotifyEmailEntry[]
  systemDefaultThreshold: number
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const notifyEnabled = ref(props.enabled)
const customThreshold = ref<number | null>(props.threshold)
const entries = ref<NotifyEmailEntry[]>([...props.extraEmails])
const pendingEmail = ref('')
const verifyCode = ref('')
const verifyingEmail = ref('')
const savingThreshold = ref(false)
const sendingCode = ref(false)
const verifying = ref(false)

watch(() => props.enabled, (value) => {
  notifyEnabled.value = value
})
watch(() => props.threshold, (value) => {
  customThreshold.value = value
})
watch(() => props.extraEmails, (value) => {
  entries.value = [...value]
})

const thresholdPlaceholder = computed(() =>
  props.systemDefaultThreshold > 0
    ? `${t('profile.balanceNotify.systemDefault')} $${props.systemDefaultThreshold}`
    : t('profile.balanceNotify.thresholdPlaceholder')
)

async function syncUser(update: Promise<any>) {
  const updatedUser = await update
  authStore.user = updatedUser
  entries.value = [...(updatedUser.balance_notify_extra_emails || [])]
  notifyEnabled.value = updatedUser.balance_notify_enabled
  customThreshold.value = updatedUser.balance_notify_threshold
}

async function saveEnabled() {
  try {
    await syncUser(userAPI.updateProfile({ balance_notify_enabled: notifyEnabled.value }))
    appStore.showSuccess(t('profile.updateSuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('profile.updateFailed')))
    notifyEnabled.value = props.enabled
  }
}

async function saveThreshold() {
  savingThreshold.value = true
  try {
    await syncUser(userAPI.updateProfile({ balance_notify_threshold: customThreshold.value }))
    appStore.showSuccess(t('profile.balanceNotify.thresholdSaved'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('profile.updateFailed')))
  } finally {
    savingThreshold.value = false
  }
}

async function toggleEntry(entry: NotifyEmailEntry) {
  try {
    await syncUser(userAPI.toggleNotifyEmail(entry.email, !entry.disabled))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  }
}

async function removeEntry(email: string) {
  try {
    await syncUser(userAPI.removeNotifyEmail(email))
    appStore.showSuccess(t('profile.balanceNotify.removeSuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  }
}

function startVerifyExisting(email: string) {
  verifyingEmail.value = email
  verifyCode.value = ''
}

async function sendCodeForExisting() {
  if (!verifyingEmail.value) return
  sendingCode.value = true
  try {
    await userAPI.sendNotifyEmailCode(verifyingEmail.value)
    appStore.showSuccess(t('profile.balanceNotify.codeSent'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  } finally {
    sendingCode.value = false
  }
}

async function verifyExisting() {
  if (!verifyingEmail.value || verifyCode.value.length !== 6) return
  verifying.value = true
  try {
    await syncUser(userAPI.verifyNotifyEmail(verifyingEmail.value, verifyCode.value))
    verifyingEmail.value = ''
    verifyCode.value = ''
    appStore.showSuccess(t('profile.balanceNotify.verifySuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  } finally {
    verifying.value = false
  }
}

async function sendCodeForNew() {
  if (!pendingEmail.value) return
  sendingCode.value = true
  try {
    await userAPI.sendNotifyEmailCode(pendingEmail.value)
    appStore.showSuccess(t('profile.balanceNotify.codeSent'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  } finally {
    sendingCode.value = false
  }
}

async function verifyNew() {
  if (!pendingEmail.value || verifyCode.value.length !== 6) return
  verifying.value = true
  try {
    await syncUser(userAPI.verifyNotifyEmail(pendingEmail.value, verifyCode.value))
    pendingEmail.value = ''
    verifyCode.value = ''
    appStore.showSuccess(t('profile.balanceNotify.verifySuccess'))
  } catch (error) {
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  } finally {
    verifying.value = false
  }
}
</script>
