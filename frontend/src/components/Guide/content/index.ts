import type { OnboardingGuideContent, OnboardingGuideLocale } from './types'

const onboardingGuideLoaders: Record<
  OnboardingGuideLocale,
  () => Promise<{ default: OnboardingGuideContent }>
> = {
  en: () => import('./en'),
  zh: () => import('./zh')
}

export async function loadOnboardingGuideContent(locale: string): Promise<OnboardingGuideContent> {
  const normalizedLocale: OnboardingGuideLocale = locale === 'zh' ? 'zh' : 'en'
  const onboardingGuideModule = await onboardingGuideLoaders[normalizedLocale]()
  return onboardingGuideModule.default
}

export type {
  AdminOnboardingGuideContent,
  OnboardingGuideContent,
  OnboardingGuideStepContent,
  UserOnboardingGuideContent
} from './types'
