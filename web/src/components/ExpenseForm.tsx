import { useEffect, useRef, useState, type FormEvent } from 'react'
import { useCreateExpense, useUpdateExpense } from '../api/queries'
import type { components } from '../api/schema.gen'
import { categories, categoryLabels } from '../lib/categories'
import { toDateTimeLocalValue } from '../lib/dates'
import {
  validateCreateExpenseForm,
  validateUpdateExpenseForm,
  type ExpenseFormErrors,
  type ExpenseFormValues,
} from '../lib/expense-form'
import { memberDisplayName } from '../lib/members'
import { ErrorMessage } from './ErrorMessage'

type Expense = components['schemas']['Expense']
type Member = components['schemas']['Member']

type Props = {
  expense?: Expense
  members: Member[]
  onClose: () => void
}

function initialValues(expense: Expense | undefined): ExpenseFormValues {
  if (expense === undefined) {
    return { userId: '', amount: '', category: 'other', spentAt: toDateTimeLocalValue(new Date().toISOString()), comment: '' }
  }
  const rubles = Math.trunc(expense.amount / 100)
  const rest = String(expense.amount % 100).padStart(2, '0')
  return {
    userId: String(expense.user_id),
    amount: `${rubles},${rest}`,
    category: expense.category,
    spentAt: toDateTimeLocalValue(expense.spent_at),
    comment: expense.comment,
  }
}

export function ExpenseForm({ expense, members, onClose }: Props) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const [values, setValues] = useState<ExpenseFormValues>(() => initialValues(expense))
  const [errors, setErrors] = useState<ExpenseFormErrors>({})
  const create = useCreateExpense()
  const update = useUpdateExpense()
  const mutation = expense === undefined ? create : update
  const noMembers = expense === undefined && members.length === 0

  useEffect(() => {
    const dialog = dialogRef.current
    if (dialog !== null && !dialog.open) {
      dialog.showModal()
    }
  }, [])

  function change(patch: Partial<ExpenseFormValues>) {
    setValues((prev) => ({ ...prev, ...patch }))
  }

  function submit(event: FormEvent) {
    event.preventDefault()
    if (expense === undefined) {
      const result = validateCreateExpenseForm(values)
      setErrors(result.ok ? {} : result.errors)
      if (result.ok) {
        create.mutate(result.body, { onSuccess: onClose })
      }
      return
    }
    const result = validateUpdateExpenseForm(values)
    setErrors(result.ok ? {} : result.errors)
    if (result.ok) {
      update.mutate({ id: expense.id, body: result.body }, { onSuccess: onClose })
    }
  }

  return (
    <dialog ref={dialogRef} className="expense-dialog" onClose={onClose}>
      <form onSubmit={submit} noValidate>
        <h2>{expense === undefined ? 'Новая трата' : 'Изменить трату'}</h2>
        {noMembers && <p className="field-error">Сначала напишите боту: членов семьи пока нет</p>}
        <label>
          Владелец
          {expense === undefined ? (
            <select value={values.userId} onChange={(e) => change({ userId: e.target.value })}>
              <option value="">Выберите…</option>
              {members.map((m) => (
                <option key={m.telegram_id} value={String(m.telegram_id)}>
                  {memberDisplayName(members, m.telegram_id)}
                </option>
              ))}
            </select>
          ) : (
            <input value={memberDisplayName(members, expense.user_id)} readOnly />
          )}
          {errors.userId && <span className="field-error">{errors.userId}</span>}
        </label>
        <label>
          Сумма, ₽
          <input inputMode="decimal" value={values.amount} onChange={(e) => change({ amount: e.target.value })} />
          {errors.amount && <span className="field-error">{errors.amount}</span>}
        </label>
        <label>
          Категория
          <select value={values.category} onChange={(e) => change({ category: e.target.value })}>
            {categories.map((c) => (
              <option key={c} value={c}>
                {categoryLabels[c]}
              </option>
            ))}
          </select>
          {errors.category && <span className="field-error">{errors.category}</span>}
        </label>
        <label>
          Дата и время
          <input type="datetime-local" value={values.spentAt} onChange={(e) => change({ spentAt: e.target.value })} />
          {errors.spentAt && <span className="field-error">{errors.spentAt}</span>}
        </label>
        <label>
          Комментарий
          <input value={values.comment} maxLength={200} onChange={(e) => change({ comment: e.target.value })} />
        </label>
        {mutation.isError && <ErrorMessage error={mutation.error} />}
        <div className="dialog-actions">
          <button type="button" onClick={onClose}>
            Отмена
          </button>
          <button type="submit" disabled={noMembers || mutation.isPending}>
            Сохранить
          </button>
        </div>
      </form>
    </dialog>
  )
}
