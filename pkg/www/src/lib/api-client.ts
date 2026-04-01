import type { components, paths } from "@/types/api"
import { getAuthToken } from "./auth-token"

type JsonContent<T> = T extends { content: { "application/json": infer TJson } }
  ? TJson
  : never

type SigninOperation = NonNullable<paths["/auth/signin"]["post"]>
type SignupOperation = NonNullable<paths["/auth/signup"]["post"]>
type CreateGameOperation = NonNullable<paths["/games"]["post"]>
type GetGameByIDOperation = NonNullable<paths["/games/{gameID}"]["get"]>
type DeleteGameByIDOperation = NonNullable<paths["/games/{gameID}"]["delete"]>
type ListGamesForPlayerOperation = NonNullable<paths["/players/{playerID}/games"]["get"]>
type SearchPlayersOperation = NonNullable<paths["/players/search"]["get"]>
type GetAllTimeLeaderboardOperation = NonNullable<paths["/leaderboard"]["get"]>
type ListSeasonsOperation = NonNullable<paths["/seasons"]["get"]>
type CreateSeasonOperation = NonNullable<paths["/seasons"]["post"]>
type GetActiveSeasonOperation = NonNullable<paths["/seasons/active"]["get"]>
type GetSeasonLeaderboardOperation = NonNullable<paths["/seasons/{seasonID}/leaderboard"]["get"]>

export type SigninInput = JsonContent<SigninOperation["requestBody"]>
export type SignupInput = JsonContent<SignupOperation["requestBody"]>
export type AuthResult = JsonContent<SigninOperation["responses"][200]>
export type CreateGameInput = JsonContent<CreateGameOperation["requestBody"]>
export type Game = JsonContent<GetGameByIDOperation["responses"][200]>
export type PlayerGames = JsonContent<ListGamesForPlayerOperation["responses"][200]>
export type SearchPlayer = components["schemas"]["auth.User"]
export type SearchPlayersResult = JsonContent<SearchPlayersOperation["responses"][200]>
export type LeaderboardEntries = JsonContent<GetAllTimeLeaderboardOperation["responses"][200]>
export type LeaderboardEntry = LeaderboardEntries[number]
export type Seasons = JsonContent<ListSeasonsOperation["responses"][200]>
export type Season = Seasons[number]
export type CreateSeasonInput = JsonContent<CreateSeasonOperation["requestBody"]>
export type ErrorResponse = components["schemas"]["handlers.ErrorResponse"]

type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE"

interface ApiRequest<TBody> {
  path: string
  method: HttpMethod
  body?: TBody
  headers?: HeadersInit
  expectedStatus?: number[]
}

export interface ApiClientOptions {
  baseUrl?: string
  defaultHeaders?: HeadersInit
  getToken?: () => string | undefined
  fetchFn?: typeof fetch
}

export class ApiError<TError = unknown> extends Error {
  status: number
  body?: TError

  constructor(message: string, status: number, body?: TError) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.body = body
  }
}

export class ApiClient {
  private readonly baseUrl: string
  private readonly defaultHeaders?: HeadersInit
  private readonly getToken?: () => string | undefined
  private readonly fetchFn: typeof fetch

  constructor(options: ApiClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? ""
    this.defaultHeaders = options.defaultHeaders
    this.getToken = options.getToken
    const fetchImpl = options.fetchFn ?? globalThis.fetch
    this.fetchFn = fetchImpl.bind(globalThis)
  }

  async signIn(input: SigninInput): Promise<AuthResult> {
    return this.request<SigninInput, AuthResult, ErrorResponse>({
      path: "/auth/signin",
      method: "POST",
      body: input,
      expectedStatus: [200],
    })
  }

  async signUp(input: SignupInput): Promise<AuthResult> {
    return this.request<SignupInput, AuthResult, ErrorResponse>({
      path: "/auth/signup",
      method: "POST",
      body: input,
      expectedStatus: [201],
    })
  }

  async createGame(input: CreateGameInput): Promise<Game> {
    return this.request<CreateGameInput, Game, ErrorResponse>({
      path: "/games",
      method: "POST",
      body: input,
      expectedStatus: [201],
    })
  }

  async getGameById(gameID: number): Promise<Game> {
    return this.request<undefined, Game, ErrorResponse>({
      path: `/games/${gameID}`,
      method: "GET",
      expectedStatus: [200],
    })
  }

  async listGamesForPlayer(playerID: number): Promise<PlayerGames> {
    return this.request<undefined, PlayerGames, ErrorResponse>({
      path: `/players/${playerID}/games`,
      method: "GET",
      expectedStatus: [200],
    })
  }

  async deleteGame(gameID: number): Promise<void> {
    await this.request<undefined, void, ErrorResponse>({
      path: `/games/${gameID}`,
      method: "DELETE",
      expectedStatus: [204],
    })
  }

  async searchPlayers(query?: string): Promise<SearchPlayersResult> {
    const queryString = query ? `?query=${encodeURIComponent(query)}` : ""

    return this.request<undefined, SearchPlayersResult, ErrorResponse>({
      path: `/players/search${queryString}`,
      method: "GET",
      expectedStatus: [200],
    })
  }

  async getAllTimeLeaderboard(): Promise<LeaderboardEntries> {
    return this.request<undefined, LeaderboardEntries, ErrorResponse>({
      path: "/leaderboard",
      method: "GET",
      expectedStatus: [200],
    })
  }

  async listSeasons(): Promise<Seasons> {
    return this.request<undefined, Seasons, ErrorResponse>({
      path: "/seasons",
      method: "GET",
      expectedStatus: [200],
    })
  }

  async getActiveSeason(): Promise<Season> {
    return this.request<undefined, Season, ErrorResponse>({
      path: "/seasons/active",
      method: "GET",
      expectedStatus: [200],
    })
  }

  async createSeason(input: CreateSeasonInput): Promise<Season> {
    return this.request<CreateSeasonInput, Season, ErrorResponse>({
      path: "/seasons",
      method: "POST",
      body: input,
      expectedStatus: [201],
    })
  }

  async getSeasonLeaderboard(seasonID: number): Promise<LeaderboardEntries> {
    return this.request<undefined, LeaderboardEntries, ErrorResponse>({
      path: `/seasons/${seasonID}/leaderboard`,
      method: "GET",
      expectedStatus: [200],
    })
  }

  private async request<TBody, TResponse, TError = unknown>(
    request: ApiRequest<TBody>,
  ): Promise<TResponse> {
    const token = this.getToken?.()
    const headers = new Headers(this.defaultHeaders)

    if (request.body !== undefined) {
      headers.set("Content-Type", "application/json")
    }

    if (token) {
      headers.set("Authorization", `Bearer ${token}`)
    }

    if (request.headers) {
      for (const [key, value] of new Headers(request.headers).entries()) {
        headers.set(key, value)
      }
    }

    const response = await this.fetchFn(this.resolveUrl(request.path), {
      method: request.method,
      headers,
      body: request.body !== undefined ? JSON.stringify(request.body) : undefined,
    })

    const parsedBody = await this.parseResponseBody(response)

    if (!response.ok) {
      throw new ApiError<TError>(
        `Request failed with status ${response.status}`,
        response.status,
        parsedBody as TError,
      )
    }

    if (
      request.expectedStatus &&
      request.expectedStatus.length > 0 &&
      !request.expectedStatus.includes(response.status)
    ) {
      throw new ApiError<TError>(
        `Unexpected success status ${response.status}`,
        response.status,
        parsedBody as TError,
      )
    }

    return parsedBody as TResponse
  }

  private resolveUrl(path: string): string {
    if (!this.baseUrl) {
      return path
    }

    const normalizedBase = this.baseUrl.endsWith("/")
      ? this.baseUrl.slice(0, -1)
      : this.baseUrl
    const normalizedPath = path.startsWith("/") ? path : `/${path}`

    return `${normalizedBase}${normalizedPath}`
  }

  private async parseResponseBody(response: Response): Promise<unknown> {
    if (response.status === 204) {
      return undefined
    }

    const contentType = response.headers.get("content-type") ?? ""

    if (contentType.includes("application/json")) {
      return response.json()
    }

    const text = await response.text()
    return text || undefined
  }
}

export function createApiClient(baseUrl?: string): ApiClient {
  return new ApiClient({
    baseUrl,
    getToken: getAuthToken,
  })
}