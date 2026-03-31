import { useEffect, useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { createApiClient, type LeaderboardEntry } from "@/lib/api-client"

import LeaderboardTable from "./LeaderboardTable"

type FilterMode = "all-time" | "season"

interface LeaderboardWorkspaceProps {
  apiBaseUrl: string
}

function buildTitle(mode: FilterMode, seasonID: number | null): string {
  if (mode === "all-time") {
    return "All-Time Leaderboard"
  }

  if (seasonID) {
    return `Season ${seasonID} Leaderboard`
  }

  return "Season Leaderboard"
}

export default function LeaderboardWorkspace({ apiBaseUrl }: LeaderboardWorkspaceProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])

  const [mode, setMode] = useState<FilterMode>("all-time")
  const [seasonInput, setSeasonInput] = useState("1")
  const [appliedSeasonID, setAppliedSeasonID] = useState<number | null>(1)
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const handleModeChange = (nextMode: FilterMode) => {
    setMode(nextMode)
    setErrorMessage(null)
  }

  const handleApplySeason = () => {
    const parsed = Number.parseInt(seasonInput, 10)
    if (!Number.isFinite(parsed) || parsed <= 0) {
      setErrorMessage("Enter a valid season ID greater than 0.")
      setEntries([])
      setAppliedSeasonID(null)
      return
    }

    setErrorMessage(null)
    setAppliedSeasonID(parsed)
  }

  useEffect(() => {
    let isMounted = true

    const loadLeaderboard = async () => {
      try {
        setIsLoading(true)
        setErrorMessage(null)

        if (mode === "all-time") {
          const result = await apiClient.getAllTimeLeaderboard()
          if (isMounted) {
            setEntries(result)
            setAppliedSeasonID(null)
          }
          return
        }

        if (!appliedSeasonID || appliedSeasonID <= 0) {
          if (isMounted) {
            setErrorMessage("Choose a season ID and apply the filter.")
            setEntries([])
          }
          return
        }

        const result = await apiClient.getSeasonLeaderboard(appliedSeasonID)
        if (isMounted) {
          setEntries(result)
        }
      } catch {
        if (isMounted) {
          setErrorMessage("Could not load leaderboard data right now.")
          setEntries([])
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadLeaderboard()

    return () => {
      isMounted = false
    }
  }, [apiClient, mode, appliedSeasonID])

  const title = buildTitle(mode, appliedSeasonID)

  return (
    <section className="w-full space-y-6">
      <div className="space-y-3">
        <h1 className="text-2xl font-semibold">Leaderboard</h1>
        <p className="text-sm text-muted-foreground">
          Rankings are derived from game score difference. Filter to all-time or a specific season.
        </p>
      </div>

      <Card>
        <CardContent className="p-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant={mode === "all-time" ? "default" : "outline"}
              onClick={() => handleModeChange("all-time")}
            >
              All Time
            </Button>
            <Button
              type="button"
              variant={mode === "season" ? "default" : "outline"}
              onClick={() => handleModeChange("season")}
            >
              Season
            </Button>
          </div>

          {mode === "season" ? (
            <div className="flex w-full max-w-xs items-center gap-2">
              <Label htmlFor="season-id" className="sr-only">
                Season ID
              </Label>
              <Input
                id="season-id"
                inputMode="numeric"
                pattern="[0-9]*"
                value={seasonInput}
                onChange={(event) => setSeasonInput(event.target.value)}
                placeholder="Season ID"
                aria-label="Season ID"
              />
              <Button type="button" variant="outline" onClick={handleApplySeason}>
                Apply
              </Button>
            </div>
          ) : null}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-muted-foreground">Loading leaderboard...</p>
          </CardContent>
        </Card>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">{title}</h2>
            <p className="text-xs text-muted-foreground">Sorted by score difference, wins, username</p>
          </div>

          {errorMessage ? (
            <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {errorMessage}
            </p>
          ) : null}

          {!errorMessage && entries.length === 0 ? (
            <Card>
              <CardContent className="p-6">
                <p className="text-sm text-muted-foreground">No leaderboard data available for this filter.</p>
              </CardContent>
            </Card>
          ) : null}

          {!errorMessage && entries.length > 0 ? <LeaderboardTable entries={entries} /> : null}
        </>
      )}
    </section>
  )
}
