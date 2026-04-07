<template>
  <div v-if="form.platform === 'anthropic'" class="group-model-routing-section border-t pt-4">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="group-model-routing-section__title text-sm font-medium">
        {{ t('admin.groups.modelRouting.title') }}
      </label>
      <GroupSectionInfoTooltip
        :text="t('admin.groups.modelRouting.tooltip')"
        width-class="w-80"
      />
    </div>

    <div class="mb-3 flex items-center gap-3">
      <Toggle v-model="form.model_routing_enabled" />
      <span class="group-model-routing-section__status text-sm">
        {{
          form.model_routing_enabled
            ? t('admin.groups.modelRouting.enabled')
            : t('admin.groups.modelRouting.disabled')
        }}
      </span>
    </div>

    <p v-if="!form.model_routing_enabled" class="group-model-routing-section__hint mb-3 text-xs">
      {{ t('admin.groups.modelRouting.disabledHint') }}
    </p>
    <template v-else>
      <p class="group-model-routing-section__hint mb-3 text-xs">
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
        class="group-model-routing-section__add mt-3 flex items-center gap-1.5 text-sm"
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
import type { CreateGroupForm, EditGroupForm, ModelRoutingRule, SimpleAccount } from './groupsForm'
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

<style scoped>
.group-model-routing-section {
  border-color: var(--theme-page-border);
}

.group-model-routing-section__title {
  color: var(--theme-page-text);
}

.group-model-routing-section__status,
.group-model-routing-section__hint {
  color: var(--theme-page-muted);
}

.group-model-routing-section__add {
  color: var(--theme-accent);
  transition: color 0.2s ease;
}

.group-model-routing-section__add:hover {
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}
</style>
