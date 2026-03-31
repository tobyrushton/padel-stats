import { useEffect, useState } from "react"

import { Button } from "@/components/ui/button"
import { clearAuthToken, getAuthToken } from "@/lib/auth-token"
import { clearAuthUser } from "@/lib/auth-user"

export default function Header() {
  const [isLoggedIn, setIsLoggedIn] = useState(false)

  useEffect(() => {
    const syncAuthState = () => {
      setIsLoggedIn(Boolean(getAuthToken()))
    }

    syncAuthState()

    window.addEventListener("storage", syncAuthState)
    window.addEventListener("focus", syncAuthState)

    return () => {
      window.removeEventListener("storage", syncAuthState)
      window.removeEventListener("focus", syncAuthState)
    }
  }, [])

  const handleLogout = () => {
    clearAuthToken()
    clearAuthUser()
    setIsLoggedIn(false)
    window.location.assign("/auth/signin")
  }

  return (
    <header className="border-b border-border bg-background/90 backdrop-blur">
      <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-4 py-3">
        <a href="/" className="text-sm font-semibold tracking-wide">
          Padel Stats
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
