import { describe, expect, it } from 'vitest'
import { useAccountsViewDialogs } from '../useAccountsViewDialogs'

describe('useAccountsViewDialogs', () => {
  it('opens export dialog with proxy export enabled by default', () => {
    const dialogs = useAccountsViewDialogs()
    dialogs.includeProxyOnExport.value = false

    dialogs.openExportDataDialog()

    expect(dialogs.includeProxyOnExport.value).toBe(true)
    expect(dialogs.showExportDataDialog.value).toBe(true)
  })

  it('opens and syncs action menu state', () => {
    const dialogs = useAccountsViewDialogs()
    const button = {
      getBoundingClientRect: () => ({
        left: 400,
        top: 100,
        bottom: 120,
        width: 40,
        height: 20
      })
    } as unknown as HTMLElement

    dialogs.openMenu(
      {
        id: 7,
        name: 'Account',
        platform: 'openai',
        type: 'oauth'
      } as any,
      {
        currentTarget: button,
        clientX: 520,
        clientY: 120
      } as MouseEvent
    )

    expect(dialogs.menu.show).toBe(true)
    expect(dialogs.menu.acc?.id).toBe(7)
    expect(dialogs.menu.pos).toBeTruthy()

    dialogs.syncMenuAccount({
      id: 7,
      name: 'Account Updated',
      platform: 'openai',
      type: 'oauth'
    } as any)
    expect(dialogs.menu.acc?.name).toBe('Account Updated')

    dialogs.closeActionMenu()
    expect(dialogs.menu.show).toBe(false)
  })
})
