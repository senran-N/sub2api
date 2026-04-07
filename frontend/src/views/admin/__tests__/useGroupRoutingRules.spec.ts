import { describe, expect, it, vi } from 'vitest'
import {
  appendRoutingRuleAccount,
  closeAllRoutingRuleDropdowns,
  createEmptyRoutingRule,
  hydrateRoutingRulesFromApi,
  pushRoutingRule,
  removeRoutingRuleAccount,
  removeRoutingRuleByReference
} from '../groups/useGroupRoutingRules'

describe('group routing rule helpers', () => {
  it('creates, appends, and removes rules by reference', () => {
    const rules = [createEmptyRoutingRule()]

    pushRoutingRule(rules)
    expect(rules).toHaveLength(2)
    expect(rules[1]).toEqual({
      pattern: '',
      accounts: []
    })

    expect(removeRoutingRuleByReference(rules, rules[0])).toBe(true)
    expect(rules).toHaveLength(1)
    expect(removeRoutingRuleByReference(rules, createEmptyRoutingRule())).toBe(false)
  })

  it('adds unique accounts and removes selected accounts', () => {
    const rule = createEmptyRoutingRule()

    expect(appendRoutingRuleAccount(rule, { id: 3, name: 'alpha' })).toBe(true)
    expect(appendRoutingRuleAccount(rule, { id: 3, name: 'alpha again' })).toBe(false)
    expect(rule.accounts).toEqual([{ id: 3, name: 'alpha' }])

    removeRoutingRuleAccount(rule, 3)
    expect(rule.accounts).toEqual([])
  })

  it('hydrates rules from api data and preserves missing accounts as #id placeholders', async () => {
    const loadAccount = vi.fn(async (id: number) => {
      if (id === 7) {
        throw new Error('missing')
      }
      return { id, name: `account-${id}` }
    })

    await expect(
      hydrateRoutingRulesFromApi(
        {
          'claude-*': [1, 7]
        },
        loadAccount
      )
    ).resolves.toEqual([
      {
        pattern: 'claude-*',
        accounts: [
          { id: 1, name: 'account-1' },
          { id: 7, name: '#7' }
        ]
      }
    ])
  })

  it('closes all routing dropdowns in place', () => {
    const dropdowns = {
      a: true,
      b: false,
      c: true
    }

    closeAllRoutingRuleDropdowns(dropdowns)
    expect(dropdowns).toEqual({
      a: false,
      b: false,
      c: false
    })
  })
})
