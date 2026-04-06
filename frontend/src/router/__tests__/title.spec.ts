import { describe, expect, it } from 'vitest'
import type { CustomMenuItem } from '@/types'
import { resolveDocumentTitle, resolveRouteDocumentTitle } from '@/router/title'

function createCustomMenuItem(overrides: Partial<CustomMenuItem> = {}): CustomMenuItem {
  return {
    id: 'docs',
    label: 'Docs',
    icon_svg: '',
    url: 'https://example.com',
    visibility: 'user',
    sort_order: 0,
    ...overrides
  }
}

describe('resolveDocumentTitle', () => {
  it('路由存在标题时，使用“路由标题 - 站点名”格式', () => {
    expect(resolveDocumentTitle('Usage Records', 'My Site')).toBe('Usage Records - My Site')
  })

  it('路由无标题时，回退到站点名', () => {
    expect(resolveDocumentTitle(undefined, 'My Site')).toBe('My Site')
  })

  it('站点名为空时，回退默认站点名', () => {
    expect(resolveDocumentTitle('Dashboard', '')).toBe('Dashboard - Sub2API')
    expect(resolveDocumentTitle(undefined, '   ')).toBe('Sub2API')
  })

  it('站点名变更时仅影响后续路由标题计算', () => {
    const before = resolveDocumentTitle('Admin Dashboard', 'Alpha')
    const after = resolveDocumentTitle('Admin Dashboard', 'Beta')

    expect(before).toBe('Admin Dashboard - Alpha')
    expect(after).toBe('Admin Dashboard - Beta')
  })
})

describe('resolveRouteDocumentTitle', () => {
  it('自定义页面优先使用菜单标题', () => {
    const route = {
      name: 'CustomPage',
      meta: { title: 'Custom Page' },
      params: { id: 'docs' }
    } as Parameters<typeof resolveRouteDocumentTitle>[0]

    expect(
      resolveRouteDocumentTitle(route, {
        siteName: 'My Site',
        publicCustomMenuItems: [createCustomMenuItem()],
        adminCustomMenuItems: [],
        isAdmin: false
      })
    ).toBe('Docs - My Site')
  })
})
