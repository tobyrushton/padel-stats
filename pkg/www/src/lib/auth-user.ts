import { getUserIDFromAuthToken } from "@/lib/auth-token"

const AUTH_USER_KEY = "auth_user"

interface StoredAuthUser {
  id: number
  username?: string
  firstName?: string
  lastName?: string
  isAdmin?: boolean
  isAcceptedByAdmin?: boolean
}

function canUseDocument(): boolean {
  return typeof document !== "undefined"
}

function getCookie(name: string): string | undefined {
  if (!canUseDocument()) {
    return undefined
  }

  const encodedName = `${encodeURIComponent(name)}=`
  const cookies = document.cookie ? document.cookie.split(";") : []

  for (const cookie of cookies) {
    const trimmed = cookie.trim()
    if (!trimmed.startsWith(encodedName)) {
      continue
    }

    return decodeURIComponent(trimmed.slice(encodedName.length))
  }

  return undefined
}

function setCookie(name: string, value: string): void {
  if (!canUseDocument()) {
    return
  }

  const secure = window.location.protocol === "https:" ? "; Secure" : ""
  document.cookie = `${encodeURIComponent(name)}=${encodeURIComponent(value)}; Path=/; SameSite=Lax${secure}`
}

function clearCookie(name: string): void {
  if (!canUseDocument()) {
    return
  }

  const secure = window.location.protocol === "https:" ? "; Secure" : ""
  document.cookie = `${encodeURIComponent(name)}=; Path=/; Max-Age=0; SameSite=Lax${secure}`
}

function isStoredAuthUser(value: unknown): value is StoredAuthUser {
  if (!value || typeof value !== "object") {
    return false
  }

  const candidate = value as { id?: unknown }
  return typeof candidate.id === "number" && Number.isInteger(candidate.id) && candidate.id > 0
}

export function getAuthUser(): StoredAuthUser | undefined {
  const stored = getCookie(AUTH_USER_KEY)
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
  isAdmin?: boolean
  isAcceptedByAdmin?: boolean
} | null | undefined): void {
  if (!user?.id || !Number.isInteger(user.id) || user.id <= 0) {
    clearCookie(AUTH_USER_KEY)
    return
  }

  const payload: StoredAuthUser = {
    id: user.id,
    username: user.username,
    firstName: user.firstName,
    lastName: user.lastName,
    isAdmin: user.isAdmin,
    isAcceptedByAdmin: user.isAcceptedByAdmin,
  }

  setCookie(AUTH_USER_KEY, JSON.stringify(payload))
}

export function clearAuthUser(): void {
  clearCookie(AUTH_USER_KEY)
}

export function getCurrentUserID(): number | undefined {
  const authUserID = getAuthUser()?.id
  if (typeof authUserID === "number" && Number.isInteger(authUserID) && authUserID > 0) {
    return authUserID
  }

  return getUserIDFromAuthToken()
}
