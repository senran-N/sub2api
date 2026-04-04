import { describe, expect, it } from 'vitest'
import type { Proxy } from '@/types'
import {
  buildCreateProxyRequest,
  buildUpdateProxyRequest,
  createDefaultProxyBatchParseState,
  createDefaultProxyCreateForm,
  createDefaultProxyEditForm,
  getProxyFormValidationError,
  hydrateProxyEditForm,
  resetProxyBatchParseState,
  resetProxyCreateForm,
  resetProxyEditForm
} from '../proxyForm'

function createProxy(overrides: Partial<Proxy> = {}): Proxy {
  return {
    id: 1,
    name: 'Proxy',
    protocol: 'socks5',
    host: 'proxy.local',
    port: 1080,
    username: 'alice',
    password: 'secret',
    status: 'inactive',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    ...overrides
  }
}

describe('proxyForm helpers', () => {
  it('creates and resets create, edit, and batch parse state', () => {
    const createForm = createDefaultProxyCreateForm()
    createForm.name = 'Changed'
    resetProxyCreateForm(createForm)
    expect(createForm).toEqual(createDefaultProxyCreateForm())

    const editForm = createDefaultProxyEditForm()
    editForm.status = 'inactive'
    resetProxyEditForm(editForm)
    expect(editForm).toEqual(createDefaultProxyEditForm())

    const batchState = createDefaultProxyBatchParseState()
    batchState.valid = 2
    batchState.proxies = [{ protocol: 'http', host: 'a', port: 80, username: '', password: '' }]
    resetProxyBatchParseState(batchState)
    expect(batchState).toEqual(createDefaultProxyBatchParseState())
  })

  it('hydrates edit form from proxy data', () => {
    const editForm = createDefaultProxyEditForm()

    hydrateProxyEditForm(editForm, createProxy())

    expect(editForm).toEqual({
      name: 'Proxy',
      protocol: 'socks5',
      host: 'proxy.local',
      port: 1080,
      username: 'alice',
      password: 'secret',
      status: 'inactive'
    })
  })

  it('validates required proxy fields before submission', () => {
    expect(
      getProxyFormValidationError({
        name: ' ',
        host: 'proxy.local',
        port: 8080
      })
    ).toBe('admin.proxies.nameRequired')

    expect(
      getProxyFormValidationError({
        name: 'proxy',
        host: '',
        port: 8080
      })
    ).toBe('admin.proxies.hostRequired')

    expect(
      getProxyFormValidationError({
        name: 'proxy',
        host: 'proxy.local',
        port: 70000
      })
    ).toBe('admin.proxies.portInvalid')

    expect(
      getProxyFormValidationError({
        name: 'proxy',
        host: 'proxy.local',
        port: 8080
      })
    ).toBeNull()
  })

  it('builds create and update payloads with trimmed optional credentials', () => {
    const createForm = createDefaultProxyCreateForm()
    Object.assign(createForm, {
      name: '  Edge  ',
      protocol: 'https',
      host: ' proxy.local ',
      port: 443,
      username: ' alice ',
      password: ' secret '
    })

    expect(buildCreateProxyRequest(createForm)).toEqual({
      name: 'Edge',
      protocol: 'https',
      host: 'proxy.local',
      port: 443,
      username: 'alice',
      password: 'secret'
    })

    const editForm = createDefaultProxyEditForm()
    Object.assign(editForm, {
      name: '  Edge  ',
      protocol: 'http',
      host: ' proxy.local ',
      port: 80,
      username: ' ',
      password: ' ',
      status: 'active'
    })

    expect(buildUpdateProxyRequest(editForm, false)).toEqual({
      name: 'Edge',
      protocol: 'http',
      host: 'proxy.local',
      port: 80,
      username: null,
      status: 'active'
    })

    expect(buildUpdateProxyRequest(editForm, true)).toEqual({
      name: 'Edge',
      protocol: 'http',
      host: 'proxy.local',
      port: 80,
      username: null,
      password: null,
      status: 'active'
    })
  })
})
