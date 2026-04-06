export interface OnboardingGuideStepContent {
  title: string
  description: string
  nextBtn?: string
  prevBtn?: string
}

export interface AdminOnboardingGuideContent {
  welcome: OnboardingGuideStepContent
  groupManage: OnboardingGuideStepContent
  createGroup: OnboardingGuideStepContent
  groupName: OnboardingGuideStepContent
  groupPlatform: OnboardingGuideStepContent
  groupMultiplier: OnboardingGuideStepContent
  groupExclusive: OnboardingGuideStepContent
  groupSubmit: OnboardingGuideStepContent
  accountManage: OnboardingGuideStepContent
  createAccount: OnboardingGuideStepContent
  accountName: OnboardingGuideStepContent
  accountPlatform: OnboardingGuideStepContent
  accountType: OnboardingGuideStepContent
  accountPriority: OnboardingGuideStepContent
  accountGroups: OnboardingGuideStepContent
  accountSubmit: OnboardingGuideStepContent
  keyManage: OnboardingGuideStepContent
  createKey: OnboardingGuideStepContent
  keyName: OnboardingGuideStepContent
  keyGroup: OnboardingGuideStepContent
  keySubmit: OnboardingGuideStepContent
}

export interface UserOnboardingGuideContent {
  welcome: OnboardingGuideStepContent
  keyManage: OnboardingGuideStepContent
  createKey: OnboardingGuideStepContent
  keyName: OnboardingGuideStepContent
  keyGroup: OnboardingGuideStepContent
  keySubmit: OnboardingGuideStepContent
}

export interface OnboardingGuideContent {
  admin: AdminOnboardingGuideContent
  user: UserOnboardingGuideContent
}

export type OnboardingGuideLocale = 'en' | 'zh'
