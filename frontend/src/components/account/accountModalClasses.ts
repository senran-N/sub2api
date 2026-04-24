type ModalTone = 'rose' | 'orange' | 'purple' | 'amber' | 'green' | 'blue' | 'emerald' | 'danger'
type ModeTone = 'accent' | 'purple' | 'danger'

function joinClassNames(classNames: Array<string | false | null | undefined>) {
  return classNames.filter(Boolean).join(' ')
}

function prefixedClass(prefix: string, block: string, suffix: string) {
  return `${prefix}__${block}--${suffix}`
}

export function getAccountModalModeToggleClasses(
  prefix: string,
  isSelected: boolean,
  tone: ModeTone
) {
  return joinClassNames([
    `${prefix}__mode-toggle ${prefix}__mode-toggle-control flex-1 text-sm font-medium transition-all`,
    prefixedClass(prefix, 'mode-toggle', isSelected ? tone : 'idle')
  ])
}

export function getAccountModalSwitchTrackClasses(prefix: string, isEnabled: boolean) {
  return joinClassNames([
    `${prefix}__switch relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none`,
    prefixedClass(prefix, 'switch', isEnabled ? 'enabled' : 'disabled')
  ])
}

export function getAccountModalSwitchThumbClasses(prefix: string, isEnabled: boolean) {
  return joinClassNames([
    `${prefix}__switch-thumb pointer-events-none inline-block h-5 w-5 transform rounded-full shadow ring-0 transition duration-200 ease-in-out`,
    isEnabled ? 'translate-x-5' : 'translate-x-0'
  ])
}

export function getAccountModalStatusChipClasses(
  prefix: string,
  isSelected: boolean,
  tone: ModeTone = 'danger'
) {
  return joinClassNames([
    `${prefix}__status-chip ${prefix}__status-chip-control text-sm font-medium transition-colors`,
    prefixedClass(prefix, 'status-chip', isSelected ? tone : 'idle')
  ])
}

export function getCreateChoiceCardClasses(
  isSelected: boolean,
  tone: ModalTone,
  isDisabled = false
) {
  return joinClassNames([
    'create-account-modal__choice-card create-account-modal__choice-card-control flex items-center gap-3 border-2 text-left transition-all',
    prefixedClass('create-account-modal', 'choice-card', isSelected ? tone : 'idle'),
    isDisabled && 'create-account-modal__choice-card--disabled'
  ])
}

export function getCreateChoiceIconClasses(isSelected: boolean, tone: ModalTone) {
  return joinClassNames([
    'create-account-modal__choice-icon create-account-modal__choice-icon-control flex h-8 w-8 shrink-0 items-center justify-center',
    prefixedClass('create-account-modal', 'choice-icon', isSelected ? tone : 'idle')
  ])
}

export function getCreateToneTagClasses(tone: ModalTone) {
  return joinClassNames([
    'create-account-modal__tone-tag create-account-modal__tone-tag-control text-[10px] font-semibold',
    prefixedClass('create-account-modal', 'tone-tag', tone)
  ])
}

export function getCreateValidationInputClasses(hasError: boolean, extraClassName = '') {
  return joinClassNames(['input', extraClassName, hasError && 'input-error'])
}

export function getCreateRadioOptionClasses(isSelected: boolean) {
  return joinClassNames([
    'create-account-modal__radio-option',
    isSelected && 'create-account-modal__radio-option--active'
  ])
}

export function getCreateSegmentOptionClasses(isSelected: boolean) {
  return joinClassNames([
    'create-account-modal__segment-option',
    prefixedClass('create-account-modal', 'segment-option', isSelected ? 'active' : 'idle')
  ])
}

export function getEditToneNoticeClasses(tone: Extract<ModalTone, 'purple' | 'amber' | 'blue' | 'danger'>) {
  return joinClassNames([
    'edit-account-modal__notice edit-account-modal__notice-card border',
    prefixedClass('edit-account-modal', 'notice', tone)
  ])
}
