import { useEffect, useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { createApiClient, type LeaderboardEntry } from "@/lib/api-client"

import LeaderboardTable from "./LeaderboardTable"

type FilterMode = "all-time" | "season"

interface LeaderboardWorkspaceProps {
  apiBaseUrl: string
}

interface SeasonOption {
  id: number
  name: string
}

function buildTitle(mode: FilterMode, seasonName: string | null): string {
  if (mode === "all-time") {
    return "All-Time Leaderboard"
  }

  if (seasonName) {
    return `${seasonName} Leaderboard`
  }

  return "Season Leaderboard"
}

export default function LeaderboardWorkspace({ apiBaseUrl }: LeaderboardWorkspaceProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])

  const [mode, setMode] = useState<FilterMode>("all-time")
  const [seasons, setSeasons] = useState<SeasonOption[]>([])
  const [selectedSeasonID, setSelectedSeasonID] = useState<number | null>(null)
  const [isLoadingSeasons, setIsLoadingSeasons] = useState(true)
  const [seasonErrorMessage, setSeasonErrorMessage] = useState<string | null>(null)
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  const handleModeChange = (nextMode: FilterMode) => {
    setMode(nextMode)
    setErrorMessage(null)
  }

  useEffect(() => {
    let isMounted = true

    const loadSeasons = async () => {
      try {
        setIsLoadingSeasons(true)
        setSeasonErrorMessage(null)

        const result = await apiClient.listSeasons()
        const normalized = result
          .filter((season) => typeof season.id === "number" && season.id > 0)
          .map((season) => ({
            id: season.id as number,
            name: typeof season.name === "string" && season.name.trim().length > 0
              ? season.name.trim()
              : `Season ${season.id as number}`,
          }))
          .sort((a, b) => b.id - a.id)

        if (!isMounted) {
          return
        }

        setSeasons(normalized)
        setSelectedSeasonID(normalized.length > 0 ? normalized[0].id : null)
      } catch {
        if (isMounted) {
          setSeasonErrorMessage("Could not load seasons right now.")
          setSeasons([])
          setSelectedSeasonID(null)
        }
      } finally {
        if (isMounted) {
          setIsLoadingSeasons(false)
        }
      }
    }

    void loadSeasons()

    return () => {
      isMounted = false
    }
  }, [apiClient])

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
          }
          return
        }

        if (isLoadingSeasons) {
          if (isMounted) {
            setEntries([])
          }
          return
        }

        if (seasonErrorMessage) {
          if (isMounted) {
            setEntries([])
          }
          return
        }

        if (!selectedSeasonID || selectedSeasonID <= 0) {
          if (isMounted) {
            setEntries([])
          }
          return
        }

        const result = await apiClient.getSeasonLeaderboard(selectedSeasonID)
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
  }, [apiClient, isLoadingSeasons, mode, seasonErrorMessage, selectedSeasonID])

  const selectedSeasonName =
    selectedSeasonID !== null
      ? seasons.find((season) => season.id === selectedSeasonID)?.name ?? null
      : null

  const title = buildTitle(mode, selectedSeasonName)

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
            <div className="w-full max-w-xs space-y-2">
              <Label htmlFor="season-select">
                Season
              </Label>
              <Select
                value={selectedSeasonID !== null ? String(selectedSeasonID) : undefined}
                onValueChange={(value) => setSelectedSeasonID(Number.parseInt(value, 10))}
                disabled={isLoadingSeasons || !!seasonErrorMessage || seasons.length === 0}
              >
                <SelectTrigger id="season-select" className="w-full">
                  <SelectValue placeholder={isLoadingSeasons ? "Loading seasons..." : "Select season"} />
                </SelectTrigger>
                <SelectContent>
                  {seasons.map((season) => (
                    <SelectItem key={season.id} value={String(season.id)}>
                      {season.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {seasonErrorMessage ? (
                <p className="text-sm text-destructive">{seasonErrorMessage}</p>
              ) : null}

              {!isLoadingSeasons && !seasonErrorMessage && seasons.length === 0 ? (
                <p className="text-sm text-muted-foreground">No seasons available yet.</p>
              ) : null}
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
