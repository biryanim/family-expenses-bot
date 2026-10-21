import type { components } from '../api/schema.gen'
import { isCategory } from './categories'
import { toRfc3339Local } from './dates'
import { parseAmount } from './money'

type ExpenseCreate = components['schemas']['ExpenseCreate']
type ExpenseUpdate = components['schemas']['ExpenseUpdate']

export type ExpenseFormValues = {
  userId: string
  amount: string
  category: string
  spentAt: string
  comment: string
}

export type ExpenseFormErrors = Partial<Record<keyof ExpenseFormValues, string>>

type Result<T> = { ok: true; body: T } | { ok: false; errors: ExpenseFormErrors }

function validateCommon(values: ExpenseFormValues, errors: ExpenseFormErrors) {
  const amount = parseAmount(values.amount)
  if (!amount.ok) {
    errors.amount = amount.error
  }
  if (!isCategory(values.category)) {
    errors.category = 'Выберите категорию'
  }
  if (values.spentAt !== '' && Number.isNaN(new Date(values.spentAt).getTime())) {
    errors.spentAt = 'Некорректная дата'
  }
  return amount
}

export function validateCreateExpenseForm(values: ExpenseFormValues): Result<ExpenseCreate> {
  const errors: ExpenseFormErrors = {}
  if (!/^\d+$/.test(values.userId)) {
    errors.userId = 'Выберите владельца'
  }
  const amount = validateCommon(values, errors)
  if (Object.keys(errors).length > 0 || !amount.ok || !isCategory(values.category)) {
    return { ok: false, errors }
  }
  const body: ExpenseCreate = {
    user_id: Number(values.userId),
    amount: amount.kopecks,
    category: values.category,
    comment: values.comment,
  }
  if (values.spentAt !== '') {
    body.spent_at = toRfc3339Local(new Date(values.spentAt))
  }
  return { ok: true, body }
}

export function validateUpdateExpenseForm(values: ExpenseFormValues): Result<ExpenseUpdate> {
  const errors: ExpenseFormErrors = {}
  const amount = validateCommon(values, errors)
  if (values.spentAt === '') {
    errors.spentAt = 'Укажите дату'
  }
  if (Object.keys(errors).length > 0 || !amount.ok || !isCategory(values.category)) {
    return { ok: false, errors }
  }
  return {
    ok: true,
    body: {
      amount: amount.kopecks,
      category: values.category,
      spent_at: toRfc3339Local(new Date(values.spentAt)),
      comment: values.comment,
    },
  }
}
