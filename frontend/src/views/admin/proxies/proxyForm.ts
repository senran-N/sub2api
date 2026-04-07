import type {
  CreateProxyRequest,
  Proxy,
  ProxyProtocol,
  UpdateProxyRequest
} from '@/types'
import type { ProxyBatchEntry, ProxyBatchParseResult } from './proxyUtils'

export interface ProxyCreateForm {
  name: string
  protocol: ProxyProtocol
  host: string
  port: number
  username: string
  password: string
}

export interface ProxyEditForm extends ProxyCreateForm {
  status: 'active' | 'inactive'
}

export interface ProxyBatchParseState extends ProxyBatchParseResult {
  proxies: ProxyBatchEntry[]
}

export function createDefaultProxyCreateForm(): ProxyCreateForm {
  return {
    name: '',
    protocol: 'http',
    host: '',
    port: 8080,
    username: '',
    password: ''
  }
}

export function createDefaultProxyEditForm(): ProxyEditForm {
  return {
    ...createDefaultProxyCreateForm(),
    status: 'active'
  }
}

export function createDefaultProxyBatchParseState(): ProxyBatchParseState {
  return {
    total: 0,
    valid: 0,
    invalid: 0,
    duplicate: 0,
    proxies: []
  }
}

export function resetProxyCreateForm(form: ProxyCreateForm): void {
  Object.assign(form, createDefaultProxyCreateForm())
}

export function resetProxyEditForm(form: ProxyEditForm): void {
  Object.assign(form, createDefaultProxyEditForm())
}

export function resetProxyBatchParseState(state: ProxyBatchParseState): void {
  Object.assign(state, createDefaultProxyBatchParseState())
}

export function hydrateProxyEditForm(form: ProxyEditForm, proxy: Proxy): void {
  Object.assign(form, createDefaultProxyEditForm(), {
    name: proxy.name,
    protocol: proxy.protocol,
    host: proxy.host,
    port: proxy.port,
    username: proxy.username || '',
    password: proxy.password || '',
    status: proxy.status
  })
}

export function getProxyFormValidationError(
  form: Pick<ProxyCreateForm, 'name' | 'host' | 'port'>
): string | null {
  if (!form.name.trim()) {
    return 'admin.proxies.nameRequired'
  }
  if (!form.host.trim()) {
    return 'admin.proxies.hostRequired'
  }
  if (form.port < 1 || form.port > 65535) {
    return 'admin.proxies.portInvalid'
  }
  return null
}

export function buildCreateProxyRequest(form: ProxyCreateForm): CreateProxyRequest {
  return {
    name: form.name.trim(),
    protocol: form.protocol,
    host: form.host.trim(),
    port: form.port,
    username: form.username.trim() || null,
    password: form.password.trim() || null
  }
}

export function buildUpdateProxyRequest(
  form: ProxyEditForm,
  passwordDirty: boolean
): UpdateProxyRequest {
  const request: UpdateProxyRequest = {
    name: form.name.trim(),
    protocol: form.protocol,
    host: form.host.trim(),
    port: form.port,
    username: form.username.trim() || null,
    status: form.status
  }

  if (passwordDirty) {
    request.password = form.password.trim() || null
  }

  return request
}
