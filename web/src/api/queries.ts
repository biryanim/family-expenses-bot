import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError, apiErrorFromException, apiErrorFromResponse } from './api-error'
import { client } from './client'
import type { components } from './schema.gen'
import type { ExpenseQuery } from '../lib/filters'
import { readTotal } from '../lib/pagination'

export function useMembers() {
  return useQuery({
    queryKey: ['members'],
    queryFn: async () => {
      const { data, error, response } = await client.GET('/members').catch((e: unknown) => {
        throw apiErrorFromException(e)
      })
      if (data === undefined) {
        throw apiErrorFromResponse(response.status, error)
      }
      return data
    },
  })
}

export function useExpenses(params: ExpenseQuery, enabled: boolean) {
  return useQuery({
    queryKey: ['expenses', params],
    enabled,
    queryFn: async () => {
      const { data, error, response } = await client.GET('/expenses', { params: { query: params } }).catch((e: unknown) => {
        throw apiErrorFromException(e)
      })
      if (data === undefined) {
        throw apiErrorFromResponse(response.status, error)
      }
      return { items: data, total: readTotal(response.headers) }
    },
  })
}

type ExpenseCreate = components['schemas']['ExpenseCreate']
type ExpenseUpdate = components['schemas']['ExpenseUpdate']

function useInvalidateExpensesOnSettled() {
  const queryClient = useQueryClient()
  return (error: unknown) => {
    if (error === null || (error instanceof ApiError && error.status === 404)) {
      return queryClient.invalidateQueries({ queryKey: ['expenses'] })
    }
  }
}

export function useCreateExpense() {
  const invalidate = useInvalidateExpensesOnSettled()
  return useMutation({
    mutationFn: async (body: ExpenseCreate) => {
      const { data, error, response } = await client.POST('/expenses', { body }).catch((e: unknown) => {
        throw apiErrorFromException(e)
      })
      if (data === undefined) {
        throw apiErrorFromResponse(response.status, error)
      }
      return data
    },
    onSettled: (_data, error) => invalidate(error),
  })
}

export function useUpdateExpense() {
  const invalidate = useInvalidateExpensesOnSettled()
  return useMutation({
    mutationFn: async ({ id, body }: { id: number; body: ExpenseUpdate }) => {
      const { data, error, response } = await client
        .PUT('/expenses/{id}', { params: { path: { id } }, body })
        .catch((e: unknown) => {
          throw apiErrorFromException(e)
        })
      if (data === undefined) {
        throw apiErrorFromResponse(response.status, error)
      }
      return data
    },
    onSettled: (_data, error) => invalidate(error),
  })
}

export function useDeleteExpense() {
  const invalidate = useInvalidateExpensesOnSettled()
  return useMutation({
    mutationFn: async (id: number) => {
      const { error, response } = await client.DELETE('/expenses/{id}', { params: { path: { id } } }).catch((e: unknown) => {
        throw apiErrorFromException(e)
      })
      if (!response.ok) {
        throw apiErrorFromResponse(response.status, error)
      }
    },
    onSettled: (_data, error) => invalidate(error),
  })
}
