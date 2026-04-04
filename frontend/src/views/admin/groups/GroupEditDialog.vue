<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.editGroup')"
    width="normal"
    @close="emit('close')"
  >
    <form
      v-if="editingGroup"
      id="edit-group-form"
      class="space-y-5"
      @submit.prevent="emit('submit')"
    >
      <GroupBaseFieldsSection
        :form="form"
        :platform-options="platformOptions"
        :copy-accounts-options="copyAccountsGroupOptions"
        :copy-accounts-tooltip-text="t('admin.groups.copyAccounts.tooltipEdit')"
        :copy-accounts-hint-text="t('admin.groups.copyAccounts.hintEdit')"
        :platform-hint="t('admin.groups.platformNotEditable')"
        :platform-disabled="true"
        name-tour-target="edit-group-form-name"
        platform-tour-target="group-form-platform"
        rate-multiplier-tour-target="group-form-multiplier"
        @add-group="emit('add-copy-group', $event)"
        @remove-group="emit('remove-copy-group', $event)"
      />
      <GroupExclusiveSection :form="form" />
      <GroupEditStatusField v-model="form.status" :status-options="editStatusOptions" />

      <GroupSubscriptionSection
        :form="form"
        :subscription-type-options="subscriptionTypeOptions"
        :subscription-type-hint="t('admin.groups.subscription.typeNotEditable')"
        :subscription-type-disabled="true"
      />

      <GroupImagePricingSection :form="form" />
      <GroupSoraPricingSection :form="form" />
      <GroupSupportedScopesSection :form="form" @toggle-scope="emit('toggle-scope', $event)" />
      <GroupMcpXmlSection :form="form" />

      <GroupClaudeCodeSection
        :form="form"
        :fallback-group-options="fallbackGroupOptions"
      />

      <GroupOpenAIMessagesSection :form="form" />
      <GroupAccountFilterSection :form="form" />

      <GroupInvalidRequestFallbackSection
        :form="form"
        :options="invalidRequestFallbackOptions"
      />

      <GroupModelRoutingSection
        :form="form"
        :rules="rules"
        :account-search-keyword="accountSearchKeyword"
        :account-search-results="accountSearchResults"
        :show-account-dropdown="showAccountDropdown"
        :get-rule-render-key="getRuleRenderKey"
        :get-rule-search-key="getRuleSearchKey"
        :search-accounts-by-rule="searchAccountsByRule"
        :select-account="selectAccount"
        :remove-selected-account="removeSelectedAccount"
        :on-account-search-focus="onAccountSearchFocus"
        :add-routing-rule="addRoutingRule"
        :remove-routing-rule="removeRoutingRule"
      />
    </form>

    <template #footer>
      <GroupDialogFooter
        form-id="edit-group-form"
        :submitting="submitting"
        :submitting-text="t('admin.groups.updating')"
        :submit-text="t('common.update')"
        @close="emit('close')"
      />
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SelectOption } from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { AdminGroup } from '@/types'
import type {
  EditGroupForm,
  ModelRoutingRule,
  NullableNumberSelectOption,
  NumberSelectOption,
  SimpleAccount
} from '../groupsForm'
import GroupAccountFilterSection from './GroupAccountFilterSection.vue'
import GroupBaseFieldsSection from './GroupBaseFieldsSection.vue'
import GroupClaudeCodeSection from './GroupClaudeCodeSection.vue'
import GroupDialogFooter from './GroupDialogFooter.vue'
import GroupEditStatusField from './GroupEditStatusField.vue'
import GroupExclusiveSection from './GroupExclusiveSection.vue'
import GroupImagePricingSection from './GroupImagePricingSection.vue'
import GroupInvalidRequestFallbackSection from './GroupInvalidRequestFallbackSection.vue'
import GroupMcpXmlSection from './GroupMcpXmlSection.vue'
import GroupModelRoutingSection from './GroupModelRoutingSection.vue'
import GroupOpenAIMessagesSection from './GroupOpenAIMessagesSection.vue'
import GroupSoraPricingSection from './GroupSoraPricingSection.vue'
import GroupSubscriptionSection from './GroupSubscriptionSection.vue'
import GroupSupportedScopesSection from './GroupSupportedScopesSection.vue'

defineProps<{
  show: boolean
  submitting: boolean
  editingGroup: AdminGroup | null
  form: EditGroupForm
  platformOptions: SelectOption[]
  copyAccountsGroupOptions: NumberSelectOption[]
  editStatusOptions: SelectOption[]
  subscriptionTypeOptions: SelectOption[]
  fallbackGroupOptions: NullableNumberSelectOption[]
  invalidRequestFallbackOptions: NullableNumberSelectOption[]
  rules: ModelRoutingRule[]
  accountSearchKeyword: Record<string, string>
  accountSearchResults: Record<string, SimpleAccount[]>
  showAccountDropdown: Record<string, boolean>
  getRuleRenderKey: (rule: ModelRoutingRule) => string
  getRuleSearchKey: (rule: ModelRoutingRule) => string
  searchAccountsByRule: (rule: ModelRoutingRule) => void
  selectAccount: (rule: ModelRoutingRule, account: SimpleAccount) => void
  removeSelectedAccount: (rule: ModelRoutingRule, accountId: number) => void
  onAccountSearchFocus: (rule: ModelRoutingRule) => void
  addRoutingRule: () => void
  removeRoutingRule: (rule: ModelRoutingRule) => void
}>()

const emit = defineEmits<{
  close: []
  submit: []
  'add-copy-group': [groupId: number]
  'remove-copy-group': [groupId: number]
  'toggle-scope': [scope: string]
}>()

const { t } = useI18n()
</script>
