import { useEffect, useMemo, useState } from "react"

import LeaderboardTable from "@/components/leaderboard/LeaderboardTable"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { createApiClient, type LeaderboardEntry } from "@/lib/api-client"

interface CurrentSeasonLeaderboardProps {
  apiBaseUrl: string
}

interface SeasonOption {
  id: number
  name: string
}

function normalizeSeasonName(season: SeasonOption): string {
  return season.name.trim().length > 0 ? season.name.trim() : `Season ${season.id}`
}

export default function CurrentSeasonLeaderboard({ apiBaseUrl }: CurrentSeasonLeaderboardProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])

  const [season, setSeason] = useState<SeasonOption | null>(null)
  const [entries, setEntries] = useState<LeaderboardEntry[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)

  useEffect(() => {
    let isMounted = true

    const loadCurrentSeasonLeaderboard = async () => {
      try {
        setIsLoading(true)
        setErrorMessage(null)

        const seasons = await apiClient.listSeasons()
        const latestSeason = seasons
          .filter((item) => typeof item.id === "number" && item.id > 0)
          .map((item) => ({
            id: item.id as number,
            name: typeof item.name === "string" ? item.name : "",
          }))
          .sort((a, b) => b.id - a.id)[0]

        if (!isMounted) {
          return
        }

        if (!latestSeason) {
          setSeason(null)
          setEntries([])
          return
        }

        setSeason(latestSeason)

        const leaderboard = await apiClient.getSeasonLeaderboard(latestSeason.id)
        if (isMounted) {
          setEntries(leaderboard)
        }
      } catch {
        if (isMounted) {
          setErrorMessage("Could not load current season leaderboard right now.")
          setEntries([])
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadCurrentSeasonLeaderboard()

    return () => {
      isMounted = false
    }
  }, [apiClient])

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="space-y-2">
          <Badge variant="outline">Current Season</Badge>
          <h2 className="text-2xl font-semibold tracking-tight">Leaderboard Snapshot</h2>
          <p className="text-sm text-muted-foreground">
            Latest rankings from the active season table, sorted by score difference.
          </p>
        </div>
        <Button asChild variant="outline" size="sm">
          <a href="/leaderboard">View full leaderboard</a>
        </Button>
      </div>

      {isLoading ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Loading season standings</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Skeleton className="h-10 w-52" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </CardContent>
        </Card>
      ) : null}

      {!isLoading && errorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {errorMessage}
        </p>
      ) : null}

      {!isLoading && !errorMessage && !season ? (
        <Card>
          <CardContent className="p-6">
            <p className="text-sm text-muted-foreground">No seasons have been created yet.</p>
          </CardContent>
        </Card>
      ) : null}

      {!isLoading && !errorMessage && season ? (
        <div className="space-y-3">
          <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <Badge>{normalizeSeasonName(season)}</Badge>
            <span>Top players this season</span>
          </div>
          {entries.length > 0 ? (
            <LeaderboardTable entries={entries} />
          ) : (
            <Card>
              <CardContent className="p-6">
                <p className="text-sm text-muted-foreground">No leaderboard data available for this season.</p>
              </CardContent>
            </Card>
          )}
        </div>
      ) : null}
    </section>
  )
}