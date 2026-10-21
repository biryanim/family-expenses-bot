export const pageSize = 50

export type PageInfo = {
  from: number
  to: number
  total?: number
  hasPrev: boolean
  hasNext: boolean
}

export function parsePage(raw: string): number {
  return /^[1-9]\d{0,5}$/.test(raw) ? Number(raw) : 1
}

export function pageToOffset(page: number): number {
  return (page - 1) * pageSize
}

export function readTotal(headers: Headers): number | undefined {
  const raw = headers.get('X-Total-Count')
  return raw !== null && /^\d+$/.test(raw) ? Number(raw) : undefined
}

export function describePage(page: number, itemsCount: number, total?: number): PageInfo {
  const offset = pageToOffset(page)
  return {
    from: itemsCount === 0 ? 0 : offset + 1,
    to: offset + itemsCount,
    total,
    hasPrev: page > 1,
    hasNext: total !== undefined ? offset + itemsCount < total : itemsCount === pageSize,
  }
}

export function pageAfterDelete(page: number, itemsCount: number): number {
  return itemsCount === 1 && page > 1 ? page - 1 : page
}
