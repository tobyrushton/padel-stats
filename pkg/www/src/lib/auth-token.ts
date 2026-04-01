const AUTH_TOKEN_KEY = "auth_token"

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

export function getAuthToken(): string | undefined {
  return getCookie(AUTH_TOKEN_KEY)
}

export function setAuthToken(token: string): void {
  setCookie(AUTH_TOKEN_KEY, token)
}

export function clearAuthToken(): void {
  clearCookie(AUTH_TOKEN_KEY)
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
