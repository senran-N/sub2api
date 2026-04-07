import { applyInterceptWarmup, applyTempUnschedConfig } from './credentialsBuilder'
import type { ModelMapping, TempUnschedRuleForm } from './credentialsBuilder'
import type { Translate } from './accountModalShared'

export interface ApplySharedAccountCredentialsStateOptions {
  interceptWarmupRequests: boolean
  tempUnschedEnabled: boolean
  tempUnschedRules: TempUnschedRuleForm[]
  showError: (message: string) => void
  t: Translate
  interceptContext?: 'create' | 'edit'
}

export interface ApplyTempUnschedCredentialsStateOptions {
  tempUnschedEnabled: boolean
  tempUnschedRules: TempUnschedRuleForm[]
  showError: (message: string) => void
  t: Translate
}

export function applyTempUnschedCredentialsState(
  credentials: Record<string, unknown>,
  options: ApplyTempUnschedCredentialsStateOptions
): boolean {
  if (!applyTempUnschedConfig(credentials, options.tempUnschedEnabled, options.tempUnschedRules)) {
    options.showError(options.t('admin.accounts.tempUnschedulable.rulesInvalid'))
    return false
  }
  return true
}

export function applySharedAccountCredentialsState(
  credentials: Record<string, unknown>,
  options: ApplySharedAccountCredentialsStateOptions
): boolean {
  applyInterceptWarmup(credentials, options.interceptWarmupRequests, options.interceptContext ?? 'edit')
  return applyTempUnschedCredentialsState(credentials, options)
}

export function appendEmptyModelMapping(target: ModelMapping[]) {
  target.push({ from: '', to: '' })
}

export function removeModelMappingAt(target: ModelMapping[], index: number) {
  target.splice(index, 1)
}

export function appendPresetModelMapping(
  target: ModelMapping[],
  from: string,
  to: string,
  onDuplicate: (model: string) => void
) {
  if (target.some((mapping) => mapping.from === from)) {
    onDuplicate(from)
    return
  }
  target.push({ from, to })
}

export function confirmCustomErrorCodeSelection(
  code: number,
  confirmFn: (message: string) => boolean,
  t: Translate
) {
  if (code === 429) {
    return confirmFn(t('admin.accounts.customErrorCodes429Warning'))
  }
  if (code === 529) {
    return confirmFn(t('admin.accounts.customErrorCodes529Warning'))
  }
  return true
}
