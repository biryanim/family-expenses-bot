import { describe, expect, it } from 'vitest'
import { formatAmount, parseAmount } from './money'

describe('formatAmount', () => {
  it.each([
    [120000, '1 200,00 ₽'],
    [35050, '350,50 ₽'],
    [5, '0,05 ₽'],
  ])('formats %i kopecks', (kopecks, expected) => {
    expect(formatAmount(kopecks)).toBe(expected)
  })
})

describe('parseAmount', () => {
  it.each([
    ['350,5', 35050],
    ['350.50', 35050],
    ['1 200', 120000],
    ['350', 35000],
    ['0,05', 5],
  ])('parses %s', (input, kopecks) => {
    expect(parseAmount(input)).toEqual({ ok: true, kopecks })
  })

  it.each(['', '0', '0,00', '-5', '12,345', 'abc', '1,', '90071992547409.92'])('rejects %j', (input) => {
    expect(parseAmount(input).ok).toBe(false)
  })
})
