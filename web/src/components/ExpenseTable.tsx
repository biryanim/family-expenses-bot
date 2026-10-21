import { categoryLabels } from '../lib/categories'
import { formatDateTime } from '../lib/dates'
import { memberDisplayName } from '../lib/members'
import { formatAmount } from '../lib/money'
import type { components } from '../api/schema.gen'

type Props = {
  expenses: components['schemas']['Expense'][]
  members: components['schemas']['Member'][]
  onEdit: (expense: components['schemas']['Expense']) => void
  onDelete: (expense: components['schemas']['Expense']) => void
}

export function ExpenseTable({ expenses, members, onEdit, onDelete }: Props) {
  return (
    <table className="expenses">
      <thead>
        <tr>
          <th>Владелец</th>
          <th>Сумма</th>
          <th>Категория</th>
          <th>Дата</th>
          <th>Комментарий</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {expenses.map((e) => (
          <tr key={e.id}>
            <td>{memberDisplayName(members, e.user_id)}</td>
            <td className="amount">{formatAmount(e.amount)}</td>
            <td>{categoryLabels[e.category]}</td>
            <td>{formatDateTime(e.spent_at)}</td>
            <td>{e.comment}</td>
            <td className="actions">
              <button type="button" onClick={() => onEdit(e)}>
                Изменить
              </button>
              <button type="button" onClick={() => onDelete(e)}>
                Удалить
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
