<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <SettingsLoadingState v-if="loading" />

      <form v-else @submit.prevent="saveSettings" class="space-y-6" novalidate>
        <SettingsTabsNav
          :active-tab="activeTab"
          :tabs="settingsTabs"
          @update:active-tab="setActiveTab"
        />

        <div v-if="activeTab === 'security'" class="space-y-6">
          <SettingsAdminApiKeyCard
            :loading="adminApiKeyLoading"
            :exists="adminApiKeyExists"
            :masked-key="adminApiKeyMasked"
            :operating="adminApiKeyOperating"
            :new-key="newAdminApiKey"
            @create="createAdminApiKey"
            @regenerate="regenerateAdminApiKey"
            @delete="deleteAdminApiKey"
            @copy="copyNewKey"
          />
        </div>

        <div v-if="activeTab === 'gateway'" class="space-y-6">
          <SettingsOverloadCooldownCard
            :loading="overloadCooldownLoading"
            :saving="overloadCooldownSaving"
            :form="overloadCooldownForm"
            @save="saveOverloadCooldownSettings"
          />

          <SettingsStreamTimeoutCard
            :loading="streamTimeoutLoading"
            :saving="streamTimeoutSaving"
            :form="streamTimeoutForm"
            @save="saveStreamTimeoutSettings"
          />

          <SettingsRectifierCard
            :loading="rectifierLoading"
            :saving="rectifierSaving"
            :form="rectifierForm"
            @save="saveRectifierSettings"
          />

          <SettingsBetaPolicyCard
            :loading="betaPolicyLoading"
            :saving="betaPolicySaving"
            :rules="betaPolicyForm.rules"
            :action-options="betaPolicyActionOptions"
            :scope-options="betaPolicyScopeOptions"
            :get-display-name="getBetaDisplayName"
            @save="saveBetaPolicySettings"
          />
        </div>

        <div v-if="activeTab === 'security'" class="space-y-6">
          <SettingsRegistrationCard
            :form="form"
            :tags="registrationEmailSuffixWhitelistTags"
            v-model:draft="registrationEmailSuffixWhitelistDraft"
            @remove-tag="removeRegistrationEmailSuffixWhitelistTag"
            @draft-input="handleRegistrationEmailSuffixWhitelistDraftInput"
            @draft-keydown="handleRegistrationEmailSuffixWhitelistDraftKeydown"
            @commit-draft="commitRegistrationEmailSuffixWhitelistDraft"
            @draft-paste="handleRegistrationEmailSuffixWhitelistPaste"
          />

          <SettingsTurnstileCard :form="form" />

          <SettingsLinuxdoCard
            :form="form"
            :redirect-url-suggestion="linuxdoRedirectUrlSuggestion"
            @quick-set-copy="setAndCopyLinuxdoRedirectUrl"
          />

          <SettingsWechatCard
            :form="form"
            :redirect-url-suggestion="wechatRedirectUrlSuggestion"
            @quick-set-copy="setAndCopyWeChatRedirectUrl"
          />

          <SettingsOidcCard
            :form="form"
            :redirect-url-suggestion="oidcRedirectUrlSuggestion"
            @quick-set-copy="setAndCopyOidcRedirectUrl"
          />
        </div>

        <div v-if="activeTab === 'users'" class="space-y-6">
          <SettingsDefaultsCard
            :form="form"
            :default-subscription-group-options="defaultSubscriptionGroupOptions"
            :to-default-subscription-group-option="toDefaultSubscriptionGroupOption"
            @add-default-subscription="addDefaultSubscription"
            @remove-default-subscription="removeDefaultSubscription"
          />

          <SettingsAuthSourceDefaultsCard
            :form="form"
            :default-subscription-group-options="defaultSubscriptionGroupOptions"
            :to-default-subscription-group-option="toDefaultSubscriptionGroupOption"
            @add-auth-source-default-subscription="addAuthSourceDefaultSubscription"
            @remove-auth-source-default-subscription="removeAuthSourceDefaultSubscription"
          />
        </div>

        <div v-if="activeTab === 'gateway'" class="space-y-6">
          <SettingsModelRoutingCard :form="form" />

          <SettingsClaudeCodeCard :form="form" />

          <SettingsSchedulingCard :form="form" />

          <SettingsGatewayForwardingCard :form="form" />
        </div>

        <div v-if="activeTab === 'general'" class="space-y-6">
          <SettingsSiteCard
            :form="form"
            @add-endpoint="addEndpoint"
            @remove-endpoint="removeEndpoint"
          />
          <SettingsPurchaseCard :form="form" />

          <SettingsCustomMenuCard
            :form="form"
            @add-item="addMenuItem"
            @remove-item="removeMenuItem"
            @move-item="moveMenuItem"
          />
        </div>

        <div v-if="activeTab === 'payment'" class="space-y-6">
          <div class="card">
            <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.settings.payment.title') }}
              </h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.payment.description') }}
              </p>
            </div>

            <div class="space-y-6 p-6">
              <div v-if="paymentConfigLoading" class="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('common.loading') }}
              </div>

              <template v-else>
                <div class="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3 dark:border-dark-700">
                  <div>
                    <label class="font-medium text-gray-900 dark:text-white">
                      {{ t('admin.settings.payment.enabled') }}
                    </label>
                    <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                      {{ t('admin.settings.payment.enabledHint') }}
                    </p>
                  </div>
                  <button
                    type="button"
                    :class="[
                      'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out',
                      paymentConfig.enabled ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                    ]"
                    @click="paymentConfig.enabled = !paymentConfig.enabled"
                  >
                    <span
                      :class="[
                        'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        paymentConfig.enabled ? 'translate-x-5' : 'translate-x-0'
                      ]"
                    />
                  </button>
                </div>

                <template v-if="paymentConfig.enabled">
                  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.productNamePrefix') }}</label>
                      <input v-model="paymentConfig.product_name_prefix" type="text" class="input" placeholder="Sub2API" />
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.productNameSuffix') }}</label>
                      <input v-model="paymentConfig.product_name_suffix" type="text" class="input" placeholder="CNY" />
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.preview') }}</label>
                      <div class="input flex items-center bg-gray-50 text-sm text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                        {{ paymentPreviewName }}
                      </div>
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.minAmount') }}</label>
                      <input v-model.number="paymentConfig.min_amount" type="number" step="0.01" min="0" class="input" :placeholder="t('admin.settings.payment.noLimit')" />
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.maxAmount') }}</label>
                      <input v-model.number="paymentConfig.max_amount" type="number" step="0.01" min="0" class="input" :placeholder="t('admin.settings.payment.noLimit')" />
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.dailyLimit') }}</label>
                      <input v-model.number="paymentConfig.daily_limit" type="number" step="0.01" min="0" class="input" :placeholder="t('admin.settings.payment.noLimit')" />
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.balanceRechargeMultiplier') }}</label>
                      <input v-model.number="paymentConfig.balance_recharge_multiplier" type="number" step="0.01" min="0.01" class="input" />
                      <p class="mt-1 text-xs text-primary-600 dark:text-primary-400">
                        {{ t('admin.settings.payment.balanceRechargePreview', { usd: paymentMultiplierPreview }) }}
                      </p>
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.rechargeFeeRate') }}</label>
                      <input v-model.number="paymentConfig.recharge_fee_rate" type="number" step="0.01" min="0" max="100" class="input" />
                      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                        {{ t('admin.settings.payment.rechargeFeeRateHint') }}
                      </p>
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.orderTimeout') }}</label>
                      <input v-model.number="paymentConfig.order_timeout_minutes" type="number" min="1" class="input" />
                      <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                        {{ t('admin.settings.payment.orderTimeoutHint') }}
                      </p>
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.maxPendingOrders') }}</label>
                      <input v-model.number="paymentConfig.max_pending_orders" type="number" min="1" class="input" />
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.loadBalanceStrategy') }}</label>
                      <select v-model="paymentConfig.load_balance_strategy" class="input">
                        <option
                          v-for="option in loadBalanceOptions"
                          :key="option.value"
                          :value="option.value"
                        >
                          {{ option.label }}
                        </option>
                      </select>
                    </div>
                    <div class="rounded-lg border border-gray-100 px-4 py-3 dark:border-dark-700">
                      <div class="flex items-center justify-between gap-4">
                        <div>
                          <label class="font-medium text-gray-900 dark:text-white">
                            {{ t('admin.settings.payment.balancePaymentDisabled') }}
                          </label>
                          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {{ t('admin.settings.payment.balanceRechargeMultiplierHint') }}
                          </p>
                        </div>
                        <button
                          type="button"
                          :class="[
                            'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out',
                            paymentConfig.balance_disabled ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                          ]"
                          @click="paymentConfig.balance_disabled = !paymentConfig.balance_disabled"
                        >
                          <span
                            :class="[
                              'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                              paymentConfig.balance_disabled ? 'translate-x-5' : 'translate-x-0'
                            ]"
                          />
                        </button>
                      </div>
                    </div>
                  </div>

                  <div class="space-y-3 rounded-lg border border-gray-100 px-4 py-4 dark:border-dark-700">
                    <div class="flex items-center justify-between gap-4">
                      <div>
                        <label class="font-medium text-gray-900 dark:text-white">
                          {{ t('admin.settings.payment.cancelRateLimit') }}
                        </label>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                          {{ t('admin.settings.payment.cancelRateLimitHint') }}
                        </p>
                      </div>
                      <button
                        type="button"
                        :class="[
                          'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out',
                          paymentConfig.cancel_rate_limit_enabled ? 'bg-primary-500' : 'bg-gray-300 dark:bg-dark-600'
                        ]"
                        @click="paymentConfig.cancel_rate_limit_enabled = !paymentConfig.cancel_rate_limit_enabled"
                      >
                        <span
                          :class="[
                            'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                            paymentConfig.cancel_rate_limit_enabled ? 'translate-x-5' : 'translate-x-0'
                          ]"
                        />
                      </button>
                    </div>

                    <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
                      <div>
                        <label class="input-label">{{ t('admin.settings.payment.cancelRateLimitWindowMode') }}</label>
                        <select v-model="paymentConfig.cancel_rate_limit_window_mode" class="input" :disabled="!paymentConfig.cancel_rate_limit_enabled">
                          <option
                            v-for="option in cancelRateLimitModeOptions"
                            :key="option.value"
                            :value="option.value"
                          >
                            {{ option.label }}
                          </option>
                        </select>
                      </div>
                      <div>
                        <label class="input-label">{{ t('admin.settings.payment.cancelRateLimitWindow') }}</label>
                        <input v-model.number="paymentConfig.cancel_rate_limit_window" type="number" min="1" class="input" :disabled="!paymentConfig.cancel_rate_limit_enabled" />
                      </div>
                      <div>
                        <label class="input-label">{{ t('admin.settings.payment.cancelRateLimitUnit') }}</label>
                        <select v-model="paymentConfig.cancel_rate_limit_unit" class="input" :disabled="!paymentConfig.cancel_rate_limit_enabled">
                          <option
                            v-for="option in cancelRateLimitUnitOptions"
                            :key="option.value"
                            :value="option.value"
                          >
                            {{ option.label }}
                          </option>
                        </select>
                      </div>
                      <div>
                        <label class="input-label">{{ t('admin.settings.payment.cancelRateLimitMax') }}</label>
                        <input v-model.number="paymentConfig.cancel_rate_limit_max" type="number" min="1" class="input" :disabled="!paymentConfig.cancel_rate_limit_enabled" />
                      </div>
                    </div>
                  </div>

                  <div>
                    <label class="input-label">{{ t('admin.settings.payment.enabledPaymentTypes') }}</label>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <button
                        v-for="paymentType in allPaymentTypes"
                        :key="paymentType.value"
                        type="button"
                        @click="togglePaymentType(paymentType.value)"
                        :class="[
                          'rounded-lg border px-3 py-1.5 text-sm font-medium transition-all',
                          isPaymentTypeEnabled(paymentType.value)
                            ? 'border-primary-500 bg-primary-500 text-white shadow-sm'
                            : 'border-gray-300 bg-white text-gray-600 hover:border-gray-400 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:border-dark-500'
                        ]"
                      >
                        {{ paymentType.label }}
                      </button>
                    </div>
                    <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                      {{ t('admin.settings.payment.enabledPaymentTypesHint') }}
                      <a
                        :href="locale === 'zh'
                          ? 'https://github.com/senran-N/sub2api/blob/main/docs/PAYMENT_CN.md#%E6%94%AF%E6%8C%81%E7%9A%84%E6%94%AF%E4%BB%98%E6%96%B9%E5%BC%8F'
                          : 'https://github.com/senran-N/sub2api/blob/main/docs/PAYMENT.md#supported-payment-methods'"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="ml-1 text-primary-500 hover:text-primary-600 dark:text-primary-400 dark:hover:text-primary-300"
                      >
                        {{ t('admin.settings.payment.findProvider') }}
                      </a>
                    </p>
                  </div>

                  <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.helpImage') }}</label>
                      <ImageUpload
                        v-model="paymentConfig.help_image_url"
                        :upload-label="t('admin.settings.site.uploadImage')"
                        :remove-label="t('admin.settings.site.remove')"
                        :hint="t('admin.settings.payment.helpImagePlaceholder')"
                      />
                    </div>
                    <div>
                      <label class="input-label">{{ t('admin.settings.payment.helpText') }}</label>
                      <textarea
                        v-model="paymentConfig.help_text"
                        rows="2"
                        class="input resize-none"
                        :placeholder="t('admin.settings.payment.helpTextPlaceholder')"
                      ></textarea>
                    </div>
                  </div>

                  <div class="flex justify-end">
                    <button
                      type="button"
                      class="btn btn-primary"
                      :disabled="paymentConfigSaving"
                      @click="savePaymentConfig"
                    >
                      {{ paymentConfigSaving ? t('admin.settings.saving') : t('admin.settings.saveSettings') }}
                    </button>
                  </div>
                </template>
              </template>
            </div>
          </div>

          <PaymentProviderList
            v-if="paymentConfig.enabled"
            :providers="providers"
            :loading="providersLoading"
            :can-create="hasAnyPaymentTypeEnabled"
            :enabled-payment-types="paymentConfig.enabled_payment_types"
            :all-payment-types="allPaymentTypes"
            :redirect-label="t('admin.settings.payment.easypayRedirect')"
            @refresh="loadProviders"
            @create="openCreateProvider"
            @edit="openEditProvider"
            @delete="confirmDeleteProvider"
            @toggle-field="handleToggleField"
            @toggle-type="handleToggleType"
            @reorder="handleReorderProviders"
          />
        </div>

        <div v-if="activeTab === 'email'" class="space-y-6">
          <SettingsEmailDisabledCard v-if="!form.email_verify_enabled" />

          <SettingsSmtpCard
            v-if="form.email_verify_enabled"
            :form="form"
            :testing="testingSmtp"
            :disabled="loadFailed"
            @test-connection="testSmtpConnection"
            @password-interaction="smtpPasswordManuallyEdited = true"
          />

          <SettingsTestEmailCard
            v-if="form.email_verify_enabled"
            v-model="testEmailAddress"
            :sending="sendingTestEmail"
            :disabled="loadFailed"
            @send="sendTestEmail"
          />

          <SettingsNotifyCard :form="form" />
        </div>

        <div v-if="activeTab === 'backup'">
          <BackupSettings />
        </div>

        <SettingsSaveBar
          v-if="activeTab !== 'backup' && activeTab !== 'payment'"
          :saving="saving"
          :disabled="loadFailed"
        />
      </form>

      <PaymentProviderDialog
        ref="providerDialogRef"
        :show="showProviderDialog"
        :saving="providerSaving"
        :editing="editingProvider"
        :all-key-options="providerKeyOptions"
        :enabled-key-options="enabledProviderKeyOptions"
        :all-payment-types="allPaymentTypes"
        :redirect-label="t('admin.settings.payment.easypayRedirect')"
        @close="showProviderDialog = false"
        @save="handleSaveProvider"
      />

      <ConfirmDialog
        :show="showDeleteProviderDialog"
        :title="t('admin.settings.payment.deleteProvider')"
        :message="t('admin.settings.payment.deleteProviderConfirm')"
        :confirm-text="t('common.delete')"
        danger
        @confirm="handleDeleteProvider"
        @cancel="showDeleteProviderDialog = false"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import ImageUpload from '@/components/common/ImageUpload.vue'
import PaymentProviderDialog from '@/components/payment/PaymentProviderDialog.vue'
import PaymentProviderList from '@/components/payment/PaymentProviderList.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import type { GroupPlatform, SubscriptionType } from '@/types'
import type { AdminPaymentConfig } from '@/api/admin/payment'
import type { ProviderInstance } from '@/types/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import SettingsAdminApiKeyCard from './settings/SettingsAdminApiKeyCard.vue'
import SettingsBetaPolicyCard from './settings/SettingsBetaPolicyCard.vue'
import SettingsClaudeCodeCard from './settings/SettingsClaudeCodeCard.vue'
import SettingsCustomMenuCard from './settings/SettingsCustomMenuCard.vue'
import SettingsDefaultsCard from './settings/SettingsDefaultsCard.vue'
import SettingsAuthSourceDefaultsCard from './settings/SettingsAuthSourceDefaultsCard.vue'
import SettingsEmailDisabledCard from './settings/SettingsEmailDisabledCard.vue'
import SettingsGatewayForwardingCard from './settings/SettingsGatewayForwardingCard.vue'
import SettingsLinuxdoCard from './settings/SettingsLinuxdoCard.vue'
import SettingsWechatCard from './settings/SettingsWechatCard.vue'
import SettingsOidcCard from './settings/SettingsOidcCard.vue'
import SettingsModelRoutingCard from './settings/SettingsModelRoutingCard.vue'
import SettingsOverloadCooldownCard from './settings/SettingsOverloadCooldownCard.vue'
import SettingsPurchaseCard from './settings/SettingsPurchaseCard.vue'
import SettingsRegistrationCard from './settings/SettingsRegistrationCard.vue'
import SettingsRectifierCard from './settings/SettingsRectifierCard.vue'
import SettingsSchedulingCard from './settings/SettingsSchedulingCard.vue'
import SettingsSaveBar from './settings/SettingsSaveBar.vue'
import SettingsNotifyCard from './settings/SettingsNotifyCard.vue'
import SettingsSiteCard from './settings/SettingsSiteCard.vue'
import SettingsSmtpCard from './settings/SettingsSmtpCard.vue'
import SettingsStreamTimeoutCard from './settings/SettingsStreamTimeoutCard.vue'
import SettingsTestEmailCard from './settings/SettingsTestEmailCard.vue'
import SettingsLoadingState from './settings/SettingsLoadingState.vue'
import SettingsTabsNav from './settings/SettingsTabsNav.vue'
import SettingsTurnstileCard from './settings/SettingsTurnstileCard.vue'
import { useSettingsViewForm } from './settings/useSettingsViewForm'
import { useSettingsViewPolicies } from './settings/useSettingsViewPolicies'

const BackupSettings = defineAsyncComponent(() => import('@/views/admin/BackupView.vue'))

const { t, locale } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()

type SettingsTab = 'general' | 'security' | 'users' | 'gateway' | 'payment' | 'email' | 'backup'

interface DefaultSubscriptionGroupOptionView {
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

const activeTab = ref<SettingsTab>('general')
const settingsTabs = [
  { key: 'general'  as SettingsTab, icon: 'home'   as const },
  { key: 'security' as SettingsTab, icon: 'shield' as const },
  { key: 'users'    as SettingsTab, icon: 'user'   as const },
  { key: 'gateway'  as SettingsTab, icon: 'server' as const },
  { key: 'payment'  as SettingsTab, icon: 'creditCard' as const },
  { key: 'email'    as SettingsTab, icon: 'mail'   as const },
  { key: 'backup'   as SettingsTab, icon: 'database' as const },
]
const { copyToClipboard } = useClipboard()

function createDefaultPaymentConfig(): AdminPaymentConfig {
  return {
    enabled: false,
    min_amount: 1,
    max_amount: 0,
    daily_limit: 0,
    order_timeout_minutes: 30,
    max_pending_orders: 3,
    enabled_payment_types: [],
    balance_disabled: false,
    balance_recharge_multiplier: 1,
    recharge_fee_rate: 0,
    load_balance_strategy: 'round-robin',
    product_name_prefix: '',
    product_name_suffix: '',
    help_image_url: '',
    help_text: '',
    cancel_rate_limit_enabled: false,
    cancel_rate_limit_max: 10,
    cancel_rate_limit_window: 1,
    cancel_rate_limit_unit: 'day',
    cancel_rate_limit_window_mode: 'rolling',
  }
}

const paymentConfigLoading = ref(false)
const paymentConfigSaving = ref(false)
const paymentConfig = reactive<AdminPaymentConfig>(createDefaultPaymentConfig())
const lastSavedEnabledPaymentTypes = ref<string[]>([])
const providersLoading = ref(false)
const providerSaving = ref(false)
const providers = ref<ProviderInstance[]>([])
const showProviderDialog = ref(false)
const showDeleteProviderDialog = ref(false)
const editingProvider = ref<ProviderInstance | null>(null)
const deletingProviderId = ref<number | null>(null)
const providerDialogRef = ref<InstanceType<typeof PaymentProviderDialog> | null>(null)

const paymentPreviewName = computed(() => {
  const prefix = paymentConfig.product_name_prefix || 'Sub2API'
  const suffix = paymentConfig.product_name_suffix || 'CNY'
  return `${prefix} 100 ${suffix}`
})

const paymentMultiplierPreview = computed(() =>
  (Number(paymentConfig.balance_recharge_multiplier) || 1).toFixed(2)
)

const allPaymentTypes = computed(() => [
  { value: 'easypay', label: t('payment.methods.easypay') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
])

const hasAnyPaymentTypeEnabled = computed(() => paymentConfig.enabled_payment_types.length > 0)

const providerKeyOptions = computed(() => [
  { value: 'easypay', label: t('admin.settings.payment.providerEasypay') },
  { value: 'alipay', label: t('admin.settings.payment.providerAlipay') },
  { value: 'wxpay', label: t('admin.settings.payment.providerWxpay') },
  { value: 'stripe', label: t('admin.settings.payment.providerStripe') },
])

const enabledProviderKeyOptions = computed(() =>
  providerKeyOptions.value.filter((option) =>
    paymentConfig.enabled_payment_types.includes(option.value)
  )
)

const loadBalanceOptions = computed(() => [
  { value: 'round-robin', label: t('admin.settings.payment.strategyRoundRobin') },
  { value: 'least-amount', label: t('admin.settings.payment.strategyLeastAmount') },
])

const cancelRateLimitUnitOptions = computed(() => [
  { value: 'minute', label: t('admin.settings.payment.cancelRateLimitUnitMinute') },
  { value: 'hour', label: t('admin.settings.payment.cancelRateLimitUnitHour') },
  { value: 'day', label: t('admin.settings.payment.cancelRateLimitUnitDay') },
])

const cancelRateLimitModeOptions = computed(() => [
  { value: 'rolling', label: t('admin.settings.payment.cancelRateLimitWindowModeRolling') },
  { value: 'fixed', label: t('admin.settings.payment.cancelRateLimitWindowModeFixed') },
])

function setActiveTab(tab: string) {
  activeTab.value = tab as SettingsTab
}

function toDefaultSubscriptionGroupOption(option: unknown): DefaultSubscriptionGroupOptionView {
  return option as DefaultSubscriptionGroupOptionView
}

const {
  loading,
  loadFailed,
  saving,
  testingSmtp,
  sendingTestEmail,
  smtpPasswordManuallyEdited,
  testEmailAddress,
  registrationEmailSuffixWhitelistTags,
  registrationEmailSuffixWhitelistDraft,
  form,
  defaultSubscriptionGroupOptions,
  linuxdoRedirectUrlSuggestion,
  wechatRedirectUrlSuggestion,
  oidcRedirectUrlSuggestion,
  removeRegistrationEmailSuffixWhitelistTag,
  commitRegistrationEmailSuffixWhitelistDraft,
  handleRegistrationEmailSuffixWhitelistDraftInput,
  handleRegistrationEmailSuffixWhitelistDraftKeydown,
  handleRegistrationEmailSuffixWhitelistPaste,
  setAndCopyLinuxdoRedirectUrl,
  setAndCopyWeChatRedirectUrl,
  setAndCopyOidcRedirectUrl,
  addMenuItem,
  removeMenuItem,
  moveMenuItem,
  addEndpoint,
  removeEndpoint,
  loadSettings,
  loadSubscriptionGroups,
  addDefaultSubscription,
  removeDefaultSubscription,
  addAuthSourceDefaultSubscription,
  removeAuthSourceDefaultSubscription,
  saveSettings,
  testSmtpConnection,
  sendTestEmail
} = useSettingsViewForm({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess,
  refreshPublicSettings: appStore.fetchPublicSettings,
  refreshAdminSettings: adminSettingsStore.fetch,
  copyToClipboard
})

const {
  adminApiKeyLoading,
  adminApiKeyExists,
  adminApiKeyMasked,
  adminApiKeyOperating,
  newAdminApiKey,
  overloadCooldownLoading,
  overloadCooldownSaving,
  overloadCooldownForm,
  streamTimeoutLoading,
  streamTimeoutSaving,
  streamTimeoutForm,
  rectifierLoading,
  rectifierSaving,
  rectifierForm,
  betaPolicyLoading,
  betaPolicySaving,
  betaPolicyForm,
  betaPolicyActionOptions,
  betaPolicyScopeOptions,
  getBetaDisplayName,
  loadAdminApiKey,
  createAdminApiKey,
  regenerateAdminApiKey,
  deleteAdminApiKey,
  copyNewKey,
  loadOverloadCooldownSettings,
  saveOverloadCooldownSettings,
  loadStreamTimeoutSettings,
  saveStreamTimeoutSettings,
  loadRectifierSettings,
  saveRectifierSettings,
  loadBetaPolicySettings,
  saveBetaPolicySettings
} = useSettingsViewPolicies({
  t,
  showError: appStore.showError,
  showSuccess: appStore.showSuccess,
  confirm: (message?: string) => window.confirm(message),
  copyToClipboard
})

function isPaymentTypeEnabled(type: string): boolean {
  return paymentConfig.enabled_payment_types.includes(type)
}

function togglePaymentType(type: string) {
  if (paymentConfig.enabled_payment_types.includes(type)) {
    paymentConfig.enabled_payment_types = paymentConfig.enabled_payment_types.filter(
      (value) => value !== type
    )
    return
  }

  paymentConfig.enabled_payment_types = [...paymentConfig.enabled_payment_types, type]
}

async function loadPaymentConfig() {
  paymentConfigLoading.value = true
  try {
    const response = await adminAPI.payment.getConfig()
    Object.assign(paymentConfig, createDefaultPaymentConfig(), response.data)
    paymentConfig.enabled_payment_types = [...(response.data.enabled_payment_types || [])]
    lastSavedEnabledPaymentTypes.value = [...paymentConfig.enabled_payment_types]
  } catch (error: unknown) {
    appStore.showError(
      `${t('admin.settings.failedToLoad')}: ${extractI18nErrorMessage(error, t, 'payment.errors', t('common.error'))}`
    )
  } finally {
    paymentConfigLoading.value = false
  }
}

async function savePaymentConfig() {
  paymentConfigSaving.value = true
  const removedTypes = lastSavedEnabledPaymentTypes.value.filter(
    (type) => !paymentConfig.enabled_payment_types.includes(type)
  )

  try {
    await adminAPI.payment.updateConfig({
      enabled: paymentConfig.enabled,
      min_amount: Number(paymentConfig.min_amount) || 0,
      max_amount: Number(paymentConfig.max_amount) || 0,
      daily_limit: Number(paymentConfig.daily_limit) || 0,
      order_timeout_minutes: Math.max(1, Number(paymentConfig.order_timeout_minutes) || 30),
      max_pending_orders: Math.max(1, Number(paymentConfig.max_pending_orders) || 3),
      enabled_payment_types: [...paymentConfig.enabled_payment_types],
      balance_disabled: paymentConfig.balance_disabled,
      balance_recharge_multiplier: Math.max(0.01, Number(paymentConfig.balance_recharge_multiplier) || 1),
      recharge_fee_rate: Math.max(0, Number(paymentConfig.recharge_fee_rate) || 0),
      load_balance_strategy: paymentConfig.load_balance_strategy,
      product_name_prefix: paymentConfig.product_name_prefix,
      product_name_suffix: paymentConfig.product_name_suffix,
      help_image_url: paymentConfig.help_image_url,
      help_text: paymentConfig.help_text,
      cancel_rate_limit_enabled: paymentConfig.cancel_rate_limit_enabled,
      cancel_rate_limit_max: Math.max(1, Number(paymentConfig.cancel_rate_limit_max) || 10),
      cancel_rate_limit_window: Math.max(1, Number(paymentConfig.cancel_rate_limit_window) || 1),
      cancel_rate_limit_unit: paymentConfig.cancel_rate_limit_unit,
      cancel_rate_limit_window_mode: paymentConfig.cancel_rate_limit_window_mode,
    })

    await Promise.allSettled(removedTypes.map((type) => disableProvidersByType(type)))
    lastSavedEnabledPaymentTypes.value = [...paymentConfig.enabled_payment_types]
    await loadProviders()
    appStore.showSuccess(t('admin.settings.settingsSaved'))
  } catch (error: unknown) {
    appStore.showError(
      `${t('admin.settings.failedToSave')}: ${extractI18nErrorMessage(error, t, 'payment.errors', t('common.error'))}`
    )
  } finally {
    paymentConfigSaving.value = false
  }
}

async function loadProviders() {
  providersLoading.value = true
  try {
    const response = await adminAPI.payment.getProviders()
    providers.value = response.data || []
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    providersLoading.value = false
  }
}

async function disableProvidersByType(type: string) {
  const matchingProviders = providers.value.filter(
    (provider) => provider.provider_key === type && provider.enabled
  )

  await Promise.all(
    matchingProviders.map(async (provider) => {
      await adminAPI.payment.updateProvider(provider.id, { enabled: false })
      provider.enabled = false
    })
  )
}

function openCreateProvider() {
  editingProvider.value = null
  providerDialogRef.value?.reset(enabledProviderKeyOptions.value[0]?.value || 'easypay')
  showProviderDialog.value = true
}

function openEditProvider(provider: ProviderInstance) {
  editingProvider.value = provider
  providerDialogRef.value?.loadProvider(provider)
  showProviderDialog.value = true
}

async function handleSaveProvider(payload: Partial<ProviderInstance>) {
  providerSaving.value = true
  try {
    if (editingProvider.value) {
      await adminAPI.payment.updateProvider(editingProvider.value.id, payload)
    } else {
      await adminAPI.payment.createProvider(payload)
    }
    showProviderDialog.value = false
    await loadProviders()
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  } finally {
    providerSaving.value = false
  }
}

async function handleToggleField(
  provider: ProviderInstance,
  field: 'enabled' | 'refund_enabled' | 'allow_user_refund'
) {
  const nextValue =
    field === 'enabled'
      ? !provider.enabled
      : field === 'refund_enabled'
        ? !provider.refund_enabled
        : !provider.allow_user_refund

  const payload: Record<string, boolean> = { [field]: nextValue }
  if (field === 'refund_enabled' && !nextValue) {
    payload.allow_user_refund = false
  }

  try {
    await adminAPI.payment.updateProvider(provider.id, payload)
    if (field === 'enabled') {
      provider.enabled = nextValue
      return
    }
    if (field === 'refund_enabled') {
      provider.refund_enabled = nextValue
      if (!nextValue) {
        provider.allow_user_refund = false
      }
      return
    }
    provider.allow_user_refund = nextValue
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  }
}

async function handleToggleType(provider: ProviderInstance, type: string) {
  const supportedTypes = provider.supported_types.includes(type)
    ? provider.supported_types.filter((value) => value !== type)
    : [...provider.supported_types, type]

  try {
    await adminAPI.payment.updateProvider(provider.id, { supported_types: supportedTypes })
    provider.supported_types = supportedTypes
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  }
}

async function handleReorderProviders(updates: { id: number; sort_order: number }[]) {
  try {
    await Promise.all(
      updates.map((update) =>
        adminAPI.payment.updateProvider(update.id, { sort_order: update.sort_order })
      )
    )

    for (const update of updates) {
      const provider = providers.value.find((item) => item.id === update.id)
      if (provider) {
        provider.sort_order = update.sort_order
      }
    }
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
    await loadProviders()
  }
}

function confirmDeleteProvider(provider: ProviderInstance) {
  deletingProviderId.value = provider.id
  showDeleteProviderDialog.value = true
}

async function handleDeleteProvider() {
  if (!deletingProviderId.value) {
    return
  }

  try {
    await adminAPI.payment.deleteProvider(deletingProviderId.value)
    showDeleteProviderDialog.value = false
    deletingProviderId.value = null
    appStore.showSuccess(t('common.deleted'))
    await loadProviders()
  } catch (error: unknown) {
    appStore.showError(extractI18nErrorMessage(error, t, 'payment.errors', t('common.error')))
  }
}

onMounted(() => {
  void loadSettings()
  void loadSubscriptionGroups()
  void loadAdminApiKey()
  void loadOverloadCooldownSettings()
  void loadStreamTimeoutSettings()
  void loadRectifierSettings()
  void loadBetaPolicySettings()
  void loadPaymentConfig()
  void loadProviders()
})
</script>

<style scoped>
.default-sub-group-select :deep(.select-trigger) {
  @apply h-[42px];
}

.default-sub-delete-btn {
  @apply h-[42px];
}
</style>
