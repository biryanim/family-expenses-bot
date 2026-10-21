import { categories, categoryLabels, isCategory } from '../lib/categories'
import { emptyFilters, type Filters as FiltersValue } from '../lib/filters'
import { memberDisplayName } from '../lib/members'
import type { components } from '../api/schema.gen'

type Props = {
  filters: FiltersValue
  members: components['schemas']['Member'][]
  periodError?: string
  onChange: (filters: FiltersValue) => void
}

export function Filters({ filters, members, periodError, onChange }: Props) {
  return (
    <form className="filters" onSubmit={(e) => e.preventDefault()}>
      <label>
        Член семьи
        <select value={filters.member} onChange={(e) => onChange({ ...filters, member: e.target.value })}>
          <option value="">Все</option>
          {members.map((m) => (
            <option key={m.telegram_id} value={String(m.telegram_id)}>
              {memberDisplayName(members, m.telegram_id)}
            </option>
          ))}
        </select>
      </label>
      <label>
        С
        <input type="date" value={filters.from} onChange={(e) => onChange({ ...filters, from: e.target.value })} />
      </label>
      <label>
        По
        <input type="date" value={filters.to} onChange={(e) => onChange({ ...filters, to: e.target.value })} />
      </label>
      <label>
        Категория
        <select
          value={filters.category}
          onChange={(e) => onChange({ ...filters, category: isCategory(e.target.value) ? e.target.value : '' })}
        >
          <option value="">Все</option>
          {categories.map((c) => (
            <option key={c} value={c}>
              {categoryLabels[c]}
            </option>
          ))}
        </select>
      </label>
      <button type="button" onClick={() => onChange(emptyFilters)}>
        Сбросить
      </button>
      {periodError && (
        <p className="field-error" role="alert">
          {periodError}
        </p>
      )}
    </form>
  )
}
