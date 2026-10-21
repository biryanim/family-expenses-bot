import type { components } from './schema.gen'

type ErrorCode = components['schemas']['ErrorCode']

const codeMessages: Record<ErrorCode, string> = {
  invalid_request: 'Некорректный запрос',
  invalid_amount: 'Некорректная сумма',
  invalid_category: 'Некорректная категория',
  invalid_date: 'Некорректная дата',
  unknown_member: 'Такого члена семьи нет',
  not_found: 'Трата не найдена',
  internal: 'Ошибка сервера, попробуйте позже',
}

export const serverUnavailableMessage = 'Сервер недоступен'

export class ApiError extends Error {
  readonly status: number
  readonly code?: string

  constructor(status: number, message: string, code?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export function apiErrorFromResponse(status: number, body: unknown): ApiError {
  if (
    typeof body === 'object' &&
    body !== null &&
    'code' in body &&
    'message' in body &&
    typeof body.code === 'string' &&
    typeof body.message === 'string'
  ) {
    const known = codeMessages[body.code as ErrorCode] as string | undefined
    return new ApiError(status, known ?? body.message, body.code)
  }
  return new ApiError(status, serverUnavailableMessage)
}

export function apiErrorFromException(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error
  }
  return new ApiError(0, serverUnavailableMessage)
}
