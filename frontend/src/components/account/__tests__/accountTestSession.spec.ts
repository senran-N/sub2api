import { describe, expect, it, vi } from 'vitest'
import { parseSSEDataChunk, sortTestModels } from '../accountTestSession'

describe('accountTestSession helpers', () => {
  it('sorts gemini models by predefined priority', () => {
    const sorted = sortTestModels([
      { id: 'gemini-2.0-flash', display_name: 'Gemini 2.0 Flash' } as any,
      { id: 'gemini-3.1-flash-image', display_name: 'Gemini 3.1 Flash Image' } as any,
      { id: 'gemini-2.5-pro', display_name: 'Gemini 2.5 Pro' } as any,
      { id: 'custom-model', display_name: 'Custom Model' } as any
    ])

    expect(sorted.map((model) => model.id)).toEqual([
      'gemini-3.1-flash-image',
      'gemini-2.5-pro',
      'gemini-2.0-flash',
      'custom-model'
    ])
  })

  it('parses sse chunks while preserving an unfinished trailing buffer', () => {
    const first = parseSSEDataChunk(
      '',
      'data: {"type":"test_start","model":"gemini-2.5"}\ndata: {"type":"content","text":"hel'
    )
    expect(first.events).toEqual([{ type: 'test_start', model: 'gemini-2.5' }])

    const second = parseSSEDataChunk(
      first.buffer,
      'lo"}\ndata: {"type":"image","image_url":"data:image/png;base64,AA=="}\n'
    )
    expect(second.events).toEqual([
      { type: 'content', text: 'hello' },
      { type: 'image', image_url: 'data:image/png;base64,AA==' }
    ])
    expect(second.buffer).toBe('')
  })

  it('reports json parse failures without breaking other events', () => {
    const onParseError = vi.fn()
    const result = parseSSEDataChunk(
      '',
      'data: {"type":"test_start"}\ndata: {"type":}\ndata: {"type":"error","error":"boom"}\n',
      onParseError
    )

    expect(onParseError).toHaveBeenCalledTimes(1)
    expect(result.events).toEqual([
      { type: 'test_start' },
      { type: 'error', error: 'boom' }
    ])
  })
})
