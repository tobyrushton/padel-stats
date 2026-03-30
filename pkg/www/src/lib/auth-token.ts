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
