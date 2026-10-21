import type { components } from '../api/schema.gen'

type Member = components['schemas']['Member']

export function memberDisplayName(members: Member[], userId: number): string {
  const member = members.find((m) => m.telegram_id === userId)
  if (member === undefined) {
    return String(userId)
  }
  const fullName = `${member.first_name} ${member.last_name}`.trim()
  if (fullName !== '') {
    return fullName
  }
  if (member.username !== '') {
    return `@${member.username}`
  }
  return String(member.telegram_id)
}
