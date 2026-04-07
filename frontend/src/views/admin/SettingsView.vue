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
        </div>

        <div v-if="activeTab === 'users'" class="space-y-6">
          <SettingsDefaultsCard
            :form="form"
            :default-subscription-group-options="defaultSubscriptionGroupOptions"
            :to-default-subscription-group-option="toDefaultSubscriptionGroupOption"
            @add-default-subscription="addDefaultSubscription"
            @remove-default-subscription="removeDefaultSubscription"
          />
        </div>

        <div v-if="activeTab === 'gateway'" class="space-y-6">
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

          <SettingsSoraClientCard :form="form" />

          <SettingsCustomMenuCard
            :form="form"
            @add-item="addMenuItem"
            @remove-item="removeMenuItem"
            @move-item="moveMenuItem"
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
        </div>

        <div v-if="activeTab === 'backup'">
          <BackupSettings />
        </div>

        <div v-if="activeTab === 'data'">
          <DataManagementSettings />
        </div>

        <SettingsSaveBar
          v-if="activeTab !== 'backup' && activeTab !== 'data'"
          :saving="saving"
          :disabled="loadFailed"
        />
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { defineAsyncComponent, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import type { GroupPlatform, SubscriptionType } from '@/types'
import SettingsAdminApiKeyCard from './settings/SettingsAdminApiKeyCard.vue'
import SettingsBetaPolicyCard from './settings/SettingsBetaPolicyCard.vue'
import SettingsClaudeCodeCard from './settings/SettingsClaudeCodeCard.vue'
import SettingsCustomMenuCard from './settings/SettingsCustomMenuCard.vue'
import SettingsDefaultsCard from './settings/SettingsDefaultsCard.vue'
import SettingsEmailDisabledCard from './settings/SettingsEmailDisabledCard.vue'
import SettingsGatewayForwardingCard from './settings/SettingsGatewayForwardingCard.vue'
import SettingsLinuxdoCard from './settings/SettingsLinuxdoCard.vue'
import SettingsOverloadCooldownCard from './settings/SettingsOverloadCooldownCard.vue'
import SettingsPurchaseCard from './settings/SettingsPurchaseCard.vue'
import SettingsRegistrationCard from './settings/SettingsRegistrationCard.vue'
import SettingsRectifierCard from './settings/SettingsRectifierCard.vue'
import SettingsSchedulingCard from './settings/SettingsSchedulingCard.vue'
import SettingsSaveBar from './settings/SettingsSaveBar.vue'
import SettingsSiteCard from './settings/SettingsSiteCard.vue'
import SettingsSmtpCard from './settings/SettingsSmtpCard.vue'
import SettingsSoraClientCard from './settings/SettingsSoraClientCard.vue'
import SettingsStreamTimeoutCard from './settings/SettingsStreamTimeoutCard.vue'
import SettingsTestEmailCard from './settings/SettingsTestEmailCard.vue'
import SettingsLoadingState from './settings/SettingsLoadingState.vue'
import SettingsTabsNav from './settings/SettingsTabsNav.vue'
import SettingsTurnstileCard from './settings/SettingsTurnstileCard.vue'
import { useSettingsViewForm } from './settings/useSettingsViewForm'
import { useSettingsViewPolicies } from './settings/useSettingsViewPolicies'

const BackupSettings = defineAsyncComponent(() => import('@/views/admin/BackupView.vue'))
const DataManagementSettings = defineAsyncComponent(() => import('@/views/admin/DataManagementView.vue'))

const { t } = useI18n()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()

type SettingsTab = 'general' | 'security' | 'users' | 'gateway' | 'email' | 'backup' | 'data'

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
  { key: 'email'    as SettingsTab, icon: 'mail'   as const },
  { key: 'backup'   as SettingsTab, icon: 'database' as const },
  { key: 'data'     as SettingsTab, icon: 'cube'     as const },
]
const { copyToClipboard } = useClipboard()

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
  removeRegistrationEmailSuffixWhitelistTag,
  commitRegistrationEmailSuffixWhitelistDraft,
  handleRegistrationEmailSuffixWhitelistDraftInput,
  handleRegistrationEmailSuffixWhitelistDraftKeydown,
  handleRegistrationEmailSuffixWhitelistPaste,
  setAndCopyLinuxdoRedirectUrl,
  addMenuItem,
  removeMenuItem,
  moveMenuItem,
  addEndpoint,
  removeEndpoint,
  loadSettings,
  loadSubscriptionGroups,
  addDefaultSubscription,
  removeDefaultSubscription,
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
  confirm: window.confirm,
  copyToClipboard
})

onMounted(() => {
  void loadSettings()
  void loadSubscriptionGroups()
  void loadAdminApiKey()
  void loadOverloadCooldownSettings()
  void loadStreamTimeoutSettings()
  void loadRectifierSettings()
  void loadBetaPolicySettings()
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
