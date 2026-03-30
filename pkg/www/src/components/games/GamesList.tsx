import { useEffect, useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { createApiClient, type Game } from "@/lib/api-client"
import { getCurrentUserID } from "@/lib/auth-user"

interface GamesListProps {
  apiBaseUrl: string
  refreshKey: number
  lastCreatedGameID?: number
  onAddRequest: () => void
}

function formatPlayer(game: Game, team: 1 | 2, slot: 1 | 2): string {
  const player =
    team === 1
      ? slot === 1
        ? game.team1Player1
        : game.team1Player2
      : slot === 1
        ? game.team2Player1
        : game.team2Player2

  if (player) {
    const fullName = [player.firstName, player.lastName].filter(Boolean).join(" ").trim()
    return fullName || player.username || `Player ${player.id}`
  }

  const id =
    team === 1
      ? slot === 1
        ? game.team1Player1Id
        : game.team1Player2Id
      : slot === 1
        ? game.team2Player1Id
        : game.team2Player2Id

  return `Player ${id}`
}

function formatPlayedAt(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return "Unknown date"
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date)
}

export default function GamesList({
  apiBaseUrl,
  refreshKey,
  lastCreatedGameID,
  onAddRequest,
}: GamesListProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const [games, setGames] = useState<Game[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [deletingGameID, setDeletingGameID] = useState<number | null>(null)

  useEffect(() => {
    let isMounted = true

    const loadGames = async () => {
      const currentUserID = getCurrentUserID()
      if (!currentUserID) {
        setErrorMessage("Unable to determine your player ID. Please sign in again.")
        setIsLoading(false)
        return
      }

      try {
        setIsLoading(true)
        setErrorMessage(null)
        const result = await apiClient.listGamesForPlayer(currentUserID)
        if (isMounted) {
          setGames(result)
        }
      } catch {
        if (isMounted) {
          setErrorMessage("Could not load your games right now.")
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadGames()

    return () => {
      isMounted = false
    }
  }, [apiClient, refreshKey])

  const handleDelete = async (gameID: number) => {
    const approved = window.confirm("Delete this game? This cannot be undone.")
    if (!approved) {
      return
    }

    try {
      setDeletingGameID(gameID)
      await apiClient.deleteGame(gameID)
      setGames((current) => current.filter((game) => game.id !== gameID))
    } catch {
      setErrorMessage("Could not delete the game. Please try again.")
    } finally {
      setDeletingGameID(null)
    }
  }

  if (isLoading) {
    return (
      <section className="rounded-xl border border-border bg-card p-6 text-card-foreground shadow-sm">
        <p className="text-sm text-muted-foreground">Loading your games...</p>
      </section>
    )
  }

  return (
    <section className="space-y-4">
      {errorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {errorMessage}
        </p>
      ) : null}

      {games.length === 0 ? (
        <div className="rounded-xl border border-border bg-card p-6 text-card-foreground shadow-sm">
          <h2 className="text-lg font-semibold">No games yet</h2>
          <p className="mt-1 text-sm text-muted-foreground">Record your first game to see stats here.</p>
          <Button type="button" className="mt-4" onClick={onAddRequest}>
            Add your first game
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {games.map((game) => {
            const isNew = game.id === lastCreatedGameID

            return (
              <article
                key={game.id}
                className="rounded-xl border border-border bg-card p-4 text-card-foreground shadow-sm"
              >
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="space-y-2">
                    <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                      Game #{game.id} {isNew ? "- Just Added" : ""}
                    </p>
                    <p className="text-base font-medium">
                      {formatPlayer(game, 1, 1)} &amp; {formatPlayer(game, 1, 2)}
                      <span className="mx-2 text-muted-foreground">vs</span>
                      {formatPlayer(game, 2, 1)} &amp; {formatPlayer(game, 2, 2)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Score: <span className="font-medium text-foreground">{game.team1Score}</span>-
                      <span className="font-medium text-foreground">{game.team2Score}</span>
                    </p>
                    <p className="text-sm text-muted-foreground">Played: {formatPlayedAt(game.playedAt)}</p>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button asChild type="button" variant="outline" size="sm">
                      <a href={`/games/${game.id}`}>View</a>
                    </Button>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      onClick={() => void handleDelete(game.id)}
                      disabled={deletingGameID === game.id}
                    >
                      {deletingGameID === game.id ? "Deleting..." : "Delete"}
                    </Button>
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      )}
    </section>
  )
}
