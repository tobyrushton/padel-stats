import type { LeaderboardEntry } from "@/lib/api-client"
import { Card, CardContent } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

interface LeaderboardTableProps {
  entries: LeaderboardEntry[]
}

function signedDifference(value: number): string {
  if (value > 0) {
    return `+${value}`
  }

  return `${value}`
}

export default function LeaderboardTable({ entries }: LeaderboardTableProps) {
  return (
    <Card>
      <CardContent className="p-0">
        <Table className="min-w-[700px] text-left">
          <TableHeader>
            <TableRow className="bg-muted/50 text-xs uppercase tracking-wide text-muted-foreground hover:bg-muted/50">
              <TableHead className="px-4 py-3">Rank</TableHead>
              <TableHead className="px-4 py-3">Player</TableHead>
              <TableHead className="px-4 py-3 text-right">Diff</TableHead>
              <TableHead className="px-4 py-3 text-right">Wins</TableHead>
              <TableHead className="px-4 py-3 text-right">Losses</TableHead>
              <TableHead className="px-4 py-3 text-right">Games</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => {
              const fullName = [entry.firstName, entry.lastName].filter(Boolean).join(" ").trim()
              const displayName = fullName || entry.username || `Player ${entry.playerId}`
              const scoreDifference = entry.scoreDifference ?? 0
              const wins = entry.wins ?? 0
              const losses = entry.losses ?? 0
              const gamesPlayed = entry.gamesPlayed ?? 0
              const diffClass =
                scoreDifference > 0
                  ? "text-emerald-600"
                  : scoreDifference < 0
                    ? "text-red-600"
                    : "text-muted-foreground"

              return (
                <TableRow key={entry.playerId} className="border-border/80">
                  <TableCell className="px-4 py-3 text-sm font-semibold">#{entry.rank}</TableCell>
                  <TableCell className="px-4 py-3 text-sm">
                    <p className="font-medium">{displayName}</p>
                    <p className="text-xs text-muted-foreground">@{entry.username || `id:${entry.playerId}`}</p>
                  </TableCell>
                  <TableCell className={`px-4 py-3 text-right text-sm font-semibold ${diffClass}`}>
                    {signedDifference(scoreDifference)}
                  </TableCell>
                  <TableCell className="px-4 py-3 text-right text-sm">{wins}</TableCell>
                  <TableCell className="px-4 py-3 text-right text-sm">{losses}</TableCell>
                  <TableCell className="px-4 py-3 text-right text-sm">{gamesPlayed}</TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
