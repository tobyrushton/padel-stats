const AUTH_TOKEN_KEY = "auth_token"

function canUseStorage(): boolean {
  return typeof window !== "undefined" && typeof window.sessionStorage !== "undefined"
}

export function getAuthToken(): string | undefined {
  if (!canUseStorage()) {
    return undefined
  }

  const token = window.sessionStorage.getItem(AUTH_TOKEN_KEY)
  return token ?? undefined
}

export function setAuthToken(token: string): void {
  if (!canUseStorage()) {
    return
  }

  window.sessionStorage.setItem(AUTH_TOKEN_KEY, token)
}

export function clearAuthToken(): void {
  if (!canUseStorage()) {
    return
  }

  window.sessionStorage.removeItem(AUTH_TOKEN_KEY)
}

function decodeJwtPayload(token: string): Record<string, unknown> | undefined {
  const parts = token.split(".")
  if (parts.length < 2) {
    return undefined
  }

  try {
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/")
    const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4)
    const payload = JSON.parse(atob(padded)) as Record<string, unknown>
    return payload
  } catch {
    return undefined
  }
}

export function getUserIDFromAuthToken(): number | undefined {
  const token = getAuthToken()
  if (!token) {
    return undefined
  }

  const payload = decodeJwtPayload(token)
  const uid = payload?.uid

  if (typeof uid === "number" && Number.isInteger(uid) && uid > 0) {
    return uid
  }

  if (typeof uid === "string") {
    const parsed = Number(uid)
    if (Number.isInteger(parsed) && parsed > 0) {
      return parsed
    }
  }

  return undefined
}
