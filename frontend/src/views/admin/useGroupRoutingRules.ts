import { ref } from 'vue'
import { adminAPI } from '@/api/admin'
import { useKeyedDebouncedSearch } from '@/composables/useKeyedDebouncedSearch'
import { createStableObjectKeyResolver } from '@/utils/stableObjectKey'
import type { ModelRoutingRule, SimpleAccount } from './groupsForm'

export function createEmptyRoutingRule(): ModelRoutingRule {
  return {
    pattern: '',
    accounts: []
  }
}

export function pushRoutingRule(rules: ModelRoutingRule[]): void {
  rules.push(createEmptyRoutingRule())
}

export function removeRoutingRuleByReference(
  rules: ModelRoutingRule[],
  rule: ModelRoutingRule
): boolean {
  const index = rules.indexOf(rule)
  if (index === -1) {
    return false
  }
  rules.splice(index, 1)
  return true
}

export function appendRoutingRuleAccount(
  rule: ModelRoutingRule,
  account: SimpleAccount
): boolean {
  if (rule.accounts.some((item) => item.id === account.id)) {
    return false
  }

  rule.accounts.push(account)
  return true
}

export function removeRoutingRuleAccount(rule: ModelRoutingRule, accountId: number): void {
  rule.accounts = rule.accounts.filter((account) => account.id !== accountId)
}

export function closeAllRoutingRuleDropdowns(state: Record<string, boolean>): void {
  Object.keys(state).forEach((key) => {
    state[key] = false
  })
}

export async function hydrateRoutingRulesFromApi(
  apiFormat: Record<string, number[]> | null,
  loadAccountById: (id: number) => Promise<SimpleAccount>
): Promise<ModelRoutingRule[]> {
  if (!apiFormat) {
    return []
  }

  const rules: ModelRoutingRule[] = []
  for (const [pattern, accountIds] of Object.entries(apiFormat)) {
    const accounts: SimpleAccount[] = []
    for (const id of accountIds) {
      try {
        accounts.push(await loadAccountById(id))
      } catch {
        accounts.push({ id, name: `#${id}` })
      }
    }
    rules.push({ pattern, accounts })
  }
  return rules
}

export function useGroupRoutingRules(scope: 'create' | 'edit') {
  const rules = ref<ModelRoutingRule[]>([])
  const accountSearchKeyword = ref<Record<string, string>>({})
  const accountSearchResults = ref<Record<string, SimpleAccount[]>>({})
  const showAccountDropdown = ref<Record<string, boolean>>({})

  const resolveRuleKey = createStableObjectKeyResolver<ModelRoutingRule>(`${scope}-rule`)

  const getRuleRenderKey = (rule: ModelRoutingRule) => resolveRuleKey(rule)
  const getRuleSearchKey = (rule: ModelRoutingRule) => `${scope}-${resolveRuleKey(rule)}`

  const clearAccountSearchStateByKey = (key: string) => {
    delete accountSearchKeyword.value[key]
    delete accountSearchResults.value[key]
    delete showAccountDropdown.value[key]
  }

  const clearAllAccountSearchState = () => {
    accountSearchKeyword.value = {}
    accountSearchResults.value = {}
    showAccountDropdown.value = {}
  }

  const accountSearchRunner = useKeyedDebouncedSearch<SimpleAccount[]>({
    delay: 300,
    search: async (keyword, { signal }) => {
      const response = await adminAPI.accounts.list(
        1,
        20,
        {
          search: keyword,
          platform: 'anthropic'
        },
        { signal }
      )
      return response.items.map((account) => ({ id: account.id, name: account.name }))
    },
    onSuccess: (key, result) => {
      accountSearchResults.value[key] = result
    },
    onError: (key) => {
      accountSearchResults.value[key] = []
    }
  })

  const searchAccounts = (key: string) => {
    accountSearchRunner.trigger(key, accountSearchKeyword.value[key] || '')
  }

  const searchAccountsByRule = (rule: ModelRoutingRule) => {
    searchAccounts(getRuleSearchKey(rule))
  }

  const selectAccount = (rule: ModelRoutingRule, account: SimpleAccount) => {
    appendRoutingRuleAccount(rule, account)
    const key = getRuleSearchKey(rule)
    accountSearchKeyword.value[key] = ''
    showAccountDropdown.value[key] = false
  }

  const removeSelectedAccount = (rule: ModelRoutingRule, accountId: number) => {
    removeRoutingRuleAccount(rule, accountId)
  }

  const onAccountSearchFocus = (rule: ModelRoutingRule) => {
    const key = getRuleSearchKey(rule)
    showAccountDropdown.value[key] = true
    if (!accountSearchResults.value[key]?.length) {
      searchAccounts(key)
    }
  }

  const addRoutingRule = () => {
    pushRoutingRule(rules.value)
  }

  const removeRoutingRule = (rule: ModelRoutingRule) => {
    const key = getRuleSearchKey(rule)
    accountSearchRunner.clearKey(key)
    clearAccountSearchStateByKey(key)
    removeRoutingRuleByReference(rules.value, rule)
  }

  const hideAllDropdowns = () => {
    closeAllRoutingRuleDropdowns(showAccountDropdown.value)
  }

  const loadRulesFromApi = async (apiFormat: Record<string, number[]> | null) => {
    rules.value = await hydrateRoutingRulesFromApi(apiFormat, async (id) => {
      const account = await adminAPI.accounts.getById(id)
      return { id: account.id, name: account.name }
    })
  }

  const reset = () => {
    rules.value.forEach((rule) => {
      accountSearchRunner.clearKey(getRuleSearchKey(rule))
    })
    accountSearchRunner.clearAll()
    clearAllAccountSearchState()
    rules.value = []
  }

  return {
    rules,
    accountSearchKeyword,
    accountSearchResults,
    showAccountDropdown,
    getRuleRenderKey,
    getRuleSearchKey,
    searchAccountsByRule,
    selectAccount,
    removeSelectedAccount,
    onAccountSearchFocus,
    addRoutingRule,
    removeRoutingRule,
    hideAllDropdowns,
    loadRulesFromApi,
    reset
  }
}
