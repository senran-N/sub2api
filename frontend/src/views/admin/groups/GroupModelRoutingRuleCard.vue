<template>
  <div class="rounded-lg border border-gray-200 p-3 dark:border-dark-600">
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
              class="inline-flex items-center gap-1 rounded-full bg-primary-100 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
            >
              {{ account.name }}
              <button
                type="button"
                class="ml-0.5 text-primary-500 hover:text-primary-700 dark:hover:text-primary-200"
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
              class="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-lg border bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800"
            >
              <button
                v-for="account in accountSearchResults[searchKey]"
                :key="account.id"
                type="button"
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-dark-700"
                :class="{ 'opacity-50': isSelected(account.id) }"
                :disabled="isSelected(account.id)"
                @click="selectAccount(rule, account)"
              >
                <span>{{ account.name }}</span>
                <span class="ml-2 text-xs text-gray-400">#{{ account.id }}</span>
              </button>
            </div>
          </div>
          <p class="mt-1 text-xs text-gray-400">{{ t('admin.groups.modelRouting.accountsHint') }}</p>
        </div>
      </div>
      <button
        type="button"
        class="mt-5 p-1.5 text-gray-400 transition-colors hover:text-red-500"
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
