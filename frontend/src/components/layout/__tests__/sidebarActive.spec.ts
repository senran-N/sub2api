import { describe, expect, it } from 'vitest'
import { isSidebarItemActive } from '../sidebarActive'

describe('isSidebarItemActive', () => {
  it('matches exact sidebar paths', () => {
    expect(isSidebarItemActive('/admin/subscriptions', '/admin/subscriptions')).toBe(true)
    expect(isSidebarItemActive('/admin/orders', '/admin/orders')).toBe(true)
  })

  it('keeps nested routes active for regular sidebar sections', () => {
    expect(isSidebarItemActive('/admin/subscriptions/assign', '/admin/subscriptions')).toBe(true)
    expect(isSidebarItemActive('/orders/123', '/orders')).toBe(true)
  })

  it('does not activate order management for sibling payment routes', () => {
    expect(isSidebarItemActive('/admin/orders/plans', '/admin/orders')).toBe(false)
    expect(isSidebarItemActive('/admin/orders/dashboard', '/admin/orders')).toBe(false)
    expect(isSidebarItemActive('/admin/orders/plans', '/admin/orders/plans')).toBe(true)
  })
})
