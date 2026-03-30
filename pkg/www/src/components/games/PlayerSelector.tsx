import { useEffect, useMemo, useRef, useState } from "react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { createApiClient, type SearchPlayer } from "@/lib/api-client"

interface PlayerSelectorProps {
  apiBaseUrl: string
  inputId: string
  label: string
  selectedPlayerId: string
  excludedPlayerIDs: number[]
  disabled?: boolean
  focusSignal?: number
  onPlayerSelect: (player: SearchPlayer | null) => void
}

function formatPlayerLabel(player: SearchPlayer): string {
  const fullName = [player.firstName, player.lastName].filter(Boolean).join(" ").trim()
  if (fullName) {
    return `${fullName} (${player.username || `id:${player.id}`})`
  }

  if (player.username) {
    return `${player.username} (id:${player.id})`
  }

  return `Player ${player.id}`
}

export default function PlayerSelector({
  apiBaseUrl,
  inputId,
  label,
  selectedPlayerId,
  excludedPlayerIDs,
  disabled,
  focusSignal,
  onPlayerSelect,
}: PlayerSelectorProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const inputRef = useRef<HTMLInputElement | null>(null)
  const [query, setQuery] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [options, setOptions] = useState<SearchPlayer[]>([])

  useEffect(() => {
    if (focusSignal !== undefined) {
      inputRef.current?.focus()
    }
  }, [focusSignal])

  useEffect(() => {
    if (!selectedPlayerId) {
      setQuery("")
    }
  }, [selectedPlayerId])

  useEffect(() => {
    if (!isOpen || disabled) {
      return
    }

    let isMounted = true

    const loadPlayers = async () => {
      try {
        setIsLoading(true)
        const result = await apiClient.searchPlayers(query.trim())
        if (isMounted) {
          setOptions(result.players ?? [])
        }
      } catch {
        if (isMounted) {
          setOptions([])
        }
      } finally {
        if (isMounted) {
          setIsLoading(false)
        }
      }
    }

    void loadPlayers()

    return () => {
      isMounted = false
    }
  }, [apiClient, disabled, isOpen, query])

  const takenPlayerIDSet = useMemo(() => new Set(excludedPlayerIDs), [excludedPlayerIDs])

  return (
    <div className="space-y-2">
      <Label htmlFor={inputId}>{label}</Label>
      <div className="relative">
        <Input
          ref={inputRef}
          id={inputId}
          type="text"
          placeholder="Search players"
          value={query}
          onFocus={() => setIsOpen(true)}
          onChange={(event) => {
            setQuery(event.currentTarget.value)
            setIsOpen(true)
          }}
          disabled={disabled}
          autoComplete="off"
        />

        {selectedPlayerId ? (
          <Button
            type="button"
            variant="ghost"
            size="xs"
            className="absolute right-1 top-1"
            onClick={() => {
              onPlayerSelect(null)
              setQuery("")
              setIsOpen(true)
            }}
            disabled={disabled}
          >
            Clear
          </Button>
        ) : null}

        {isOpen ? (
          <div className="absolute z-50 mt-1 max-h-48 w-full overflow-y-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md">
            {isLoading ? (
              <p className="px-2 py-2 text-xs text-muted-foreground">Loading players...</p>
            ) : options.length === 0 ? (
              <p className="px-2 py-2 text-xs text-muted-foreground">No players found.</p>
            ) : (
              <div className="space-y-1">
                {options.map((player) => {
                  const playerID = player.id
                  if (!playerID || !Number.isInteger(playerID) || playerID <= 0) {
                    return null
                  }

                  const isTaken = takenPlayerIDSet.has(playerID)

                  return (
                    <button
                      key={`${inputId}-${playerID}`}
                      type="button"
                      className="w-full rounded-sm px-2 py-1 text-left text-sm hover:bg-accent hover:text-accent-foreground disabled:cursor-not-allowed disabled:opacity-50"
                      disabled={isTaken}
                      onClick={() => {
                        onPlayerSelect(player)
                        setQuery(formatPlayerLabel(player))
                        setIsOpen(false)
                      }}
                    >
                      {formatPlayerLabel(player)}
                    </button>
                  )
                })}
              </div>
            )}
          </div>
        ) : null}
      </div>
    </div>
  )
}
