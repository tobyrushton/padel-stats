import { useEffect, useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  ApiError,
  createApiClient,
  type CreateGameInput,
  type ErrorResponse,
  type Game,
  type Season,
  type SearchPlayer,
} from "@/lib/api-client"
import PlayerSelector from "@/components/games/PlayerSelector"

type CreateGameFormState = {
  team1Player1Id: string
  team1Player2Id: string
  team2Player1Id: string
  team2Player2Id: string
  team1Score: string
  team2Score: string
  playedAt: string
}

type SeasonOption = {
  id: number
  name: string
}

function getTodayDateInputValue(): string {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, "0")
  const day = String(now.getDate()).padStart(2, "0")
  return `${year}-${month}-${day}`
}

const INITIAL_STATE: CreateGameFormState = {
  team1Player1Id: "",
  team1Player2Id: "",
  team2Player1Id: "",
  team2Player2Id: "",
  team1Score: "",
  team2Score: "",
  playedAt: getTodayDateInputValue(),
}

interface CreateGameFormProps {
  apiBaseUrl: string
  onCreated: (game: Game) => void
  focusSignal: number
  onCancel?: () => void
}

function getCreateGameErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const body = error.body as ErrorResponse | undefined
    if (body?.error) {
      return body.error
    }

    if (error.status === 400) {
      return "Please review the game details and try again."
    }
  }

  return "Could not create the game. Please try again."
}

function parsePositiveInteger(value: string): number | undefined {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return undefined
  }
  return parsed
}

function parseNonNegativeInteger(value: string): number | undefined {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < 0) {
    return undefined
  }
  return parsed
}

function validateGameInput(form: CreateGameFormState, selectedSeasonID: string): string | null {
  const seasonID = parsePositiveInteger(selectedSeasonID)
  if (!seasonID) {
    return "Season is required."
  }

  const playerIDs = [
    parsePositiveInteger(form.team1Player1Id),
    parsePositiveInteger(form.team1Player2Id),
    parsePositiveInteger(form.team2Player1Id),
    parsePositiveInteger(form.team2Player2Id),
  ]

  if (playerIDs.some((playerID) => !playerID)) {
    return "All player IDs must be positive integers."
  }

  const uniquePlayerIDs = new Set(playerIDs as number[])
  if (uniquePlayerIDs.size !== playerIDs.length) {
    return "Player IDs must be unique across both teams."
  }

  const team1Score = parseNonNegativeInteger(form.team1Score)
  const team2Score = parseNonNegativeInteger(form.team2Score)
  if (team1Score === undefined || team2Score === undefined) {
    return "Scores must be whole numbers of 0 or greater."
  }

  if (!form.playedAt) {
    return "Played date is required."
  }

  const playedAtDate = new Date(form.playedAt)
  if (Number.isNaN(playedAtDate.getTime())) {
    return "Played date is invalid."
  }

  return null
}

function toSeasonOptions(seasons: Season[]): SeasonOption[] {
  return seasons
    .filter((season): season is Season & { id: number; name: string } => (
      typeof season.id === "number" &&
      season.id > 0 &&
      typeof season.name === "string" &&
      season.name.trim().length > 0
    ))
    .map((season) => ({ id: season.id, name: season.name.trim() }))
    .sort((a, b) => b.id - a.id)
}

function toCreateGameInput(form: CreateGameFormState, selectedSeasonID: string): CreateGameInput {
  const seasonID = parsePositiveInteger(selectedSeasonID)

  if (!seasonID) {
    throw new Error("Season is required")
  }

  return {
    seasonId: seasonID,
    team1Player1Id: Number(form.team1Player1Id),
    team1Player2Id: Number(form.team1Player2Id),
    team2Player1Id: Number(form.team2Player1Id),
    team2Player2Id: Number(form.team2Player2Id),
    team1Score: Number(form.team1Score),
    team2Score: Number(form.team2Score),
    playedAt: `${form.playedAt}T00:00:00.000Z`,
  }
}

export default function CreateGameForm({ apiBaseUrl, onCreated, focusSignal, onCancel }: CreateGameFormProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const [form, setForm] = useState<CreateGameFormState>(INITIAL_STATE)
  const [seasons, setSeasons] = useState<SeasonOption[]>([])
  const [selectedSeasonID, setSelectedSeasonID] = useState("")
  const [isLoadingSeasons, setIsLoadingSeasons] = useState(true)
  const [seasonsErrorMessage, setSeasonsErrorMessage] = useState<string | null>(null)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    let isMounted = true

    const loadSeasons = async () => {
      try {
        setIsLoadingSeasons(true)
        setSeasonsErrorMessage(null)

        const response = await apiClient.listSeasons()
        if (!isMounted) {
          return
        }

        const seasonOptions = toSeasonOptions(response)
        setSeasons(seasonOptions)
        setSelectedSeasonID((current) => current || String(seasonOptions[0]?.id ?? ""))

        if (seasonOptions.length === 0) {
          setSeasonsErrorMessage("No seasons are available yet. Create a season before adding games.")
        }
      } catch {
        if (isMounted) {
          setSeasons([])
          setSelectedSeasonID("")
          setSeasonsErrorMessage("Could not load seasons right now.")
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

  const selectedPlayerIDs = [
    parsePositiveInteger(form.team1Player1Id),
    parsePositiveInteger(form.team1Player2Id),
    parsePositiveInteger(form.team2Player1Id),
    parsePositiveInteger(form.team2Player2Id),
  ].filter((id): id is number => Boolean(id))

  const selectPlayer = (field: keyof Pick<CreateGameFormState, "team1Player1Id" | "team1Player2Id" | "team2Player1Id" | "team2Player2Id">, player: SearchPlayer | null) => {
    const playerID = player?.id
    if (!playerID || !Number.isInteger(playerID) || playerID <= 0) {
      handleChange(field, "")
      return
    }

    handleChange(field, String(playerID))
  }

  const handleChange = (field: keyof CreateGameFormState, value: string) => {
    setForm((current) => ({
      ...current,
      [field]: value,
    }))
  }

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setErrorMessage(null)
    setSuccessMessage(null)

    const validationError = validateGameInput(form, selectedSeasonID)
    if (validationError) {
      setErrorMessage(validationError)
      return
    }

    try {
      setIsSubmitting(true)
      const game = await apiClient.createGame(toCreateGameInput(form, selectedSeasonID))
      setForm(() => ({
        ...INITIAL_STATE,
      }))
      setSuccessMessage(`Game #${game.id} added successfully.`)
      onCreated(game)
    } catch (error) {
      setErrorMessage(getCreateGameErrorMessage(error))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="w-full space-y-4 rounded-xl border border-border bg-card p-4 text-card-foreground shadow-sm sm:space-y-5 sm:p-6">
      <div className="space-y-1">
        <h2 className="text-xl font-semibold">Add game</h2>
        <p className="text-sm text-muted-foreground">Enter players, score, and match date to record a game.</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="game-season">Season</Label>
        <Select
          value={selectedSeasonID}
          onValueChange={setSelectedSeasonID}
          disabled={isSubmitting || isLoadingSeasons || seasons.length === 0}
        >
          <SelectTrigger id="game-season" className="w-full">
            <SelectValue placeholder={isLoadingSeasons ? "Loading seasons..." : "Select a season"} />
          </SelectTrigger>
          <SelectContent>
            {seasons.map((season) => (
              <SelectItem key={season.id} value={String(season.id)}>
                {season.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <PlayerSelector
          apiBaseUrl={apiBaseUrl}
          inputId="game-team1-player1-id"
          label="Team 1 Player 1"
          selectedPlayerId={form.team1Player1Id}
          excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== parsePositiveInteger(form.team1Player1Id))}
          onPlayerSelect={(player) => selectPlayer("team1Player1Id", player)}
          disabled={isSubmitting}
          focusSignal={focusSignal}
        />

        <PlayerSelector
          apiBaseUrl={apiBaseUrl}
          inputId="game-team1-player2-id"
          label="Team 1 Player 2"
          selectedPlayerId={form.team1Player2Id}
          excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== parsePositiveInteger(form.team1Player2Id))}
          onPlayerSelect={(player) => selectPlayer("team1Player2Id", player)}
          disabled={isSubmitting}
        />

        <PlayerSelector
          apiBaseUrl={apiBaseUrl}
          inputId="game-team2-player1-id"
          label="Team 2 Player 1"
          selectedPlayerId={form.team2Player1Id}
          excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== parsePositiveInteger(form.team2Player1Id))}
          onPlayerSelect={(player) => selectPlayer("team2Player1Id", player)}
          disabled={isSubmitting}
        />

        <PlayerSelector
          apiBaseUrl={apiBaseUrl}
          inputId="game-team2-player2-id"
          label="Team 2 Player 2"
          selectedPlayerId={form.team2Player2Id}
          excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== parsePositiveInteger(form.team2Player2Id))}
          onPlayerSelect={(player) => selectPlayer("team2Player2Id", player)}
          disabled={isSubmitting}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="game-team1-score">Team 1 Score</Label>
          <Input
            id="game-team1-score"
            type="number"
            min={0}
            value={form.team1Score}
            onChange={(event) => handleChange("team1Score", event.currentTarget.value)}
            disabled={isSubmitting}
            required
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="game-team2-score">Team 2 Score</Label>
          <Input
            id="game-team2-score"
            type="number"
            min={0}
            value={form.team2Score}
            onChange={(event) => handleChange("team2Score", event.currentTarget.value)}
            disabled={isSubmitting}
            required
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="game-played-at">Played at</Label>
        <Input
          id="game-played-at"
          type="date"
          value={form.playedAt}
          onChange={(event) => handleChange("playedAt", event.currentTarget.value)}
          disabled={isSubmitting}
          required
        />
      </div>

      {errorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {errorMessage}
        </p>
      ) : null}

      {seasonsErrorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {seasonsErrorMessage}
        </p>
      ) : null}

      {successMessage ? (
        <p className="rounded-md border border-primary/30 bg-primary/10 px-3 py-2 text-sm text-primary">
          {successMessage}
        </p>
      ) : null}

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Button
          type="submit"
          disabled={isSubmitting || isLoadingSeasons || seasons.length === 0 || !selectedSeasonID}
          className="w-full sm:w-auto"
        >
          {isSubmitting ? "Adding game..." : "Add game"}
        </Button>
        {onCancel ? (
          <Button type="button" variant="outline" onClick={onCancel} className="w-full sm:w-auto">
            Cancel
          </Button>
        ) : null}
      </div>
    </form>
  )
}
