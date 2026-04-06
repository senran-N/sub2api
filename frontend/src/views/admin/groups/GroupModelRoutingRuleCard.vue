<template>
  <div class="group-model-routing-rule-card border">
    <div class="flex items-start gap-3">
      <div class="flex-1 space-y-2">
        <div>
          <label class="input-label text-xs">{{ t('admin.groups.modelRouting.modelPattern') }}</label>
          <input
            v-model="rule.pattern"
            type="text"
            class="input text-sm"
            :placeholder="t('admin.groups.modelRouting.modelPatternPlaceholder')"
          />
        </div>
        <div>
          <label class="input-label text-xs">{{ t('admin.groups.modelRouting.accounts') }}</label>
          <div v-if="rule.accounts.length > 0" class="mb-2 flex flex-wrap gap-1.5">
            <span
              v-for="account in rule.accounts"
              :key="account.id"
              class="group-model-routing-rule-card__account-chip inline-flex items-center gap-1 text-xs font-medium"
            >
              {{ account.name }}
              <button
                type="button"
                class="group-model-routing-rule-card__account-chip-remove ml-0.5"
                @click="removeSelectedAccount(rule, account.id)"
              >
                <Icon name="x" size="xs" />
              </button>
            </span>
          </div>
          <div class="relative account-search-container">
            <input
              v-model="accountSearchKeyword[searchKey]"
              type="text"
              class="input text-sm"
              :placeholder="t('admin.groups.modelRouting.searchAccountPlaceholder')"
              @input="searchAccountsByRule(rule)"
              @focus="onAccountSearchFocus(rule)"
            />
            <div
              v-if="showAccountDropdown[searchKey] && accountSearchResults[searchKey]?.length"
              class="group-model-routing-rule-card__dropdown absolute z-50 mt-1 w-full overflow-auto border"
            >
              <button
                v-for="account in accountSearchResults[searchKey]"
                :key="account.id"
                type="button"
                class="group-model-routing-rule-card__dropdown-option w-full text-left text-sm"
                :class="{ 'opacity-50': isSelected(account.id) }"
                :disabled="isSelected(account.id)"
                @click="selectAccount(rule, account)"
              >
                <span>{{ account.name }}</span>
                <span class="group-model-routing-rule-card__dropdown-meta ml-2 text-xs">#{{ account.id }}</span>
              </button>
            </div>
          </div>
          <p class="group-model-routing-rule-card__hint mt-1 text-xs">{{ t('admin.groups.modelRouting.accountsHint') }}</p>
        </div>
      </div>
      <button
        type="button"
        class="group-model-routing-rule-card__remove mt-5 transition-colors"
        :title="t('admin.groups.modelRouting.removeRule')"
        @click="removeRoutingRule(rule)"
      >
        <Icon name="trash" size="sm" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ModelRoutingRule, SimpleAccount } from '../groupsForm'

const props = defineProps<{
  rule: ModelRoutingRule
  accountSearchKeyword: Record<string, string>
  accountSearchResults: Record<string, SimpleAccount[]>
  showAccountDropdown: Record<string, boolean>
  getRuleSearchKey: (rule: ModelRoutingRule) => string
  searchAccountsByRule: (rule: ModelRoutingRule) => void
  selectAccount: (rule: ModelRoutingRule, account: SimpleAccount) => void
  removeSelectedAccount: (rule: ModelRoutingRule, accountId: number) => void
  onAccountSearchFocus: (rule: ModelRoutingRule) => void
  removeRoutingRule: (rule: ModelRoutingRule) => void
}>()

const { t } = useI18n()

const searchKey = computed(() => props.getRuleSearchKey(props.rule))

const isSelected = (accountId: number) => {
  return props.rule.accounts.some((account) => account.id === accountId)
}
</script>

<style scoped>
.group-model-routing-rule-card {
  border-radius: var(--theme-group-routing-card-radius);
  padding: var(--theme-group-routing-card-padding);
  border-color: color-mix(in srgb, var(--theme-card-border) 84%, transparent);
}

.group-model-routing-rule-card__account-chip {
  border-radius: 999px;
  padding: var(--theme-group-routing-chip-padding-y) var(--theme-group-routing-chip-padding-x);
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}

.group-model-routing-rule-card__account-chip-remove,
.group-model-routing-rule-card__dropdown-meta,
.group-model-routing-rule-card__hint {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group-model-routing-rule-card__account-chip-remove:hover {
  color: color-mix(in srgb, var(--theme-accent) 92%, var(--theme-page-text));
}

.group-model-routing-rule-card__dropdown {
  max-height: var(--theme-group-routing-dropdown-max-height);
  border-radius: var(--theme-group-routing-dropdown-radius);
  border-color: color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
}

.group-model-routing-rule-card__dropdown-option {
  padding: var(--theme-group-routing-dropdown-option-padding-y)
    var(--theme-group-routing-dropdown-option-padding-x);
  color: var(--theme-page-text);
  transition: background-color 0.2s ease;
}

.group-model-routing-rule-card__dropdown-option:hover:not(:disabled) {
  background: var(--theme-dropdown-item-hover-bg);
}

.group-model-routing-rule-card__remove {
  padding: var(--theme-group-routing-remove-padding);
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group-model-routing-rule-card__remove:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
