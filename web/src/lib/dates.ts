const dateTimeFormat = new Intl.DateTimeFormat('ru-RU', { dateStyle: 'short', timeStyle: 'short' })

export type PeriodRange = { ok: true; from?: string; to?: string } | { ok: false; error: string }

function pad(value: number): string {
  return String(value).padStart(2, '0')
}

export function isValidDay(day: string): boolean {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(day)
  if (match === null) {
    return false
  }
  const [year, month, date] = [Number(match[1]), Number(match[2]), Number(match[3])]
  const parsed = new Date(year, month - 1, date)
  return parsed.getFullYear() === year && parsed.getMonth() === month - 1 && parsed.getDate() === date
}

export function toRfc3339Local(date: Date): string {
  const offset = -date.getTimezoneOffset()
  const sign = offset >= 0 ? '+' : '-'
  const abs = Math.abs(offset)
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}` +
    `${sign}${pad(Math.floor(abs / 60))}:${pad(abs % 60)}`
  )
}

export function periodToRange(fromDay: string, toDay: string): PeriodRange {
  if (fromDay !== '' && toDay !== '' && fromDay > toDay) {
    return { ok: false, error: 'Дата начала позже даты конца' }
  }
  const range: { ok: true; from?: string; to?: string } = { ok: true }
  if (fromDay !== '') {
    const [year, month, day] = fromDay.split('-').map(Number) as [number, number, number]
    range.from = toRfc3339Local(new Date(year, month - 1, day))
  }
  if (toDay !== '') {
    const [year, month, day] = toDay.split('-').map(Number) as [number, number, number]
    range.to = toRfc3339Local(new Date(year, month - 1, day + 1))
  }
  return range
}

export function formatDateTime(iso: string): string {
  return dateTimeFormat.format(new Date(iso))
}

export function toDateTimeLocalValue(iso: string): string {
  const date = new Date(iso)
  return (
    `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}` +
    `T${pad(date.getHours())}:${pad(date.getMinutes())}`
  )
}
