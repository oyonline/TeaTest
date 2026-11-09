export const ENTRY_TEA = 'tea'
export const ENTRY_PRACTICE = 'practice'

const ENTRY_HOME = {
  [ENTRY_TEA]: '/welcome',
  [ENTRY_PRACTICE]: '/practice/banks'
}

const ENTRY_LOGIN = {
  [ENTRY_TEA]: '/',
  [ENTRY_PRACTICE]: '/practice'
}

export function normalizeEntry(value) {
  return value === ENTRY_PRACTICE ? ENTRY_PRACTICE : ENTRY_TEA
}

export function entryFromRoute(route) {
  return normalizeEntry(route?.meta?.entry || route?.query?.entry)
}

export function entryHome(entry) {
  return ENTRY_HOME[normalizeEntry(entry)]
}

export function entryLogin(entry) {
  return ENTRY_LOGIN[normalizeEntry(entry)]
}

export function entryFromLocation() {
  const fromQuery = new URLSearchParams(window.location.search).get('entry')
  if (fromQuery === ENTRY_PRACTICE) return ENTRY_PRACTICE
  if (fromQuery === ENTRY_TEA) return ENTRY_TEA
  return window.location.pathname.startsWith('/practice') ? ENTRY_PRACTICE : ENTRY_TEA
}

export function isPracticeEntry(entry) {
  return normalizeEntry(entry) === ENTRY_PRACTICE
}
