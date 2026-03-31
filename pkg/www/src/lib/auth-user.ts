import { getUserIDFromAuthToken } from "@/lib/auth-token"

const AUTH_USER_KEY = "auth_user"

interface StoredAuthUser {
  id: number
  username?: string
  firstName?: string
  lastName?: string
}

function canUseStorage(): boolean {
  return typeof window !== "undefined" && typeof window.sessionStorage !== "undefined"
}

function isStoredAuthUser(value: unknown): value is StoredAuthUser {
  if (!value || typeof value !== "object") {
    return false
  }

  const candidate = value as { id?: unknown }
  return typeof candidate.id === "number" && Number.isInteger(candidate.id) && candidate.id > 0
}

export function getAuthUser(): StoredAuthUser | undefined {
  if (!canUseStorage()) {
    return undefined
  }

  const stored = window.sessionStorage.getItem(AUTH_USER_KEY)
  if (!stored) {
    return undefined
  }

  try {
    const parsed = JSON.parse(stored) as unknown
    if (!isStoredAuthUser(parsed)) {
      return undefined
    }
    return parsed
  } catch {
    return undefined
  }
}

export function setAuthUser(user: {
  id?: number
  username?: string
  firstName?: string
  lastName?: string
} | null | undefined): void {
  if (!canUseStorage()) {
    return
  }

  if (!user?.id || !Number.isInteger(user.id) || user.id <= 0) {
    window.sessionStorage.removeItem(AUTH_USER_KEY)
    return
  }

  const payload: StoredAuthUser = {
    id: user.id,
    username: user.username,
    firstName: user.firstName,
    lastName: user.lastName,
  }

  window.sessionStorage.setItem(AUTH_USER_KEY, JSON.stringify(payload))
}

export function clearAuthUser(): void {
  if (!canUseStorage()) {
    return
  }

  window.sessionStorage.removeItem(AUTH_USER_KEY)
}

export function getCurrentUserID(): number | undefined {
  const authUserID = getAuthUser()?.id
  if (typeof authUserID === "number" && Number.isInteger(authUserID) && authUserID > 0) {
    return authUserID
  }

  return getUserIDFromAuthToken()
}
