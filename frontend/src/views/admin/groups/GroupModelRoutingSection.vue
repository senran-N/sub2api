<template>
  <div v-if="form.platform === 'anthropic'" class="border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('admin.groups.modelRouting.title') }}
      </label>
      <GroupSectionInfoTooltip
        :text="t('admin.groups.modelRouting.tooltip')"
        width-class="w-80"
      />
    </div>

    <div class="mb-3 flex items-center gap-3">
      <Toggle v-model="form.model_routing_enabled" />
      <span class="text-sm text-gray-500 dark:text-gray-400">
        {{
          form.model_routing_enabled
            ? t('admin.groups.modelRouting.enabled')
            : t('admin.groups.modelRouting.disabled')
        }}
      </span>
    </div>

    <p v-if="!form.model_routing_enabled" class="mb-3 text-xs text-gray-500 dark:text-gray-400">
      {{ t('admin.groups.modelRouting.disabledHint') }}
    </p>
    <template v-else>
      <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
        {{ t('admin.groups.modelRouting.noRulesHint') }}
      </p>

      <div v-if="rules.length > 0" class="space-y-3">
        <GroupModelRoutingRuleCard
          v-for="rule in rules"
          :key="getRuleRenderKey(rule)"
          :rule="rule"
          :account-search-keyword="accountSearchKeyword"
          :account-search-results="accountSearchResults"
          :show-account-dropdown="showAccountDropdown"
          :get-rule-search-key="getRuleSearchKey"
          :search-accounts-by-rule="searchAccountsByRule"
          :select-account="selectAccount"
          :remove-selected-account="removeSelectedAccount"
          :on-account-search-focus="onAccountSearchFocus"
          :remove-routing-rule="removeRoutingRule"
        />
      </div>

      <button
        type="button"
        class="mt-3 flex items-center gap-1.5 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
        @click="addRoutingRule"
      >
        <Icon name="plus" size="sm" />
        {{ t('admin.groups.modelRouting.addRule') }}
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import type { CreateGroupForm, EditGroupForm, ModelRoutingRule, SimpleAccount } from '../groupsForm'
import GroupModelRoutingRuleCard from './GroupModelRoutingRuleCard.vue'
import GroupSectionInfoTooltip from './GroupSectionInfoTooltip.vue'

defineProps<{
  form: CreateGroupForm | EditGroupForm
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

const { t } = useI18n()
</script>
