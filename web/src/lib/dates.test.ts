import { describe, expect, it } from 'vitest'
import { isValidDay, periodToRange, toDateTimeLocalValue, toRfc3339Local } from './dates'

describe('periodToRange', () => {
  it('converts a week to local midnights', () => {
    expect(periodToRange('2026-10-01', '2026-10-07')).toEqual({
      ok: true,
      from: '2026-10-01T00:00:00+03:00',
      to: '2026-10-08T00:00:00+03:00',
    })
  })

  it('handles only start', () => {
    expect(periodToRange('2026-10-01', '')).toEqual({ ok: true, from: '2026-10-01T00:00:00+03:00' })
  })

  it('handles only end', () => {
    expect(periodToRange('', '2026-10-31')).toEqual({ ok: true, to: '2026-11-01T00:00:00+03:00' })
  })

  it('handles empty period', () => {
    expect(periodToRange('', '')).toEqual({ ok: true })
  })

  it('rejects inverted period', () => {
    expect(periodToRange('2026-10-10', '2026-10-01').ok).toBe(false)
  })
})

describe('toRfc3339Local', () => {
  it('formats with offset and zero seconds', () => {
    expect(toRfc3339Local(new Date(2026, 9, 21, 14, 35))).toBe('2026-10-21T14:35:00+03:00')
  })
})

describe('toDateTimeLocalValue', () => {
  it('formats in local time without seconds', () => {
    expect(toDateTimeLocalValue('2026-10-21T11:35:00Z')).toBe('2026-10-21T14:35')
  })
})

describe('isValidDay', () => {
  it.each([
    ['2026-10-21', true],
    ['2026-13-40', false],
    ['2026-02-30', false],
    ['abc', false],
    ['', false],
  ])('%j -> %s', (day, expected) => {
    expect(isValidDay(day)).toBe(expected)
  })
})
