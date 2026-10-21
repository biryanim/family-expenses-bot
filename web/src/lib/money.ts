const rubleFormat = new Intl.NumberFormat('ru-RU', { useGrouping: true })

export type ParsedAmount = { ok: true; kopecks: number } | { ok: false; error: string }

export function formatAmount(kopecks: number): string {
  const rubles = Math.trunc(kopecks / 100)
  const rest = kopecks % 100
  return `${rubleFormat.format(rubles)},${String(rest).padStart(2, '0')}\u00a0₽`
}

export function parseAmount(input: string): ParsedAmount {
  const cleaned = input.replace(/\s/g, '')
  if (cleaned === '') {
    return { ok: false, error: 'Введите сумму' }
  }
  const match = /^(\d+)(?:[.,](\d{1,2}))?$/.exec(cleaned)
  if (match === null) {
    return { ok: false, error: 'Некорректная сумма' }
  }
  const rubles = Number(match[1])
  const fraction = Number((match[2] ?? '').padEnd(2, '0'))
  const kopecks = rubles * 100 + fraction
  if (kopecks === 0) {
    return { ok: false, error: 'Сумма должна быть больше нуля' }
  }
  if (!Number.isSafeInteger(kopecks)) {
    return { ok: false, error: 'Слишком большая сумма' }
  }
  return { ok: true, kopecks }
}
