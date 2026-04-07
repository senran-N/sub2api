import { beforeEach, describe, expect, it, vi } from 'vitest'
import { computed } from 'vue'
import { useAccountsViewColumns } from '../accounts/useAccountsViewColumns'

describe('useAccountsViewColumns', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('loads saved visibility state, toggles columns, and exposes filtered columns', async () => {
    localStorage.setItem('account-hidden-columns', JSON.stringify(['proxy']))
    const onUsageColumnShown = vi.fn().mockResolvedValue(undefined)
    const state = useAccountsViewColumns({
      t: (key: string) => key,
      isSimpleMode: computed(() => false),
      onUsageColumnShown
    })

    expect(state.isColumnVisible('proxy')).toBe(false)
    expect(state.toggleableColumns.value.some((column) => column.key === 'proxy')).toBe(true)
    expect(state.cols.value.some((column) => column.key === 'proxy')).toBe(false)

    state.toggleColumn('proxy')
    expect(state.isColumnVisible('proxy')).toBe(true)
    expect(JSON.parse(localStorage.getItem('account-hidden-columns') || '[]')).not.toContain(
      'proxy'
    )

    state.toggleColumn('usage')
    state.toggleColumn('usage')
    await Promise.resolve()
    expect(onUsageColumnShown).toHaveBeenCalledTimes(1)
  })

  it('requests stats refresh only when showing usage-related columns', async () => {
    localStorage.setItem('account-hidden-columns', JSON.stringify(['usage', 'today_stats']))
    const onUsageColumnShown = vi.fn().mockResolvedValue(undefined)
    const state = useAccountsViewColumns({
      t: (key: string) => key,
      isSimpleMode: computed(() => true),
      onUsageColumnShown
    })

    state.toggleColumn('notes')
    state.toggleColumn('notes')
    state.toggleColumn('usage')
    state.toggleColumn('today_stats')
    await Promise.resolve()

    expect(onUsageColumnShown).toHaveBeenCalledTimes(2)
  })
})
