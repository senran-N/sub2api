import type { DriveStep } from 'driver.js'
import type {
  AdminOnboardingGuideContent,
  UserOnboardingGuideContent
} from '@/components/Guide/content/types'

/**
 * 管理员完整引导流程
 * 交互式引导：指引用户实际操作
 * @param content 按语言加载的引导文案
 * @param isSimpleMode 是否为简易模式（简易模式下会过滤分组相关步骤）
 */
export const getAdminSteps = (
  content: AdminOnboardingGuideContent,
  isSimpleMode = false
): DriveStep[] => {
  const allSteps: DriveStep[] = [
    {
      popover: {
        title: content.welcome.title,
        description: content.welcome.description,
        align: 'center',
        nextBtnText: content.welcome.nextBtn,
        prevBtnText: content.welcome.prevBtn
      }
    },
    {
      element: '#sidebar-group-manage',
      popover: {
        title: content.groupManage.title,
        description: content.groupManage.description,
        side: 'right',
        align: 'center',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="groups-create-btn"]',
      popover: {
        title: content.createGroup.title,
        description: content.createGroup.description,
        side: 'bottom',
        align: 'end',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="group-form-name"]',
      popover: {
        title: content.groupName.title,
        description: content.groupName.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="group-form-platform"]',
      popover: {
        title: content.groupPlatform.title,
        description: content.groupPlatform.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="group-form-multiplier"]',
      popover: {
        title: content.groupMultiplier.title,
        description: content.groupMultiplier.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="group-form-exclusive"]',
      popover: {
        title: content.groupExclusive.title,
        description: content.groupExclusive.description,
        side: 'top',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="group-form-submit"]',
      popover: {
        title: content.groupSubmit.title,
        description: content.groupSubmit.description,
        side: 'left',
        align: 'center',
        showButtons: ['close']
      }
    },
    {
      element: '#sidebar-channel-manage',
      popover: {
        title: content.accountManage.title,
        description: content.accountManage.description,
        side: 'right',
        align: 'center',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="accounts-create-btn"]',
      popover: {
        title: content.createAccount.title,
        description: content.createAccount.description,
        side: 'bottom',
        align: 'end',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="account-form-name"]',
      popover: {
        title: content.accountName.title,
        description: content.accountName.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="account-form-platform"]',
      popover: {
        title: content.accountPlatform.title,
        description: content.accountPlatform.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="account-form-type"]',
      popover: {
        title: content.accountType.title,
        description: content.accountType.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="account-form-priority"]',
      popover: {
        title: content.accountPriority.title,
        description: content.accountPriority.description,
        side: 'top',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="account-form-groups"]',
      popover: {
        title: content.accountGroups.title,
        description: content.accountGroups.description,
        side: 'top',
        align: 'center',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="account-form-submit"]',
      popover: {
        title: content.accountSubmit.title,
        description: content.accountSubmit.description,
        side: 'left',
        align: 'center',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="sidebar-my-keys"]',
      popover: {
        title: content.keyManage.title,
        description: content.keyManage.description,
        side: 'right',
        align: 'center',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="keys-create-btn"]',
      popover: {
        title: content.createKey.title,
        description: content.createKey.description,
        side: 'bottom',
        align: 'end',
        showButtons: ['close']
      }
    },
    {
      element: '[data-tour="key-form-name"]',
      popover: {
        title: content.keyName.title,
        description: content.keyName.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="key-form-group"]',
      popover: {
        title: content.keyGroup.title,
        description: content.keyGroup.description,
        side: 'right',
        align: 'start',
        showButtons: ['next', 'previous']
      }
    },
    {
      element: '[data-tour="key-form-submit"]',
      popover: {
        title: content.keySubmit.title,
        description: content.keySubmit.description,
        side: 'left',
        align: 'center',
        showButtons: ['close']
      }
    }
  ]

  if (isSimpleMode) {
    return allSteps.filter((step) => {
      const element = step.element as string | undefined
      return !element || (
        !element.includes('sidebar-group-manage') &&
        !element.includes('groups-create-btn') &&
        !element.includes('group-form-') &&
        !element.includes('account-form-groups')
      )
    })
  }

  return allSteps
}

/**
 * 普通用户引导流程
 */
export const getUserSteps = (content: UserOnboardingGuideContent): DriveStep[] => [
  {
    popover: {
      title: content.welcome.title,
      description: content.welcome.description,
      align: 'center',
      nextBtnText: content.welcome.nextBtn,
      prevBtnText: content.welcome.prevBtn
    }
  },
  {
    element: '[data-tour="sidebar-my-keys"]',
    popover: {
      title: content.keyManage.title,
      description: content.keyManage.description,
      side: 'right',
      align: 'center',
      showButtons: ['close']
    }
  },
  {
    element: '[data-tour="keys-create-btn"]',
    popover: {
      title: content.createKey.title,
      description: content.createKey.description,
      side: 'bottom',
      align: 'end',
      showButtons: ['close']
    }
  },
  {
    element: '[data-tour="key-form-name"]',
    popover: {
      title: content.keyName.title,
      description: content.keyName.description,
      side: 'right',
      align: 'start',
      showButtons: ['next', 'previous']
    }
  },
  {
    element: '[data-tour="key-form-group"]',
    popover: {
      title: content.keyGroup.title,
      description: content.keyGroup.description,
      side: 'right',
      align: 'start',
      showButtons: ['next', 'previous']
    }
  },
  {
    element: '[data-tour="key-form-submit"]',
    popover: {
      title: content.keySubmit.title,
      description: content.keySubmit.description,
      side: 'left',
      align: 'center',
      showButtons: ['close']
    }
  }
]
