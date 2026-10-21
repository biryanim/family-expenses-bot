import type { components } from '../api/schema.gen'

export type Category = components['schemas']['Category']

export const categoryLabels: Record<Category, string> = {
  food: 'еда',
  transport: 'транспорт',
  housing: 'жильё',
  entertainment: 'развлечения',
  health: 'здоровье',
  shopping: 'покупки',
  other: 'другое',
}

export const categories = Object.keys(categoryLabels) as Category[]

export function isCategory(value: string): value is Category {
  return Object.hasOwn(categoryLabels, value)
}
