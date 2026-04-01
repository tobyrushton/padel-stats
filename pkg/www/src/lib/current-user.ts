import { ApiError, createApiClient, type User } from "@/lib/api-client"
import { clearAuthToken, getAuthToken } from "@/lib/auth-token"
import { clearAuthUser } from "@/lib/auth-user"

export async function fetchCurrentUserFromServer(apiBaseUrl?: string): Promise<User | undefined> {
  const token = getAuthToken()
  if (!token) {
    return undefined
  }

  const apiClient = createApiClient(apiBaseUrl)

  try {
    return await apiClient.getCurrentUser()
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      clearAuthToken()
      clearAuthUser()
      return undefined
    }

    throw error
  }
}
