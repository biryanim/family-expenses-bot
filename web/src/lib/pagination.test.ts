import { describe, expect, it } from 'vitest'
import { describePage, pageAfterDelete, pageToOffset, parsePage, readTotal } from './pagination'

describe('parsePage', () => {
  it('reads positive integers', () => {
    expect(parsePage('3')).toBe(3)
  })

  it('falls back to the first page on garbage', () => {
    for (const raw of ['', '0', '-1', 'abc', '1.5', '1000000']) {
      expect(parsePage(raw)).toBe(1)
    }
  })
})

describe('pageToOffset', () => {
  it('multiplies by page size', () => {
    expect(pageToOffset(1)).toBe(0)
    expect(pageToOffset(3)).toBe(100)
  })
})

describe('readTotal', () => {
  it('reads the header', () => {
    expect(readTotal(new Headers({ 'X-Total-Count': '120' }))).toBe(120)
  })

  it('returns undefined when the header is missing or broken', () => {
    expect(readTotal(new Headers())).toBeUndefined()
    expect(readTotal(new Headers({ 'X-Total-Count': 'abc' }))).toBeUndefined()
  })
})

describe('describePage', () => {
  it('describes a middle page', () => {
    expect(describePage(2, 50, 120)).toEqual({ from: 51, to: 100, total: 120, hasPrev: true, hasNext: true })
  })

  it('has no next page on the last one', () => {
    expect(describePage(3, 20, 120)).toEqual({ from: 101, to: 120, total: 120, hasPrev: true, hasNext: false })
  })

  it('has no prev page on the first one', () => {
    expect(describePage(1, 50, 120).hasPrev).toBe(false)
  })

  it('describes an empty page', () => {
    expect(describePage(1, 0, 0)).toEqual({ from: 0, to: 0, total: 0, hasPrev: false, hasNext: false })
  })

  it('guesses the next page from the page fill when total is unknown', () => {
    expect(describePage(1, 50).hasNext).toBe(true)
    expect(describePage(1, 49).hasNext).toBe(false)
  })
})

describe('pageAfterDelete', () => {
  it('goes back when the last item of the page is deleted', () => {
    expect(pageAfterDelete(3, 1)).toBe(2)
  })

  it('stays on the first page', () => {
    expect(pageAfterDelete(1, 1)).toBe(1)
  })

  it('stays when items remain', () => {
    expect(pageAfterDelete(3, 2)).toBe(3)
  })
})
