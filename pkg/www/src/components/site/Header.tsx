import { useEffect, useState } from "react"

import { Button } from "@/components/ui/button"
import { clearAuthToken, getAuthToken } from "@/lib/auth-token"
import { clearAuthUser, getAuthUser } from "@/lib/auth-user"
import { fetchCurrentUserFromServer } from "@/lib/current-user"

export default function Header() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    let cancelled = false

    const syncAuthState = async () => {
      const token = getAuthToken()
      if (!token) {
        if (!cancelled) {
          setIsLoggedIn(false)
          setIsAdmin(false)
        }
        return
      }

      try {
        const user = await fetchCurrentUserFromServer()
        if (cancelled) {
          return
        }

        setIsLoggedIn(Boolean(user))
        setIsAdmin(Boolean(user?.isAdmin))
      } catch {
        if (!cancelled) {
          setIsLoggedIn(true)
          setIsAdmin(Boolean(getAuthUser()?.isAdmin))
        }
      }
    }

    void syncAuthState()

    const handleAuthChange = () => {
      void syncAuthState()
    }

    window.addEventListener("storage", handleAuthChange)
    window.addEventListener("focus", handleAuthChange)

    return () => {
      cancelled = true
      window.removeEventListener("storage", handleAuthChange)
      window.removeEventListener("focus", handleAuthChange)
    }
  }, [])

  const handleLogout = () => {
    clearAuthToken()
    clearAuthUser()
    setIsLoggedIn(false)
    setIsAdmin(false)
    window.location.assign("/auth/signin")
  }

  return (
    <header className="border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 py-3">
        <a href="/" className="text-sm font-semibold tracking-wide">
          Bexbox Premier Padel
        </a>

        <nav className="flex items-center gap-2">
          {isLoggedIn ? (
            <>
              <Button asChild variant="ghost" size="sm">
                <a href="/games">Games</a>
              </Button>
              <Button asChild variant="ghost" size="sm">
                <a href="/leaderboard">Leaderboard</a>
              </Button>
              {isAdmin ? (
                <Button asChild variant="ghost" size="sm">
                  <a href="/admin/users">Admin</a>
                </Button>
              ) : null}
              <Button type="button" variant="outline" size="sm" onClick={handleLogout}>
                Log out
              </Button>
            </>
          ) : (
            <>
              <Button asChild variant="ghost" size="sm">
                <a href="/leaderboard">Leaderboard</a>
              </Button>
              <Button asChild variant="ghost" size="sm">
                <a href="/auth/signin">Log in</a>
              </Button>
              <Button asChild size="sm">
                <a href="/auth/signup">Sign up</a>
              </Button>
            </>
          )}
        </nav>
      </div>
    </header>
  )
}
