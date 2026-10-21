import { describe, expect, it } from 'vitest'
import { ApiError, apiErrorFromException, apiErrorFromResponse } from './api-error'

describe('apiErrorFromResponse', () => {
  it.each([
    ['invalid_request', 'Некорректный запрос'],
    ['invalid_amount', 'Некорректная сумма'],
    ['invalid_category', 'Некорректная категория'],
    ['invalid_date', 'Некорректная дата'],
    ['invalid_pagination', 'Некорректная страница'],
    ['invalid_comment', 'Комментарий слишком длинный (не больше 200 символов)'],
    ['unknown_member', 'Такого члена семьи нет'],
    ['not_found', 'Трата не найдена'],
    ['internal', 'Ошибка сервера, попробуйте позже'],
  ])('maps %s to a russian message', (code, message) => {
    const err = apiErrorFromResponse(400, { code, message: 'raw server text' })
    expect(err.message).toBe(message)
    expect(err.code).toBe(code)
    expect(err.status).toBe(400)
  })

  it('uses server message for unknown code', () => {
    const err = apiErrorFromResponse(409, { code: 'conflict', message: 'Конфликт версий' })
    expect(err.message).toBe('Конфликт версий')
    expect(err.code).toBe('conflict')
  })

  it.each([
    ['html body', '<html>502 Bad Gateway</html>'],
    ['empty body', undefined],
    ['object without code', { message: 'x' }],
    ['non-string code', { code: 1, message: 'x' }],
  ])('falls back to server unavailable for %s', (_, body) => {
    const err = apiErrorFromResponse(502, body)
    expect(err.message).toBe('Сервер недоступен')
    expect(err.code).toBeUndefined()
    expect(err.status).toBe(502)
  })
})

describe('apiErrorFromException', () => {
  it('maps network failure to server unavailable', () => {
    const err = apiErrorFromException(new TypeError('Failed to fetch'))
    expect(err.message).toBe('Сервер недоступен')
    expect(err.status).toBe(0)
  })

  it('keeps ApiError as is', () => {
    const original = new ApiError(404, 'Трата не найдена', 'not_found')
    expect(apiErrorFromException(original)).toBe(original)
  })
})
