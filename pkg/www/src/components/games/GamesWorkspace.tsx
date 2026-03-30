import { useState } from "react"
import { Plus } from "lucide-react"

import { Button } from "@/components/ui/button"
import type { Game } from "@/lib/api-client"

import CreateGameForm from "@/components/games/CreateGameForm"
import GamesList from "@/components/games/GamesList"

interface GamesWorkspaceProps {
  apiBaseUrl: string
}

export default function GamesWorkspace({ apiBaseUrl }: GamesWorkspaceProps) {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [lastCreatedGameID, setLastCreatedGameID] = useState<number | undefined>(undefined)
  const [openCount, setOpenCount] = useState(0)

  const handleCreated = (game: Game) => {
    setLastCreatedGameID(game.id)
    setRefreshKey((current) => current + 1)
    setIsCreateModalOpen(false)
  }

  const handleAddClick = () => {
    setOpenCount((current) => current + 1)
    setIsCreateModalOpen(true)
  }

  const handleCloseCreateModal = () => {
    setIsCreateModalOpen(false)
  }

  return (
    <section className="w-full space-y-6 pb-24">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Games</h1>
          <p className="text-sm text-muted-foreground">Add new matches and review your game history.</p>
        </div>
      </div>

      <GamesList
        apiBaseUrl={apiBaseUrl}
        refreshKey={refreshKey}
        lastCreatedGameID={lastCreatedGameID}
        onAddRequest={handleAddClick}
      />

      {isCreateModalOpen ? (
        <div className="fixed left-0 top-0 z-50 h-screen min-h-[100dvh] w-screen">
          <div className="absolute inset-0 bg-black/50" />
          <div className="relative flex h-full w-full items-center justify-center p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))]">
            <div className="w-full max-w-2xl space-y-3 rounded-xl border border-border bg-background p-3 shadow-xl max-h-[90dvh] overflow-y-auto sm:space-y-4 sm:p-5">
              <div className="flex items-center justify-between gap-4">
                <h2 className="text-lg font-semibold">Add Game</h2>
                <Button type="button" variant="ghost" size="sm" onClick={handleCloseCreateModal}>
                  Close
                </Button>
              </div>

              <CreateGameForm
                apiBaseUrl={apiBaseUrl}
                onCreated={handleCreated}
                focusSignal={openCount}
                onCancel={handleCloseCreateModal}
              />
            </div>
          </div>
        </div>
      ) : null}

      <div className="pointer-events-none fixed inset-x-0 bottom-4 z-40 flex justify-center px-4">
        <div className="pointer-events-auto rounded-full border border-border bg-background/95 p-2 shadow-lg backdrop-blur">
          <Button type="button" size="sm" className="rounded-full px-4" onClick={handleAddClick}>
            <Plus className="size-4" aria-hidden="true" />
            <span>Add Game</span>
          </Button>
        </div>
      </div>
    </section>
  )
}
