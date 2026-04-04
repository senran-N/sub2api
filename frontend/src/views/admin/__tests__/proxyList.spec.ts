import { describe, expect, it } from 'vitest'
import {
  applyProxyPageChange,
  applyProxyPageSizeChange,
  buildProxyListFilters,
  resetProxyListPage
} from '../proxyList'

describe('proxyList helpers', () => {
  it('normalizes list filters', () => {
    expect(
      buildProxyListFilters(
        {
          protocol: 'socks5',
          status: 'active'
        },
        '  edge  '
      )
    ).toEqual({
      protocol: 'socks5',
      status: 'active',
      search: 'edge'
    })

    expect(
      buildProxyListFilters(
        {
          protocol: '',
          status: ''
        },
        '   '
      )
    ).toEqual({
      protocol: undefined,
      status: undefined,
      search: undefined
    })
  })

  it('applies pagination mutations', () => {
    const pagination = {
      page: 3,
      page_size: 20
    }

    applyProxyPageChange(pagination, 5)
    expect(pagination.page).toBe(5)

    applyProxyPageSizeChange(pagination, 50)
    expect(pagination).toEqual({
      page: 1,
      page_size: 50
    })

    pagination.page = 9
    resetProxyListPage(pagination)
    expect(pagination.page).toBe(1)
  })
})
