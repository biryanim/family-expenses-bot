import { describe, expect, it } from 'vitest'
import { filtersToQuery, readFilters } from './filters'

describe('readFilters', () => {
  it('reads valid params', () => {
    expect(readFilters('?member=123&from=2026-10-01&to=2026-10-07&category=food')).toEqual({
      member: '123',
      from: '2026-10-01',
      to: '2026-10-07',
      category: 'food',
      page: 1,
    })
  })

  it('reads the page', () => {
    expect(readFilters('?page=3').page).toBe(3)
  })

  it('ignores garbage', () => {
    expect(readFilters('?member=abc&category=travel&from=2026-13-40')).toEqual({
      member: '',
      from: '',
      to: '',
      category: '',
      page: 1,
    })
  })
})

describe('filtersToQuery', () => {
  it('omits empty filters', () => {
    expect(filtersToQuery({ member: '', from: '', to: '', category: '', page: 1 })).toEqual({
      ok: true,
      query: { limit: 50, offset: 0 },
    })
  })

  it('builds combined query', () => {
    expect(filtersToQuery({ member: '123', from: '', to: '', category: 'food', page: 1 })).toEqual({
      ok: true,
      query: { user_id: 123, category: 'food', limit: 50, offset: 0 },
    })
  })

  it('converts period to range', () => {
    expect(filtersToQuery({ member: '', from: '2026-10-01', to: '2026-10-07', category: '', page: 1 })).toEqual({
      ok: true,
      query: { from: '2026-10-01T00:00:00+03:00', to: '2026-10-08T00:00:00+03:00', limit: 50, offset: 0 },
    })
  })

  it('converts the page to an offset', () => {
    expect(filtersToQuery({ member: '', from: '', to: '', category: '', page: 3 })).toEqual({
      ok: true,
      query: { limit: 50, offset: 100 },
    })
  })

  it('fails on inverted period', () => {
    expect(filtersToQuery({ member: '', from: '2026-10-10', to: '2026-10-01', category: '', page: 1 }).ok).toBe(false)
  })
})
