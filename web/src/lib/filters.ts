import { isValidDay, periodToRange } from './dates'
import { isCategory, type Category } from './categories'
import { pageSize, pageToOffset, parsePage } from './pagination'

export type Filters = {
  member: string
  from: string
  to: string
  category: Category | ''
  page: number
}

export type ExpenseQuery = {
  user_id?: number
  from?: string
  to?: string
  category?: Category
  limit: number
  offset: number
}

export type QueryResult = { ok: true; query: ExpenseQuery } | { ok: false; error: string }

export const emptyFilters: Filters = { member: '', from: '', to: '', category: '', page: 1 }

export function readFilters(search: string): Filters {
  const params = new URLSearchParams(search)
  const member = params.get('member') ?? ''
  const from = params.get('from') ?? ''
  const to = params.get('to') ?? ''
  const category = params.get('category') ?? ''
  return {
    member: /^\d+$/.test(member) ? member : '',
    from: isValidDay(from) ? from : '',
    to: isValidDay(to) ? to : '',
    category: isCategory(category) ? category : '',
    page: parsePage(params.get('page') ?? ''),
  }
}

export function writeFilters(filters: Filters): void {
  const params = new URLSearchParams()
  const { page, ...rest } = filters
  for (const [key, value] of Object.entries(rest)) {
    if (value !== '') {
      params.set(key, value)
    }
  }
  if (page > 1) {
    params.set('page', String(page))
  }
  const search = params.toString()
  history.replaceState(null, '', search === '' ? location.pathname : `?${search}`)
}

export function filtersToQuery(filters: Filters): QueryResult {
  const range = periodToRange(filters.from, filters.to)
  if (!range.ok) {
    return range
  }
  const query: ExpenseQuery = { limit: pageSize, offset: pageToOffset(filters.page) }
  if (filters.member !== '') {
    query.user_id = Number(filters.member)
  }
  if (range.from !== undefined) {
    query.from = range.from
  }
  if (range.to !== undefined) {
    query.to = range.to
  }
  if (filters.category !== '') {
    query.category = filters.category
  }
  return { ok: true, query }
}
