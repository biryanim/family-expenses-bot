import type { PageInfo } from '../lib/pagination'

type Props = {
  page: number
  info: PageInfo
  onPage: (page: number) => void
}

export function Pagination({ page, info, onPage }: Props) {
  const range = `${info.from}–${info.to}`
  return (
    <nav className="pagination" aria-label="Страницы">
      <button type="button" onClick={() => onPage(page - 1)} disabled={!info.hasPrev}>
        Назад
      </button>
      <span className="status">
        {info.total !== undefined ? `${range} из ${info.total}` : range}, страница {page}
      </span>
      <button type="button" onClick={() => onPage(page + 1)} disabled={!info.hasNext}>
        Вперёд
      </button>
    </nav>
  )
}
