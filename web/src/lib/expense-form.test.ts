import { describe, expect, it } from 'vitest'
import { validateCreateExpenseForm, validateUpdateExpenseForm, type ExpenseFormValues } from './expense-form'

const base: ExpenseFormValues = {
  userId: '123',
  amount: '350',
  category: 'food',
  spentAt: '',
  comment: 'кофе',
}

describe('validateCreateExpenseForm', () => {
  it('builds the body', () => {
    expect(validateCreateExpenseForm(base)).toEqual({
      ok: true,
      body: { user_id: 123, amount: 35000, category: 'food', comment: 'кофе' },
    })
  })

  it('sends spent_at with offset when set', () => {
    const result = validateCreateExpenseForm({ ...base, spentAt: '2026-10-21T14:35' })
    expect(result).toMatchObject({ ok: true, body: { spent_at: '2026-10-21T14:35:00+03:00' } })
  })

  it('requires owner', () => {
    const result = validateCreateExpenseForm({ ...base, userId: '' })
    expect(result).toMatchObject({ ok: false, errors: { userId: expect.any(String) } })
  })

  it('requires amount', () => {
    const result = validateCreateExpenseForm({ ...base, amount: '' })
    expect(result).toMatchObject({ ok: false, errors: { amount: expect.any(String) } })
  })

  it('rejects invalid amount', () => {
    const result = validateCreateExpenseForm({ ...base, amount: '12,345' })
    expect(result).toMatchObject({ ok: false, errors: { amount: expect.any(String) } })
  })
})

describe('validateUpdateExpenseForm', () => {
  const edit = { ...base, spentAt: '2026-10-21T14:35' }

  it('builds the body', () => {
    expect(validateUpdateExpenseForm(edit)).toEqual({
      ok: true,
      body: { amount: 35000, category: 'food', spent_at: '2026-10-21T14:35:00+03:00', comment: 'кофе' },
    })
  })

  it('sends empty comment', () => {
    expect(validateUpdateExpenseForm({ ...edit, comment: '' })).toMatchObject({ ok: true, body: { comment: '' } })
  })

  it('requires date', () => {
    const result = validateUpdateExpenseForm({ ...edit, spentAt: '' })
    expect(result).toMatchObject({ ok: false, errors: { spentAt: expect.any(String) } })
  })
})
