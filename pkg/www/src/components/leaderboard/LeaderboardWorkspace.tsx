import { useEffect, useMemo, useState } from "react"

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

const ALL_TIME_SCOPE = "all-time"

type ScopeValue = typeof ALL_TIME_SCOPE | `season:${number}`

interface LeaderboardWorkspaceProps {
  apiBaseUrl: string
}

interface SeasonOption {
  id: number
  name: string
}

function toSeasonScope(seasonID: number): ScopeValue {
  return `season:${seasonID}`
}

function parseSeasonScope(scope: ScopeValue): number | null {
  if (scope === ALL_TIME_SCOPE) {
    return null
  }

  const parsed = Number.parseInt(scope.replace("season:", ""), 10)
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null
}

function buildTitle(scope: ScopeValue, seasonName: string | null): string {
  if (scope === ALL_TIME_SCOPE) {
    return "All-Time Leaderboard"
  }

  if (seasonName) {
    return `${seasonName} Leaderboard`
  }

  return "Season Leaderboard"
}

export default function LeaderboardWorkspace({ apiBaseUrl }: LeaderboardWorkspaceProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])

  const [selectedScope, setSelectedScope] = useState<ScopeValue>(ALL_TIME_SCOPE)
  const [seasons, setSeasons] = useState<SeasonOption[]>([])
  const [activeSeasonID, setActiveSeasonID] = useState<number | null>(null)
  const [isLoadingSeasons, setIsLoadingSeasons] = useState(true)
  const [seasonErrorMessage, setSeasonErrorMessage] = useState<string | null>(null)
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true

    const loadSeasons = async () => {
      try {
        setIsLoadingSeasons(true)
        setSeasonErrorMessage(null)

        const [seasonResult, activeSeasonResult] = await Promise.all([
          apiClient.listSeasons(),
          apiClient.getActiveSeason().catch(() => null),
        ])

        const normalized = seasonResult
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

        const activeID =
          activeSeasonResult && typeof activeSeasonResult.id === "number" && activeSeasonResult.id > 0
            ? activeSeasonResult.id
            : null
        const activeExists = activeID !== null && normalized.some((season) => season.id === activeID)

        setSeasons(normalized)
        setActiveSeasonID(activeID)
        if (activeExists && activeID !== null) {
          setSelectedScope(toSeasonScope(activeID))
        } else if (normalized.length > 0) {
          setSelectedScope(toSeasonScope(normalized[0].id))
        } else {
          setSelectedScope(ALL_TIME_SCOPE)
        }

        if (activeID === null && normalized.length > 0) {
          setSeasonErrorMessage("Active season unavailable. Defaulted to the latest season.")
        }
      } catch {
        if (isMounted) {
          setSeasonErrorMessage("Could not load seasons right now. All-time leaderboard is still available.")
          setSeasons([])
          setActiveSeasonID(null)
          setSelectedScope(ALL_TIME_SCOPE)
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

        if (selectedScope === ALL_TIME_SCOPE) {
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

        const selectedSeasonID = parseSeasonScope(selectedScope)
        if (!selectedSeasonID) {
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
  }, [apiClient, isLoadingSeasons, seasonErrorMessage, selectedScope])

  const selectedSeasonID = parseSeasonScope(selectedScope)
  const selectedSeasonName =
    selectedSeasonID !== null
      ? seasons.find((season) => season.id === selectedSeasonID)?.name ?? null
      : null

  const title = buildTitle(selectedScope, selectedSeasonName)
  const currentScopeLabel =
    selectedScope === ALL_TIME_SCOPE
      ? "All Time"
      : selectedSeasonName ?? "Season"

  const selectPlaceholder = isLoadingSeasons ? "Loading scopes..." : "Select leaderboard scope"

  return (
    <section className="w-full space-y-6">
      <Card className="animate-in fade-in-0 slide-in-from-top-2 duration-300 overflow-hidden border-border/80">
        <CardContent className="relative p-5 sm:p-6">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_right,hsl(var(--primary)/0.18),transparent_45%),radial-gradient(circle_at_bottom_left,hsl(var(--accent)/0.2),transparent_50%)]"
          />

          <div className="relative space-y-5">
            <div className="space-y-3">
              <p className="text-xs font-medium uppercase tracking-[0.14em] text-muted-foreground">
                Live Rankings
              </p>
              <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">Leaderboard</h1>
              <p className="max-w-2xl text-sm text-muted-foreground">
                Score difference drives rankings. Switch between all-time and season snapshots from one selector.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2 text-xs">
              <span className="rounded-full border border-border/80 bg-background/80 px-3 py-1 font-medium">
                Scope: {currentScopeLabel}
              </span>
              {!isLoading && !errorMessage ? (
                <span className="rounded-full border border-border/80 bg-background/80 px-3 py-1 font-medium">
                  {entries.length} players ranked
                </span>
              ) : null}
            </div>

            <div className="w-full max-w-sm space-y-2">
              <Label htmlFor="leaderboard-scope-select">Leaderboard scope</Label>
              <Select
                value={selectedScope}
                onValueChange={(value) => {
                  setSelectedScope(value as ScopeValue)
                  setErrorMessage(null)
                }}
                disabled={isLoadingSeasons && seasons.length === 0}
              >
                <SelectTrigger id="leaderboard-scope-select" className="w-full bg-background/90">
                  <SelectValue placeholder={selectPlaceholder} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={ALL_TIME_SCOPE}>All Time</SelectItem>
                  {seasons.map((season) => {
                    const isCurrent = activeSeasonID !== null && season.id === activeSeasonID
                    const label = isCurrent ? `${season.name} (Current)` : season.name

                    return (
                      <SelectItem key={season.id} value={toSeasonScope(season.id)}>
                        {label}
                      </SelectItem>
                    )
                  })}
                </SelectContent>
              </Select>

              <p className="text-xs text-muted-foreground">
                Ranking order: score difference, wins, then username.
              </p>
            </div>

            {seasonErrorMessage ? (
              <p className="rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-sm text-amber-700">
                {seasonErrorMessage}
              </p>
            ) : null}

            {!isLoadingSeasons && !seasonErrorMessage && seasons.length === 0 ? (
              <p className="text-sm text-muted-foreground">No seasons available yet. Showing all-time by default.</p>
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
          <div className="flex flex-wrap items-center justify-between gap-2">
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
