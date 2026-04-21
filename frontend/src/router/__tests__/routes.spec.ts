import { describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    isAuthenticated: false,
    isAdmin: false,
    isSimpleMode: false,
    checkAuth: vi.fn()
  })
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    siteName: 'Sub2API',
    backendModeEnabled: false,
    cachedPublicSettings: null
  })
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: []
  })
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false }
  })
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn()
  })
}))

const { default: router } = await import('@/router')

describe('router admin channel route meta', () => {
  it('为渠道管理配置 i18n 标题和描述键', () => {
    const route = router.resolve('/admin/channels')

    expect(route.meta.titleKey).toBe('admin.channels.title')
    expect(route.meta.descriptionKey).toBe('admin.channels.description')
  })

  it('为登录页配置存在的 i18n 标题键', () => {
    const route = router.resolve('/login')

    expect(route.meta.titleKey).toBe('auth.signIn')
  })
})
