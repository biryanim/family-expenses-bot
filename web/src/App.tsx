import { useState } from 'react'
import { useDeleteExpense, useExpenses, useMembers } from './api/queries'
import { ErrorMessage } from './components/ErrorMessage'
import { ExpenseForm } from './components/ExpenseForm'
import { ExpenseTable } from './components/ExpenseTable'
import { Filters } from './components/Filters'
import { categoryLabels } from './lib/categories'
import { formatDateTime } from './lib/dates'
import { formatAmount } from './lib/money'
import type { components } from './api/schema.gen'
import { filtersToQuery, readFilters, writeFilters, type Filters as FiltersValue } from './lib/filters'

export function App() {
  const [filters, setFilters] = useState<FiltersValue>(() => readFilters(location.search))
  const [formState, setFormState] = useState<{ expense?: components['schemas']['Expense'] } | null>(null)
  const members = useMembers()
  const deleteExpense = useDeleteExpense()
  const queryResult = filtersToQuery(filters)
  const expenses = useExpenses(queryResult.ok ? queryResult.query : {}, queryResult.ok)

  function changeFilters(next: FiltersValue) {
    setFilters(next)
    writeFilters(next)
  }

  function confirmDelete(expense: components['schemas']['Expense']) {
    const text = `Удалить трату ${formatAmount(expense.amount)} (${categoryLabels[expense.category]}, ${formatDateTime(expense.spent_at)})?`
    if (window.confirm(text)) {
      deleteExpense.mutate(expense.id)
    }
  }

  const memberList = members.data ?? []

  return (
    <main className="page">
      <h1>Семейные расходы</h1>
      {members.isError && <ErrorMessage error={members.error} onRetry={() => void members.refetch()} />}
      <Filters
        filters={filters}
        members={memberList}
        periodError={queryResult.ok ? undefined : queryResult.error}
        onChange={changeFilters}
      />
      <div className="toolbar">
        <button type="button" onClick={() => setFormState({})}>
          Добавить трату
        </button>
        <button type="button" onClick={() => void expenses.refetch()} disabled={!queryResult.ok}>
          Обновить
        </button>
      </div>
      {deleteExpense.isError && <ErrorMessage error={deleteExpense.error} />}
      {queryResult.ok && expenses.isPending && <p className="status">Загрузка…</p>}
      {queryResult.ok && expenses.isError && (
        <ErrorMessage error={expenses.error} onRetry={() => void expenses.refetch()} />
      )}
      {queryResult.ok && expenses.isSuccess && expenses.data.length === 0 && (
        <p className="status">Трат по выбранным фильтрам нет</p>
      )}
      {queryResult.ok && expenses.isSuccess && expenses.data.length > 0 && (
        <ExpenseTable
          expenses={expenses.data}
          members={memberList}
          onEdit={(expense) => setFormState({ expense })}
          onDelete={confirmDelete}
        />
      )}
      {formState !== null && (
        <ExpenseForm expense={formState.expense} members={memberList} onClose={() => setFormState(null)} />
      )}
    </main>
  )
}
