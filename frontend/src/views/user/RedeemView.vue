<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <RedeemBalanceCard
        :balance-text="balanceText"
        :concurrency="user?.concurrency || 0"
        :concurrency-label="t('redeem.concurrency')"
        :current-balance-label="t('redeem.currentBalance')"
        :requests-label="t('redeem.requests')"
      />

      <div class="card">
        <div class="redeem-view__card-content">
          <form class="space-y-5" @submit.prevent="handleRedeem">
            <div>
              <label for="code" class="input-label">
                {{ t('redeem.redeemCodeLabel') }}
              </label>
              <div class="relative mt-1">
                <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                  <Icon name="gift" size="md" class="redeem-view__gift-icon" />
                </div>
                <input
                  id="code"
                  v-model="redeemCode"
                  type="text"
                  required
                  :placeholder="t('redeem.redeemCodePlaceholder')"
                  :disabled="submitting"
                  class="input redeem-view__code-input text-lg"
                />
              </div>
              <p class="input-hint">
                {{ t('redeem.redeemCodeHint') }}
              </p>
            </div>

            <button
              type="submit"
              :disabled="!redeemCode || submitting"
              class="btn btn-primary redeem-view__submit w-full"
            >
              <svg
                v-if="submitting"
                class="-ml-1 mr-2 h-5 w-5 animate-spin"
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
              <Icon v-else name="checkCircle" size="md" class="mr-2" />
              {{ submitting ? t('redeem.redeeming') : t('redeem.redeemButton') }}
            </button>
          </form>
        </div>
      </div>

      <transition name="fade">
        <div
          v-if="redeemResult"
          class="redeem-view__result-card redeem-view__result-card--success card"
        >
          <div class="redeem-view__card-content">
            <div class="flex items-start gap-4">
              <div
                class="redeem-view__result-icon-shell redeem-view__result-icon-shell--success flex flex-shrink-0 items-center justify-center"
              >
                <Icon name="checkCircle" size="md" class="redeem-view__result-tone redeem-view__result-tone--success" />
              </div>
              <div class="flex-1">
                <h3 class="redeem-view__result-title redeem-view__result-title--success text-sm font-semibold">
                  {{ t('redeem.redeemSuccess') }}
                </h3>
                <div class="redeem-view__result-body redeem-view__result-body--success mt-2 text-sm">
                  <p>{{ redeemResult.message }}</p>
                  <div class="mt-3 space-y-1">
                    <p v-if="redeemResult.type === 'balance'" class="font-medium">
                      {{ t('redeem.added') }}: ${{ redeemResult.value.toFixed(2) }}
                    </p>
                    <p v-else-if="redeemResult.type === 'concurrency'" class="font-medium">
                      {{ t('redeem.added') }}: {{ redeemResult.value }}
                      {{ t('redeem.concurrentRequests') }}
                    </p>
                    <p v-else-if="redeemResult.type === 'subscription'" class="font-medium">
                      {{ t('redeem.subscriptionAssigned') }}
                      <span v-if="redeemResult.group_name"> - {{ redeemResult.group_name }}</span>
                      <span v-if="redeemResult.validity_days">
                        ({{ t('redeem.subscriptionDays', { days: redeemResult.validity_days }) }})
                      </span>
                    </p>
                    <p v-if="redeemResult.new_balance !== undefined">
                      {{ t('redeem.newBalance') }}:
                      <span class="font-semibold">${{ redeemResult.new_balance.toFixed(2) }}</span>
                    </p>
                    <p v-if="redeemResult.new_concurrency !== undefined">
                      {{ t('redeem.newConcurrency') }}:
                      <span class="font-semibold">
                        {{ redeemResult.new_concurrency }} {{ t('redeem.requests') }}
                      </span>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <transition name="fade">
        <div
          v-if="errorMessage"
          class="redeem-view__result-card redeem-view__result-card--danger card"
        >
          <div class="redeem-view__card-content">
            <div class="flex items-start gap-4">
              <div
                class="redeem-view__result-icon-shell redeem-view__result-icon-shell--danger flex flex-shrink-0 items-center justify-center"
              >
                <Icon name="exclamationCircle" size="md" class="redeem-view__result-tone redeem-view__result-tone--danger" />
              </div>
              <div class="flex-1">
                <h3 class="redeem-view__result-title redeem-view__result-title--danger text-sm font-semibold">
                  {{ t('redeem.redeemFailed') }}
                </h3>
                <p class="redeem-view__result-body redeem-view__result-body--danger mt-2 text-sm">
                  {{ errorMessage }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <RedeemInfoCard
        :about-codes-label="t('redeem.aboutCodes')"
        :code-rule1="t('redeem.codeRule1')"
        :code-rule2="t('redeem.codeRule2')"
        :code-rule3="t('redeem.codeRule3')"
        :code-rule4="t('redeem.codeRule4')"
        :contact-info="contactInfo"
      />

      <RedeemHistoryList
        :admin-adjustment-label="t('redeem.adminAdjustment')"
        :empty-label="t('redeem.historyWillAppear')"
        :history="history"
        :loading="loadingHistory"
        :title="t('redeem.recentActivity')"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { redeemAPI, type RedeemHistoryItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useSubscriptionStore } from '@/stores/subscriptions'
import RedeemBalanceCard from './redeem/RedeemBalanceCard.vue'
import RedeemHistoryList from './redeem/RedeemHistoryList.vue'
import RedeemInfoCard from './redeem/RedeemInfoCard.vue'
import {
  formatRedeemBalance,
  resolveRedeemErrorMessage,
  type RedeemResultData
} from './redeem/redeemView'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const subscriptionStore = useSubscriptionStore()

const user = computed(() => authStore.user)
const balanceText = computed(() => formatRedeemBalance(user.value?.balance))

const redeemCode = ref('')
const submitting = ref(false)
const redeemResult = ref<RedeemResultData | null>(null)
const errorMessage = ref('')
const history = ref<RedeemHistoryItem[]>([])
const loadingHistory = ref(false)
const contactInfo = ref('')

const fetchHistory = async () => {
  loadingHistory.value = true

  try {
    history.value = await redeemAPI.getHistory()
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loadingHistory.value = false
  }
}

const handleRedeem = async () => {
  const trimmedCode = redeemCode.value.trim()
  if (!trimmedCode) {
    appStore.showError(t('redeem.pleaseEnterCode'))
    return
  }

  submitting.value = true
  errorMessage.value = ''
  redeemResult.value = null

  try {
    const result = await redeemAPI.redeem(trimmedCode)
    redeemResult.value = result

    await authStore.refreshUser()

    if (result.type === 'subscription') {
      try {
        await subscriptionStore.fetchActiveSubscriptions(true)
      } catch (error) {
        console.error('Failed to refresh subscriptions after redeem:', error)
        appStore.showWarning(t('redeem.subscriptionRefreshFailed'))
      }
    }

    redeemCode.value = ''
    await fetchHistory()
    appStore.showSuccess(t('redeem.codeRedeemSuccess'))
  } catch (error: unknown) {
    errorMessage.value = resolveRedeemErrorMessage(error, t('redeem.failedToRedeem'))
    appStore.showError(t('redeem.redeemFailed'))
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  void fetchHistory()

  const cachedContactInfo = appStore.cachedPublicSettings?.contact_info
  if (cachedContactInfo) {
    contactInfo.value = cachedContactInfo
    return
  }

  try {
    const settings = await appStore.fetchPublicSettings()
    contactInfo.value = settings?.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
})
</script>

<style scoped>
.redeem-view__gift-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.redeem-view__card-content {
  padding: var(--theme-redeem-card-padding);
}

.redeem-view__code-input {
  padding-block: var(--theme-redeem-code-input-padding-y);
  padding-inline-start: var(--theme-redeem-code-input-padding-start);
}

.redeem-view__submit {
  padding-block: var(--theme-redeem-submit-padding-y);
}

.redeem-view__result-card--success {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.redeem-view__result-card--danger {
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
}

.redeem-view__result-icon-shell--success {
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 14%, var(--theme-surface));
}

.redeem-view__result-icon-shell--danger {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 14%, var(--theme-surface));
}

.redeem-view__result-icon-shell {
  width: var(--theme-redeem-result-icon-size);
  height: var(--theme-redeem-result-icon-size);
  border-radius: var(--theme-redeem-result-icon-radius);
}

.redeem-view__result-tone--success,
.redeem-view__result-title--success,
.redeem-view__result-body--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.redeem-view__result-tone--danger,
.redeem-view__result-title--danger,
.redeem-view__result-body--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>

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
</style>
