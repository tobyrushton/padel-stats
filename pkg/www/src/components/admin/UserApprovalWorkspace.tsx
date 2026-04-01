import { useEffect, useMemo, useState } from "react"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
  ApiError,
  createApiClient,
  type ErrorResponse,
  type User,
} from "@/lib/api-client"
import { getAuthToken } from "@/lib/auth-token"
import { getAuthUser } from "@/lib/auth-user"

interface UserApprovalWorkspaceProps {
  apiBaseUrl: string
}

function getApiErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    const body = error.body as ErrorResponse | undefined
    if (body?.error) {
      return body.error
    }

    if (error.status === 401) {
      return "You must be logged in to approve users."
    }

    if (error.status === 403) {
      return "Admin access is required to approve users."
    }
  }

  return "Something went wrong. Please try again."
}

function isValidUser(user: User): user is User & { id: number } {
  return typeof user.id === "number" && user.id > 0
}

export default function UserApprovalWorkspace({ apiBaseUrl }: UserApprovalWorkspaceProps) {
  const apiClient = useMemo(() => createApiClient(apiBaseUrl), [apiBaseUrl])
  const [query, setQuery] = useState("")
  const [users, setUsers] = useState<(User & { id: number })[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmittingByID, setIsSubmittingByID] = useState<Record<number, boolean>>({})
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const isLoggedIn = Boolean(getAuthToken())
  const isAdmin = Boolean(getAuthUser()?.isAdmin)

  const pendingUsers = users.filter((user) => user.isAcceptedByAdmin !== true)

  const loadUsers = async (searchQuery?: string) => {
    try {
      setIsLoading(true)
      setErrorMessage(null)

      const result = await apiClient.searchPlayers(searchQuery)
      const normalized = (result.players ?? []).filter(isValidUser)
      setUsers(normalized)
    } catch (error) {
      setErrorMessage(getApiErrorMessage(error))
      setUsers([])
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    void loadUsers()
  }, [apiClient])

  const handleSearch = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSuccessMessage(null)
    await loadUsers(query.trim())
  }

  const handleApprove = async (userID: number) => {
    try {
      setErrorMessage(null)
      setSuccessMessage(null)
      setIsSubmittingByID((current) => ({ ...current, [userID]: true }))

      await apiClient.approveUser(userID)
      setUsers((current) => current.filter((user) => user.id !== userID))
      setSuccessMessage(`User ${userID} approved.`)
    } catch (error) {
      setErrorMessage(getApiErrorMessage(error))
    } finally {
      setIsSubmittingByID((current) => ({ ...current, [userID]: false }))
    }
  }

  if (!isLoggedIn) {
    return (
      <Card className="w-full">
        <CardContent className="p-6">
          <h1 className="text-xl font-semibold">Admin approvals</h1>
          <p className="mt-2 text-sm text-muted-foreground">You need to sign in to access this page.</p>
        </CardContent>
      </Card>
    )
  }

  if (!isAdmin) {
    return (
      <Card className="w-full">
        <CardContent className="p-6">
          <h1 className="text-xl font-semibold">Admin approvals</h1>
          <p className="mt-2 text-sm text-muted-foreground">You do not have admin access.</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <section className="w-full space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold">Admin approvals</h1>
        <p className="text-sm text-muted-foreground">Search users and approve pending accounts.</p>
      </div>

      <Card>
        <CardContent className="p-4">
          <form onSubmit={handleSearch} className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <Input
              value={query}
              onChange={(event) => setQuery(event.currentTarget.value)}
              placeholder="Search by username or name"
              disabled={isLoading}
            />
            <Button type="submit" disabled={isLoading}>
              {isLoading ? "Searching..." : "Search"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {errorMessage ? (
        <p className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {errorMessage}
        </p>
      ) : null}

      {successMessage ? (
        <p className="rounded-md border border-emerald-300/60 bg-emerald-100/60 px-3 py-2 text-sm text-emerald-800">
          {successMessage}
        </p>
      ) : null}

      <Card>
        <CardContent className="p-4">
          <h2 className="text-base font-semibold">Pending users</h2>

          {isLoading ? (
            <p className="mt-3 text-sm text-muted-foreground">Loading users...</p>
          ) : null}

          {!isLoading && pendingUsers.length === 0 ? (
            <p className="mt-3 text-sm text-muted-foreground">No pending users found.</p>
          ) : null}

          {!isLoading && pendingUsers.length > 0 ? (
            <div className="mt-3 overflow-x-auto">
              <table className="w-full min-w-[520px] text-left text-sm">
                <thead>
                  <tr className="border-b border-border text-muted-foreground">
                    <th className="py-2 pr-3">User</th>
                    <th className="py-2 pr-3">Username</th>
                    <th className="py-2 pr-3">User ID</th>
                    <th className="py-2 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {pendingUsers.map((user) => {
                    const fullName = [user.firstName, user.lastName].filter(Boolean).join(" ").trim() || "-"
                    const isSubmitting = Boolean(isSubmittingByID[user.id])

                    return (
                      <tr key={user.id} className="border-b border-border/70">
                        <td className="py-3 pr-3">{fullName}</td>
                        <td className="py-3 pr-3">@{user.username ?? "-"}</td>
                        <td className="py-3 pr-3">{user.id}</td>
                        <td className="py-3 text-right">
                          <Button
                            type="button"
                            size="sm"
                            disabled={isSubmitting}
                            onClick={() => handleApprove(user.id)}
                          >
                            {isSubmitting ? "Approving..." : "Approve"}
                          </Button>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </section>
  )
}
