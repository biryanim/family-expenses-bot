import { apiErrorFromException } from '../api/api-error'

type Props = {
  error: unknown
  onRetry?: () => void
}

export function ErrorMessage({ error, onRetry }: Props) {
  return (
    <div className="error" role="alert">
      <span>{apiErrorFromException(error).message}</span>
      {onRetry && (
        <button type="button" onClick={onRetry}>
          Повторить
        </button>
      )}
    </div>
  )
}
