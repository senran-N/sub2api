<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.createGroup')"
    width="normal"
    @close="emit('close')"
  >
    <form id="create-group-form" class="space-y-5" @submit.prevent="emit('submit')">
      <GroupFormSections
        mode="create"
        :form="form"
        :platform-options="platformOptions"
        :copy-accounts-group-options="copyAccountsGroupOptions"
        :edit-status-options="[]"
        :subscription-type-options="subscriptionTypeOptions"
        :fallback-group-options="fallbackGroupOptions"
        :invalid-request-fallback-options="invalidRequestFallbackOptions"
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
        @add-copy-group="emit('add-copy-group', $event)"
        @remove-copy-group="emit('remove-copy-group', $event)"
        @toggle-scope="emit('toggle-scope', $event)"
      />
    </form>

    <template #footer>
      <GroupDialogFooter
        form-id="create-group-form"
        :submitting="submitting"
        :submitting-text="t('admin.groups.creating')"
        :submit-text="t('common.create')"
        @close="emit('close')"
      />
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SelectOption } from '@/components/common/Select.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type {
  CreateGroupForm,
  ModelRoutingRule,
  NullableNumberSelectOption,
  NumberSelectOption,
  SimpleAccount
} from './groupsForm'
import GroupDialogFooter from './GroupDialogFooter.vue'
import GroupFormSections from './GroupFormSections.vue'

defineProps<{
  show: boolean
  submitting: boolean
  form: CreateGroupForm
  platformOptions: SelectOption[]
  copyAccountsGroupOptions: NumberSelectOption[]
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
