import { useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  ApiError,
  createApiClient,
  type CreateGameInput,
  type ErrorResponse,
  type Game,
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

type PlayerFieldKey = keyof Pick<
  CreateGameFormState,
  "team1Player1Id" | "team1Player2Id" | "team2Player1Id" | "team2Player2Id"
>

interface TeamPlayerGroupProps {
  apiBaseUrl: string
  teamName: string
  firstField: {
    key: PlayerFieldKey
    inputId: string
    label: string
    value: string
  }
  secondField: {
    key: PlayerFieldKey
    inputId: string
    label: string
    value: string
  }
  selectedPlayerIDs: number[]
  isSubmitting: boolean
  focusSignal?: number
  scoreInputId: string
  scoreLabel: string
  scoreValue: string
  onScoreChange: (value: string) => void
  onPlayerSelect: (field: PlayerFieldKey, player: SearchPlayer | null) => void
}

function TeamPlayerGroup({
  apiBaseUrl,
  teamName,
  firstField,
  secondField,
  selectedPlayerIDs,
  isSubmitting,
  focusSignal,
  scoreInputId,
  scoreLabel,
  scoreValue,
  onScoreChange,
  onPlayerSelect,
}: TeamPlayerGroupProps) {
  const firstSelected = parsePositiveInteger(firstField.value)
  const secondSelected = parsePositiveInteger(secondField.value)

  return (
    <Card size="sm" className="overflow-visible border border-border/70 bg-background/40">
      <CardHeader className="border-b pb-3">
        <CardTitle className="text-sm tracking-wide uppercase text-muted-foreground">{teamName}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3 pt-3">
        <div className="grid gap-3 sm:grid-cols-[1fr_auto_1fr] sm:items-end">
          <PlayerSelector
            apiBaseUrl={apiBaseUrl}
            inputId={firstField.inputId}
            label={firstField.label}
            selectedPlayerId={firstField.value}
            excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== firstSelected)}
            onPlayerSelect={(player) => onPlayerSelect(firstField.key, player)}
            disabled={isSubmitting}
            focusSignal={focusSignal}
          />

          <p className="text-center text-sm font-semibold text-muted-foreground sm:pb-2">&amp;</p>

          <PlayerSelector
            apiBaseUrl={apiBaseUrl}
            inputId={secondField.inputId}
            label={secondField.label}
            selectedPlayerId={secondField.value}
            excludedPlayerIDs={selectedPlayerIDs.filter((id) => id !== secondSelected)}
            onPlayerSelect={(player) => onPlayerSelect(secondField.key, player)}
            disabled={isSubmitting}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor={scoreInputId}>{scoreLabel}</Label>
          <Input
            id={scoreInputId}
            type="number"
            min={0}
            value={scoreValue}
            onChange={(event) => onScoreChange(event.currentTarget.value)}
            disabled={isSubmitting}
            required
          />
        </div>
      </CardContent>
    </Card>
  )
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

function validateGameInput(form: CreateGameFormState): string | null {
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

function toCreateGameInput(form: CreateGameFormState): CreateGameInput {
  return {
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
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const selectedPlayerIDs = [
    parsePositiveInteger(form.team1Player1Id),
    parsePositiveInteger(form.team1Player2Id),
    parsePositiveInteger(form.team2Player1Id),
    parsePositiveInteger(form.team2Player2Id),
  ].filter((id): id is number => Boolean(id))

  const selectPlayer = (field: PlayerFieldKey, player: SearchPlayer | null) => {
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

    const validationError = validateGameInput(form)
    if (validationError) {
      setErrorMessage(validationError)
      return
    }

    try {
      setIsSubmitting(true)
      const game = await apiClient.createGame(toCreateGameInput(form))
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

      <div className="grid gap-3 lg:grid-cols-[1fr_auto_1fr] lg:items-center">
        <TeamPlayerGroup
          apiBaseUrl={apiBaseUrl}
          teamName="Team A"
          firstField={{
            key: "team1Player1Id",
            inputId: "game-team-a-player-a-id",
            label: "Player A",
            value: form.team1Player1Id,
          }}
          secondField={{
            key: "team1Player2Id",
            inputId: "game-team-a-player-b-id",
            label: "Player B",
            value: form.team1Player2Id,
          }}
          selectedPlayerIDs={selectedPlayerIDs}
          onPlayerSelect={selectPlayer}
          isSubmitting={isSubmitting}
          focusSignal={focusSignal}
          scoreInputId="game-team1-score"
          scoreLabel="Team 1 Score"
          scoreValue={form.team1Score}
          onScoreChange={(value) => handleChange("team1Score", value)}
        />

        <div className="flex items-center justify-center">
          <span className="inline-flex h-9 min-w-9 items-center justify-center rounded-full border border-border bg-muted px-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            vs
          </span>
        </div>

        <TeamPlayerGroup
          apiBaseUrl={apiBaseUrl}
          teamName="Team B"
          firstField={{
            key: "team2Player1Id",
            inputId: "game-team-b-player-c-id",
            label: "Player C",
            value: form.team2Player1Id,
          }}
          secondField={{
            key: "team2Player2Id",
            inputId: "game-team-b-player-d-id",
            label: "Player D",
            value: form.team2Player2Id,
          }}
          selectedPlayerIDs={selectedPlayerIDs}
          onPlayerSelect={selectPlayer}
          isSubmitting={isSubmitting}
          scoreInputId="game-team2-score"
          scoreLabel="Team 2 Score"
          scoreValue={form.team2Score}
          onScoreChange={(value) => handleChange("team2Score", value)}
        />
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

      {successMessage ? (
        <p className="rounded-md border border-primary/30 bg-primary/10 px-3 py-2 text-sm text-primary">
          {successMessage}
        </p>
      ) : null}

      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Button
          type="submit"
          disabled={isSubmitting}
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
