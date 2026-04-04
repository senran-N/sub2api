import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import type { Proxy } from '@/types'
import {
  createDefaultProxyBatchParseState,
  createDefaultProxyCreateForm,
  createDefaultProxyEditForm
} from '../proxyForm'
import { useProxyFormActions } from '../useProxyFormActions'

const { batchCreate, createProxy, updateProxy } = vi.hoisted(() => ({
  batchCreate: vi.fn(),
  createProxy: vi.fn(),
  updateProxy: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    proxies: {
      batchCreate,
      create: createProxy,
      update: updateProxy
    }
  }
}))

function createProxyRecord(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy',
    protocol: 'http',
    host: 'proxy.local',
    port: 8080,
    username: null,
    password: null,
    status: 'active',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

function createComposable() {
  const showCreateModal = ref(true)
  const createMode = ref<'standard' | 'batch'>('batch')
  const createForm = createDefaultProxyCreateForm()
  const createPasswordVisible = ref(true)
  const batchInput = ref('http://proxy.local:8080')
  const batchParseResult = createDefaultProxyBatchParseState()
  const showImportData = ref(true)
  const editingProxy = ref<Proxy | null>(null)
  const editForm = createDefaultProxyEditForm()
  const showEditModal = ref(false)
  const editPasswordVisible = ref(true)
  const editPasswordDirty = ref(true)
  const submitting = ref(false)
  const loadProxies = vi.fn(async () => {})
  const showSuccess = vi.fn()
  const showError = vi.fn()
  const showInfo = vi.fn()

  const composable = useProxyFormActions({
    showCreateModal,
    createMode,
    createForm,
    createPasswordVisible,
    batchInput,
    batchParseResult,
    showImportData,
    editingProxy,
    editForm,
    showEditModal,
    editPasswordVisible,
    editPasswordDirty,
    submitting,
    loadProxies,
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
    showSuccess,
    showError,
    showInfo
  })

  return {
    batchInput,
    batchParseResult,
    composable,
    createForm,
    createMode,
    editForm,
    editingProxy,
    loadProxies,
    showCreateModal,
    showEditModal,
    editPasswordDirty,
    showError,
    showImportData,
    showInfo,
    showSuccess
  }
}

describe('useProxyFormActions', () => {
  beforeEach(() => {
    batchCreate.mockReset()
    createProxy.mockReset()
    updateProxy.mockReset()
  })

  it('resets create modal state and parses batch input', () => {
    const setup = createComposable()
    Object.assign(setup.createForm, {
      name: 'Edge',
      host: 'proxy.local',
      username: 'alice'
    })

    setup.composable.parseBatchInput()
    expect(setup.batchParseResult.valid).toBe(1)

    setup.composable.closeCreateModal()
    expect(setup.showCreateModal.value).toBe(false)
    expect(setup.createMode.value).toBe('standard')
    expect(setup.createForm.name).toBe('')
    expect(setup.batchInput.value).toBe('')
    expect(setup.batchParseResult.valid).toBe(0)
  })

  it('creates a proxy after validation and reloads the list', async () => {
    const setup = createComposable()
    Object.assign(setup.createForm, {
      name: ' Edge ',
      protocol: 'https',
      host: ' proxy.local ',
      port: 443
    })
    createProxy.mockResolvedValue(createProxyRecord())

    await setup.composable.handleCreateProxy()

    expect(createProxy).toHaveBeenCalledWith({
      name: 'Edge',
      protocol: 'https',
      host: 'proxy.local',
      port: 443,
      username: null,
      password: null
    })
    expect(setup.showSuccess).toHaveBeenCalledWith('admin.proxies.proxyCreated')
    expect(setup.loadProxies).toHaveBeenCalledTimes(1)
  })

  it('hydrates edit state and updates proxy payloads', async () => {
    const setup = createComposable()
    const proxy = createProxyRecord({
      id: 7,
      protocol: 'socks5',
      host: 'edge.local',
      port: 1080,
      username: 'alice',
      password: 'secret',
      status: 'inactive'
    })
    updateProxy.mockResolvedValue(proxy)

    setup.composable.handleEdit(proxy)
    expect(setup.editingProxy.value?.id).toBe(7)
    expect(setup.showEditModal.value).toBe(true)
    expect(setup.editForm.host).toBe('edge.local')

    setup.editForm.password = ' next '
    setup.editForm.status = 'active'
    setup.editPasswordDirty.value = true
    await setup.composable.handleUpdateProxy()

    expect(updateProxy).toHaveBeenCalledWith(7, {
      name: 'Proxy',
      protocol: 'socks5',
      host: 'edge.local',
      port: 1080,
      username: 'alice',
      password: 'next',
      status: 'active'
    })
    expect(setup.loadProxies).toHaveBeenCalledTimes(1)
  })

  it('batch creates parsed proxies and closes import dialog after external import', async () => {
    const setup = createComposable()
    const proxies = [
      { protocol: 'http', host: 'a.local', port: 80, username: '', password: '' },
      { protocol: 'https', host: 'b.local', port: 443, username: 'alice', password: 'secret' }
    ]
    setup.batchParseResult.valid = 2
    setup.batchParseResult.proxies = proxies
    batchCreate.mockResolvedValue({
      created: 2,
      skipped: 0
    })

    await setup.composable.handleBatchCreate()
    setup.composable.handleDataImported()

    expect(batchCreate).toHaveBeenCalledWith(proxies)
    expect(setup.showSuccess).toHaveBeenCalledWith(
      'admin.proxies.batchImportSuccess:{"created":2,"skipped":0}'
    )
    expect(setup.showImportData.value).toBe(false)
    expect(setup.loadProxies).toHaveBeenCalledTimes(2)
  })
})
