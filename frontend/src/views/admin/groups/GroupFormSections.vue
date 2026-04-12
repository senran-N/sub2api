<template>
  <GroupBaseFieldsSection
    :form="form"
    :platform-options="platformOptions"
    :copy-accounts-options="copyAccountsGroupOptions"
    :copy-accounts-tooltip-text="copyAccountsTooltipText"
    :copy-accounts-hint-text="copyAccountsHintText"
    :platform-hint="platformHintText"
    :platform-disabled="isEditMode"
    :name-placeholder="namePlaceholder"
    :description-placeholder="descriptionPlaceholder"
    :rate-multiplier-hint="rateMultiplierHint"
    :reset-copy-accounts-on-platform-change="isCreateMode"
    :name-tour-target="nameTourTarget"
    platform-tour-target="group-form-platform"
    rate-multiplier-tour-target="group-form-multiplier"
    @add-group="emit('add-copy-group', $event)"
    @remove-group="emit('remove-copy-group', $event)"
  />
  <GroupExclusiveSection :form="form" :tour-target="exclusiveTourTarget" />
  <GroupEditStatusField
    v-if="isEditMode"
    v-model="editStatusModel"
    :status-options="editStatusOptions"
  />

  <GroupSubscriptionSection
    :form="form"
    :subscription-type-options="subscriptionTypeOptions"
    :subscription-type-hint="subscriptionTypeHint"
    :subscription-type-disabled="isEditMode"
  />

  <GroupImagePricingSection :form="form" />
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { SelectOption } from '@/components/common/Select.vue'
import type {
  EditGroupForm,
  GroupDialogForm,
  ModelRoutingRule,
  NullableNumberSelectOption,
  NumberSelectOption,
  SimpleAccount
} from './groupsForm'
import GroupAccountFilterSection from './GroupAccountFilterSection.vue'
import GroupBaseFieldsSection from './GroupBaseFieldsSection.vue'
import GroupClaudeCodeSection from './GroupClaudeCodeSection.vue'
import GroupEditStatusField from './GroupEditStatusField.vue'
import GroupExclusiveSection from './GroupExclusiveSection.vue'
import GroupImagePricingSection from './GroupImagePricingSection.vue'
import GroupInvalidRequestFallbackSection from './GroupInvalidRequestFallbackSection.vue'
import GroupMcpXmlSection from './GroupMcpXmlSection.vue'
import GroupModelRoutingSection from './GroupModelRoutingSection.vue'
import GroupOpenAIMessagesSection from './GroupOpenAIMessagesSection.vue'
import GroupSubscriptionSection from './GroupSubscriptionSection.vue'
import GroupSupportedScopesSection from './GroupSupportedScopesSection.vue'

const props = defineProps<{
  mode: 'create' | 'edit'
  form: GroupDialogForm
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
  'add-copy-group': [groupId: number]
  'remove-copy-group': [groupId: number]
  'toggle-scope': [scope: string]
}>()

const { t } = useI18n()

const isCreateMode = computed(() => props.mode === 'create')
const isEditMode = computed(() => props.mode === 'edit')

const copyAccountsTooltipText = computed(() =>
  isEditMode.value
    ? t('admin.groups.copyAccounts.tooltipEdit')
    : t('admin.groups.copyAccounts.tooltip')
)
const copyAccountsHintText = computed(() =>
  isEditMode.value
    ? t('admin.groups.copyAccounts.hintEdit')
    : t('admin.groups.copyAccounts.hint')
)
const platformHintText = computed(() =>
  isEditMode.value ? t('admin.groups.platformNotEditable') : t('admin.groups.platformHint')
)
const namePlaceholder = computed(() =>
  isCreateMode.value ? t('admin.groups.enterGroupName') : undefined
)
const descriptionPlaceholder = computed(() =>
  isCreateMode.value ? t('admin.groups.optionalDescription') : undefined
)
const rateMultiplierHint = computed(() =>
  isCreateMode.value ? t('admin.groups.rateMultiplierHint') : undefined
)
const nameTourTarget = computed(() =>
  isCreateMode.value ? 'group-form-name' : 'edit-group-form-name'
)
const exclusiveTourTarget = computed(() =>
  isCreateMode.value ? 'group-form-exclusive' : undefined
)
const subscriptionTypeHint = computed(() =>
  isEditMode.value
    ? t('admin.groups.subscription.typeNotEditable')
    : t('admin.groups.subscription.typeHint')
)

const editStatusModel = computed({
  get: () => (props.form as EditGroupForm).status,
  set: (value: EditGroupForm['status']) => {
    (props.form as EditGroupForm).status = value
  }
})
</script>
