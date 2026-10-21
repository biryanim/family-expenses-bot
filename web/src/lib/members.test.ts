import { describe, expect, it } from 'vitest'
import { memberDisplayName } from './members'

const members = [
  { telegram_id: 1, username: 'anna', first_name: 'Анна', last_name: 'Иванова' },
  { telegram_id: 2, username: 'boris', first_name: 'Борис', last_name: '' },
  { telegram_id: 3, username: 'vera', first_name: '', last_name: '' },
  { telegram_id: 4, username: '', first_name: '', last_name: '' },
]

describe('memberDisplayName', () => {
  it('uses first and last name', () => {
    expect(memberDisplayName(members, 1)).toBe('Анна Иванова')
  })

  it('uses first name without last name', () => {
    expect(memberDisplayName(members, 2)).toBe('Борис')
  })

  it('falls back to username', () => {
    expect(memberDisplayName(members, 3)).toBe('@vera')
  })

  it('falls back to telegram id', () => {
    expect(memberDisplayName(members, 4)).toBe('4')
  })

  it('uses user id for unknown member', () => {
    expect(memberDisplayName(members, 789)).toBe('789')
  })
})
