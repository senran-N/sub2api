import { beforeEach, describe, expect, it, vi } from 'vitest'
import {
  buildKeysGroupDropdownPosition,
  buildKeysMoreMenuPosition
} from '../keys/keysOverlays'

describe('keysOverlays helpers', () => {
  beforeEach(() => {
    vi.stubGlobal('innerWidth', 1280)
    vi.stubGlobal('innerHeight', 800)
  })

  it('positions more menu within viewport bounds', () => {
    const button = {
      getBoundingClientRect: () => ({
        top: 100,
        bottom: 130,
        left: 1200
      })
    } as unknown as HTMLElement

    expect(buildKeysMoreMenuPosition(button)).toEqual({
      top: 134,
      left: 1080
    })
  })

  it('positions group selector above or below depending on space', () => {
    const belowButton = {
      getBoundingClientRect: () => ({
        top: 100,
        bottom: 130,
        left: 200
      })
    } as unknown as HTMLElement
    const aboveButton = {
      getBoundingClientRect: () => ({
        top: 700,
        bottom: 730,
        left: 300
      })
    } as unknown as HTMLElement

    expect(buildKeysGroupDropdownPosition(belowButton)).toEqual({
      top: 134,
      left: 200
    })
    expect(buildKeysGroupDropdownPosition(aboveButton)).toEqual({
      bottom: 104,
      left: 300
    })
  })
})
